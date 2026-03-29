package rrpitBidirectionalSession

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lunixbochs/struc"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferLane"
)

type bidirectionalIntegrationConfig struct {
	name                   string
	channelCount           int
	shardSize              int
	maxDataShardsPerLane   int
	maxBufferedLanes       int
	channelHistoryCapacity int
	aOddChannelIDs         bool
	bOddChannelIDs         bool
}

type bidirectionalIntegrationHarness struct {
	t         *testing.T
	config    bidirectionalIntegrationConfig
	timestamp uint64

	a *bidirectionalIntegrationPeer
	b *bidirectionalIntegrationPeer
}

type bidirectionalIntegrationPeer struct {
	name string

	session *BidirectionalSession

	outboundWriters []*recordingWriteCloser
	inboundChannels []*rrpitMaterializedTransferChannel.ChannelRx
	txChannelIDs    []uint64
	nextRead        []int

	received    [][]byte
	seenControl []rriptMonoDirectionSession.ControlMessage
}

type bidirectionalIntegrationFrame struct {
	from         *bidirectionalIntegrationPeer
	to           *bidirectionalIntegrationPeer
	channelIndex int
	raw          []byte
	packetKind   uint8
	data         *sessionDataPacketForIntegrationTest
	control      *sessionControlPacketForTest
}

type sessionDataPacketForIntegrationTest struct {
	PacketKind uint8
	LaneID     uint64
	Transfer   rrpitTransferLane.TransferData
}

type bidirectionalTickedNetworkProfile struct {
	maxSourceLossPerLane    int
	maxRepairLossPerLane    int
	maxControlLossPerSender int
	maxLatencyTicks         int
}

type bidirectionalScheduledFrame struct {
	deliverAt uint64
	frame     bidirectionalIntegrationFrame
}

type bidirectionalTickedNetwork struct {
	profile  bidirectionalTickedNetworkProfile
	schedule []byte
	cursor   int

	sourceLossBudget  map[string]int
	repairLossBudget  map[string]int
	controlLossBudget map[string]int
	pending           []bidirectionalScheduledFrame
}

const materializedChannelSequenceFieldLengthForIntegration = 8

func TestBidirectionalSessionIntegrationMatrix(t *testing.T) {
	configs := []bidirectionalIntegrationConfig{
		{
			name:                   "single_channel_single_shard",
			channelCount:           1,
			shardSize:              16,
			maxDataShardsPerLane:   1,
			maxBufferedLanes:       8,
			channelHistoryCapacity: 16,
			aOddChannelIDs:         true,
			bOddChannelIDs:         false,
		},
		{
			name:                   "dual_channel_dual_shard",
			channelCount:           2,
			shardSize:              24,
			maxDataShardsPerLane:   2,
			maxBufferedLanes:       8,
			channelHistoryCapacity: 16,
			aOddChannelIDs:         false,
			bOddChannelIDs:         true,
		},
		{
			name:                   "triple_channel_triple_shard",
			channelCount:           3,
			shardSize:              32,
			maxDataShardsPerLane:   3,
			maxBufferedLanes:       8,
			channelHistoryCapacity: 16,
			aOddChannelIDs:         true,
			bOddChannelIDs:         false,
		},
	}

	scenarios := []struct {
		name string
		run  func(*testing.T, *bidirectionalIntegrationHarness)
	}{
		{name: "control_learning_and_hints", run: runBidirectionalControlLearningAndHints},
		{name: "ordered_round_trip", run: runBidirectionalOrderedRoundTrip},
		{name: "reverse_delivery_round_trip", run: runBidirectionalReverseRoundTrip},
		{name: "loss_repair_and_duplicate_control", run: runBidirectionalLossRepairAndDuplicateControl},
		{name: "eventual_delivery_with_loss_and_latency", run: runBidirectionalEventualDeliveryWithLossAndLatency},
	}

	for _, config := range configs {
		config := config
		t.Run(config.name, func(t *testing.T) {
			for _, scenario := range scenarios {
				scenario := scenario
				t.Run(scenario.name, func(t *testing.T) {
					h := newBidirectionalIntegrationHarness(t, config)
					scenario.run(t, h)
				})
			}
		})
	}
}

