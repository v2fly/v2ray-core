package rrpitChannelManager

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
)

var ErrSharedSendQuotaReached = errors.New("rrpit shared send quota reached")

type (
	Listener func(payload []byte) error
	Sniffer  func(payload []byte)
)

type Config struct {
	OddChannelIDs                  bool
	MaxRewindableTimestampNum      int
	MaxRewindableControlMessageNum int
}

type channelState struct {
	config rriptMonoDirectionSession.ChannelConfig
	status rriptMonoDirectionSession.ChannelRateControlStatus
	tx     *rrpitMaterializedTransferChannel.ChannelTx
	rx     *rrpitMaterializedTransferChannel.ChannelRx
}

type ChannelManager struct {
	mu   sync.Mutex
	cond *sync.Cond

	config            Config
	nextChannelID     uint64
	currentTimestamp  uint64
	channels          []*channelState
	listeners         map[uint8][]Listener
	sniffers          map[uint8][]Sniffer
	totalSent         uint64
	blockOnNoChannels bool
	closed            bool
}

func New(config Config) (*ChannelManager, error) {
	manager := &ChannelManager{
		config:    config,
		listeners: make(map[uint8][]Listener),
		sniffers:  make(map[uint8][]Sniffer),
	}
	manager.cond = sync.NewCond(&manager.mu)
	return manager, nil
}

func (m *ChannelManager) RegisterListener(kind uint8, handler Listener) {
	if m == nil || handler == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners[kind] = append(m.listeners[kind], handler)
}

func (m *ChannelManager) AddSniffer(kind uint8, tap Sniffer) {
	if m == nil || tap == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sniffers[kind] = append(m.sniffers[kind], tap)
}

func (m *ChannelManager) AttachChannelWithConfig(
	writer io.WriteCloser,
	config rriptMonoDirectionSession.ChannelConfig,
) (int, error) {
	if m == nil {
		return 0, io.ErrClosedPipe
	}
	if writer == nil {
		return 0, io.ErrClosedPipe
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	if config.Weight == 0 {
		config.Weight = 1
	}
	channelID := m.allocateChannelIDLocked()
	tx, err := rrpitMaterializedTransferChannel.NewChannelTx(
		channelID,
		writer,
		m.config.MaxRewindableTimestampNum,
		m.config.MaxRewindableControlMessageNum,
	)
	if err != nil {
		return 0, err
	}
	state := &channelState{config: config, tx: tx}
	rx, err := rrpitMaterializedTransferChannel.NewChannelRx(0, func(data []byte) error {
		return m.onLogicalPacket(state, data)
	})
	if err != nil {
		return 0, err
	}
	state.rx = rx
	m.channels = append(m.channels, state)
	if m.cond != nil {
		m.cond.Broadcast()
	}
	return len(m.channels) - 1, nil
}

func (m *ChannelManager) SetBlockOnNoChannels(enabled bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.blockOnNoChannels = enabled
	if m.cond != nil {
		m.cond.Broadcast()
	}
	m.mu.Unlock()
}

func (m *ChannelManager) DetachChannel(channelIndex int) error {
	if m == nil {
		return io.ErrClosedPipe
	}

	m.mu.Lock()
	if channelIndex < 0 || channelIndex >= len(m.channels) {
		m.mu.Unlock()
		return io.ErrClosedPipe
	}
	channel := m.channels[channelIndex]
	m.channels[channelIndex] = nil
	if m.cond != nil {
		m.cond.Broadcast()
	}
	m.mu.Unlock()

	if channel == nil {
		return nil
	}
	var firstErr error
	if channel.rx != nil {
		if err := channel.rx.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if channel.tx != nil {
		if err := channel.tx.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *ChannelManager) OnNewMessageArrived(channelIndex int, payload []byte) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return io.ErrClosedPipe
	}
	if channelIndex < 0 || channelIndex >= len(m.channels) || m.channels[channelIndex] == nil || m.channels[channelIndex].rx == nil {
		m.mu.Unlock()
		return io.ErrClosedPipe
	}
	rx := m.channels[channelIndex].rx
	m.mu.Unlock()
	return rx.OnNewMessageArrived(payload)
}

func (m *ChannelManager) OnNewTimestamp(timestamp uint64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTimestamp = timestamp
	for _, channel := range m.channels {
		if channel == nil {
			continue
		}
		if channel.status.TimestampLastSent != timestamp {
			channel.status.PacketSentCurrentTimestamp = 0
			channel.status.EnforcedPacketSentCurrentTimestamp = 0
		}
	}
}

func (m *ChannelManager) HasRemainingQuota() bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.chooseEnforcedChannelLocked() >= 0
}

