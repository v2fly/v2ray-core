package rrpitBidirectionalSession

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

type BidirectionalSession struct {
	mu   sync.Mutex
	cond *sync.Cond

	rx *rriptMonoDirectionSession.SessionRx
	tx *rriptMonoDirectionSession.SessionTx

	localSessionInstanceID rriptMonoDirectionSession.SessionInstanceID
	TimestampInterval      time.Duration

	nextTimestamp                         uint64
	lastControlPayload                    []byte
	lastControlTimestamp                  uint64
	managerHostedControlKeepaliveInterval int
	sourcePacketActivity                  uint64
	autoTickStop                          chan struct{}
	autoTickDone                          chan struct{}
	closeOnce                             sync.Once
	closed                                bool
}

type Config struct {
	Rx rriptMonoDirectionSession.SessionRxConfig
	Tx rriptMonoDirectionSession.SessionTxConfig

	LocalSessionInstanceID                     rriptMonoDirectionSession.SessionInstanceID
	ValidateRemoteControl                      func(rriptMonoDirectionSession.ControlMessage) error
	TimestampInterval                          time.Duration
	ManagerHostedControlKeepaliveIntervalTicks int
}

const managerHostedControlKeepaliveIntervalTicks = 32

func New(config Config) (*BidirectionalSession, error) {
	session := &BidirectionalSession{
		TimestampInterval:      config.TimestampInterval,
		localSessionInstanceID: config.LocalSessionInstanceID,
	}
	session.cond = sync.NewCond(&session.mu)

	tx, err := rriptMonoDirectionSession.NewSessionTx(config.Tx)
	if err != nil {
		return nil, err
	}
	session.tx = tx

	rxConfig := config.Rx
	userRemoteControlHandler := rxConfig.OnRemoteControlMsg
	validateRemoteControl := config.ValidateRemoteControl
	rxConfig.OnRemoteControlMsg = func(ctrl rriptMonoDirectionSession.ControlMessage) error {
		if validateRemoteControl != nil {
			if err := validateRemoteControl(ctrl); err != nil {
				return err
			}
		}
		session.mu.Lock()
		if session.closed {
			session.mu.Unlock()
			return io.ErrClosedPipe
		}
		if err := tx.AcceptRemoteControlMessage(ctrl); err != nil {
			session.mu.Unlock()
			return err
		}
		if session.cond != nil {
			session.cond.Broadcast()
		}
		session.mu.Unlock()
		if userRemoteControlHandler != nil {
			return userRemoteControlHandler(ctrl)
		}
		return nil
	}

	rx, err := rriptMonoDirectionSession.NewSessionRx(rxConfig)
	if err != nil {
		return nil, err
	}
	session.rx = rx
	if config.ManagerHostedControlKeepaliveIntervalTicks <= 0 {
		config.ManagerHostedControlKeepaliveIntervalTicks = managerHostedControlKeepaliveIntervalTicks
	}
	session.managerHostedControlKeepaliveInterval = config.ManagerHostedControlKeepaliveIntervalTicks
	session.startAutoTick()
	return session, nil
}

func (s *BidirectionalSession) SendMessage(data []byte) error {
	if s == nil || s.tx == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		if s.closed {
			return io.ErrClosedPipe
		}
		err := s.tx.SendMessage(data)
		if err == nil {
			s.sourcePacketActivity += 1
			return nil
		}
		if !errors.Is(err, rriptMonoDirectionSession.ErrTxLaneBufferFull) {
			return err
		}
		if s.cond == nil {
			return err
		}
		s.cond.Wait()
	}
}

func (s *BidirectionalSession) OnNewTimestamp(timestamp uint64) error {
	_, err := s.OnNewTimestampWithStats(timestamp)
	return err
}

func (s *BidirectionalSession) OnNewTimestampWithStats(timestamp uint64) (rriptMonoDirectionSession.TickStats, error) {
	if s == nil || s.tx == nil || s.rx == nil {
		return rriptMonoDirectionSession.TickStats{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if timestamp > s.nextTimestamp {
		s.nextTimestamp = timestamp
	}
	return s.onNewTimestampLocked(timestamp)
}

func (s *BidirectionalSession) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		if s.cond != nil {
			s.cond.Broadcast()
		}
		s.mu.Unlock()
		if s.autoTickStop != nil {
			close(s.autoTickStop)
			<-s.autoTickDone
		}
	})
	return nil
}

func (s *BidirectionalSession) onNewTimestampLocked(timestamp uint64) (rriptMonoDirectionSession.TickStats, error) {
	stats, err := s.tx.OnNewTimestampWithStats(timestamp)
	if err != nil {
		return stats, err
	}
	if s.tx.SendIgnoreQuota != nil {
		ctrl, err := s.rx.GenerateStrippedControlMessage()
		if err != nil {
			return stats, err
		}
		ctrl = s.withLocalSessionControl(ctrl)
		payload, err := rriptMonoDirectionSession.MarshalSessionControlPacket(s.tx.ControlPacketKind, ctrl)
		if err != nil {
			return stats, err
		}
		if !s.shouldSendManagerHostedControlLocked(timestamp, payload, stats) {
			return stats, nil
		}
		stats.ControlPacketsGenerated += 1
		if err := s.tx.SendIgnoreQuota(s.tx.ControlPacketKind, payload); err != nil {
			return stats, err
		}
		stats.ControlPacketsSent += 1
		s.lastControlPayload = append(s.lastControlPayload[:0], payload...)
		s.lastControlTimestamp = timestamp
		return stats, nil
	}
	stats.ControlPacketsGenerated += 1
	if err := s.tx.FloodControlMessages(func(currentChannelID uint64) (rriptMonoDirectionSession.ControlMessage, error) {
		ctrl, err := s.rx.GenerateControlMessage(currentChannelID)
		if err != nil {
			return rriptMonoDirectionSession.ControlMessage{}, err
		}
		return s.withLocalSessionControl(ctrl), nil
	}); err != nil {
		return stats, err
	}
	stats.ControlPacketsSent += uint32(s.tx.ChannelCount())
	return stats, nil
}

