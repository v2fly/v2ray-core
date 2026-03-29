package rriptMonoDirectionSession

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lunixbochs/struc"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferLane"
)

const materializedChannelSequenceFieldLength = 8

func TestSessionTxAttachChannelAndSendMessage(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		OddChannelIDs:                  true,
	})

	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}
	if channelID != 1 {
		t.Fatalf("expected first odd channel id 1, got %d", channelID)
	}

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 1 {
		t.Fatalf("expected 1 wire message, got %d", len(writer.writes))
	}

	channelSeq, payload := splitMaterializedWire(t, writer.writes[0])
	if channelSeq != 0 {
		t.Fatalf("expected first materialized channel seq 0, got %d", channelSeq)
	}

	packet := mustUnmarshalSessionDataPacket(t, payload)
	if packet.PacketKind != PacketKind_DATA {
		t.Fatalf("expected session packet kind DATA, got %d", packet.PacketKind)
	}
	if packet.LaneID != 0 {
		t.Fatalf("expected first lane id 0, got %d", packet.LaneID)
	}
	if packet.Transfer.Seq != 0 {
		t.Fatalf("expected first transfer seq 0, got %d", packet.Transfer.Seq)
	}
	if packet.Transfer.TotalDataShards != 0 {
		t.Fatalf("expected data packet to omit total shard count, got %d", packet.Transfer.TotalDataShards)
	}
	if string(packet.Transfer.Data) != "hello" {
		t.Fatalf("expected payload hello, got %q", string(packet.Transfer.Data))
	}
}

func TestSessionTxOnNewTimestampSendsRepairPacket(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	if _, err := tx.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("beta")); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(77); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected 3 wire messages, got %d", len(writer.writes))
	}

	_, payload := splitMaterializedWire(t, writer.writes[2])
	packet := mustUnmarshalSessionDataPacket(t, payload)
	if packet.LaneID != 0 {
		t.Fatalf("expected repair packet for lane 0, got %d", packet.LaneID)
	}
	if packet.Transfer.TotalDataShards != 2 {
		t.Fatalf("expected repair packet to announce 2 total data shards, got %d", packet.Transfer.TotalDataShards)
	}
	if packet.Transfer.Seq != 2 {
		t.Fatalf("expected first repair packet seq 2, got %d", packet.Transfer.Seq)
	}
	if len(packet.Transfer.Data) != 16 {
		t.Fatalf("expected repair symbol size 16, got %d", len(packet.Transfer.Data))
	}
	if tx.txChannelsConfig[0].Status.TimestampLastSent != 77 {
		t.Fatalf("expected channel rate timestamp 77, got %d", tx.txChannelsConfig[0].Status.TimestampLastSent)
	}
}

func TestSessionTxOnNewTimestampSendsConfiguredInitialRepairPackets(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           4,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			InitialRepairShardRatio: 1.5,
		},
	})
	if _, err := tx.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("beta")); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(77); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 5 {
		t.Fatalf("expected 5 wire messages, got %d", len(writer.writes))
	}

	for i, wantSeq := range []uint32{2, 3, 4} {
		_, payload := splitMaterializedWire(t, writer.writes[2+i])
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.LaneID != 0 {
			t.Fatalf("expected configured repair packet for lane 0, got %d", packet.LaneID)
		}
		if packet.Transfer.TotalDataShards != 2 {
			t.Fatalf("expected repair packet to announce 2 total data shards, got %d", packet.Transfer.TotalDataShards)
		}
		if packet.Transfer.Seq != wantSeq {
			t.Fatalf("expected repair packet seq %d, got %d", wantSeq, packet.Transfer.Seq)
		}
	}
}

func TestSessionTxLaneRepairWeightUsesPeerSeenChunks(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           4,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			LaneRepairWeight: []float64{0.5},
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("beta")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 2 {
		t.Fatalf("expected only the source packets before control feedback, got %d writes", len(writer.writes))
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 0},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 4 {
		t.Fatalf("expected two weighted repair packets after control feedback against the K+1 target, got %d writes", len(writer.writes))
	}

	for i, wantSeq := range []uint32{2, 3} {
		_, payload := splitMaterializedWire(t, writer.writes[2+i])
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.Transfer.Seq != wantSeq {
			t.Fatalf("expected weighted repair packet seq %d, got %d", wantSeq, packet.Transfer.Seq)
		}
	}
}

func TestSessionTxSecondaryRepairResendsAfterConfiguredTicks(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           4,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:      1,
			TimeResendSecondaryRepairShard: 2,
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 1 {
		t.Fatalf("expected no repair before remote feedback, got %d writes", len(writer.writes))
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 0},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(3); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected the first scheduled secondary repair burst to send 2 repair shards against the K+1 target, got %d writes", len(writer.writes))
	}
	if err := tx.OnNewTimestamp(4); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(5); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 5 {
		t.Fatalf("expected the next scheduled secondary repair burst after another two ticks, got %d writes", len(writer.writes))
	}

	for i, wantSeq := range []uint32{1, 2, 3, 4} {
		_, payload := splitMaterializedWire(t, writer.writes[1+i])
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.Transfer.Seq != wantSeq {
			t.Fatalf("expected repair packet seq %d, got %d", wantSeq, packet.Transfer.Seq)
		}
	}
}

