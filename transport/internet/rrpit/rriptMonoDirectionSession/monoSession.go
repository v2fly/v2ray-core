package rriptMonoDirectionSession

import (
	"bytes"
	"errors"
	"io"
	"math"
	"sync"
	"time"

	"github.com/lunixbochs/struc"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferLane"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type SessionRx struct {
	mu sync.Mutex

	rxLanes
	rxChannels

	DataPacketKind    uint8
	ControlPacketKind uint8

	OnMessage              func([]byte) error
	OnRemoteControlMessage func(ControlMessage) error
}

type SessionTx struct {
	txLanes
	txChannels
	txChannelsConfig  []ChannelStatus
	OddChannelIDs     bool
	Reconstruction    SessionTxReconstructionConfig
	DataPacketKind    uint8
	ControlPacketKind uint8
	SendEnforced      func(kind uint8, payload []byte) error
	SendIgnoreQuota   func(kind uint8, payload []byte) error
	HasRemainingQuota func() bool

	currentTimestamp                               uint64
	currentTimestampInitialized                    bool
	dynamicRestrictSourceDataWhenOldestLaneStalled bool
}

type ChannelStatus struct {
	Config             ChannelConfig
	Status             ChannelRateControlStatus
	MaterializeChannel *rrpitMaterializedTransferChannel.ChannelTx
}

type ChannelConfig struct {
	Weight          int
	MaxSendingSpeed int
}

type ChannelRateControlStatus struct {
	TimestampLastSent                  uint64
	PacketSentCurrentTimestamp         uint64
	EnforcedPacketSentCurrentTimestamp uint64
}

type SessionTxConfig struct {
	LaneShardSize                  int
	MaxDataShardsPerLane           int
	MaxBufferedLanes               int
	MaxRewindableTimestampNum      int
	MaxRewindableControlMessageNum int
	OddChannelIDs                  bool
	DataPacketKind                 uint8
	ControlPacketKind              uint8
	SendEnforced                   func(kind uint8, payload []byte) error
	SendIgnoreQuota                func(kind uint8, payload []byte) error
	HasRemainingQuota              func() bool
	Reconstruction                 SessionTxReconstructionConfig
}

type SessionTxReconstructionConfig struct {
	InitialRepairShardRatio                       float64
	LaneRepairWeight                              []float64
	SecondaryRepairShardRatio                     float64
	TimeResendSecondaryRepairShard                int
	StaleLaneFinalizedAgeThresholdTicks           int
	StaleLaneProgressStallThresholdTicks          int
	SecondaryRepairMinBurst                       int
	AlwaysRestrictSourceDataWhenOldestLaneStalled bool
}

type SessionRxConfig struct {
	LaneShardSize              int
	MaxBufferedLanes           int
	RemoteMaxDataShardsPerLane int
	DataPacketKind             uint8
	ControlPacketKind          uint8
	OnMessage                  func([]byte) error
	OnRemoteControlMsg         func(ControlMessage) error
}

type TickStats struct {
	RepairPacketsGenerated    uint32
	RepairPacketsSent         uint32
	ControlPacketsGenerated   uint32
	ControlPacketsSent        uint32
	BlockedBySharedSendBudget bool
}

type txLanes struct {
	laneShardSize        int
	maxDataShardsPerLane int
	maxBufferedLanes     int

	firstLaneID int64
	lanes       []*txLane
}

type txLane struct {
	LaneID uint64

	TransferLane *rrpitTransferLane.TransferLaneTx

	DataShards                     uint32
	TotalDataShards                uint32
	Finalized                      bool
	PeerSeenChunks                 uint16
	PeerSeenChunksKnown            bool
	RepairPackets                  uint32
	InitialRepairPacketsPending    uint32
	SecondaryRepairPacketsPending  uint32
	SecondaryRepairPacketsPerBurst uint32
	NextSecondaryRepairTimestamp   uint64
	PeerReconstructed              bool
	CreatedAtTimestamp             uint64
	FinalizedAtTimestamp           uint64
	LastProgressTimestamp          uint64
}

type txChannels struct {
	nextChannelID                  uint64
	maxRewindableTimestampNum      int
	maxRewindableControlMessageNum int
}

type rxLanes struct {
	laneShardSize              int
	maxBufferedLanes           int
	remoteMaxDataShardsPerLane int

	firstLaneID int64
	lanes       []*rxLane
}

type rxLane struct {
	LaneID uint64

	TransferLane     *rrpitTransferLane.TransferLaneRx
	Reconstructed    []rrpitTransferLane.ReconstructionData
	Ready            bool
	NextDeliverIndex int
}

type rxChannels struct {
	channels []*rxChannel
}

type rxChannel struct {
	MaterializeChannel *rrpitMaterializedTransferChannel.ChannelRx
}

type sessionDataPacket struct {
	PacketKind uint8
	LaneID     uint64
	Transfer   rrpitTransferLane.TransferData
}

type sessionControlPacket struct {
	PacketKind uint8
	Control    ControlMessage
}

const (
	defaultLaneShardSize                  = 1200
	defaultMaxDataShardsPerLane           = 32
	defaultMaxBufferedLanes               = 64
	defaultMaxRewindableTimestampNum      = 256
	defaultMaxRewindableControlMessageNum = 256
	defaultStaleLaneFinalizedAgeTicks     = 8
	defaultStaleLaneProgressStallTicks    = 8
)

var ErrTxLaneBufferFull = errors.New("too many buffered transfer lanes")

func NewSessionTx(config SessionTxConfig) (*SessionTx, error) {
	tx := &SessionTx{
		OddChannelIDs:     config.OddChannelIDs,
		Reconstruction:    config.Reconstruction,
		DataPacketKind:    config.DataPacketKind,
		ControlPacketKind: config.ControlPacketKind,
		SendEnforced:      config.SendEnforced,
		SendIgnoreQuota:   config.SendIgnoreQuota,
		HasRemainingQuota: config.HasRemainingQuota,
	}
	if len(config.Reconstruction.LaneRepairWeight) > 0 {
		tx.Reconstruction.LaneRepairWeight = append([]float64(nil), config.Reconstruction.LaneRepairWeight...)
	}
	tx.laneShardSize = config.LaneShardSize
	tx.maxDataShardsPerLane = config.MaxDataShardsPerLane
	tx.maxBufferedLanes = config.MaxBufferedLanes
	tx.maxRewindableTimestampNum = config.MaxRewindableTimestampNum
	tx.maxRewindableControlMessageNum = config.MaxRewindableControlMessageNum
	if err := tx.ensureDefaults(); err != nil {
		return nil, err
	}
	return tx, nil
}