func (s *BidirectionalSession) withLocalSessionControl(ctrl rriptMonoDirectionSession.ControlMessage) rriptMonoDirectionSession.ControlMessage {
	ctrl.Session = rriptMonoDirectionSession.SessionControlMessage{
		InstanceID: s.localSessionInstanceID,
	}
	return ctrl
}

func (s *BidirectionalSession) shouldSendManagerHostedControlLocked(
	timestamp uint64,
	payload []byte,
	stats rriptMonoDirectionSession.TickStats,
) bool {
	if len(s.lastControlPayload) == 0 {
		return true
	}
	if !bytes.Equal(s.lastControlPayload, payload) {
		return true
	}
	if stats.RepairPacketsGenerated > 0 || stats.RepairPacketsSent > 0 || stats.BlockedBySharedSendBudget {
		return true
	}
	if s.lastControlTimestamp == 0 {
		return true
	}
	return timestamp-s.lastControlTimestamp >= uint64(s.managerHostedControlKeepaliveIntervalTicks())
}

func (s *BidirectionalSession) managerHostedControlKeepaliveIntervalTicks() int {
	if s == nil || s.cond == nil {
		return managerHostedControlKeepaliveIntervalTicks
	}
	if s.tx == nil {
		return managerHostedControlKeepaliveIntervalTicks
	}
	if s.managerHostedControlKeepaliveInterval == 0 {
		return managerHostedControlKeepaliveIntervalTicks
	}
	return s.managerHostedControlKeepaliveInterval
}

func (s *BidirectionalSession) AttachTxChannel(closer io.WriteCloser) (channelID uint64, err error) {
	return s.AttachTxChannelWithConfig(closer, rriptMonoDirectionSession.ChannelConfig{Weight: 1})
}

func (s *BidirectionalSession) ConsumeSourcePacketActivity() uint64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := s.sourcePacketActivity
	s.sourcePacketActivity = 0
	return count
}

func (s *BidirectionalSession) SetDynamicRestrictSourceDataWhenOldestLaneStalled(enabled bool) {
	if s == nil || s.tx == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tx.SetDynamicRestrictSourceDataWhenOldestLaneStalled(enabled)
}

func (s *BidirectionalSession) RestrictSourceDataWhenOldestLaneStalledEnabled() bool {
	if s == nil || s.tx == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tx.RestrictSourceDataWhenOldestLaneStalledEnabled()
}

func (s *BidirectionalSession) AttachTxChannelWithConfig(
	closer io.WriteCloser,
	config rriptMonoDirectionSession.ChannelConfig,
) (channelID uint64, err error) {
	if s == nil || s.tx == nil || s.rx == nil {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	channelID, err = s.tx.AttachTxChannelWithConfig(closer, config)
	if err != nil {
		return 0, err
	}
	return channelID, nil
}

func (s *BidirectionalSession) AttachRxChannel() (rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	if s == nil || s.rx == nil {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	rxChannel, err = s.rx.AttachRxChannel(0)
	if err != nil {
		return nil, err
	}
	return rxChannel, nil
}

func (s *BidirectionalSession) AttachChannel(closer io.WriteCloser) (channelID uint64, rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	return s.AttachChannelWithConfig(closer, rriptMonoDirectionSession.ChannelConfig{Weight: 1})
}

func (s *BidirectionalSession) AttachChannelWithConfig(
	closer io.WriteCloser,
	config rriptMonoDirectionSession.ChannelConfig,
) (channelID uint64, rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	if s == nil {
		return 0, nil, nil
	}
	channelID, err = s.AttachTxChannelWithConfig(closer, config)
	if err != nil {
		return 0, nil, err
	}
	rxChannel, err = s.AttachRxChannel()
	if err != nil {
		return 0, nil, err
	}
	return channelID, rxChannel, nil
}

func (s *BidirectionalSession) Rx() *rriptMonoDirectionSession.SessionRx {
	if s == nil {
		return nil
	}
	return s.rx
}

func (s *BidirectionalSession) OnLogicalPacket(data []byte) error {
	if s == nil || s.rx == nil {
		return io.ErrClosedPipe
	}
	return s.rx.OnLogicalPacket(data)
}

func (s *BidirectionalSession) Tx() *rriptMonoDirectionSession.SessionTx {
	if s == nil {
		return nil
	}
	return s.tx
}

func (s *BidirectionalSession) MaxMessageSize() (int, error) {
	if s == nil || s.tx == nil {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tx.MaxMessageSize()
}

func (s *BidirectionalSession) startAutoTick() {
	if s == nil || s.TimestampInterval <= 0 {
		return
	}

	s.autoTickStop = make(chan struct{})
	s.autoTickDone = make(chan struct{})
	go func() {
		ticker := time.NewTicker(s.TimestampInterval)
		defer func() {
			ticker.Stop()
			close(s.autoTickDone)
		}()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				s.nextTimestamp += 1
				_, err := s.onNewTimestampLocked(s.nextTimestamp)
				s.mu.Unlock()
				if err != nil {
					return
				}
			case <-s.autoTickStop:
				return
			}
		}
	}()
}