func TestSessionTxSecondaryRepairContinuesForOldestLaneAfterSeenChunksReachTotal(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       1,
			StaleLaneFinalizedAgeThresholdTicks:  2,
			StaleLaneProgressStallThresholdTicks: 2,
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected the first scheduled secondary repair burst at tick 1 to send 2 repair shards against the K+1 target, got %d writes", len(writer.writes))
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 1},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(3); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 7 {
		t.Fatalf("expected stale oldest lane to keep receiving a tail of 2 repairs per tick, got %d writes", len(writer.writes))
	}

	for i, wantSeq := range []uint32{3, 4, 5, 6} {
		_, payload := splitMaterializedWire(t, writer.writes[3+i])
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.Transfer.Seq != wantSeq {
			t.Fatalf("expected repair packet seq %d, got %d", wantSeq, packet.Transfer.Seq)
		}
	}
}

func TestSessionTxSecondaryRepairRecomputesMissingAtResendTime(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           4,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:      1,
			TimeResendSecondaryRepairShard: 2,
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	for _, payload := range [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")} {
		if err := tx.SendMessage(payload); err != nil {
			t.Fatal(err)
		}
	}
	if len(writer.writes) != 4 {
		t.Fatalf("expected 4 source writes, got %d", len(writer.writes))
	}

	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 4 {
		t.Fatalf("expected no secondary repair before scheduled resend, got %d writes", len(writer.writes))
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 3},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 6 {
		t.Fatalf("expected exactly two repair shards at the scheduled resend tick against the K+1 target, got %d writes", len(writer.writes))
	}

	if err := tx.OnNewTimestamp(3); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 6 {
		t.Fatalf("expected no additional resend before the next interval, got %d writes", len(writer.writes))
	}

	for i, wantSeq := range []uint32{4, 5} {
		_, payload := splitMaterializedWire(t, writer.writes[4+i])
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.Transfer.Seq != wantSeq {
			t.Fatalf("expected scheduled repair packet seq %d, got %d", wantSeq, packet.Transfer.Seq)
		}
	}
}

