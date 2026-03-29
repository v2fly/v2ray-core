package rriptMonoDirectionSession

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

type directionalIntegrationConfig struct {
	name                   string
	channelCount           int
	shardSize              int
	maxDataShardsPerLane   int
	maxBufferedLanes       int
	oddChannelIDs          bool
	channelHistoryCapacity int
}

type directionalIntegrationHarness struct {
	t *testing.T

	tx *SessionTx
	rx *SessionRx

	writers      []*recordingWriteCloser
	rxChannels   []*rrpitMaterializedTransferChannel.ChannelRx
	txChannelIDs []uint64
	nextRead     []int

	received  [][]byte
	timestamp uint64

	maxDataShardsPerLane int
}

type directionalIntegrationFrame struct {
	channelIndex int
	raw          []byte
	packetKind   uint8
	data         *sessionDataPacket
	control      *sessionControlPacket
}

func TestDirectionalSessionIntegrationMatrix(t *testing.T) {
	configs := []directionalIntegrationConfig{
		{
			name:                   "single_channel_single_shard_odd",
			channelCount:           1,
			shardSize:              16,
			maxDataShardsPerLane:   1,
			maxBufferedLanes:       8,
			oddChannelIDs:          true,
			channelHistoryCapacity: 16,
		},
		{
			name:                   "single_channel_three_shards_even",
			channelCount:           1,
			shardSize:              24,
			maxDataShardsPerLane:   3,
			maxBufferedLanes:       8,
			oddChannelIDs:          false,
			channelHistoryCapacity: 16,
		},
		{
			name:                   "dual_channel_two_shards_odd",
			channelCount:           2,
			shardSize:              24,
			maxDataShardsPerLane:   2,
			maxBufferedLanes:       8,
			oddChannelIDs:          true,
			channelHistoryCapacity: 16,
		},
		{
			name:                   "triple_channel_three_shards_even",
			channelCount:           3,
			shardSize:              32,
			maxDataShardsPerLane:   3,
			maxBufferedLanes:       8,
			oddChannelIDs:          false,
			channelHistoryCapacity: 16,
		},
	}

	scenarios := []struct {
		name string
		run  func(*testing.T, *directionalIntegrationHarness)
	}{
		{name: "ordered_no_loss", run: runDirectionalOrderedNoLoss},
		{name: "reverse_delivery_no_loss", run: runDirectionalReverseDelivery},
		{name: "drop_one_source_per_lane_repair_recovery", run: runDirectionalLossRecovery},
		{name: "source_first_then_repair_and_duplicate_control", run: runDirectionalPartialControlAndDuplicates},
	}

	for _, config := range configs {
		config := config
		t.Run(config.name, func(t *testing.T) {
			for _, scenario := range scenarios {
				scenario := scenario
				t.Run(scenario.name, func(t *testing.T) {
					harness := newDirectionalIntegrationHarness(t, config)
					scenario.run(t, harness)
				})
			}
		})
	}
}

func TestDirectionalSessionIntegrationBackpressureRelease(t *testing.T) {
	harness := newDirectionalIntegrationHarness(t, directionalIntegrationConfig{
		name:                   "backpressure_single_lane",
		channelCount:           1,
		shardSize:              16,
		maxDataShardsPerLane:   1,
		maxBufferedLanes:       1,
		oddChannelIDs:          true,
		channelHistoryCapacity: 16,
	})

	first := []byte("bp-00")
	second := []byte("bp-01")

	harness.sendMessages(first)
	harness.advanceTicks(1)

	if err := harness.tx.SendMessage(second); err == nil {
		t.Fatal("expected sender window to block a second lane before control acknowledgment")
	}

	harness.deliverFrames(harness.takePendingFrames())
	harness.applyControl(harness.generateControl(), 1)
	harness.assertSenderDrained()

	if err := harness.tx.SendMessage(second); err != nil {
		t.Fatalf("expected sender window to reopen after control acknowledgment, got %v", err)
	}
	harness.advanceTicks(1)
	harness.deliverFrames(harness.takePendingFrames())
	harness.applyControl(harness.generateControl(), 1)

	harness.assertReceived([][]byte{first, second})
	harness.assertSenderDrained()
}

func runDirectionalOrderedNoLoss(t *testing.T, harness *directionalIntegrationHarness) {
	messages := makeDirectionalIntegrationMessages(harness.maxDataShardsPerLane * 3)
	harness.sendMessages(messages...)
	harness.advanceTicks(harness.expectedLaneCount(len(messages)))

	frames := harness.takePendingFrames()
	harness.deliverFrames(frames)
	harness.applyControl(harness.generateControl(), 1)

	harness.assertReceived(messages)
	harness.assertSenderDrained()
}