func NewSessionRx(config SessionRxConfig) (*SessionRx, error) {
	rx := &SessionRx{
		DataPacketKind:         config.DataPacketKind,
		ControlPacketKind:      config.ControlPacketKind,
		OnMessage:              config.OnMessage,
		OnRemoteControlMessage: config.OnRemoteControlMsg,
	}
	rx.laneShardSize = config.LaneShardSize
	rx.maxBufferedLanes = config.MaxBufferedLanes
	rx.remoteMaxDataShardsPerLane = config.RemoteMaxDataShardsPerLane
	if err := rx.ensureDefaults(); err != nil {
		return nil, err
	}
	return rx, nil
}

func (t *SessionTx) OnNewTimestamp(timestamp uint64) error {
	_, err := t.onNewTimestamp(timestamp)
	return err
}

func (t *SessionTx) OnNewTimestampWithStats(timestamp uint64) (TickStats, error) {
	return t.onNewTimestamp(timestamp)
}

func (t *SessionTx) FloodControlMessageToAllChannels() error {
	return t.FloodControlMessages(func(uint64) (ControlMessage, error) {
		return ControlMessage{}, nil
	})
}

func (t *SessionTx) FloodControlMessages(generator func(uint64) (ControlMessage, error)) error {
	if generator == nil {
		return newError("nil control message generator")
	}
	return t.floodControlMessages(generator)
}

func (t *SessionTx) SendMessage(data []byte) error {
	return t.sendMessage(data)
}

func (t *SessionTx) MaxMessageSize() (int, error) {
	if t == nil {
		return 0, nil
	}
	if err := t.ensureDefaults(); err != nil {
		return 0, err
	}
	maxMessageSize := t.laneShardSize - rrpitTransferLane.ReconstructionLengthFieldSize
	if maxMessageSize <= 0 {
		return 0, newError("lane shard size too small")
	}
	return maxMessageSize, nil
}

func (t *SessionTx) AcceptRemoteControlMessage(ctrl ControlMessage) error {
	return t.acceptRemoteControlMessage(ctrl)
}

func (t *SessionTx) SetDynamicRestrictSourceDataWhenOldestLaneStalled(enabled bool) {
	if t == nil {
		return
	}
	t.dynamicRestrictSourceDataWhenOldestLaneStalled = enabled
}

func (t *SessionTx) RestrictSourceDataWhenOldestLaneStalledEnabled() bool {
	if t == nil {
		return false
	}
	return t.Reconstruction.AlwaysRestrictSourceDataWhenOldestLaneStalled || t.dynamicRestrictSourceDataWhenOldestLaneStalled
}

func (t *SessionTx) ChannelCount() int {
	if t == nil {
		return 0
	}
	return len(t.txChannelsConfig)
}

func (t *SessionTx) AttachTxChannel(closer io.WriteCloser) (channelID uint64, err error) {
	return t.AttachTxChannelWithConfig(closer, ChannelConfig{Weight: 1})
}

func (t *SessionTx) AttachTxChannelWithConfig(closer io.WriteCloser, config ChannelConfig) (channelID uint64, err error) {
	if closer == nil {
		return 0, newError("nil channel writer")
	}
	if err := t.ensureDefaults(); err != nil {
		return 0, err
	}

	channelID = t.allocateChannelID()
	channel, err := rrpitMaterializedTransferChannel.NewChannelTx(
		channelID,
		closer,
		t.maxRewindableTimestampNum,
		t.maxRewindableControlMessageNum,
	)
	if err != nil {
		return 0, err
	}

	if config.Weight == 0 {
		config.Weight = 1
	}
	t.txChannelsConfig = append(t.txChannelsConfig, ChannelStatus{
		Config:             config,
		MaterializeChannel: channel,
	})
	return channelID, nil
}

func (r *SessionRx) AttachRxChannel(channelID uint64) (*rrpitMaterializedTransferChannel.ChannelRx, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureDefaults(); err != nil {
		return nil, err
	}
	if channelID != 0 && r.hasAttachedChannelID(channelID, nil) {
		return nil, newError("duplicate rx channel id")
	}

	channelState := &rxChannel{}
	channel, err := rrpitMaterializedTransferChannel.NewChannelRx(channelID, func(data []byte) error {
		return r.onChannelData(channelState, data)
	})
	if err != nil {
		return nil, err
	}
	channelState.MaterializeChannel = channel
	r.channels = append(r.channels, channelState)
	return channel, nil
}

func (r *SessionRx) GenerateControlMessage(currentChannelID uint64) (ControlMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureDefaults(); err != nil {
		return ControlMessage{}, err
	}

	ctrl := ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: currentChannelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo: r.firstLaneID - 1,
		},
	}

	if len(r.lanes) > int(^uint16(0)) {
		return ControlMessage{}, newError("lane control count exceeds control field capacity")
	}
	ctrl.Lane.LaneControl = make([]rrpitTransferLane.TransferControl, len(r.lanes))
	for i, lane := range r.lanes {
		if lane == nil {
			continue
		}
		laneCtrl, err := lane.TransferLane.GenerateControl()
		if err != nil {
			return ControlMessage{}, err
		}
		ctrl.Lane.LaneControl[i] = laneCtrl
	}
	ctrl.Lane.LenLaneControl = uint16(len(ctrl.Lane.LaneControl))

	if len(r.channels) > int(^uint16(0)) {
		return ControlMessage{}, newError("channel control count exceeds control field capacity")
	}
	ctrl.Channel.ChannelControl = make([]rrpitTransferChannel.ChannelControlMessage, 0, len(r.channels))
	for _, channel := range r.channels {
		if channel == nil || channel.MaterializeChannel == nil || channel.MaterializeChannel.ChannelID == 0 {
			continue
		}
		channelCtrl, err := channel.MaterializeChannel.CreateControlMessage()
		if err != nil {
			return ControlMessage{}, err
		}
		ctrl.Channel.ChannelControl = append(ctrl.Channel.ChannelControl, *channelCtrl)
	}
	if len(ctrl.Channel.ChannelControl) > int(^uint16(0)) {
		return ControlMessage{}, newError("channel control count exceeds control field capacity")
	}
	ctrl.Channel.LenChannelControl = uint16(len(ctrl.Channel.ChannelControl))
	return ctrl, nil
}