func TestSessionTxAcceptRemoteControlMessage(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(100); err != nil {
		t.Fatal(err)
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo: 0,
		},
		Channel: SessionChannelControlMessage{
			LenChannelControl: 1,
			ChannelControl: []rrpitTransferChannel.ChannelControlMessage{
				{
					ChannelID:                  channelID,
					TotalPacketReceived:        2,
					LastSequenceNumberReceived: 1,
				},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(tx.lanes) != 0 {
		t.Fatalf("expected acknowledged lane to be dropped, still have %d lanes", len(tx.lanes))
	}
	if tx.firstLaneID != 1 {
		t.Fatalf("expected first lane id to advance to 1, got %d", tx.firstLaneID)
	}

	if _, err := tx.txChannelsConfig[0].MaterializeChannel.RemoteLastSeenMessageSenderTimestamp(); err != nil {
		t.Fatalf("expected channel control to be accepted, got %v", err)
	}
}

func TestSessionTxSeenChunksDoesNotCompleteLane(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       1,
			StaleLaneFinalizedAgeThresholdTicks:  2,
			StaleLaneProgressStallThresholdTicks: 2,
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 8},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(tx.lanes) != 1 {
		t.Fatalf("expected lane to stay active until LaneACKTo advances, have %d lanes", len(tx.lanes))
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 5 {
		t.Fatalf("expected a stale-lane repair tail of 2 after high SeenChunks without a completion sentinel, got %d total writes", len(writer.writes))
	}
}

func TestSessionTxFinalizeResetsProgressTimestamp(t *testing.T) {
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	tx.currentTimestampInitialized = true
	tx.currentTimestamp = 3

	lane, err := tx.createLane()
	if err != nil {
		t.Fatal(err)
	}
	if lane.CreatedAtTimestamp != 3 || lane.LastProgressTimestamp != 3 {
		t.Fatalf("unexpected creation timestamps: %+v", lane)
	}

	transfer, err := lane.TransferLane.AddData([]byte("ok"))
	if err != nil {
		t.Fatal(err)
	}
	if transfer.Seq != 0 {
		t.Fatalf("expected first source seq 0, got %d", transfer.Seq)
	}
	lane.DataShards += 1

	tx.currentTimestamp = 9
	tx.finalizeLane(lane)
	if lane.FinalizedAtTimestamp != 9 {
		t.Fatalf("expected finalized timestamp 9, got %d", lane.FinalizedAtTimestamp)
	}
	if lane.LastProgressTimestamp != 9 {
		t.Fatalf("expected progress timestamp to reset on finalize, got %d", lane.LastProgressTimestamp)
	}
}

func TestSessionTxLaneDoesNotBecomeStaleImmediatelyAfterFinalize(t *testing.T) {
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       1,
			StaleLaneFinalizedAgeThresholdTicks:  4,
			StaleLaneProgressStallThresholdTicks: 4,
		},
	})
	tx.currentTimestampInitialized = true
	tx.currentTimestamp = 1

	lane, err := tx.createLane()
	if err != nil {
		t.Fatal(err)
	}

	tx.currentTimestamp = 20
	if _, err := lane.TransferLane.AddData([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	lane.DataShards += 1
	tx.finalizeLane(lane)

	if tx.isStaleOldestLane(lane) {
		t.Fatal("lane should not become stale immediately after finalize")
	}
}

func TestSessionTxLaneMissingShardsTargetsDataShardsPlusOne(t *testing.T) {
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           4,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	lane := &txLane{
		TotalDataShards:     4,
		PeerSeenChunksKnown: true,
		PeerSeenChunks:      3,
	}

	if missing := tx.laneMissingShards(lane); missing != 2 {
		t.Fatalf("expected missing shards to target K+1, got %d", missing)
	}
}

func TestSessionTxStaleOldestLaneSecondaryPriorityBeatsLaterInitialRepair(t *testing.T) {
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       1,
			StaleLaneFinalizedAgeThresholdTicks:  1,
			StaleLaneProgressStallThresholdTicks: 1,
		},
	})
	tx.currentTimestampInitialized = true
	tx.currentTimestamp = 5
	tx.lanes = []*txLane{
		{
			LaneID:                        0,
			Finalized:                     true,
			TotalDataShards:               1,
			PeerSeenChunksKnown:           true,
			PeerSeenChunks:                1,
			SecondaryRepairPacketsPending: 2,
			FinalizedAtTimestamp:          1,
			LastProgressTimestamp:         1,
		},
		{
			LaneID:                      1,
			Finalized:                   true,
			TotalDataShards:             1,
			InitialRepairPacketsPending: 1,
			FinalizedAtTimestamp:        5,
			LastProgressTimestamp:       5,
		},
	}

	lane, kind, index := tx.nextConfiguredRepair(nil)
	if lane == nil {
		t.Fatal("expected a scheduled repair lane")
	}
	if kind != repairSendSecondary || index != 0 || lane.LaneID != 0 {
		t.Fatalf("expected stale oldest lane secondary repair first, got kind=%d index=%d lane=%d", kind, index, lane.LaneID)
	}
}

func TestSessionTxBecomingOldestBootstrapsStaleMonitoring(t *testing.T) {
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       5,
			StaleLaneFinalizedAgeThresholdTicks:  2,
			StaleLaneProgressStallThresholdTicks: 2,
		},
	})
	tx.currentTimestampInitialized = true
	tx.currentTimestamp = 40
	tx.firstLaneID = 100
	tx.lanes = []*txLane{
		{
			LaneID:                100,
			Finalized:             true,
			TotalDataShards:       1,
			PeerSeenChunksKnown:   true,
			PeerSeenChunks:        65535,
			PeerReconstructed:     true,
			CreatedAtTimestamp:    1,
			FinalizedAtTimestamp:  2,
			LastProgressTimestamp: 3,
		},
		{
			LaneID:                101,
			Finalized:             true,
			TotalDataShards:       20,
			PeerSeenChunksKnown:   true,
			PeerSeenChunks:        20,
			CreatedAtTimestamp:    4,
			FinalizedAtTimestamp:  5,
			LastProgressTimestamp: 6,
		},
	}

	tx.dropLanesThrough(100)
	if tx.firstLaneID != 101 {
		t.Fatalf("expected first lane id to advance to 101, got %d", tx.firstLaneID)
	}
	lane := tx.lanes[0]
	if lane.NextSecondaryRepairTimestamp != 0 {
		t.Fatalf("expected no preexisting secondary timer, got %d", lane.NextSecondaryRepairTimestamp)
	}

	tx.scheduleSecondaryRepairResends(40)
	if lane.NextSecondaryRepairTimestamp != 45 {
		t.Fatalf("expected stale-monitoring timer to bootstrap at 45, got %d", lane.NextSecondaryRepairTimestamp)
	}
}