func (m *ChannelManager) Send(kind uint8, payload []byte) error {
	if m == nil {
		return io.ErrClosedPipe
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return io.ErrClosedPipe
	}
	channelIndex, err := m.waitForEnforcedChannelLocked()
	if err != nil {
		return err
	}
	return m.sendOnChannelLocked(channelIndex, kind, payload, true)
}

func (m *ChannelManager) SendIgnoreQuota(kind uint8, payload []byte) error {
	if m == nil {
		return io.ErrClosedPipe
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return io.ErrClosedPipe
	}
	if rriptMonoDirectionSession.IsControlPacketKind(kind) {
		if err := m.waitForAnyTxChannelLocked(); err != nil {
			return err
		}
		return m.sendFloodControlLocked(kind, payload)
	}
	channelIndex, err := m.waitForBypassChannelLocked()
	if err != nil {
		return err
	}
	return m.sendOnChannelLocked(channelIndex, kind, payload, false)
}

func (m *ChannelManager) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	if m.cond != nil {
		m.cond.Broadcast()
	}
	channels := append([]*channelState(nil), m.channels...)
	m.channels = nil
	m.mu.Unlock()

	var firstErr error
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		if channel.rx != nil {
			if err := channel.rx.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if channel.tx != nil {
			if err := channel.tx.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (m *ChannelManager) onLogicalPacket(channel *channelState, data []byte) error {
	if len(data) == 0 {
		return io.ErrUnexpectedEOF
	}
	kind := data[0]
	m.emitSniffers(kind, data)

	switch {
	case rriptMonoDirectionSession.IsControlPacketKind(kind):
		ctrl, err := rriptMonoDirectionSession.UnmarshalSessionControlPacket(data)
		if err != nil {
			return err
		}
		m.mu.Lock()
		if channel != nil && channel.rx != nil && ctrl.FloodChannel.CurrentChannelID != 0 {
			if m.hasAssignedChannelIDLocked(ctrl.FloodChannel.CurrentChannelID, channel) {
				m.mu.Unlock()
				return fmt.Errorf("duplicate rx channel id %d", ctrl.FloodChannel.CurrentChannelID)
			}
			if err := channel.rx.AssignChannelID(ctrl.FloodChannel.CurrentChannelID); err != nil {
				m.mu.Unlock()
				return err
			}
		}
		for _, channelControl := range ctrl.Channel.ChannelControl {
			if tx := m.txChannelByIDLocked(channelControl.ChannelID); tx != nil {
				if err := tx.AcceptControlMessage(channelControl); err != nil {
					m.mu.Unlock()
					return err
				}
			}
		}
		ctrl.FloodChannel = rriptMonoDirectionSession.SessionFloodChannelControlMessage{}
		ctrl.Channel = rriptMonoDirectionSession.SessionChannelControlMessage{}
		listeners := append([]Listener(nil), m.listeners[kind]...)
		m.mu.Unlock()
		stripped, err := rriptMonoDirectionSession.MarshalSessionControlPacket(kind, ctrl)
		if err != nil {
			return err
		}
		return dispatchListeners(listeners, stripped)
	case rriptMonoDirectionSession.IsDataPacketKind(kind):
		m.mu.Lock()
		listeners := append([]Listener(nil), m.listeners[kind]...)
		m.mu.Unlock()
		return dispatchListeners(listeners, append([]byte(nil), data...))
	default:
		return io.ErrUnexpectedEOF
	}
}

func (m *ChannelManager) sendFloodControlLocked(kind uint8, payload []byte) error {
	ctrl, err := rriptMonoDirectionSession.UnmarshalSessionControlPacket(payload)
	if err != nil {
		return err
	}
	sharedControl, err := m.buildSharedChannelControlLocked()
	if err != nil {
		return err
	}
	for index, channel := range m.channels {
		if channel == nil || channel.tx == nil {
			continue
		}
		cloned := ctrl
		cloned.FloodChannel = rriptMonoDirectionSession.SessionFloodChannelControlMessage{
			CurrentChannelID: channel.tx.ChannelID,
		}
		cloned.Channel = sharedControl
		stamped, err := rriptMonoDirectionSession.MarshalSessionControlPacket(kind, cloned)
		if err != nil {
			return err
		}
		if err := channel.tx.SendDataMessage(stamped); err != nil {
			return err
		}
		m.markChannelSentLocked(index, false)
		m.emitSniffersLocked(kind, stamped)
	}
	return nil
}

func (m *ChannelManager) buildSharedChannelControlLocked() (rriptMonoDirectionSession.SessionChannelControlMessage, error) {
	channelControl := rriptMonoDirectionSession.SessionChannelControlMessage{
		ChannelControl: make([]rrpitTransferChannel.ChannelControlMessage, 0, len(m.channels)),
	}
	for _, channel := range m.channels {
		if channel == nil || channel.rx == nil || channel.rx.ChannelID == 0 {
			continue
		}
		ctrl, err := channel.rx.CreateControlMessage()
		if err != nil {
			return rriptMonoDirectionSession.SessionChannelControlMessage{}, err
		}
		channelControl.ChannelControl = append(channelControl.ChannelControl, *ctrl)
	}
	channelControl.LenChannelControl = uint16(len(channelControl.ChannelControl))
	return channelControl, nil
}

func (m *ChannelManager) sendOnChannelLocked(channelIndex int, kind uint8, payload []byte, enforced bool) error {
	if channelIndex < 0 || channelIndex >= len(m.channels) {
		return io.ErrClosedPipe
	}
	channel := m.channels[channelIndex]
	if channel == nil || channel.tx == nil {
		return io.ErrClosedPipe
	}
	if err := channel.tx.SendDataMessage(payload); err != nil {
		return err
	}
	m.markChannelSentLocked(channelIndex, enforced)
	m.emitSniffersLocked(kind, payload)
	return nil
}

func (m *ChannelManager) waitForEnforcedChannelLocked() (int, error) {
	for {
		if m.closed {
			return -1, io.ErrClosedPipe
		}
		channelIndex := m.chooseEnforcedChannelLocked()
		if channelIndex >= 0 {
			return channelIndex, nil
		}
		if m.activeTxChannelCountLocked() == 0 {
			if !m.blockOnNoChannels || m.cond == nil {
				return -1, io.ErrClosedPipe
			}
			m.cond.Wait()
			continue
		}
		return -1, ErrSharedSendQuotaReached
	}
}

func (m *ChannelManager) waitForBypassChannelLocked() (int, error) {
	for {
		if m.closed {
			return -1, io.ErrClosedPipe
		}
		channelIndex := m.chooseBypassChannelLocked()
		if channelIndex >= 0 {
			return channelIndex, nil
		}
		if !m.blockOnNoChannels || m.cond == nil {
			return -1, io.ErrClosedPipe
		}
		m.cond.Wait()
	}
}

func (m *ChannelManager) waitForAnyTxChannelLocked() error {
	for {
		if m.closed {
			return io.ErrClosedPipe
		}
		if m.activeTxChannelCountLocked() > 0 {
			return nil
		}
		if !m.blockOnNoChannels || m.cond == nil {
			return io.ErrClosedPipe
		}
		m.cond.Wait()
	}
}

func (m *ChannelManager) chooseEnforcedChannelLocked() int {
	best := -1
	for i := range m.channels {
		channel := m.channels[i]
		if channel == nil || channel.tx == nil || m.channelRateLimitedLocked(i) {
			continue
		}
		if best == -1 || m.channelHasLessLoadLocked(i, best) {
			best = i
		}
	}
	return best
}

func (m *ChannelManager) chooseBypassChannelLocked() int {
	best := -1
	bestOversubscribe := 0.0
	for i, channel := range m.channels {
		if channel == nil || channel.tx == nil {
			continue
		}
		if channel.config.MaxSendingSpeed <= 0 {
			continue
		}
		oversubscribe := m.channelOversubscribePercentageLocked(i)
		if best == -1 || oversubscribe < bestOversubscribe {
			best = i
			bestOversubscribe = oversubscribe
		}
	}
	if best != -1 {
		return best
	}
	for i, channel := range m.channels {
		if channel != nil && channel.tx != nil {
			return i
		}
	}
	return -1
}

func (m *ChannelManager) activeTxChannelCountLocked() int {
	count := 0
	for _, channel := range m.channels {
		if channel != nil && channel.tx != nil {
			count += 1
		}
	}
	return count
}

func (m *ChannelManager) channelOversubscribePercentageLocked(index int) float64 {
	channel := m.channels[index]
	if channel == nil || channel.config.MaxSendingSpeed <= 0 {
		return 0
	}
	return float64(m.channelWindowSendCountLocked(index)) / float64(channel.config.MaxSendingSpeed)
}

func (m *ChannelManager) channelRateLimitedLocked(index int) bool {
	channel := m.channels[index]
	if channel == nil {
		return true
	}
	maxSpeed := channel.config.MaxSendingSpeed
	if maxSpeed <= 0 {
		return false
	}
	status := channel.status
	if status.TimestampLastSent != m.effectiveTimestampLocked() {
		return false
	}
	return int(status.EnforcedPacketSentCurrentTimestamp) >= maxSpeed
}

func (m *ChannelManager) channelHasLessLoadLocked(candidateIndex int, currentBestIndex int) bool {
	candidateWeight := m.channels[candidateIndex].config.Weight
	if candidateWeight <= 0 {
		candidateWeight = 1
	}
	bestWeight := m.channels[currentBestIndex].config.Weight
	if bestWeight <= 0 {
		bestWeight = 1
	}

	candidateSent := m.channelWindowSendCountLocked(candidateIndex)
	bestSent := m.channelWindowSendCountLocked(currentBestIndex)
	left := candidateSent * uint64(bestWeight)
	right := bestSent * uint64(candidateWeight)
	if left != right {
		return left < right
	}
	return candidateIndex < currentBestIndex
}

func (m *ChannelManager) channelWindowSendCountLocked(index int) uint64 {
	channel := m.channels[index]
	if channel == nil {
		return 0
	}
	if channel.status.TimestampLastSent != m.effectiveTimestampLocked() {
		return 0
	}
	return channel.status.PacketSentCurrentTimestamp
}

func (m *ChannelManager) markChannelSentLocked(index int, enforced bool) {
	channel := m.channels[index]
	if channel == nil {
		return
	}
	timestamp := m.effectiveTimestampLocked()
	if channel.status.TimestampLastSent != timestamp {
		channel.status.PacketSentCurrentTimestamp = 0
		channel.status.EnforcedPacketSentCurrentTimestamp = 0
	}
	channel.status.TimestampLastSent = timestamp
	channel.status.PacketSentCurrentTimestamp += 1
	if enforced {
		channel.status.EnforcedPacketSentCurrentTimestamp += 1
	}
	m.totalSent += 1
}

func (m *ChannelManager) effectiveTimestampLocked() uint64 {
	if m.currentTimestamp == 0 {
		m.currentTimestamp = uint64(time.Now().UnixNano())
	}
	return m.currentTimestamp
}

func (m *ChannelManager) txChannelByIDLocked(channelID uint64) *rrpitMaterializedTransferChannel.ChannelTx {
	for _, channel := range m.channels {
		if channel != nil && channel.tx != nil && channel.tx.ChannelID == channelID {
			return channel.tx
		}
	}
	return nil
}

func (m *ChannelManager) hasAssignedChannelIDLocked(channelID uint64, exclude *channelState) bool {
	for _, channel := range m.channels {
		if channel == nil || channel == exclude || channel.rx == nil {
			continue
		}
		if channel.rx.ChannelID == channelID {
			return true
		}
	}
	return false
}

func (m *ChannelManager) allocateChannelIDLocked() uint64 {
	if m.nextChannelID == 0 {
		if m.config.OddChannelIDs {
			m.nextChannelID = 1
		} else {
			m.nextChannelID = 2
		}
	}
	channelID := m.nextChannelID
	m.nextChannelID += 2
	return channelID
}

func (m *ChannelManager) emitSniffers(kind uint8, payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emitSniffersLocked(kind, payload)
}

func (m *ChannelManager) emitSniffersLocked(kind uint8, payload []byte) {
	taps := append([]Sniffer(nil), m.sniffers[kind]...)
	if len(taps) == 0 {
		return
	}
	cloned := append([]byte(nil), payload...)
	for _, tap := range taps {
		tap := tap
		go tap(cloned)
	}
}

func dispatchListeners(listeners []Listener, payload []byte) error {
	for _, listener := range listeners {
		if listener == nil {
			continue
		}
		if err := listener(append([]byte(nil), payload...)); err != nil {
			return err
		}
	}
	return nil
}