func (r *SessionRx) GenerateStrippedControlMessage() (ControlMessage, error) {
	ctrl, err := r.GenerateControlMessage(0)
	if err != nil {
		return ControlMessage{}, err
	}
	ctrl.FloodChannel = SessionFloodChannelControlMessage{}
	ctrl.Channel = SessionChannelControlMessage{}
	return ctrl, nil
}

func (r *SessionRx) ensureDefaults() error {
	if r.OnMessage == nil {
		return newError("nil OnMessage callback")
	}
	if r.DataPacketKind == 0 {
		r.DataPacketKind = PacketKind_InteractiveStreamData
	}
	if r.ControlPacketKind == 0 {
		r.ControlPacketKind = PacketKind_InteractiveStreamControl
	}
	if r.laneShardSize == 0 {
		r.laneShardSize = defaultLaneShardSize
	}
	if r.maxBufferedLanes == 0 {
		r.maxBufferedLanes = defaultMaxBufferedLanes
	}
	if r.maxBufferedLanes < 0 {
		return newError("invalid max buffered lanes")
	}
	if _, err := rrpitTransferLane.NewTransferLaneRx(r.laneShardSize, r.remoteMaxDataShardsPerLane); err != nil {
		return err
	}
	return nil
}

func (t *SessionTx) ensureDefaults() error {
	if t.DataPacketKind == 0 {
		t.DataPacketKind = PacketKind_InteractiveStreamData
	}
	if t.ControlPacketKind == 0 {
		t.ControlPacketKind = PacketKind_InteractiveStreamControl
	}
	if t.laneShardSize == 0 {
		t.laneShardSize = defaultLaneShardSize
	}
	if t.maxDataShardsPerLane == 0 {
		t.maxDataShardsPerLane = defaultMaxDataShardsPerLane
	}
	if t.maxBufferedLanes == 0 {
		t.maxBufferedLanes = defaultMaxBufferedLanes
	}
	if t.maxRewindableTimestampNum == 0 {
		t.maxRewindableTimestampNum = defaultMaxRewindableTimestampNum
	}
	if t.maxRewindableControlMessageNum == 0 {
		t.maxRewindableControlMessageNum = defaultMaxRewindableControlMessageNum
	}

	if t.maxBufferedLanes < 0 {
		return newError("invalid max buffered lanes")
	}
	if t.Reconstruction.InitialRepairShardRatio < 0 {
		return newError("invalid initial repair shard ratio")
	}
	if t.Reconstruction.SecondaryRepairShardRatio < 0 {
		return newError("invalid secondary repair shard ratio")
	}
	if t.Reconstruction.TimeResendSecondaryRepairShard < 0 {
		return newError("invalid secondary repair resend interval")
	}
	if t.Reconstruction.StaleLaneFinalizedAgeThresholdTicks == 0 {
		t.Reconstruction.StaleLaneFinalizedAgeThresholdTicks = defaultStaleLaneFinalizedAgeTicks
	}
	if t.Reconstruction.StaleLaneProgressStallThresholdTicks == 0 {
		t.Reconstruction.StaleLaneProgressStallThresholdTicks = defaultStaleLaneProgressStallTicks
	}
	if t.Reconstruction.StaleLaneFinalizedAgeThresholdTicks < 0 {
		return newError("invalid stale lane finalized age threshold")
	}
	if t.Reconstruction.StaleLaneProgressStallThresholdTicks < 0 {
		return newError("invalid stale lane progress stall threshold")
	}
	if t.Reconstruction.SecondaryRepairMinBurst < 0 {
		return newError("invalid secondary repair minimum burst")
	}
	for _, weight := range t.Reconstruction.LaneRepairWeight {
		if weight < 0 {
			return newError("invalid lane repair weight")
		}
	}
	if _, err := rrpitTransferLane.NewTransferLaneTx(t.laneShardSize, t.maxDataShardsPerLane); err != nil {
		return err
	}
	if _, err := rrpitTransferChannel.NewChannelTx(0, t.maxRewindableTimestampNum, t.maxRewindableControlMessageNum); err != nil {
		return err
	}
	return nil
}

func (t *SessionTx) onNewTimestamp(timestamp uint64) (TickStats, error) {
	if err := t.ensureDefaults(); err != nil {
		return TickStats{}, err
	}
	stats := TickStats{}
	t.currentTimestamp = timestamp
	t.currentTimestampInitialized = true
	t.resetChannelRateWindow(timestamp)

	if t.hasCustomReconstructionConfig() {
		t.finalizeLatestOpenLane()
		t.scheduleSecondaryRepairResends(timestamp)

		opportunisticBudget := t.buildOpportunisticRepairBudget()
		for {
			lane, kind, index := t.nextConfiguredRepair(opportunisticBudget)
			if lane == nil {
				return stats, nil
			}
			if !t.hasRepairSendCapacity() {
				stats.BlockedBySharedSendBudget = true
				return stats, nil
			}
			stats.RepairPacketsGenerated += 1
			if err := t.sendRepairPacket(lane); err != nil {
				return stats, err
			}
			stats.RepairPacketsSent += 1
			switch kind {
			case repairSendInitial:
				lane.InitialRepairPacketsPending -= 1
			case repairSendSecondary:
				lane.SecondaryRepairPacketsPending -= 1
				if lane.SecondaryRepairPacketsPending == 0 {
					t.scheduleNextSecondaryRepair(lane)
				}
			case repairSendOpportunistic:
				opportunisticBudget[index] -= 1
			}
		}
	}

	lane := t.nextLaneForRepair()
	if lane == nil {
		return stats, nil
	}
	if !t.hasRepairSendCapacity() {
		stats.BlockedBySharedSendBudget = true
		return stats, nil
	}
	stats.RepairPacketsGenerated = 1
	if err := t.sendRepairPacket(lane); err != nil {
		return stats, err
	}
	stats.RepairPacketsSent = 1
	return stats, nil
}

func (r *SessionRx) onChannelData(channel *rxChannel, data []byte) error {
	return r.onPacketData(channel, data)
}

func (r *SessionRx) OnLogicalPacket(data []byte) error {
	return r.onPacketData(nil, data)
}