func TestSessionTxAllowsNewSourceDataWhenOldestLaneIsStaleUntilLaneBufferFull(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               2,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			InitialRepairShardRatio:              1,
			SecondaryRepairShardRatio:            1,
			TimeResendSecondaryRepairShard:       1,
			StaleLaneFinalizedAgeThresholdTicks:  2,
			StaleLaneProgressStallThresholdTicks: 2,
		},
	})
	if _, err := tx.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("a")); err != nil {
		t.Fatal(err)
	}
	for ts := uint64(1); ts <= 4; ts++ {
		if err := tx.OnNewTimestamp(ts); err != nil {
			t.Fatal(err)
		}
	}

	if err := tx.SendMessage([]byte("b")); err != nil {
		t.Fatalf("expected stale oldest lane to still allow running ahead before buffer is full, got %v", err)
	}
	if err := tx.SendMessage([]byte("c")); !errors.Is(err, ErrTxLaneBufferFull) {
		t.Fatalf("expected max buffered lanes to be the backpressure limit, got %v", err)
	}
}

func TestSessionTxBackpressuresNewSourceDataWhenOldestLaneIsStaleIfConfigured(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               16,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		Reconstruction: SessionTxReconstructionConfig{
			InitialRepairShardRatio:                       1,
			SecondaryRepairShardRatio:                     1,
			TimeResendSecondaryRepairShard:                1,
			StaleLaneFinalizedAgeThresholdTicks:           2,
			StaleLaneProgressStallThresholdTicks:          2,
			AlwaysRestrictSourceDataWhenOldestLaneStalled: true,
		},
	})
	if _, err := tx.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("a")); err != nil {
		t.Fatal(err)
	}
	for ts := uint64(1); ts <= 4; ts++ {
		if err := tx.OnNewTimestamp(ts); err != nil {
			t.Fatal(err)
		}
	}

	if err := tx.SendMessage([]byte("b")); !errors.Is(err, ErrTxLaneBufferFull) {
		t.Fatalf("expected configured stale-oldest-lane backpressure, got %v", err)
	}
}

func TestSessionTxReturnsErrTxLaneBufferFullAtMaxBufferedLanesLimit(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               2,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	if _, err := tx.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("b")); err != nil {
		t.Fatal(err)
	}

	if got := len(tx.lanes); got != 2 {
		t.Fatalf("expected exactly 2 buffered lanes before limit, got %d", got)
	}

	if err := tx.SendMessage([]byte("c")); !errors.Is(err, ErrTxLaneBufferFull) {
		t.Fatalf("expected ErrTxLaneBufferFull when maxBufferedLanes is reached, got %v", err)
	}
	if got := len(tx.lanes); got != 2 {
		t.Fatalf("expected lane count to stay capped at 2 after limit error, got %d", got)
	}
}

func TestSessionTxCompletionSentinelStopsRepairImmediately(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:      1,
			TimeResendSecondaryRepairShard: 1,
		},
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected source and two repairs before completion sentinel under the K+1 target, got %d writes", len(writer.writes))
	}

	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: rrpitTransferLane.SeenChunksCompletionSentinel},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	lane := tx.lanes[0]
	if !lane.PeerReconstructed {
		t.Fatal("expected completion sentinel to mark peer reconstruction complete")
	}
	if lane.InitialRepairPacketsPending != 0 || lane.SecondaryRepairPacketsPending != 0 ||
		lane.SecondaryRepairPacketsPerBurst != 0 || lane.NextSecondaryRepairTimestamp != 0 {
		t.Fatalf("expected completion sentinel to clear repair state, got %+v", lane)
	}

	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected no further repair after completion sentinel, got %d writes", len(writer.writes))
	}
}

