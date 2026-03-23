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

	OnMessage              func([]byte) error
	OnRemoteControlMessage func(ControlMessage) error
}

type SessionTx struct {
	txLanes
	txChannels
	txChannelsConfig []ChannelStatus
	OddChannelIDs    bool
	Reconstruction   SessionTxReconstructionConfig

	currentTimestamp            uint64
	currentTimestampInitialized bool
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
	TimestampLastSent          uint64
	PacketSentCurrentTimestamp uint64
}

type SessionTxConfig struct {
	LaneShardSize                  int
	MaxDataShardsPerLane           int
	MaxBufferedLanes               int
	MaxRewindableTimestampNum      int
	MaxRewindableControlMessageNum int
	OddChannelIDs                  bool
	Reconstruction                 SessionTxReconstructionConfig
}

type SessionTxReconstructionConfig struct {
	InitialRepairShardRatio        float64
	LaneRepairWeight               []float64
	SecondaryRepairShardRatio      float64
	TimeResendSecondaryRepairShard int
}

type SessionRxConfig struct {
	LaneShardSize      int
	MaxBufferedLanes   int
	OnMessage          func([]byte) error
	OnRemoteControlMsg func(ControlMessage) error
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
}

type txChannels struct {
	nextChannelID                  uint64
	maxRewindableTimestampNum      int
	maxRewindableControlMessageNum int
}

type rxLanes struct {
	laneShardSize    int
	maxBufferedLanes int

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
)

var ErrTxLaneBufferFull = errors.New("too many buffered transfer lanes")

func NewSessionTx(config SessionTxConfig) (*SessionTx, error) {
	tx := &SessionTx{
		OddChannelIDs:  config.OddChannelIDs,
		Reconstruction: config.Reconstruction,
	}
	if len(config.Reconstruction.LaneRepairWeight) > 0 {
		tx.Reconstruction.LaneRepairWeight = append([]float64(nil), config.Reconstruction.LaneRepairWeight...)
	}
	tx.txLanes.laneShardSize = config.LaneShardSize
	tx.txLanes.maxDataShardsPerLane = config.MaxDataShardsPerLane
	tx.txLanes.maxBufferedLanes = config.MaxBufferedLanes
	tx.txChannels.maxRewindableTimestampNum = config.MaxRewindableTimestampNum
	tx.txChannels.maxRewindableControlMessageNum = config.MaxRewindableControlMessageNum
	if err := tx.ensureDefaults(); err != nil {
		return nil, err
	}
	return tx, nil
}

func NewSessionRx(config SessionRxConfig) (*SessionRx, error) {
	rx := &SessionRx{
		OnMessage:              config.OnMessage,
		OnRemoteControlMessage: config.OnRemoteControlMsg,
	}
	rx.rxLanes.laneShardSize = config.LaneShardSize
	rx.rxLanes.maxBufferedLanes = config.MaxBufferedLanes
	if err := rx.ensureDefaults(); err != nil {
		return nil, err
	}
	return rx, nil
}

func (t *SessionTx) OnNewTimestamp(timestamp uint64) error {
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
	maxMessageSize := t.txLanes.laneShardSize - rrpitTransferLane.ReconstructionLengthFieldSize
	if maxMessageSize <= 0 {
		return 0, newError("lane shard size too small")
	}
	return maxMessageSize, nil
}