func runDirectionalReverseDelivery(t *testing.T, harness *directionalIntegrationHarness) {
	messages := makeDirectionalIntegrationMessages(harness.maxDataShardsPerLane * 3)
	harness.sendMessages(messages...)
	harness.advanceTicks(harness.expectedLaneCount(len(messages)))

	frames := harness.takePendingFrames()
	reverseDirectionalFrames(frames)
	harness.deliverFrames(frames)
	harness.applyControl(harness.generateControl(), 2)

	harness.assertReceived(messages)
	harness.assertSenderDrained()
}

func runDirectionalLossRecovery(t *testing.T, harness *directionalIntegrationHarness) {
	messages := makeDirectionalIntegrationMessages(harness.maxDataShardsPerLane * 3)
	harness.sendMessages(messages...)
	harness.advanceTicks(harness.expectedLaneCount(len(messages)))

	frames := harness.takePendingFrames()
	deliver := make([]directionalIntegrationFrame, 0, len(frames))
	droppedByLane := map[uint64]bool{}
	for _, frame := range frames {
		if frame.packetKind == PacketKind_DATA && frame.data != nil && frame.data.Transfer.TotalDataShards == 0 && !droppedByLane[frame.data.LaneID] {
			droppedByLane[frame.data.LaneID] = true
			continue
		}
		deliver = append(deliver, frame)
	}
	reverseDirectionalFrames(deliver)
	harness.deliverFrames(deliver)
	harness.applyControl(harness.generateControl(), 1)

	harness.assertReceived(messages)
	harness.assertSenderDrained()
}

func runDirectionalPartialControlAndDuplicates(t *testing.T, harness *directionalIntegrationHarness) {
	messages := makeDirectionalIntegrationMessages(harness.maxDataShardsPerLane * 3)
	harness.sendMessages(messages...)
	harness.advanceTicks(harness.expectedLaneCount(len(messages)))

	frames := harness.takePendingFrames()
	sourceFrames := make([]directionalIntegrationFrame, 0, len(frames))
	repairFrames := make([]directionalIntegrationFrame, 0, len(frames))
	for _, frame := range frames {
		if frame.packetKind == PacketKind_DATA && frame.data != nil && frame.data.Transfer.TotalDataShards == 0 {
			sourceFrames = append(sourceFrames, frame)
			continue
		}
		repairFrames = append(repairFrames, frame)
	}

	harness.deliverFrames(sourceFrames)
	earlyControl := harness.generateControl()
	lastLaneID := int64(harness.expectedLaneCount(len(messages)) - 1)
	if earlyControl.Lane.LaneACKTo >= lastLaneID {
		t.Fatalf("expected partial control to avoid acknowledging the final lane, got ack %d for last lane %d", earlyControl.Lane.LaneACKTo, lastLaneID)
	}
	harness.applyControl(earlyControl, 1)

	duplicateRepairFrames := append([]directionalIntegrationFrame{}, repairFrames...)
	if len(repairFrames) > 0 {
		duplicateRepairFrames = append(duplicateRepairFrames, repairFrames[0])
	}
	if len(sourceFrames) > 0 {
		duplicateRepairFrames = append(duplicateRepairFrames, sourceFrames[0])
	}
	harness.deliverFrames(duplicateRepairFrames)
	harness.applyControl(harness.generateControl(), 2)

	harness.assertReceived(messages)
	harness.assertSenderDrained()
}

func newDirectionalIntegrationHarness(t *testing.T, config directionalIntegrationConfig) *directionalIntegrationHarness {
	t.Helper()

	harness := &directionalIntegrationHarness{
		t:                    t,
		maxDataShardsPerLane: config.maxDataShardsPerLane,
	}
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    config.shardSize,
		MaxBufferedLanes: config.maxBufferedLanes,
		OnMessage: func(data []byte) error {
			harness.received = append(harness.received, append([]byte(nil), data...))
			return nil
		},
	})
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  config.shardSize,
		MaxDataShardsPerLane:           config.maxDataShardsPerLane,
		MaxBufferedLanes:               config.maxBufferedLanes,
		MaxRewindableTimestampNum:      config.channelHistoryCapacity,
		MaxRewindableControlMessageNum: config.channelHistoryCapacity,
		OddChannelIDs:                  config.oddChannelIDs,
	})

	harness.tx = tx
	harness.rx = rx

	harness.writers = make([]*recordingWriteCloser, 0, config.channelCount)
	harness.rxChannels = make([]*rrpitMaterializedTransferChannel.ChannelRx, 0, config.channelCount)
	harness.txChannelIDs = make([]uint64, 0, config.channelCount)
	harness.nextRead = make([]int, config.channelCount)

	for i := 0; i < config.channelCount; i++ {
		writer := &recordingWriteCloser{}
		channelID, err := tx.AttachTxChannel(writer)
		if err != nil {
			t.Fatal(err)
		}
		rxChannel, err := rx.AttachRxChannel(0)
		if err != nil {
			t.Fatal(err)
		}

		harness.writers = append(harness.writers, writer)
		harness.rxChannels = append(harness.rxChannels, rxChannel)
		harness.txChannelIDs = append(harness.txChannelIDs, channelID)
	}

	harness.learnChannelIDs()
	return harness
}