func TestSessionTxDeferredSecondaryRepairDoesNotRefreshTicketEarly(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
		Reconstruction: SessionTxReconstructionConfig{
			SecondaryRepairShardRatio:      1,
			TimeResendSecondaryRepairShard: 1,
		},
	})
	channelID, err := tx.AttachTxChannelWithConfig(writer, ChannelConfig{Weight: 1, MaxSendingSpeed: 1})
	if err != nil {
		t.Fatal(err)
	}
	tx.currentTimestampInitialized = true
	tx.currentTimestamp = 1

	if err := tx.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tx.AcceptRemoteControlMessage(ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: channelID},
		Lane: SessionLaneControlMessage{
			LaneACKTo:      -1,
			LenLaneControl: 1,
			LaneControl: []rrpitTransferLane.TransferControl{
				{SeenChunks: 0},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	lane := tx.lanes[0]
	if lane.NextSecondaryRepairTimestamp != 2 {
		t.Fatalf("expected resend ticket at timestamp 2, got %d", lane.NextSecondaryRepairTimestamp)
	}

	tx.currentTimestamp = 2
	tx.currentTimestampInitialized = true
	tx.txChannelsConfig[0].Status.TimestampLastSent = 2
	tx.txChannelsConfig[0].Status.PacketSentCurrentTimestamp = 1
	tx.txChannelsConfig[0].Status.EnforcedPacketSentCurrentTimestamp = 1
	tx.scheduleSecondaryRepairResends(2)

	if lane.SecondaryRepairPacketsPending != 2 {
		t.Fatalf("expected pending resend burst of 2 against the K+1 target, got %d", lane.SecondaryRepairPacketsPending)
	}
	if lane.NextSecondaryRepairTimestamp != 0 {
		t.Fatalf("expected active pending resend to clear the ticket, got %d", lane.NextSecondaryRepairTimestamp)
	}

	tx.currentTimestamp = 3
	tx.txChannelsConfig[0].Status.TimestampLastSent = 3
	tx.txChannelsConfig[0].Status.PacketSentCurrentTimestamp = 1
	tx.txChannelsConfig[0].Status.EnforcedPacketSentCurrentTimestamp = 1
	tx.scheduleSecondaryRepairResends(3)
	if lane.SecondaryRepairPacketsPending != 2 || lane.NextSecondaryRepairTimestamp != 0 {
		t.Fatalf("expected deferred resend to stay active without refreshing the ticket, got pending=%d next=%d", lane.SecondaryRepairPacketsPending, lane.NextSecondaryRepairTimestamp)
	}

	tx.txChannelsConfig[0].Status.PacketSentCurrentTimestamp = 0
	tx.txChannelsConfig[0].Status.EnforcedPacketSentCurrentTimestamp = 0
	if err := tx.OnNewTimestamp(3); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 2 {
		t.Fatalf("expected deferred repair to send once capacity returned, got %d writes", len(writer.writes))
	}
	if lane.SecondaryRepairPacketsPending != 1 || lane.NextSecondaryRepairTimestamp != 0 {
		t.Fatalf("expected one repair packet to remain pending after one rate-limited tick, got pending=%d next=%d", lane.SecondaryRepairPacketsPending, lane.NextSecondaryRepairTimestamp)
	}

	if err := tx.OnNewTimestamp(4); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 3 {
		t.Fatalf("expected second deferred repair to send on the next tick, got %d writes", len(writer.writes))
	}
	if lane.NextSecondaryRepairTimestamp != 5 {
		t.Fatalf("expected next resend ticket to refresh only after the pending burst drained, got %d", lane.NextSecondaryRepairTimestamp)
	}
}

func TestSessionTxFloodControlMessageToAllChannels(t *testing.T) {
	firstWriter := &recordingWriteCloser{}
	secondWriter := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		OddChannelIDs:                  true,
	})

	firstChannelID, err := tx.AttachTxChannel(firstWriter)
	if err != nil {
		t.Fatal(err)
	}
	secondChannelID, err := tx.AttachTxChannel(secondWriter)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.FloodControlMessageToAllChannels(); err != nil {
		t.Fatal(err)
	}
	if len(firstWriter.writes) != 1 || len(secondWriter.writes) != 1 {
		t.Fatalf("expected one control packet on each channel, got %d and %d", len(firstWriter.writes), len(secondWriter.writes))
	}

	_, firstPayload := splitMaterializedWire(t, firstWriter.writes[0])
	firstPacket := mustUnmarshalSessionControlPacket(t, firstPayload)
	if firstPacket.PacketKind != PacketKind_CONTROL {
		t.Fatalf("expected control packet kind, got %d", firstPacket.PacketKind)
	}
	if firstPacket.Control.FloodChannel.CurrentChannelID != firstChannelID {
		t.Fatalf("expected first flood channel id %d, got %d", firstChannelID, firstPacket.Control.FloodChannel.CurrentChannelID)
	}

	_, secondPayload := splitMaterializedWire(t, secondWriter.writes[0])
	secondPacket := mustUnmarshalSessionControlPacket(t, secondPayload)
	if secondPacket.Control.FloodChannel.CurrentChannelID != secondChannelID {
		t.Fatalf("expected second flood channel id %d, got %d", secondChannelID, secondPacket.Control.FloodChannel.CurrentChannelID)
	}
}