func (r *SessionRx) onPacketData(channel *rxChannel, data []byte) error {
	if err := r.ensureDefaults(); err != nil {
		return err
	}
	if len(data) == 0 {
		return newError("session packet too short")
	}

	switch data[0] {
	case r.DataPacketKind:
		var packet sessionDataPacket
		if err := struc.Unpack(bytes.NewReader(data), &packet); err != nil {
			return newError("failed to unpack session data packet: ", err)
		}
		r.mu.Lock()
		payloads, err := r.onSessionDataLocked(packet)
		r.mu.Unlock()
		if err != nil {
			return err
		}
		return r.deliverMessages(payloads)
	case r.ControlPacketKind:
		var packet sessionControlPacket
		if err := struc.Unpack(bytes.NewReader(data), &packet); err != nil {
			return newError("failed to unpack session control packet: ", err)
		}
		r.mu.Lock()
		if channel != nil && channel.MaterializeChannel != nil && packet.Control.FloodChannel.CurrentChannelID != 0 {
			if r.hasAttachedChannelID(packet.Control.FloodChannel.CurrentChannelID, channel) {
				r.mu.Unlock()
				return newError("duplicate rx channel id")
			}
			if err := channel.MaterializeChannel.AssignChannelID(packet.Control.FloodChannel.CurrentChannelID); err != nil {
				r.mu.Unlock()
				return err
			}
		}
		callback := r.OnRemoteControlMessage
		r.mu.Unlock()
		if callback != nil {
			return callback(packet.Control)
		}
		return nil
	default:
		return newError("unknown session packet kind")
	}
}

func (r *SessionRx) onSessionDataLocked(packet sessionDataPacket) ([][]byte, error) {
	lane, err := r.ensureLane(packet.LaneID)
	if err != nil {
		return nil, err
	}
	if lane == nil {
		return nil, nil
	}

	done, err := lane.TransferLane.AddTransferData(packet.Transfer)
	if err != nil {
		return nil, err
	}
	if done && !lane.Ready {
		reconstructed, err := lane.TransferLane.Reconstruct()
		if err != nil {
			if rrpitTransferLane.IsNotEnoughSymbolsToReconstruct(err) {
				return nil, nil
			}
			return nil, err
		}
		lane.Reconstructed = reconstructed
		lane.Ready = true
	}
	return r.drainReadyLanesLocked()
}

func (r *SessionRx) ensureLane(laneID uint64) (*rxLane, error) {
	if laneID > ^uint64(0)>>1 {
		return nil, newError("lane id exceeds supported range")
	}
	laneIDInt := int64(laneID)
	if laneIDInt < r.firstLaneID {
		return nil, nil
	}

	index := int(laneIDInt - r.firstLaneID)
	if r.maxBufferedLanes > 0 && index >= r.maxBufferedLanes {
		return nil, newError("too many buffered transfer lanes")
	}
	for len(r.lanes) <= index {
		r.lanes = append(r.lanes, nil)
	}
	if r.lanes[index] != nil {
		return r.lanes[index], nil
	}

	transferLane, err := rrpitTransferLane.NewTransferLaneRx(r.laneShardSize, r.remoteMaxDataShardsPerLane)
	if err != nil {
		return nil, err
	}
	lane := &rxLane{
		LaneID:       laneID,
		TransferLane: transferLane,
	}
	r.lanes[index] = lane
	return lane, nil
}

func (r *SessionRx) deliverMessages(payloads [][]byte) error {
	for _, payload := range payloads {
		if err := r.OnMessage(payload); err != nil {
			return err
		}
	}
	return nil
}

func (r *SessionRx) drainReadyLanesLocked() ([][]byte, error) {
	payloads := make([][]byte, 0)
	for len(r.lanes) > 0 {
		lane := r.lanes[0]
		if lane == nil || !lane.Ready {
			return payloads, nil
		}

		for lane.NextDeliverIndex < len(lane.Reconstructed) {
			payload := append([]byte(nil), lane.Reconstructed[lane.NextDeliverIndex].Data...)
			payloads = append(payloads, payload)
			lane.NextDeliverIndex += 1
		}

		r.lanes = r.lanes[1:]
		r.firstLaneID += 1
	}
	return payloads, nil
}

func (r *SessionRx) hasAttachedChannelID(channelID uint64, exclude *rxChannel) bool {
	for _, attached := range r.channels {
		if attached == nil || attached == exclude || attached.MaterializeChannel == nil {
			continue
		}
		if attached.MaterializeChannel.ChannelID == channelID {
			return true
		}
	}
	return false
}

func (t *SessionTx) floodControlMessages(generator func(uint64) (ControlMessage, error)) error {
	if err := t.ensureDefaults(); err != nil {
		return err
	}
	for i := range t.txChannelsConfig {
		channel := t.txChannelsConfig[i].MaterializeChannel
		if channel == nil {
			continue
		}
		ctrl, err := generator(channel.ChannelID)
		if err != nil {
			return err
		}
		ctrl.FloodChannel.CurrentChannelID = channel.ChannelID
		payload, err := marshalSessionControlPacket(t.ControlPacketKind, ctrl)
		if err != nil {
			return err
		}
		if err := channel.SendDataMessage(payload); err != nil {
			return err
		}
		t.markChannelSent(i)
	}
	return nil
}

func (t *SessionTx) sendMessage(data []byte) error {
	if err := t.ensureDefaults(); err != nil {
		return err
	}
	if t.shouldBackpressureSourceData() {
		return ErrTxLaneBufferFull
	}
	if len(t.txChannelsConfig) == 0 {
		if t.SendIgnoreQuota == nil && t.SendEnforced == nil {
			return newError("no transfer channel attached")
		}
	}

	lane := t.latestAppendableLane()
	if lane == nil {
		var err error
		lane, err = t.createLane()
		if err != nil {
			return err
		}
		transfer, err := lane.TransferLane.AddData(data)
		if err != nil {
			t.removeNewestLane()
			return err
		}
		lane.DataShards += 1
		if t.hasCustomReconstructionConfig() && t.shouldFinalizeLaneAfterData(lane) {
			t.finalizeLane(lane)
		}
		if err := t.sendTransferPacket(lane.LaneID, *transfer, false); err != nil {
			return err
		}
		return t.flushInitialRepairPackets(lane)
	}

	transfer, err := lane.TransferLane.AddData(data)
	if err != nil {
		lane, err = t.createLane()
		if err != nil {
			return err
		}
		transfer, err = lane.TransferLane.AddData(data)
		if err != nil {
			t.removeNewestLane()
			return err
		}
	}

	lane.DataShards += 1
	if t.hasCustomReconstructionConfig() && t.shouldFinalizeLaneAfterData(lane) {
		t.finalizeLane(lane)
	}
	if err := t.sendTransferPacket(lane.LaneID, *transfer, false); err != nil {
		return err
	}
	return t.flushInitialRepairPackets(lane)
}