func (h *directionalIntegrationHarness) learnChannelIDs() {
	h.t.Helper()

	if err := h.tx.FloodControlMessageToAllChannels(); err != nil {
		h.t.Fatal(err)
	}

	frames := h.takePendingFrames()
	if len(frames) != len(h.rxChannels) {
		h.t.Fatalf("expected one control flood frame per attached channel, got %d", len(frames))
	}
	for _, frame := range frames {
		if frame.packetKind != PacketKind_CONTROL || frame.control == nil {
			h.t.Fatalf("expected flood learn phase to contain only control packets, got kind %d", frame.packetKind)
		}
	}

	h.deliverFrames(frames)
	for i, channel := range h.rxChannels {
		if channel.ChannelID != h.txChannelIDs[i] {
			h.t.Fatalf("expected rx channel %d to learn id %d, got %d", i, h.txChannelIDs[i], channel.ChannelID)
		}
	}

	h.advanceTicks(1)
	if pending := h.takePendingFrames(); len(pending) != 0 {
		h.t.Fatalf("expected no pending frames after empty timestamp advance, got %d", len(pending))
	}
}

func (h *directionalIntegrationHarness) sendMessages(messages ...[]byte) {
	h.t.Helper()

	for _, message := range messages {
		if err := h.tx.SendMessage(message); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *directionalIntegrationHarness) advanceTicks(count int) {
	h.t.Helper()

	for i := 0; i < count; i++ {
		h.timestamp++
		if err := h.tx.OnNewTimestamp(h.timestamp); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *directionalIntegrationHarness) takePendingFrames() []directionalIntegrationFrame {
	h.t.Helper()

	frames := make([]directionalIntegrationFrame, 0)
	for channelIndex, writer := range h.writers {
		for h.nextRead[channelIndex] < len(writer.writes) {
			raw := append([]byte(nil), writer.writes[h.nextRead[channelIndex]]...)
			_, payload := splitMaterializedWire(h.t, raw)
			if len(payload) == 0 {
				h.t.Fatal("session payload is empty")
			}

			frame := directionalIntegrationFrame{
				channelIndex: channelIndex,
				raw:          raw,
				packetKind:   payload[0],
			}
			switch frame.packetKind {
			case PacketKind_DATA:
				packet := mustUnmarshalSessionDataPacket(h.t, payload)
				frame.data = &packet
			case PacketKind_CONTROL:
				packet := mustUnmarshalSessionControlPacket(h.t, payload)
				frame.control = &packet
			default:
				h.t.Fatalf("unknown session packet kind %d", frame.packetKind)
			}

			frames = append(frames, frame)
			h.nextRead[channelIndex] += 1
		}
	}
	return frames
}

func (h *directionalIntegrationHarness) deliverFrames(frames []directionalIntegrationFrame) {
	h.t.Helper()

	for _, frame := range frames {
		if err := h.rxChannels[frame.channelIndex].OnNewMessageArrived(frame.raw); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *directionalIntegrationHarness) generateControl() ControlMessage {
	h.t.Helper()

	currentChannelID := uint64(0)
	for _, channel := range h.rxChannels {
		if channel.ChannelID != 0 {
			currentChannelID = channel.ChannelID
			break
		}
	}
	ctrl, err := h.rx.GenerateControlMessage(currentChannelID)
	if err != nil {
		h.t.Fatal(err)
	}
	if len(ctrl.Channel.ChannelControl) != len(h.rxChannels) {
		h.t.Fatalf("expected channel control for all attached channels, got %d want %d", len(ctrl.Channel.ChannelControl), len(h.rxChannels))
	}
	return ctrl
}

func (h *directionalIntegrationHarness) applyControl(ctrl ControlMessage, copies int) {
	h.t.Helper()

	for i := 0; i < copies; i++ {
		if err := h.tx.AcceptRemoteControlMessage(ctrl); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *directionalIntegrationHarness) assertReceived(expected [][]byte) {
	h.t.Helper()

	if diff := cmp.Diff(expected, h.received); diff != "" {
		h.t.Fatalf("unexpected directional integration payloads (-want +got):\n%s", diff)
	}
}

func (h *directionalIntegrationHarness) assertSenderDrained() {
	h.t.Helper()

	if len(h.tx.lanes) != 0 {
		h.t.Fatalf("expected sender lanes to be drained, still have %d", len(h.tx.lanes))
	}
}

func (h *directionalIntegrationHarness) expectedLaneCount(messageCount int) int {
	return (messageCount + h.maxDataShardsPerLane - 1) / h.maxDataShardsPerLane
}

func reverseDirectionalFrames(frames []directionalIntegrationFrame) {
	for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
		frames[i], frames[j] = frames[j], frames[i]
	}
}

func makeDirectionalIntegrationMessages(count int) [][]byte {
	messages := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		messages = append(messages, []byte(fmt.Sprintf("msg-%02d", i)))
	}
	return messages
}