func TestSessionTxFloodControlMessagesOverridesCurrentChannelIDPerChannel(t *testing.T) {
	firstWriter := &recordingWriteCloser{}
	secondWriter := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
		OddChannelIDs:                  true,
	})

	firstChannelID, err := tx.AttachTxChannel(firstWriter)
	if err != nil {
		t.Fatal(err)
	}
	secondChannelID, err := tx.AttachTxChannel(secondWriter)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.FloodControlMessages(func(uint64) (ControlMessage, error) {
		return ControlMessage{
			FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: 999},
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	_, firstPayload := splitMaterializedWire(t, firstWriter.writes[0])
	firstPacket := mustUnmarshalSessionControlPacket(t, firstPayload)
	if firstPacket.Control.FloodChannel.CurrentChannelID != firstChannelID {
		t.Fatalf("expected first flooded control to carry channel id %d, got %d", firstChannelID, firstPacket.Control.FloodChannel.CurrentChannelID)
	}

	_, secondPayload := splitMaterializedWire(t, secondWriter.writes[0])
	secondPacket := mustUnmarshalSessionControlPacket(t, secondPayload)
	if secondPacket.Control.FloodChannel.CurrentChannelID != secondChannelID {
		t.Fatalf("expected second flooded control to carry channel id %d, got %d", secondChannelID, secondPacket.Control.FloodChannel.CurrentChannelID)
	}
}

func TestSessionRxRoundTripAndGenerateControl(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	var received [][]byte
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    16,
		MaxBufferedLanes: 4,
		OnMessage: func(data []byte) error {
			received = append(received, append([]byte(nil), data...))
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(channelID)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("beta")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(10); err != nil {
		t.Fatal(err)
	}

	for _, wire := range writer.writes {
		if err := channel.OnNewMessageArrived(wire); err != nil {
			t.Fatal(err)
		}
	}

	if diff := cmp.Diff([][]byte{[]byte("alpha"), []byte("beta")}, received); diff != "" {
		t.Fatalf("unexpected delivered payloads (-want +got):\n%s", diff)
	}

	ctrl, err := rx.GenerateControlMessage(channelID)
	if err != nil {
		t.Fatal(err)
	}
	if ctrl.FloodChannel.CurrentChannelID != channelID {
		t.Fatalf("expected flood channel id %d, got %d", channelID, ctrl.FloodChannel.CurrentChannelID)
	}
	if ctrl.Lane.LaneACKTo != 0 {
		t.Fatalf("expected lane ack to 0, got %d", ctrl.Lane.LaneACKTo)
	}
	if ctrl.Lane.LenLaneControl != 0 || len(ctrl.Lane.LaneControl) != 0 {
		t.Fatalf("expected no outstanding lane control, got len field %d and %d entries", ctrl.Lane.LenLaneControl, len(ctrl.Lane.LaneControl))
	}
	if ctrl.Channel.LenChannelControl != 1 || len(ctrl.Channel.ChannelControl) != 1 {
		t.Fatalf("expected one channel control, got len field %d and %d entries", ctrl.Channel.LenChannelControl, len(ctrl.Channel.ChannelControl))
	}
	if diff := cmp.Diff(rrpitTransferChannel.ChannelControlMessage{
		ChannelID:                  channelID,
		TotalPacketReceived:        3,
		LastSequenceNumberReceived: 2,
	}, ctrl.Channel.ChannelControl[0]); diff != "" {
		t.Fatalf("unexpected channel control (-want +got):\n%s", diff)
	}
}

func TestSessionRxDeliversExactFullLaneWithoutRepairWhenRemoteMaxConfigured(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	var received [][]byte
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:              16,
		MaxBufferedLanes:           4,
		RemoteMaxDataShardsPerLane: 2,
		OnMessage: func(data []byte) error {
			received = append(received, append([]byte(nil), data...))
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(channelID)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("beta")); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 2 {
		t.Fatalf("expected 2 source writes before repair, got %d", len(writer.writes))
	}

	for _, wire := range writer.writes {
		if err := channel.OnNewMessageArrived(wire); err != nil {
			t.Fatal(err)
		}
	}

	if diff := cmp.Diff([][]byte{[]byte("alpha"), []byte("beta")}, received); diff != "" {
		t.Fatalf("unexpected delivered payloads (-want +got):\n%s", diff)
	}
}

func TestSessionRxWaitsForRepairOnShortLaneEvenWhenRemoteMaxConfigured(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           2,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	var received [][]byte
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:              16,
		MaxBufferedLanes:           4,
		RemoteMaxDataShardsPerLane: 2,
		OnMessage: func(data []byte) error {
			received = append(received, append([]byte(nil), data...))
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(channelID)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("alpha")); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 1 {
		t.Fatalf("expected 1 source write, got %d", len(writer.writes))
	}
	if err := channel.OnNewMessageArrived(writer.writes[0]); err != nil {
		t.Fatal(err)
	}
	if len(received) != 0 {
		t.Fatalf("expected short lane to stay buffered before repair, got %d payloads", len(received))
	}

	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) < 2 {
		t.Fatalf("expected repair write after tick, got %d total writes", len(writer.writes))
	}
	if err := channel.OnNewMessageArrived(writer.writes[1]); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([][]byte{[]byte("alpha")}, received); diff != "" {
		t.Fatalf("unexpected delivered payloads after repair (-want +got):\n%s", diff)
	}
}

func TestSessionRxControlPacketLearnsChannelID(t *testing.T) {
	var seen []ControlMessage
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    16,
		MaxBufferedLanes: 4,
		OnMessage:        func([]byte) error { return nil },
		OnRemoteControlMsg: func(ctrl ControlMessage) error {
			seen = append(seen, ctrl)
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(0)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := marshalSessionControlPacket(PacketKind_CONTROL, ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: 9},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := channel.OnNewMessageArrived(materializedWire(0, payload)); err != nil {
		t.Fatal(err)
	}

	if channel.ChannelID != 9 {
		t.Fatalf("expected learned channel id 9, got %d", channel.ChannelID)
	}
	if len(seen) != 1 {
		t.Fatalf("expected one seen control message, got %d", len(seen))
	}
	if seen[0].FloodChannel.CurrentChannelID != 9 {
		t.Fatalf("expected seen flood channel id 9, got %d", seen[0].FloodChannel.CurrentChannelID)
	}
	if seen[0].Lane.LenLaneControl != 0 || len(seen[0].Lane.LaneControl) != 0 {
		t.Fatalf("expected empty lane control in seen control message, got len field %d and %d entries", seen[0].Lane.LenLaneControl, len(seen[0].Lane.LaneControl))
	}
	if seen[0].Channel.LenChannelControl != 0 || len(seen[0].Channel.ChannelControl) != 0 {
		t.Fatalf("expected empty channel control in seen control message, got len field %d and %d entries", seen[0].Channel.LenChannelControl, len(seen[0].Channel.ChannelControl))
	}

	ctrl, err := rx.GenerateControlMessage(9)
	if err != nil {
		t.Fatal(err)
	}
	if ctrl.Channel.LenChannelControl != 1 || len(ctrl.Channel.ChannelControl) != 1 {
		t.Fatalf("expected one channel control, got len field %d and %d entries", ctrl.Channel.LenChannelControl, len(ctrl.Channel.ChannelControl))
	}
	if diff := cmp.Diff(rrpitTransferChannel.ChannelControlMessage{
		ChannelID:                  9,
		TotalPacketReceived:        1,
		LastSequenceNumberReceived: 0,
	}, ctrl.Channel.ChannelControl[0]); diff != "" {
		t.Fatalf("unexpected channel control after flood learn (-want +got):\n%s", diff)
	}
}

func TestSessionRxRejectsDuplicateLearnedChannelIDs(t *testing.T) {
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    16,
		MaxBufferedLanes: 4,
		OnMessage:        func([]byte) error { return nil },
	})

	firstChannel, err := rx.AttachRxChannel(0)
	if err != nil {
		t.Fatal(err)
	}
	secondChannel, err := rx.AttachRxChannel(0)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := marshalSessionControlPacket(PacketKind_CONTROL, ControlMessage{
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: 7},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := firstChannel.OnNewMessageArrived(materializedWire(0, payload)); err != nil {
		t.Fatal(err)
	}
	if err := secondChannel.OnNewMessageArrived(materializedWire(0, payload)); err == nil {
		t.Fatal("expected duplicate learned rx channel id error")
	}
}

func TestSessionControlPacketRoundTripPreservesSessionInstanceID(t *testing.T) {
	want := SessionInstanceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	payload, err := marshalSessionControlPacket(PacketKind_CONTROL, ControlMessage{
		Session: SessionControlMessage{
			InstanceID: want,
		},
		FloodChannel: SessionFloodChannelControlMessage{CurrentChannelID: 9},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := mustUnmarshalSessionControlPacket(t, payload)
	if got.Control.Session.InstanceID != want {
		t.Fatalf("unexpected session instance id: got %x want %x", got.Control.Session.InstanceID, want)
	}
}

func TestSessionRxDeliversLanesInOrder(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  16,
		MaxDataShardsPerLane:           1,
		MaxBufferedLanes:               4,
		MaxRewindableTimestampNum:      4,
		MaxRewindableControlMessageNum: 4,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	var received [][]byte
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    16,
		MaxBufferedLanes: 4,
		OnMessage: func(data []byte) error {
			received = append(received, append([]byte(nil), data...))
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(channelID)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SendMessage([]byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendMessage([]byte("second")); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if err := tx.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) != 4 {
		t.Fatalf("expected 4 writes, got %d", len(writer.writes))
	}

	for _, wireIndex := range []int{1, 2} {
		if err := channel.OnNewMessageArrived(writer.writes[wireIndex]); err != nil {
			t.Fatal(err)
		}
	}
	if len(received) != 0 {
		t.Fatalf("expected later lane to stay buffered, got %d delivered payloads", len(received))
	}

	ctrl, err := rx.GenerateControlMessage(channelID)
	if err != nil {
		t.Fatal(err)
	}
	if ctrl.Lane.LaneACKTo != -1 {
		t.Fatalf("expected no lane ack yet, got %d", ctrl.Lane.LaneACKTo)
	}
	if ctrl.Lane.LenLaneControl != 2 || len(ctrl.Lane.LaneControl) != 2 {
		t.Fatalf("expected 2 lane controls, got len field %d and %d entries", ctrl.Lane.LenLaneControl, len(ctrl.Lane.LaneControl))
	}
	if ctrl.Lane.LaneControl[0].SeenChunks != 0 || ctrl.Lane.LaneControl[1].SeenChunks != rrpitTransferLane.SeenChunksCompletionSentinel {
		t.Fatalf("unexpected lane controls: %+v", ctrl.Lane.LaneControl)
	}
	ctrlAgain, err := rx.GenerateControlMessage(channelID)
	if err != nil {
		t.Fatal(err)
	}
	if ctrlAgain.Lane.LaneControl[1].SeenChunks != rrpitTransferLane.SeenChunksCompletionSentinel {
		t.Fatalf("expected completion sentinel to persist until ack advances, got %+v", ctrlAgain.Lane.LaneControl)
	}

	for _, wireIndex := range []int{0, 3} {
		if err := channel.OnNewMessageArrived(writer.writes[wireIndex]); err != nil {
			t.Fatal(err)
		}
	}
	if diff := cmp.Diff([][]byte{[]byte("first"), []byte("second")}, received); diff != "" {
		t.Fatalf("unexpected ordered payload delivery (-want +got):\n%s", diff)
	}
}

func TestSessionRxDoesNotFailWhenDecoderSuggestsRetry(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewSessionTx(t, SessionTxConfig{
		LaneShardSize:                  32,
		MaxDataShardsPerLane:           5,
		MaxBufferedLanes:               8,
		MaxRewindableTimestampNum:      8,
		MaxRewindableControlMessageNum: 8,
	})
	channelID, err := tx.AttachTxChannel(writer)
	if err != nil {
		t.Fatal(err)
	}

	payloads := [][]byte{
		[]byte("first"),
		[]byte("second"),
		[]byte("third"),
		[]byte("fourth"),
		[]byte("fifth"),
	}
	for _, payload := range payloads {
		if err := tx.SendMessage(payload); err != nil {
			t.Fatal(err)
		}
	}
	for ts := uint64(1); ts <= 12; ts++ {
		if err := tx.OnNewTimestamp(ts); err != nil {
			t.Fatal(err)
		}
	}

	var received [][]byte
	rx := mustNewSessionRx(t, SessionRxConfig{
		LaneShardSize:    32,
		MaxBufferedLanes: 8,
		OnMessage: func(data []byte) error {
			received = append(received, append([]byte(nil), data...))
			return nil
		},
	})
	channel, err := rx.AttachRxChannel(channelID)
	if err != nil {
		t.Fatal(err)
	}

	repairDelivered := 0
	for _, wire := range writer.writes {
		_, payload := splitMaterializedWire(t, wire)
		packet := mustUnmarshalSessionDataPacket(t, payload)
		if packet.LaneID != 0 {
			continue
		}
		if packet.Transfer.TotalDataShards == 0 && (packet.Transfer.Seq == 1 || packet.Transfer.Seq == 4) {
			continue
		}
		if packet.Transfer.TotalDataShards != 0 {
			repairDelivered++
		}

		if err := channel.OnNewMessageArrived(wire); err != nil {
			t.Fatal(err)
		}
		if len(received) == len(payloads) {
			break
		}
	}

	if repairDelivered == 0 {
		t.Fatal("expected to deliver repair packets")
	}
	if diff := cmp.Diff(payloads, received); diff != "" {
		t.Fatalf("unexpected payload delivery after repair retries (-want +got):\n%s", diff)
	}
}

type recordingWriteCloser struct {
	writes [][]byte
}

func (w *recordingWriteCloser) Write(p []byte) (int, error) {
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *recordingWriteCloser) Close() error {
	return nil
}

func splitMaterializedWire(t *testing.T, wire []byte) (uint64, []byte) {
	t.Helper()

	if len(wire) < materializedChannelSequenceFieldLength {
		t.Fatalf("wire message too short: %d", len(wire))
	}
	return binary.BigEndian.Uint64(wire[:materializedChannelSequenceFieldLength]), append([]byte(nil), wire[materializedChannelSequenceFieldLength:]...)
}

func materializedWire(seq uint64, payload []byte) []byte {
	wire := make([]byte, materializedChannelSequenceFieldLength+len(payload))
	binary.BigEndian.PutUint64(wire[:materializedChannelSequenceFieldLength], seq)
	copy(wire[materializedChannelSequenceFieldLength:], payload)
	return wire
}

func mustUnmarshalSessionDataPacket(t *testing.T, payload []byte) sessionDataPacket {
	t.Helper()

	var packet sessionDataPacket
	if err := struc.Unpack(bytes.NewReader(payload), &packet); err != nil {
		t.Fatal(err)
	}
	return packet
}

func mustUnmarshalSessionControlPacket(t *testing.T, payload []byte) sessionControlPacket {
	t.Helper()

	var packet sessionControlPacket
	if err := struc.Unpack(bytes.NewReader(payload), &packet); err != nil {
		t.Fatal(err)
	}
	return packet
}

func mustNewSessionTx(t *testing.T, config SessionTxConfig) *SessionTx {
	t.Helper()

	tx, err := NewSessionTx(config)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func mustNewSessionRx(t *testing.T, config SessionRxConfig) *SessionRx {
	t.Helper()

	rx, err := NewSessionRx(config)
	if err != nil {
		t.Fatal(err)
	}
	return rx
}

var _ io.WriteCloser = (*recordingWriteCloser)(nil)