func (t *SessionTx) AcceptRemoteControlMessage(ctrl ControlMessage) error {
	return t.acceptRemoteControlMessage(ctrl)
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
		t.txChannels.maxRewindableTimestampNum,
		t.txChannels.maxRewindableControlMessageNum,
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
	r.rxChannels.channels = append(r.rxChannels.channels, channelState)
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
			LaneACKTo: r.rxLanes.firstLaneID - 1,
		},
	}

	if len(r.rxLanes.lanes) > int(^uint16(0)) {
		return ControlMessage{}, newError("lane control count exceeds control field capacity")
	}
	ctrl.Lane.LaneControl = make([]rrpitTransferLane.TransferControl, len(r.rxLanes.lanes))
	for i, lane := range r.rxLanes.lanes {
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

	if len(r.rxChannels.channels) > int(^uint16(0)) {
		return ControlMessage{}, newError("channel control count exceeds control field capacity")
	}
	ctrl.Channel.ChannelControl = make([]rrpitTransferChannel.ChannelControlMessage, 0, len(r.rxChannels.channels))
	for _, channel := range r.rxChannels.channels {
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

func (r *SessionRx) ensureDefaults() error {
	if r.OnMessage == nil {
		return newError("nil OnMessage callback")
	}
	if r.rxLanes.laneShardSize == 0 {
		r.rxLanes.laneShardSize = defaultLaneShardSize
	}
	if r.rxLanes.maxBufferedLanes == 0 {
		r.rxLanes.maxBufferedLanes = defaultMaxBufferedLanes
	}
	if r.rxLanes.maxBufferedLanes < 0 {
		return newError("invalid max buffered lanes")
	}
	if _, err := rrpitTransferLane.NewTransferLaneRx(r.rxLanes.laneShardSize); err != nil {
		return err
	}
	return nil
}

func (t *SessionTx) ensureDefaults() error {
	if t.txLanes.laneShardSize == 0 {
		t.txLanes.laneShardSize = defaultLaneShardSize
	}
	if t.txLanes.maxDataShardsPerLane == 0 {
		t.txLanes.maxDataShardsPerLane = defaultMaxDataShardsPerLane
	}
	if t.txLanes.maxBufferedLanes == 0 {
		t.txLanes.maxBufferedLanes = defaultMaxBufferedLanes
	}
	if t.txChannels.maxRewindableTimestampNum == 0 {
		t.txChannels.maxRewindableTimestampNum = defaultMaxRewindableTimestampNum
	}
	if t.txChannels.maxRewindableControlMessageNum == 0 {
		t.txChannels.maxRewindableControlMessageNum = defaultMaxRewindableControlMessageNum
	}

	if t.txLanes.maxBufferedLanes < 0 {
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
	for _, weight := range t.Reconstruction.LaneRepairWeight {
		if weight < 0 {
			return newError("invalid lane repair weight")
		}
	}
	if _, err := rrpitTransferLane.NewTransferLaneTx(t.txLanes.laneShardSize, t.txLanes.maxDataShardsPerLane); err != nil {
		return err
	}
	if _, err := rrpitTransferChannel.NewChannelTx(0, t.txChannels.maxRewindableTimestampNum, t.txChannels.maxRewindableControlMessageNum); err != nil {
		return err
	}
	return nil
}

func (t *SessionTx) onNewTimestamp(timestamp uint64) error {
	if err := t.ensureDefaults(); err != nil {
		return err
	}
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
				return nil
			}
			if err := t.sendRepairPacket(lane); err != nil {
				return err
			}
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
		return nil
	}

	return t.sendRepairPacket(lane)
}

func (r *SessionRx) onChannelData(channel *rxChannel, data []byte) error {
	if len(data) == 0 {
		return newError("session packet too short")
	}

	switch data[0] {
	case PacketKind_DATA:
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
	case PacketKind_CONTROL:
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
	if laneID > uint64(^uint64(0)>>1) {
		return nil, newError("lane id exceeds supported range")
	}
	laneIDInt := int64(laneID)
	if laneIDInt < r.rxLanes.firstLaneID {
		return nil, nil
	}

	index := int(laneIDInt - r.rxLanes.firstLaneID)
	if r.rxLanes.maxBufferedLanes > 0 && index >= r.rxLanes.maxBufferedLanes {
		return nil, newError("too many buffered transfer lanes")
	}
	for len(r.rxLanes.lanes) <= index {
		r.rxLanes.lanes = append(r.rxLanes.lanes, nil)
	}
	if r.rxLanes.lanes[index] != nil {
		return r.rxLanes.lanes[index], nil
	}

	transferLane, err := rrpitTransferLane.NewTransferLaneRx(r.rxLanes.laneShardSize)
	if err != nil {
		return nil, err
	}
	lane := &rxLane{
		LaneID:       laneID,
		TransferLane: transferLane,
	}
	r.rxLanes.lanes[index] = lane
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
	for len(r.rxLanes.lanes) > 0 {
		lane := r.rxLanes.lanes[0]
		if lane == nil || !lane.Ready {
			return payloads, nil
		}

		for lane.NextDeliverIndex < len(lane.Reconstructed) {
			payload := append([]byte(nil), lane.Reconstructed[lane.NextDeliverIndex].Data...)
			payloads = append(payloads, payload)
			lane.NextDeliverIndex += 1
		}

		r.rxLanes.lanes = r.rxLanes.lanes[1:]
		r.rxLanes.firstLaneID += 1
	}
	return payloads, nil
}

func (r *SessionRx) hasAttachedChannelID(channelID uint64, exclude *rxChannel) bool {
	for _, attached := range r.rxChannels.channels {
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
		payload, err := marshalSessionControlPacket(ctrl)
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
	if len(t.txChannelsConfig) == 0 {
		return newError("no transfer channel attached")
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
		if err := t.sendTransferPacket(lane.LaneID, *transfer); err != nil {
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
	if err := t.sendTransferPacket(lane.LaneID, *transfer); err != nil {
		return err
	}
	return t.flushInitialRepairPackets(lane)
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
		if err := lane.TransferLane.AcceptControlData(laneControl); err != nil {
			return err
		}
		if laneControl.SeenChunks > lane.PeerSeenChunks {
			lane.PeerSeenChunks = laneControl.SeenChunks
		}
		lane.PeerSeenChunksKnown = true
		if t.hasCustomReconstructionConfig() {
			t.updateLaneRepairStateAfterControl(lane, !previousKnown || lane.PeerSeenChunks > previousSeen)
		}
	}
	return nil
}

func (t *SessionTx) resetChannelRateWindow(timestamp uint64) {
	for i := range t.txChannelsConfig {
		if t.txChannelsConfig[i].Status.TimestampLastSent != timestamp {
			t.txChannelsConfig[i].Status.PacketSentCurrentTimestamp = 0
		}
	}
}

func (t *SessionTx) nextLaneForRepair() *txLane {
	if len(t.txLanes.lanes) == 0 {
		return nil
	}

	lastLane := t.txLanes.lanes[len(t.txLanes.lanes)-1]
	if !lastLane.Finalized && lastLane.DataShards > 0 {
		return lastLane
	}

	var selected *txLane
	for _, lane := range t.txLanes.lanes {
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
		t.Reconstruction.TimeResendSecondaryRepairShard > 0 ||
		len(t.Reconstruction.LaneRepairWeight) > 0
}

func (t *SessionTx) shouldFinalizeLaneAfterData(lane *txLane) bool {
	if lane == nil || lane.Finalized || t.txLanes.maxDataShardsPerLane <= 0 {
		return false
	}
	return int(lane.DataShards) >= t.txLanes.maxDataShardsPerLane
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
		if err := t.sendRepairPacket(lane); err != nil {
			return err
		}
		lane.InitialRepairPacketsPending -= 1
	}
	t.scheduleNextSecondaryRepair(lane)
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
	if err := t.sendTransferPacket(lane.LaneID, transfer); err != nil {
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

func (t *SessionTx) laneMissingShards(lane *txLane) uint32 {
	if lane == nil || !lane.PeerSeenChunksKnown || lane.TotalDataShards == 0 {
		return 0
	}
	seen := uint32(lane.PeerSeenChunks)
	if seen >= lane.TotalDataShards {
		return 0
	}
	return lane.TotalDataShards - seen
}

func (t *SessionTx) laneRepairDemand(lane *txLane) uint32 {
	if lane == nil || lane.TotalDataShards == 0 {
		return 0
	}
	if !lane.PeerSeenChunksKnown {
		return lane.TotalDataShards
	}

	missing := t.laneMissingShards(lane)
	if missing > 0 {
		return missing
	}
	if lane.LaneID != uint64(t.txLanes.firstLaneID) {
		return 0
	}

	// The receiver may need a small repair tail beyond K symbols before decode completes.
	return 1
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
		lane.NextSecondaryRepairTimestamp = 0
		return
	}

	burst := repairPacketQuota(t.Reconstruction.SecondaryRepairShardRatio, repairDemand)
	lane.SecondaryRepairPacketsPerBurst = burst
	if t.Reconstruction.TimeResendSecondaryRepairShard <= 0 {
		if !scheduleSecondary {
			if burst == 0 {
				lane.SecondaryRepairPacketsPending = 0
				lane.NextSecondaryRepairTimestamp = 0
			}
			return
		}

		lane.SecondaryRepairPacketsPending = burst
		lane.NextSecondaryRepairTimestamp = 0
		return
	}

	if burst == 0 {
		lane.SecondaryRepairPacketsPending = 0
		lane.NextSecondaryRepairTimestamp = 0
		return
	}

	if lane.SecondaryRepairPacketsPending == 0 && lane.NextSecondaryRepairTimestamp == 0 {
		lane.NextSecondaryRepairTimestamp = t.secondaryRepairScheduleBaseTimestamp() + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
	}
}

func (t *SessionTx) scheduleSecondaryRepairResends(timestamp uint64) {
	if t.Reconstruction.TimeResendSecondaryRepairShard <= 0 {
		return
	}
	for _, lane := range t.txLanes.lanes {
		if lane == nil || lane.SecondaryRepairPacketsPending != 0 {
			continue
		}
		repairDemand := t.laneRepairDemand(lane)
		if repairDemand == 0 {
			lane.SecondaryRepairPacketsPerBurst = 0
			lane.NextSecondaryRepairTimestamp = 0
			continue
		}

		if lane.NextSecondaryRepairTimestamp == 0 {
			lane.SecondaryRepairPacketsPerBurst = repairPacketQuota(t.Reconstruction.SecondaryRepairShardRatio, repairDemand)
			if lane.SecondaryRepairPacketsPerBurst == 0 {
				continue
			}
			lane.NextSecondaryRepairTimestamp = timestamp + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
			continue
		}
		if timestamp < lane.NextSecondaryRepairTimestamp {
			continue
		}

		burst := repairPacketQuota(t.Reconstruction.SecondaryRepairShardRatio, repairDemand)
		lane.SecondaryRepairPacketsPerBurst = burst
		if burst == 0 {
			lane.NextSecondaryRepairTimestamp = 0
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
		lane.NextSecondaryRepairTimestamp = 0
		lane.SecondaryRepairPacketsPerBurst = 0
		return
	}
	if lane.SecondaryRepairPacketsPending != 0 || lane.NextSecondaryRepairTimestamp != 0 {
		return
	}
	lane.NextSecondaryRepairTimestamp = t.secondaryRepairScheduleBaseTimestamp() + uint64(t.Reconstruction.TimeResendSecondaryRepairShard)
}

func (t *SessionTx) buildOpportunisticRepairBudget() []uint32 {
	if len(t.Reconstruction.LaneRepairWeight) == 0 {
		return nil
	}
	budget := make([]uint32, len(t.txLanes.lanes))
	for i, lane := range t.txLanes.lanes {
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
	for i, lane := range t.txLanes.lanes {
		if lane == nil {
			continue
		}
		if lane.PeerSeenChunksKnown && t.laneRepairDemand(lane) == 0 {
			lane.InitialRepairPacketsPending = 0
			lane.SecondaryRepairPacketsPending = 0
			continue
		}
		if lane.InitialRepairPacketsPending > 0 {
			return lane, repairSendInitial, i
		}
	}
	for i, lane := range t.txLanes.lanes {
		if lane != nil && lane.SecondaryRepairPacketsPending > 0 {
			return lane, repairSendSecondary, i
		}
	}
	for i, lane := range t.txLanes.lanes {
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
	if len(t.txLanes.lanes) == 0 {
		return nil
	}
	lane := t.txLanes.lanes[len(t.txLanes.lanes)-1]
	if lane.Finalized {
		return nil
	}
	return lane
}

func (t *SessionTx) createLane() (*txLane, error) {
	if t.txLanes.maxBufferedLanes > 0 && len(t.txLanes.lanes) >= t.txLanes.maxBufferedLanes {
		return nil, ErrTxLaneBufferFull
	}

	transferLane, err := rrpitTransferLane.NewTransferLaneTx(t.txLanes.laneShardSize, t.txLanes.maxDataShardsPerLane)
	if err != nil {
		return nil, err
	}
	laneID := uint64(t.txLanes.firstLaneID + int64(len(t.txLanes.lanes)))
	lane := &txLane{
		LaneID:       laneID,
		TransferLane: transferLane,
	}
	t.txLanes.lanes = append(t.txLanes.lanes, lane)
	return lane, nil
}

func (t *SessionTx) removeNewestLane() {
	if len(t.txLanes.lanes) == 0 {
		return
	}
	t.txLanes.lanes = t.txLanes.lanes[:len(t.txLanes.lanes)-1]
}

func (t *SessionTx) sendTransferPacket(laneID uint64, transfer rrpitTransferLane.TransferData) error {
	channelIndex, err := t.bestChannelIndex()
	if err != nil {
		return err
	}

	payload, err := marshalSessionDataPacket(laneID, transfer)
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
	return int(status.PacketSentCurrentTimestamp) >= maxSpeed
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
	}
	status.TimestampLastSent = timestamp
	status.PacketSentCurrentTimestamp += 1
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
	index := laneID - t.txLanes.firstLaneID
	if index < 0 || int(index) >= len(t.txLanes.lanes) {
		return nil
	}
	return t.txLanes.lanes[index]
}

func (t *SessionTx) dropLanesThrough(ackTo int64) {
	if len(t.txLanes.lanes) == 0 || ackTo < t.txLanes.firstLaneID {
		return
	}
	dropCount := int(ackTo - t.txLanes.firstLaneID + 1)
	if dropCount > len(t.txLanes.lanes) {
		dropCount = len(t.txLanes.lanes)
	}
	t.txLanes.lanes = t.txLanes.lanes[dropCount:]
	t.txLanes.firstLaneID += int64(dropCount)
}

func (t *SessionTx) allocateChannelID() uint64 {
	if t.txChannels.nextChannelID == 0 {
		if t.OddChannelIDs {
			t.txChannels.nextChannelID = 1
		} else {
			t.txChannels.nextChannelID = 2
		}
	}
	channelID := t.txChannels.nextChannelID
	t.txChannels.nextChannelID += 2
	return channelID
}

func marshalSessionDataPacket(laneID uint64, transfer rrpitTransferLane.TransferData) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := struc.Pack(buffer, &sessionDataPacket{
		PacketKind: PacketKind_DATA,
		LaneID:     laneID,
		Transfer:   transfer,
	}); err != nil {
		return nil, newError("failed to pack session data packet: ", err)
	}
	return buffer.Bytes(), nil
}

func marshalSessionControlPacket(ctrl ControlMessage) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := struc.Pack(buffer, &sessionControlPacket{
		PacketKind: PacketKind_CONTROL,
		Control:    ctrl,
	}); err != nil {
		return nil, newError("failed to pack session control packet: ", err)
	}
	return buffer.Bytes(), nil
}