func (t *SessionTx) shouldBackpressureSourceData() bool {
	if t == nil || !t.hasCustomReconstructionConfig() || !t.RestrictSourceDataWhenOldestLaneStalledEnabled() || len(t.lanes) == 0 {
		return false
	}
	oldest := t.lanes[0]
	return t.isStaleOldestLane(oldest)
}

func (t *SessionTx) acceptRemoteControlMessage(ctrl ControlMessage) error {
	if err := t.ensureDefaults(); err != nil {
		return err
	}
	for _, channelControl := range ctrl.Channel.ChannelControl {
		idx := t.channelIndexByID(channelControl.ChannelID)
		if idx < 0 {
			continue
		}
		channel := t.txChannelsConfig[idx].MaterializeChannel
		if channel == nil {
			continue
		}
		if err := channel.AcceptControlMessage(channelControl); err != nil {
			return err
		}
	}

	t.dropLanesThrough(ctrl.Lane.LaneACKTo)
	baseLaneID := ctrl.Lane.LaneACKTo + 1
	for i, laneControl := range ctrl.Lane.LaneControl {
		laneID := baseLaneID + int64(i)
		lane := t.laneByID(laneID)
		if lane == nil {
			continue
		}
		previousSeen := lane.PeerSeenChunks
		previousKnown := lane.PeerSeenChunksKnown
		previousComplete := lane.PeerReconstructed
		if err := lane.TransferLane.AcceptControlData(laneControl); err != nil {
			return err
		}
		if laneControl.SeenChunks == rrpitTransferLane.SeenChunksCompletionSentinel {
			lane.PeerReconstructed = true
			lane.PeerSeenChunks = laneControl.SeenChunks
		} else if laneControl.SeenChunks > lane.PeerSeenChunks {
			lane.PeerSeenChunks = laneControl.SeenChunks
		}
		lane.PeerSeenChunksKnown = true
		if !previousKnown || lane.PeerReconstructed != previousComplete || lane.PeerSeenChunks > previousSeen {
			lane.LastProgressTimestamp = t.lifecycleTimestamp()
		}
		if t.hasCustomReconstructionConfig() {
			t.updateLaneRepairStateAfterControl(lane, !previousKnown || lane.PeerReconstructed != previousComplete || lane.PeerSeenChunks > previousSeen)
		}
	}
	return nil
}

func (t *SessionTx) resetChannelRateWindow(timestamp uint64) {
	for i := range t.txChannelsConfig {
		if t.txChannelsConfig[i].Status.TimestampLastSent != timestamp {
			t.txChannelsConfig[i].Status.PacketSentCurrentTimestamp = 0
			t.txChannelsConfig[i].Status.EnforcedPacketSentCurrentTimestamp = 0
		}
	}
}

func (t *SessionTx) nextLaneForRepair() *txLane {
	if len(t.lanes) == 0 {
		return nil
	}

	lastLane := t.lanes[len(t.lanes)-1]
	if !lastLane.Finalized && lastLane.DataShards > 0 {
		return lastLane
	}

	var selected *txLane
	for _, lane := range t.lanes {
		if lane.DataShards == 0 {
			continue
		}
		if selected == nil || lane.RepairPackets < selected.RepairPackets {
			selected = lane
		}
	}
	return selected
}

type repairSendKind int

const (
	repairSendNone repairSendKind = iota
	repairSendInitial
	repairSendSecondary
	repairSendOpportunistic
)

func (t *SessionTx) hasCustomReconstructionConfig() bool {
	return t.Reconstruction.InitialRepairShardRatio > 0 ||
		t.Reconstruction.SecondaryRepairShardRatio > 0 ||
		t.Reconstruction.SecondaryRepairMinBurst > 0 ||
		t.Reconstruction.TimeResendSecondaryRepairShard > 0 ||
		len(t.Reconstruction.LaneRepairWeight) > 0
}

func (t *SessionTx) shouldFinalizeLaneAfterData(lane *txLane) bool {
	if lane == nil || lane.Finalized || t.maxDataShardsPerLane <= 0 {
		return false
	}
	return int(lane.DataShards) >= t.maxDataShardsPerLane
}

func (t *SessionTx) finalizeLatestOpenLane() {
	lane := t.latestAppendableLane()
	if lane == nil || lane.DataShards == 0 {
		return
	}
	t.finalizeLane(lane)
}

func (t *SessionTx) finalizeLane(lane *txLane) {
	if lane == nil || lane.Finalized || lane.DataShards == 0 {
		return
	}
	lane.Finalized = true
	if lane.TotalDataShards == 0 {
		lane.TotalDataShards = lane.DataShards
	}
	lane.FinalizedAtTimestamp = t.lifecycleTimestamp()
	lane.LastProgressTimestamp = lane.FinalizedAtTimestamp
	if !t.hasCustomReconstructionConfig() {
		return
	}
	lane.InitialRepairPacketsPending = repairPacketQuota(t.Reconstruction.InitialRepairShardRatio, lane.TotalDataShards)
	if lane.PeerSeenChunksKnown {
		t.updateLaneRepairStateAfterControl(lane, true)
	}
}

func (t *SessionTx) flushInitialRepairPackets(lane *txLane) error {
	if !t.hasCustomReconstructionConfig() || lane == nil {
		return nil
	}
	for lane.InitialRepairPacketsPending > 0 {
		if !t.hasRepairSendCapacity() {
			return nil
		}
		if err := t.sendRepairPacket(lane); err != nil {
			return err
		}
		lane.InitialRepairPacketsPending -= 1
	}
	if lane.InitialRepairPacketsPending == 0 {
		t.scheduleNextSecondaryRepair(lane)
	}
	return nil
}

func (t *SessionTx) sendRepairPacket(lane *txLane) error {
	if lane == nil {
		return nil
	}
	if !lane.Finalized {
		t.finalizeLane(lane)
	}
	transfer, err := lane.TransferLane.CreateReconstructionTransmissionData()
	if err != nil {
		return err
	}
	if err := t.sendTransferPacket(lane.LaneID, transfer, true); err != nil {
		return err
	}
	lane.RepairPackets += 1
	return nil
}

func repairPacketQuota(ratio float64, shardCount uint32) uint32 {
	if ratio <= 0 || shardCount == 0 {
		return 0
	}
	quota := math.Ceil(ratio * float64(shardCount))
	if quota <= 0 {
		return 0
	}
	if quota >= float64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(quota)
}

func (t *SessionTx) lifecycleTimestamp() uint64 {
	if t.currentTimestampInitialized {
		return t.currentTimestamp
	}
	return 0
}