func TestBidirectionalSessionIntegrationBackpressureRelease(t *testing.T) {
	h := newBidirectionalIntegrationHarness(t, bidirectionalIntegrationConfig{
		name:                   "backpressure_release",
		channelCount:           1,
		shardSize:              16,
		maxDataShardsPerLane:   1,
		maxBufferedLanes:       1,
		channelHistoryCapacity: 16,
		aOddChannelIDs:         true,
		bOddChannelIDs:         false,
	})

	h.bootstrapAndLearn()

	firstA := []byte("a-00")
	firstB := []byte("b-00")
	secondA := []byte("a-01")
	secondB := []byte("b-01")

	h.sendMessages(h.a, firstA)
	h.sendMessages(h.b, firstB)

	h.advanceBothTicks(1)
	h.deliverFrames(h.takeAllPendingFrames())

	aResult := make(chan error, 1)
	bResult := make(chan error, 1)
	go func() {
		aResult <- h.a.session.SendMessage(secondA)
	}()
	go func() {
		bResult <- h.b.session.SendMessage(secondB)
	}()

	select {
	case err := <-aResult:
		t.Fatalf("expected sender A to block before control acknowledgment, got %v", err)
	case <-time.After(10 * time.Millisecond):
	}
	select {
	case err := <-bResult:
		t.Fatalf("expected sender B to block before control acknowledgment, got %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	h.advanceBothTicks(1)
	h.deliverFrames(h.takeAllPendingFrames())

	select {
	case err := <-aResult:
		if err != nil {
			t.Fatalf("expected sender A to unblock successfully, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected sender A to unblock after control acknowledgment")
	}
	select {
	case err := <-bResult:
		if err != nil {
			t.Fatalf("expected sender B to unblock successfully, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected sender B to unblock after control acknowledgment")
	}

	h.advanceBothTicks(1)
	h.deliverFrames(h.takeAllPendingFrames())
	h.advanceBothTicks(1)
	h.deliverFrames(h.takeAllPendingFrames())

	h.assertReceived(h.a, [][]byte{firstB, secondB})
	h.assertReceived(h.b, [][]byte{firstA, secondA})
	h.assertControlCallbacksObserved()
}

func runBidirectionalControlLearningAndHints(t *testing.T, h *bidirectionalIntegrationHarness) {
	h.bootstrapAndLearn()

	h.advanceBothTicks(1)
	frames := h.takeAllPendingFrames()
	if len(frames) != h.config.channelCount*2 {
		t.Fatalf("expected one flooded control per direction/channel after bootstrap, got %d", len(frames))
	}
	for _, frame := range frames {
		if frame.packetKind != rriptMonoDirectionSession.PacketKind_CONTROL || frame.control == nil {
			t.Fatalf("expected control frame after empty tick, got kind %d", frame.packetKind)
		}
		if int(frame.control.Control.Channel.LenChannelControl) != h.config.channelCount {
			t.Fatalf("expected %d channel controls after learning ids, got len field %d", h.config.channelCount, frame.control.Control.Channel.LenChannelControl)
		}
	}
	reverseBidirectionalFrames(frames)
	h.deliverFrames(frames)
	h.assertControlCallbacksObserved()
}

func runBidirectionalOrderedRoundTrip(t *testing.T, h *bidirectionalIntegrationHarness) {
	h.bootstrapAndLearn()

	aToB := makeBidirectionalMessages("a", h.config.maxDataShardsPerLane*3+1)
	bToA := makeBidirectionalMessages("b", h.config.maxDataShardsPerLane*2+1)

	h.sendMessages(h.a, aToB...)
	h.sendMessages(h.b, bToA...)
	h.completeExchange(max(h.expectedLaneCount(len(aToB)), h.expectedLaneCount(len(bToA))), nil, nil)

	h.assertReceived(h.a, bToA)
	h.assertReceived(h.b, aToB)
	h.assertControlCallbacksObserved()
}

func runBidirectionalReverseRoundTrip(t *testing.T, h *bidirectionalIntegrationHarness) {
	h.bootstrapAndLearn()

	aToB := makeBidirectionalMessages("ax", h.config.maxDataShardsPerLane*3+1)
	bToA := makeBidirectionalMessages("bx", h.config.maxDataShardsPerLane*3)

	h.sendMessages(h.a, aToB...)
	h.sendMessages(h.b, bToA...)
	h.completeExchange(
		max(h.expectedLaneCount(len(aToB)), h.expectedLaneCount(len(bToA))),
		func(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
			reverseBidirectionalFrames(frames)
			return frames
		},
		func(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
			reverseBidirectionalFrames(frames)
			return frames
		},
	)

	h.assertReceived(h.a, bToA)
	h.assertReceived(h.b, aToB)
	h.assertControlCallbacksObserved()
}

func runBidirectionalLossRepairAndDuplicateControl(t *testing.T, h *bidirectionalIntegrationHarness) {
	h.bootstrapAndLearn()

	aToB := makeBidirectionalMessages("la", h.config.maxDataShardsPerLane*3)
	bToA := makeBidirectionalMessages("lb", h.config.maxDataShardsPerLane*3)

	h.sendMessages(h.a, aToB...)
	h.sendMessages(h.b, bToA...)
	h.completeExchange(
		max(h.expectedLaneCount(len(aToB)), h.expectedLaneCount(len(bToA))),
		func(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
			frames = dropFirstSourceShardPerLane(frames)
			frames = duplicateFirstRepairAndControlPerSender(frames)
			reverseBidirectionalFrames(frames)
			return frames
		},
		func(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
			frames = duplicateFirstControlPerSender(frames)
			reverseBidirectionalFrames(frames)
			return frames
		},
	)

	h.assertReceived(h.a, bToA)
	h.assertReceived(h.b, aToB)
	h.assertControlCallbacksObserved()
}

func runBidirectionalEventualDeliveryWithLossAndLatency(t *testing.T, h *bidirectionalIntegrationHarness) {
	h.bootstrapAndLearn()

	aToB := makeBidirectionalMessages("ea", h.config.maxDataShardsPerLane*3+2)
	bToA := makeBidirectionalMessages("eb", h.config.maxDataShardsPerLane*3+1)

	h.sendMessages(h.a, aToB...)
	h.sendMessages(h.b, bToA...)

	profile := bidirectionalTickedNetworkProfile{
		maxSourceLossPerLane:    2,
		maxRepairLossPerLane:    1,
		maxControlLossPerSender: 2,
		maxLatencyTicks:         3,
	}
	network := newBidirectionalTickedNetwork(profile, []byte{3, 1, 4, 1, 5, 9, 2, 6})
	settleTicks := h.settleTicksForProfile(profile)

	h.runUntilExactDelivery(
		bToA,
		aToB,
		network,
		h.maxTicksForProfile(len(aToB), len(bToA), profile, settleTicks),
		settleTicks,
	)
	h.assertControlCallbacksObserved()
}

func newBidirectionalIntegrationHarness(t *testing.T, config bidirectionalIntegrationConfig) *bidirectionalIntegrationHarness {
	t.Helper()

	h := &bidirectionalIntegrationHarness{
		t:      t,
		config: config,
	}
	h.a = h.newPeer("a", config.aOddChannelIDs)
	h.b = h.newPeer("b", config.bOddChannelIDs)

	h.a.nextRead = make([]int, config.channelCount)
	h.b.nextRead = make([]int, config.channelCount)

	for i := 0; i < config.channelCount; i++ {
		aWriter := &recordingWriteCloser{}
		aChannelID, err := h.a.session.AttachTxChannel(aWriter)
		if err != nil {
			t.Fatal(err)
		}
		bRx, err := h.b.session.AttachRxChannel()
		if err != nil {
			t.Fatal(err)
		}

		bWriter := &recordingWriteCloser{}
		bChannelID, err := h.b.session.AttachTxChannel(bWriter)
		if err != nil {
			t.Fatal(err)
		}
		aRx, err := h.a.session.AttachRxChannel()
		if err != nil {
			t.Fatal(err)
		}

		h.a.outboundWriters = append(h.a.outboundWriters, aWriter)
		h.a.txChannelIDs = append(h.a.txChannelIDs, aChannelID)
		h.a.inboundChannels = append(h.a.inboundChannels, aRx)

		h.b.outboundWriters = append(h.b.outboundWriters, bWriter)
		h.b.txChannelIDs = append(h.b.txChannelIDs, bChannelID)
		h.b.inboundChannels = append(h.b.inboundChannels, bRx)
	}

	return h
}

func (h *bidirectionalIntegrationHarness) newPeer(name string, oddChannelIDs bool) *bidirectionalIntegrationPeer {
	peer := &bidirectionalIntegrationPeer{name: name}
	peer.session = mustNewBidirectionalSession(h.t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    h.config.shardSize,
			MaxBufferedLanes: h.config.maxBufferedLanes,
			OnMessage: func(data []byte) error {
				peer.received = append(peer.received, append([]byte(nil), data...))
				return nil
			},
			OnRemoteControlMsg: func(ctrl rriptMonoDirectionSession.ControlMessage) error {
				peer.seenControl = append(peer.seenControl, ctrl)
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  h.config.shardSize,
			MaxDataShardsPerLane:           h.config.maxDataShardsPerLane,
			MaxBufferedLanes:               h.config.maxBufferedLanes,
			MaxRewindableTimestampNum:      h.config.channelHistoryCapacity,
			MaxRewindableControlMessageNum: h.config.channelHistoryCapacity,
			OddChannelIDs:                  oddChannelIDs,
		},
	})
	return peer
}

func (h *bidirectionalIntegrationHarness) bootstrapAndLearn() {
	h.t.Helper()

	h.advanceBothTicks(1)
	frames := h.takeAllPendingFrames()
	if len(frames) != h.config.channelCount*2 {
		h.t.Fatalf("expected one control flood per direction/channel during bootstrap, got %d", len(frames))
	}
	for _, frame := range frames {
		if frame.packetKind != rriptMonoDirectionSession.PacketKind_CONTROL || frame.control == nil {
			h.t.Fatalf("expected bootstrap to emit only control frames, got kind %d", frame.packetKind)
		}
	}

	h.deliverFrames(frames)
	h.assertLearnedChannelIDs()
	h.assertControlCallbacksObserved()
}

func (h *bidirectionalIntegrationHarness) completeExchange(
	repairTicks int,
	dataTransform func([]bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame,
	ackTransform func([]bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame,
) {
	h.t.Helper()

	h.advanceBothTicks(repairTicks)
	frames := h.takeAllPendingFrames()
	if dataTransform != nil {
		frames = dataTransform(frames)
	}
	h.deliverFrames(frames)

	h.advanceBothTicks(1)
	ackFrames := h.takeAllPendingFrames()
	if ackTransform != nil {
		ackFrames = ackTransform(ackFrames)
	}
	h.deliverFrames(ackFrames)
}

func (h *bidirectionalIntegrationHarness) sendMessages(peer *bidirectionalIntegrationPeer, messages ...[]byte) {
	h.t.Helper()

	for _, message := range messages {
		if err := peer.session.SendMessage(message); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *bidirectionalIntegrationHarness) advanceBothTicks(count int) {
	h.t.Helper()

	for i := 0; i < count; i++ {
		h.timestamp++
		if err := h.a.session.OnNewTimestamp(h.timestamp); err != nil {
			h.t.Fatal(err)
		}
		if err := h.b.session.OnNewTimestamp(h.timestamp); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *bidirectionalIntegrationHarness) takeAllPendingFrames() []bidirectionalIntegrationFrame {
	h.t.Helper()

	frames := h.takePendingFrames(h.a, h.b)
	frames = append(frames, h.takePendingFrames(h.b, h.a)...)
	return frames
}

func (h *bidirectionalIntegrationHarness) takePendingFrames(
	from *bidirectionalIntegrationPeer,
	to *bidirectionalIntegrationPeer,
) []bidirectionalIntegrationFrame {
	h.t.Helper()

	frames := make([]bidirectionalIntegrationFrame, 0)
	for channelIndex, writer := range from.outboundWriters {
		for from.nextRead[channelIndex] < len(writer.writes) {
			raw := append([]byte(nil), writer.writes[from.nextRead[channelIndex]]...)
			_, payload := splitMaterializedWireForIntegration(h.t, raw)
			if len(payload) == 0 {
				h.t.Fatal("empty materialized session payload")
			}

			frame := bidirectionalIntegrationFrame{
				from:         from,
				to:           to,
				channelIndex: channelIndex,
				raw:          raw,
				packetKind:   payload[0],
			}
			switch payload[0] {
			case rriptMonoDirectionSession.PacketKind_DATA:
				packet := mustUnmarshalSessionDataPacketForIntegration(h.t, payload)
				frame.data = &packet
			case rriptMonoDirectionSession.PacketKind_CONTROL:
				packet := mustUnmarshalSessionControlPacketForIntegration(h.t, payload)
				frame.control = &packet
				expectedChannelID := from.txChannelIDs[channelIndex]
				if packet.Control.FloodChannel.CurrentChannelID != expectedChannelID {
					h.t.Fatalf(
						"expected flooded control on %s channel %d to carry id %d, got %d",
						from.name,
						channelIndex,
						expectedChannelID,
						packet.Control.FloodChannel.CurrentChannelID,
					)
				}
			default:
				h.t.Fatalf("unknown session packet kind %d", payload[0])
			}

			frames = append(frames, frame)
			from.nextRead[channelIndex] += 1
		}
	}
	return frames
}

func (h *bidirectionalIntegrationHarness) deliverFrames(frames []bidirectionalIntegrationFrame) {
	h.t.Helper()

	for _, frame := range frames {
		if err := frame.to.inboundChannels[frame.channelIndex].OnNewMessageArrived(frame.raw); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *bidirectionalIntegrationHarness) assertLearnedChannelIDs() {
	h.t.Helper()

	h.assertPeerLearnedChannelIDs(h.a, h.b.txChannelIDs)
	h.assertPeerLearnedChannelIDs(h.b, h.a.txChannelIDs)
}

func (h *bidirectionalIntegrationHarness) assertPeerLearnedChannelIDs(peer *bidirectionalIntegrationPeer, expected []uint64) {
	h.t.Helper()

	seen := map[uint64]bool{}
	for i, channel := range peer.inboundChannels {
		if channel.ChannelID != expected[i] {
			h.t.Fatalf("expected %s inbound channel %d to learn id %d, got %d", peer.name, i, expected[i], channel.ChannelID)
		}
		if channel.ChannelID == 0 {
			h.t.Fatalf("expected %s inbound channel %d to learn a non-zero id", peer.name, i)
		}
		if seen[channel.ChannelID] {
			h.t.Fatalf("expected %s inbound learned ids to stay unique, got duplicate %d", peer.name, channel.ChannelID)
		}
		seen[channel.ChannelID] = true
	}
}

func (h *bidirectionalIntegrationHarness) assertReceived(peer *bidirectionalIntegrationPeer, expected [][]byte) {
	h.t.Helper()

	if diff := cmp.Diff(expected, peer.received); diff != "" {
		h.t.Fatalf("unexpected payloads received by %s (-want +got):\n%s", peer.name, diff)
	}
}

func (h *bidirectionalIntegrationHarness) assertControlCallbacksObserved() {
	h.t.Helper()

	if len(h.a.seenControl) == 0 {
		h.t.Fatal("expected peer A to observe remote control callbacks")
	}
	if len(h.b.seenControl) == 0 {
		h.t.Fatal("expected peer B to observe remote control callbacks")
	}
}

func (h *bidirectionalIntegrationHarness) runUntilExactDelivery(
	expectedAtA [][]byte,
	expectedAtB [][]byte,
	network *bidirectionalTickedNetwork,
	maxTicks int,
	settleTicks int,
) {
	h.t.Helper()

	exactTicks := 0
	for i := 0; i < maxTicks; i++ {
		h.advanceBothTicks(1)
		network.ingestFrames(h.takeAllPendingFrames(), h.timestamp)
		h.deliverFrames(network.takeDueFrames(h.timestamp))

		aExact := h.assertReceivedPrefix(h.a, expectedAtA)
		bExact := h.assertReceivedPrefix(h.b, expectedAtB)
		if aExact && bExact {
			exactTicks += 1
			if exactTicks >= settleTicks {
				return
			}
			continue
		}
		exactTicks = 0
	}

	h.assertReceived(h.a, expectedAtA)
	h.assertReceived(h.b, expectedAtB)
	h.t.Fatalf("exact delivery did not stabilize within %d ticks", maxTicks)
}

func (h *bidirectionalIntegrationHarness) assertReceivedPrefix(
	peer *bidirectionalIntegrationPeer,
	expected [][]byte,
) bool {
	h.t.Helper()

	if len(peer.received) > len(expected) {
		h.t.Fatalf(
			"peer %s received too many payloads: got %d want at most %d",
			peer.name,
			len(peer.received),
			len(expected),
		)
	}
	for i := range peer.received {
		if diff := cmp.Diff(expected[i], peer.received[i]); diff != "" {
			h.t.Fatalf(
				"peer %s received payload %d out of order or duplicated (-want +got):\n%s",
				peer.name,
				i,
				diff,
			)
		}
	}
	return len(peer.received) == len(expected)
}

func (h *bidirectionalIntegrationHarness) settleTicksForProfile(profile bidirectionalTickedNetworkProfile) int {
	return profile.maxLatencyTicks + profile.maxControlLossPerSender + 3
}

func (h *bidirectionalIntegrationHarness) maxTicksForProfile(
	aMessageCount int,
	bMessageCount int,
	profile bidirectionalTickedNetworkProfile,
	settleTicks int,
) int {
	lanes := max(h.expectedLaneCount(aMessageCount), h.expectedLaneCount(bMessageCount))
	perLaneRepairWork := h.config.maxDataShardsPerLane + profile.maxRepairLossPerLane + 2
	return 8 +
		2*profile.maxLatencyTicks +
		2*profile.maxControlLossPerSender +
		settleTicks +
		lanes*(profile.maxSourceLossPerLane+perLaneRepairWork)
}

func (h *bidirectionalIntegrationHarness) expectedLaneCount(messageCount int) int {
	return expectedLaneCountForConfig(messageCount, h.config.maxDataShardsPerLane)
}

func newBidirectionalTickedNetwork(
	profile bidirectionalTickedNetworkProfile,
	schedule []byte,
) *bidirectionalTickedNetwork {
	clonedSchedule := append([]byte(nil), schedule...)
	if len(clonedSchedule) == 0 {
		clonedSchedule = []byte{0}
	}
	return &bidirectionalTickedNetwork{
		profile:           profile,
		schedule:          clonedSchedule,
		sourceLossBudget:  map[string]int{},
		repairLossBudget:  map[string]int{},
		controlLossBudget: map[string]int{},
	}
}

func (n *bidirectionalTickedNetwork) ingestFrames(frames []bidirectionalIntegrationFrame, currentTick uint64) {
	for _, frame := range frames {
		if n.shouldDrop(frame) {
			continue
		}
		n.pending = append(n.pending, bidirectionalScheduledFrame{
			deliverAt: currentTick + uint64(n.latencyForFrame()),
			frame:     frame,
		})
	}
}

func (n *bidirectionalTickedNetwork) takeDueFrames(currentTick uint64) []bidirectionalIntegrationFrame {
	due := make([]bidirectionalIntegrationFrame, 0, len(n.pending))
	pending := make([]bidirectionalScheduledFrame, 0, len(n.pending))
	for _, scheduled := range n.pending {
		if scheduled.deliverAt <= currentTick {
			due = append(due, scheduled.frame)
			continue
		}
		pending = append(pending, scheduled)
	}
	n.pending = pending

	if len(due) > 1 && n.nextByte()&1 == 1 {
		reverseBidirectionalFrames(due)
	}
	if len(due) > 1 {
		due = rotateBidirectionalFrames(due, int(n.nextByte())%len(due))
	}
	return due
}

func (n *bidirectionalTickedNetwork) shouldDrop(frame bidirectionalIntegrationFrame) bool {
	key, budgets, maxLoss := n.lossBudgetState(frame)
	if maxLoss <= 0 {
		return false
	}
	remaining, found := budgets[key]
	if !found {
		remaining = int(n.nextByte()) % (maxLoss + 1)
		budgets[key] = remaining
	}
	if remaining == 0 {
		return false
	}
	budgets[key] = remaining - 1
	return true
}

func (n *bidirectionalTickedNetwork) lossBudgetState(
	frame bidirectionalIntegrationFrame,
) (string, map[string]int, int) {
	switch {
	case frame.packetKind == rriptMonoDirectionSession.PacketKind_CONTROL && frame.control != nil:
		return fmt.Sprintf("%s:%d:control", frame.from.name, frame.channelIndex), n.controlLossBudget, n.profile.maxControlLossPerSender
	case frame.packetKind == rriptMonoDirectionSession.PacketKind_DATA &&
		frame.data != nil &&
		frame.data.Transfer.TotalDataShards == 0:
		return fmt.Sprintf("%s:%d:source", frame.from.name, frame.data.LaneID), n.sourceLossBudget, n.profile.maxSourceLossPerLane
	case frame.packetKind == rriptMonoDirectionSession.PacketKind_DATA && frame.data != nil:
		return fmt.Sprintf("%s:%d:repair", frame.from.name, frame.data.LaneID), n.repairLossBudget, n.profile.maxRepairLossPerLane
	default:
		return "", nil, 0
	}
}

func (n *bidirectionalTickedNetwork) latencyForFrame() int {
	if n.profile.maxLatencyTicks <= 0 {
		return 0
	}
	return int(n.nextByte()) % (n.profile.maxLatencyTicks + 1)
}

func (n *bidirectionalTickedNetwork) nextByte() byte {
	if len(n.schedule) == 0 {
		return 0
	}
	value := n.schedule[n.cursor%len(n.schedule)]
	n.cursor += 1
	return value
}

func splitMaterializedWireForIntegration(t *testing.T, wire []byte) (uint64, []byte) {
	t.Helper()

	if len(wire) < materializedChannelSequenceFieldLengthForIntegration {
		t.Fatalf("materialized wire too short: %d", len(wire))
	}
	return binary.BigEndian.Uint64(wire[:materializedChannelSequenceFieldLengthForIntegration]), append([]byte(nil), wire[materializedChannelSequenceFieldLengthForIntegration:]...)
}

func mustUnmarshalSessionDataPacketForIntegration(t *testing.T, payload []byte) sessionDataPacketForIntegrationTest {
	t.Helper()

	var packet sessionDataPacketForIntegrationTest
	if err := struc.Unpack(bytes.NewReader(payload), &packet); err != nil {
		t.Fatalf("failed to unpack session data packet: %v", err)
	}
	return packet
}

func mustUnmarshalSessionControlPacketForIntegration(t *testing.T, payload []byte) sessionControlPacketForTest {
	t.Helper()

	var packet sessionControlPacketForTest
	if err := struc.Unpack(bytes.NewReader(payload), &packet); err != nil {
		t.Fatalf("failed to unpack session control packet: %v", err)
	}
	return packet
}

func reverseBidirectionalFrames(frames []bidirectionalIntegrationFrame) {
	for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
		frames[i], frames[j] = frames[j], frames[i]
	}
}

func dropFirstSourceShardPerLane(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
	seen := map[string]bool{}
	filtered := make([]bidirectionalIntegrationFrame, 0, len(frames))
	for _, frame := range frames {
		if frame.packetKind == rriptMonoDirectionSession.PacketKind_DATA && frame.data != nil && frame.data.Transfer.TotalDataShards == 0 {
			key := fmt.Sprintf("%s:%d", frame.from.name, frame.data.LaneID)
			if !seen[key] {
				seen[key] = true
				continue
			}
		}
		filtered = append(filtered, frame)
	}
	return filtered
}

func duplicateFirstRepairAndControlPerSender(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
	withDuplicates := append([]bidirectionalIntegrationFrame{}, frames...)
	seenRepair := map[string]bool{}
	seenControl := map[string]bool{}

	for _, frame := range frames {
		switch {
		case frame.packetKind == rriptMonoDirectionSession.PacketKind_CONTROL && frame.control != nil && !seenControl[frame.from.name]:
			seenControl[frame.from.name] = true
			withDuplicates = append(withDuplicates, frame)
		case frame.packetKind == rriptMonoDirectionSession.PacketKind_DATA &&
			frame.data != nil &&
			frame.data.Transfer.TotalDataShards != 0 &&
			!seenRepair[frame.from.name]:
			seenRepair[frame.from.name] = true
			withDuplicates = append(withDuplicates, frame)
		}
	}

	return withDuplicates
}

func duplicateFirstControlPerSender(frames []bidirectionalIntegrationFrame) []bidirectionalIntegrationFrame {
	withDuplicates := append([]bidirectionalIntegrationFrame{}, frames...)
	seen := map[string]bool{}
	for _, frame := range frames {
		if frame.packetKind != rriptMonoDirectionSession.PacketKind_CONTROL || frame.control == nil || seen[frame.from.name] {
			continue
		}
		seen[frame.from.name] = true
		withDuplicates = append(withDuplicates, frame)
	}
	return withDuplicates
}

func makeBidirectionalMessages(prefix string, count int) [][]byte {
	messages := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		messages = append(messages, []byte(fmt.Sprintf("%s-%02d", prefix, i)))
	}
	return messages
}

func expectedLaneCountForConfig(messageCount int, maxDataShardsPerLane int) int {
	return (messageCount + maxDataShardsPerLane - 1) / maxDataShardsPerLane
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