func (t *SessionTx) peerSeenChunksCount(lane *txLane) uint32 {
	if lane == nil {
		return 0
	}
	if lane.PeerReconstructed {
		return lane.TotalDataShards
	}
	if lane.PeerSeenChunks == rrpitTransferLane.SeenChunksCompletionSentinel {
		return lane.TotalDataShards
	}
	return uint32(lane.PeerSeenChunks)
}

func (t *SessionTx) laneSeenChunksTarget(lane *txLane) uint32 {
	if lane == nil {
		return 0
	}
	if lane.TotalDataShards == ^uint32(0) {
		return lane.TotalDataShards
	}
	return lane.TotalDataShards + 1
}

func (t *SessionTx) laneRepairDemandBase(lane *txLane) uint32 {
	if lane == nil || lane.TotalDataShards == 0 || lane.PeerReconstructed {
		return 0
	}
	if !lane.PeerSeenChunksKnown {
		return t.laneSeenChunksTarget(lane)
	}
	return t.laneMissingShards(lane)
}

func (t *SessionTx) isOldestUnackedLane(lane *txLane) bool {
	return lane != nil && lane.LaneID == uint64(t.firstLaneID)
}

func (t *SessionTx) laneNeedsStaleMonitoring(lane *txLane) bool {
	return lane != nil && lane.Finalized && !lane.PeerReconstructed && t.isOldestUnackedLane(lane)
}

func (t *SessionTx) isStaleOldestLane(lane *txLane) bool {
	if lane == nil || !lane.Finalized || lane.PeerReconstructed || !t.isOldestUnackedLane(lane) {
		return false
	}
	now := t.lifecycleTimestamp()
	finalizedAge := timestampAge(now, lane.FinalizedAtTimestamp)
	progressAge := timestampAge(now, lane.LastProgressTimestamp)
	return finalizedAge >= uint64(t.Reconstruction.StaleLaneFinalizedAgeThresholdTicks) ||
		progressAge >= uint64(t.Reconstruction.StaleLaneProgressStallThresholdTicks)
}

func timestampAge(now uint64, then uint64) uint64 {
	if now <= then {
		return 0
	}
	return now - then
}

func (t *SessionTx) secondaryRepairBurst(repairDemand uint32) uint32 {
	burst := repairPacketQuota(t.Reconstruction.SecondaryRepairShardRatio, repairDemand)
	minBurst := uint32(t.Reconstruction.SecondaryRepairMinBurst)
	if burst < minBurst {
		return minBurst
	}
	return burst
}

func (t *SessionTx) clearSecondaryRepairState(lane *txLane) {
	if lane == nil {
		return
	}
	lane.SecondaryRepairPacketsPending = 0
	lane.SecondaryRepairPacketsPerBurst = 0
	lane.NextSecondaryRepairTimestamp = 0
}

func (t *SessionTx) clearRepairState(lane *txLane) {
	if lane == nil {
		return
	}
	lane.InitialRepairPacketsPending = 0
	t.clearSecondaryRepairState(lane)
}

func (t *SessionTx) staleOldestLane() (*txLane, int) {
	if len(t.lanes) == 0 {
		return nil, -1
	}
	lane := t.lanes[0]
	if !t.isStaleOldestLane(lane) {
		return nil, -1
	}
	return lane, 0
}

func (t *SessionTx) laneMissingShards(lane *txLane) uint32 {
	if lane == nil || !lane.PeerSeenChunksKnown || lane.TotalDataShards == 0 || lane.PeerReconstructed {
		return 0
	}
	seen := t.peerSeenChunksCount(lane)
	target := t.laneSeenChunksTarget(lane)
	if seen >= target {
		return 0
	}
	return target - seen
}

func (t *SessionTx) laneRepairDemand(lane *txLane) uint32 {
	if lane == nil || lane.TotalDataShards == 0 || lane.PeerReconstructed {
		return 0
	}
	missing := t.laneRepairDemandBase(lane)
	if !t.isStaleOldestLane(lane) {
		return missing
	}
	if missing > 0 {
		if missing < 2 {
			return 2
		}
		return missing
	}
	return 2
}

func (t *SessionTx) updateLaneRepairStateAfterControl(lane *txLane, scheduleSecondary bool) {
	if lane == nil || !lane.PeerSeenChunksKnown || lane.TotalDataShards == 0 {
		return
	}
	repairDemand := t.laneRepairDemand(lane)
	if repairDemand == 0 {
		lane.InitialRepairPacketsPending = 0
		lane.SecondaryRepairPacketsPending = 0
		lane.SecondaryRepairPacketsPerBurst = 0
		if t.Reconstruction.TimeResendSecondaryRepairShard > 0 && t.laneNeedsStaleMonitoring(lane) {
			if scheduleSecondary && lane.NextSecondaryRepairTimestamp == 0 {
				lane.NextSecondaryRepairTimestamp = t.secondaryRepairScheduleBaseTimestamp() + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
			}
			return
		}
		lane.NextSecondaryRepairTimestamp = 0
		return
	}

	burst := t.secondaryRepairBurst(repairDemand)
	lane.SecondaryRepairPacketsPerBurst = burst
	if t.Reconstruction.TimeResendSecondaryRepairShard <= 0 {
		if burst == 0 {
			t.clearSecondaryRepairState(lane)
			return
		}
		if lane.SecondaryRepairPacketsPending > burst {
			lane.SecondaryRepairPacketsPending = burst
		}
		if !scheduleSecondary {
			return
		}

		lane.SecondaryRepairPacketsPending = burst
		lane.NextSecondaryRepairTimestamp = 0
		return
	}

	if burst == 0 {
		t.clearSecondaryRepairState(lane)
		return
	}

	if lane.SecondaryRepairPacketsPending > burst {
		lane.SecondaryRepairPacketsPending = burst
	}
	if lane.SecondaryRepairPacketsPending != 0 || !scheduleSecondary || lane.NextSecondaryRepairTimestamp != 0 {
		return
	}
	if lane.NextSecondaryRepairTimestamp == 0 {
		lane.NextSecondaryRepairTimestamp = t.secondaryRepairScheduleBaseTimestamp() + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
	}
}

func (t *SessionTx) scheduleSecondaryRepairResends(timestamp uint64) {
	if t.Reconstruction.TimeResendSecondaryRepairShard <= 0 {
		return
	}
	for _, lane := range t.lanes {
		if lane == nil || lane.SecondaryRepairPacketsPending != 0 {
			continue
		}
		repairDemand := t.laneRepairDemand(lane)
		if lane.NextSecondaryRepairTimestamp == 0 {
			if repairDemand > 0 {
				lane.SecondaryRepairPacketsPerBurst = t.secondaryRepairBurst(repairDemand)
				if lane.SecondaryRepairPacketsPerBurst == 0 {
					continue
				}
				lane.NextSecondaryRepairTimestamp = timestamp + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
				continue
			}
			if t.laneNeedsStaleMonitoring(lane) {
				lane.SecondaryRepairPacketsPerBurst = 0
				lane.NextSecondaryRepairTimestamp = timestamp + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
			}
			continue
		}
		if timestamp < lane.NextSecondaryRepairTimestamp {
			continue
		}
		if repairDemand == 0 {
			if t.laneNeedsStaleMonitoring(lane) {
				lane.SecondaryRepairPacketsPerBurst = 0
				lane.NextSecondaryRepairTimestamp = timestamp + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
				continue
			}
			t.clearSecondaryRepairState(lane)
			continue
		}

		burst := t.secondaryRepairBurst(repairDemand)
		lane.SecondaryRepairPacketsPerBurst = burst
		if burst == 0 {
			t.clearSecondaryRepairState(lane)
			continue
		}
		lane.SecondaryRepairPacketsPending = burst
		lane.NextSecondaryRepairTimestamp = 0
	}
}

func (t *SessionTx) secondaryRepairScheduleBaseTimestamp() uint64 {
	if t.currentTimestampInitialized {
		return t.currentTimestamp
	}
	return 0
}

func (t *SessionTx) scheduleNextSecondaryRepair(lane *txLane) {
	if lane == nil || t.Reconstruction.TimeResendSecondaryRepairShard <= 0 {
		return
	}
	if t.laneRepairDemand(lane) == 0 {
		if !t.laneNeedsStaleMonitoring(lane) {
			t.clearSecondaryRepairState(lane)
			return
		}
	}
	if lane.SecondaryRepairPacketsPending != 0 || lane.NextSecondaryRepairTimestamp != 0 {
		return
	}
	lane.SecondaryRepairPacketsPerBurst = 0
	lane.NextSecondaryRepairTimestamp = t.secondaryRepairScheduleBaseTimestamp() + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
}

func (t *SessionTx) hasRepairSendCapacity() bool {
	if t.HasRemainingQuota != nil {
		return t.HasRemainingQuota()
	}
	for i := range t.txChannelsConfig {
		if t.txChannelsConfig[i].MaterializeChannel == nil || t.channelRateLimited(i) {
			continue
		}
		return true
	}
	return false
}

func (t *SessionTx) buildOpportunisticRepairBudget() []uint32 {
	if len(t.Reconstruction.LaneRepairWeight) == 0 {
		return nil
	}
	budget := make([]uint32, len(t.lanes))
	for i, lane := range t.lanes {
		if lane == nil || !lane.Finalized {
			continue
		}
		missing := t.laneMissingShards(lane)
		if missing == 0 {
			continue
		}
		weight := t.laneRepairWeightForLane(i)
		if weight <= 0 {
			continue
		}
		budget[i] = repairPacketQuota(weight, missing)
	}
	return budget
}

func (t *SessionTx) laneRepairWeightForLane(index int) float64 {
	if index < 0 || len(t.Reconstruction.LaneRepairWeight) == 0 {
		return 0
	}
	if index < len(t.Reconstruction.LaneRepairWeight) {
		return t.Reconstruction.LaneRepairWeight[index]
	}
	return t.Reconstruction.LaneRepairWeight[len(t.Reconstruction.LaneRepairWeight)-1]
}

func (t *SessionTx) nextConfiguredRepair(opportunisticBudget []uint32) (*txLane, repairSendKind, int) {
	staleLane, staleIndex := t.staleOldestLane()
	if staleLane != nil && staleLane.SecondaryRepairPacketsPending > 0 {
		return staleLane, repairSendSecondary, staleIndex
	}
	for i, lane := range t.lanes {
		if lane == nil {
			continue
		}
		if t.laneRepairDemand(lane) == 0 {
			t.clearRepairState(lane)
			continue
		}
		if lane.InitialRepairPacketsPending > 0 {
			return lane, repairSendInitial, i
		}
	}
	for i, lane := range t.lanes {
		if lane != nil && lane.SecondaryRepairPacketsPending > 0 {
			if staleLane != nil && i == staleIndex {
				continue
			}
			return lane, repairSendSecondary, i
		}
	}
	for i, lane := range t.lanes {
		if lane == nil || i >= len(opportunisticBudget) {
			continue
		}
		if opportunisticBudget[i] > 0 {
			return lane, repairSendOpportunistic, i
		}
	}
	return nil, repairSendNone, -1
}

func (t *SessionTx) latestAppendableLane() *txLane {
	if len(t.lanes) == 0 {
		return nil
	}
	lane := t.lanes[len(t.lanes)-1]
	if lane.Finalized {
		return nil
	}
	return lane
}

func (t *SessionTx) createLane() (*txLane, error) {
	if t.maxBufferedLanes > 0 && len(t.lanes) >= t.maxBufferedLanes {
		return nil, ErrTxLaneBufferFull
	}

	transferLane, err := rrpitTransferLane.NewTransferLaneTx(t.laneShardSize, t.maxDataShardsPerLane)
	if err != nil {
		return nil, err
	}
	laneID := uint64(t.firstLaneID + int64(len(t.lanes)))
	createdAt := t.lifecycleTimestamp()
	lane := &txLane{
		LaneID:                laneID,
		TransferLane:          transferLane,
		CreatedAtTimestamp:    createdAt,
		LastProgressTimestamp: createdAt,
	}
	t.lanes = append(t.lanes, lane)
	return lane, nil
}

func (t *SessionTx) removeNewestLane() {
	if len(t.lanes) == 0 {
		return
	}
	t.lanes = t.lanes[:len(t.lanes)-1]
}

func (t *SessionTx) sendTransferPacket(laneID uint64, transfer rrpitTransferLane.TransferData, quotaBound bool) error {
	payload, err := marshalSessionDataPacket(t.DataPacketKind, laneID, transfer)
	if err != nil {
		return err
	}
	if quotaBound {
		if t.SendEnforced != nil {
			return t.SendEnforced(t.DataPacketKind, payload)
		}
	} else {
		if t.SendIgnoreQuota != nil {
			return t.SendIgnoreQuota(t.DataPacketKind, payload)
		}
		if t.SendEnforced != nil {
			return t.SendEnforced(t.DataPacketKind, payload)
		}
	}

	channelIndex, err := t.bestChannelIndex()
	if err != nil {
		return err
	}
	if err := t.txChannelsConfig[channelIndex].MaterializeChannel.SendDataMessage(payload); err != nil {
		return err
	}
	t.markChannelSent(channelIndex)
	return nil
}

func (t *SessionTx) bestChannelIndex() (int, error) {
	if len(t.txChannelsConfig) == 0 {
		return 0, newError("no transfer channel attached")
	}

	best := -1
	for i := range t.txChannelsConfig {
		if t.txChannelsConfig[i].MaterializeChannel == nil || t.channelRateLimited(i) {
			continue
		}
		if best == -1 || t.channelHasLessLoad(i, best) {
			best = i
		}
	}
	if best != -1 {
		return best, nil
	}

	for i := range t.txChannelsConfig {
		if t.txChannelsConfig[i].MaterializeChannel == nil {
			continue
		}
		if best == -1 || t.channelHasLessLoad(i, best) {
			best = i
		}
	}
	if best == -1 {
		return 0, newError("no materialized transfer channel attached")
	}
	return best, nil
}

func (t *SessionTx) channelRateLimited(index int) bool {
	maxSpeed := t.txChannelsConfig[index].Config.MaxSendingSpeed
	if maxSpeed <= 0 {
		return false
	}
	status := t.txChannelsConfig[index].Status
	if status.TimestampLastSent != t.effectiveTimestamp() {
		return false
	}
	return int(status.EnforcedPacketSentCurrentTimestamp) >= maxSpeed
}

func (t *SessionTx) channelHasLessLoad(candidateIndex int, currentBestIndex int) bool {
	candidateWeight := t.txChannelsConfig[candidateIndex].Config.Weight
	if candidateWeight <= 0 {
		candidateWeight = 1
	}
	bestWeight := t.txChannelsConfig[currentBestIndex].Config.Weight
	if bestWeight <= 0 {
		bestWeight = 1
	}

	candidateSent := t.channelWindowSendCount(candidateIndex)
	bestSent := t.channelWindowSendCount(currentBestIndex)
	left := candidateSent * uint64(bestWeight)
	right := bestSent * uint64(candidateWeight)
	if left != right {
		return left < right
	}
	return t.txChannelsConfig[candidateIndex].MaterializeChannel.ChannelID < t.txChannelsConfig[currentBestIndex].MaterializeChannel.ChannelID
}

func (t *SessionTx) channelWindowSendCount(index int) uint64 {
	status := t.txChannelsConfig[index].Status
	if status.TimestampLastSent != t.effectiveTimestamp() {
		return 0
	}
	return status.PacketSentCurrentTimestamp
}

func (t *SessionTx) markChannelSent(index int) {
	timestamp := t.effectiveTimestamp()
	status := &t.txChannelsConfig[index].Status
	if status.TimestampLastSent != timestamp {
		status.PacketSentCurrentTimestamp = 0
		status.EnforcedPacketSentCurrentTimestamp = 0
	}
	status.TimestampLastSent = timestamp
	status.PacketSentCurrentTimestamp += 1
	status.EnforcedPacketSentCurrentTimestamp += 1
}

func (t *SessionTx) effectiveTimestamp() uint64 {
	if t.currentTimestamp == 0 {
		t.currentTimestamp = uint64(time.Now().UnixNano())
	}
	return t.currentTimestamp
}

func (t *SessionTx) channelIndexByID(channelID uint64) int {
	for i := range t.txChannelsConfig {
		channel := t.txChannelsConfig[i].MaterializeChannel
		if channel != nil && channel.ChannelID == channelID {
			return i
		}
	}
	return -1
}

func (t *SessionTx) laneByID(laneID int64) *txLane {
	index := laneID - t.firstLaneID
	if index < 0 || int(index) >= len(t.lanes) {
		return nil
	}
	return t.lanes[index]
}

func (t *SessionTx) dropLanesThrough(ackTo int64) {
	if len(t.lanes) == 0 || ackTo < t.firstLaneID {
		return
	}
	dropCount := int(ackTo - t.firstLaneID + 1)
	if dropCount > len(t.lanes) {
		dropCount = len(t.lanes)
	}
	t.lanes = t.lanes[dropCount:]
	t.firstLaneID += int64(dropCount)
}

func (t *SessionTx) allocateChannelID() uint64 {
	if t.nextChannelID == 0 {
		if t.OddChannelIDs {
			t.nextChannelID = 1
		} else {
			t.nextChannelID = 2
		}
	}
	channelID := t.nextChannelID
	t.nextChannelID += 2
	return channelID
}

func marshalSessionDataPacket(kind uint8, laneID uint64, transfer rrpitTransferLane.TransferData) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := struc.Pack(buffer, &sessionDataPacket{
		PacketKind: kind,
		LaneID:     laneID,
		Transfer:   transfer,
	}); err != nil {
		return nil, newError("failed to pack session data packet: ", err)
	}
	return buffer.Bytes(), nil
}

func marshalSessionControlPacket(kind uint8, ctrl ControlMessage) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := struc.Pack(buffer, &sessionControlPacket{
		PacketKind: kind,
		Control:    ctrl,
	}); err != nil {
		return nil, newError("failed to pack session control packet: ", err)
	}
	return buffer.Bytes(), nil
}

func MarshalSessionControlPacket(kind uint8, ctrl ControlMessage) ([]byte, error) {
	return marshalSessionControlPacket(kind, ctrl)
}

func MarshalSessionDataPacket(kind uint8, laneID uint64, transfer rrpitTransferLane.TransferData) ([]byte, error) {
	return marshalSessionDataPacket(kind, laneID, transfer)
}

func UnmarshalSessionControlPacket(data []byte) (ControlMessage, error) {
	var packet sessionControlPacket
	if err := struc.Unpack(bytes.NewReader(data), &packet); err != nil {
		return ControlMessage{}, newError("failed to unpack session control packet: ", err)
	}
	return packet.Control, nil
}

func UnmarshalSessionDataPacket(data []byte) (uint64, rrpitTransferLane.TransferData, error) {
	var packet sessionDataPacket
	if err := struc.Unpack(bytes.NewReader(data), &packet); err != nil {
		return 0, rrpitTransferLane.TransferData{}, newError("failed to unpack session data packet: ", err)
	}
	return packet.LaneID, packet.Transfer, nil
}
