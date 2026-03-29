package rrpitBidirectionalSession

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lunixbochs/struc"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

func TestBidirectionalSessionCoordinatesControlFeedback(t *testing.T) {
	aToB := &recordingWriteCloser{}
	bToA := &recordingWriteCloser{}

	var receivedA [][]byte
	var receivedB [][]byte

	a := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage: func(data []byte) error {
				receivedA = append(receivedA, append([]byte(nil), data...))
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               1,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
			OddChannelIDs:                  true,
		},
	})
	b := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage: func(data []byte) error {
				receivedB = append(receivedB, append([]byte(nil), data...))
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               1,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
		},
	})

	aTxChannelID, err := a.AttachTxChannel(aToB)
	if err != nil {
		t.Fatal(err)
	}
	aRxChannel, err := a.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	bTxChannelID, err := b.AttachTxChannel(bToA)
	if err != nil {
		t.Fatal(err)
	}
	bRxChannel, err := b.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	if aTxChannelID != 1 || bTxChannelID != 2 {
		t.Fatalf("expected tx channel ids 1 and 2, got %d and %d", aTxChannelID, bTxChannelID)
	}
	if aRxChannel.ChannelID != 0 || bRxChannel.ChannelID != 0 {
		t.Fatalf("expected rx channels to start with unknown id 0, got %d and %d", aRxChannel.ChannelID, bRxChannel.ChannelID)
	}

	if err := a.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := a.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	pumpWrites(t, aToB, bRxChannel)

	if diff := cmp.Diff([][]byte{[]byte("hello")}, receivedB); diff != "" {
		t.Fatalf("unexpected payloads received by B after first round (-want +got):\n%s", diff)
	}
	if bRxChannel.ChannelID != aTxChannelID {
		t.Fatalf("expected B rx channel to learn A tx channel id %d, got %d", aTxChannelID, bRxChannel.ChannelID)
	}

	if err := b.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	pumpWrites(t, bToA, aRxChannel)
	if aRxChannel.ChannelID != bTxChannelID {
		t.Fatalf("expected A rx channel to learn B tx channel id %d, got %d", bTxChannelID, aRxChannel.ChannelID)
	}

	if err := a.SendMessage([]byte("world")); err != nil {
		t.Fatalf("expected reverse control feedback to free A tx window, got %v", err)
	}
	if err := a.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	pumpWrites(t, aToB, bRxChannel)

	if diff := cmp.Diff([][]byte{[]byte("hello"), []byte("world")}, receivedB); diff != "" {
		t.Fatalf("unexpected payloads received by B after second round (-want +got):\n%s", diff)
	}
	if len(receivedA) != 0 {
		t.Fatalf("expected no payloads delivered to A in this one-way test, got %d", len(receivedA))
	}
}

func TestBidirectionalSessionAttachRxChannelLearnsRemoteIDs(t *testing.T) {
	session := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage:        func([]byte) error { return nil },
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               4,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
			OddChannelIDs:                  true,
		},
	})

	firstChannel, err := session.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	secondChannel, err := session.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	if firstChannel.ChannelID != 0 || secondChannel.ChannelID != 0 {
		t.Fatalf("expected rx channels to start with unknown ids, got %d and %d", firstChannel.ChannelID, secondChannel.ChannelID)
	}

	firstPayload, err := marshalControlPacket(rriptMonoDirectionSession.ControlMessage{
		FloodChannel: rriptMonoDirectionSession.SessionFloodChannelControlMessage{CurrentChannelID: 11},
	})
	if err != nil {
		t.Fatal(err)
	}
	secondPayload, err := marshalControlPacket(rriptMonoDirectionSession.ControlMessage{
		FloodChannel: rriptMonoDirectionSession.SessionFloodChannelControlMessage{CurrentChannelID: 13},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := firstChannel.OnNewMessageArrived(materializedWire(0, firstPayload)); err != nil {
		t.Fatal(err)
	}
	if err := secondChannel.OnNewMessageArrived(materializedWire(0, secondPayload)); err != nil {
		t.Fatal(err)
	}

	if firstChannel.ChannelID != 11 || secondChannel.ChannelID != 13 {
		t.Fatalf("expected learned rx channel ids 11 and 13, got %d and %d", firstChannel.ChannelID, secondChannel.ChannelID)
	}
}

func TestBidirectionalSessionControlPacketsCarryLocalSessionInstanceID(t *testing.T) {
	writer := &recordingWriteCloser{}
	want := rriptMonoDirectionSession.SessionInstanceID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	session := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage:        func([]byte) error { return nil },
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               4,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
			OddChannelIDs:                  true,
		},
		LocalSessionInstanceID: want,
	})

	if _, err := session.AttachTxChannel(writer); err != nil {
		t.Fatal(err)
	}
	if err := session.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if len(writer.writes) == 0 {
		t.Fatal("expected control packet write")
	}

	var packet sessionControlPacketForTest
	if err := struc.Unpack(bytes.NewReader(writer.writes[0][8:]), &packet); err != nil {
		t.Fatal(err)
	}
	if packet.Control.Session.InstanceID != want {
		t.Fatalf("unexpected session instance id: got %x want %x", packet.Control.Session.InstanceID, want)
	}
}

func TestBidirectionalSessionAutoTicks(t *testing.T) {
	aToB := &recordingWriteCloser{}
	bToA := &recordingWriteCloser{}

	var receivedB [][]byte

	a := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage:        func([]byte) error { return nil },
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               1,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
			OddChannelIDs:                  true,
		},
		TimestampInterval: time.Millisecond,
	})
	defer func() {
		if err := a.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	b := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage: func(data []byte) error {
				receivedB = append(receivedB, append([]byte(nil), data...))
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               1,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
		},
		TimestampInterval: time.Millisecond,
	})
	defer func() {
		if err := b.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := a.AttachTxChannel(aToB); err != nil {
		t.Fatal(err)
	}
	aRxChannel, err := a.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.AttachTxChannel(bToA); err != nil {
		t.Fatal(err)
	}
	bRxChannel, err := b.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}

	if err := a.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	pumpWrites(t, aToB, bRxChannel)

	deadline := time.Now().Add(250 * time.Millisecond)
	secondResult := make(chan error, 1)
	go func() {
		secondResult <- a.SendMessage([]byte("world"))
	}()
	secondSent := false
	for time.Now().Before(deadline) {
		pumpWrites(t, bToA, aRxChannel)
		pumpWrites(t, aToB, bRxChannel)

		if !secondSent {
			select {
			case err := <-secondResult:
				if err != nil {
					t.Fatalf("expected blocked SendMessage to succeed once feedback arrives, got %v", err)
				}
				secondSent = true
			default:
			}
		}
		if secondSent && len(receivedB) == 2 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}

	if !secondSent {
		t.Fatal("expected background ticks to release tx backpressure")
	}
	if diff := cmp.Diff([][]byte{[]byte("hello"), []byte("world")}, receivedB); diff != "" {
		t.Fatalf("unexpected payloads received by B (-want +got):\n%s", diff)
	}
}

func TestBidirectionalSessionSuppressesDuplicateManagerHostedControl(t *testing.T) {
	var sent [][]byte

	session := mustNewBidirectionalSession(t, Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    16,
			MaxBufferedLanes: 4,
			OnMessage:        func([]byte) error { return nil },
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  16,
			MaxDataShardsPerLane:           1,
			MaxBufferedLanes:               1,
			MaxRewindableTimestampNum:      4,
			MaxRewindableControlMessageNum: 4,
			DataPacketKind:                 rriptMonoDirectionSession.PacketKind_InteractiveStreamData,
			ControlPacketKind:              rriptMonoDirectionSession.PacketKind_InteractiveStreamControl,
			SendIgnoreQuota: func(_ uint8, payload []byte) error {
				sent = append(sent, append([]byte(nil), payload...))
				return nil
			},
		},
		ManagerHostedControlKeepaliveIntervalTicks: 2,
	})

	firstStats, err := session.OnNewTimestampWithStats(1)
	if err != nil {
		t.Fatal(err)
	}
	if firstStats.ControlPacketsSent != 1 || len(sent) != 1 {
		t.Fatalf("expected first manager-hosted tick to send one control packet, stats=%+v sends=%d", firstStats, len(sent))
	}

	secondStats, err := session.OnNewTimestampWithStats(2)
	if err != nil {
		t.Fatal(err)
	}
	if secondStats.ControlPacketsSent != 0 || len(sent) != 1 {
		t.Fatalf("expected duplicate control to be suppressed, stats=%+v sends=%d", secondStats, len(sent))
	}

	thirdStats, err := session.OnNewTimestampWithStats(3)
	if err != nil {
		t.Fatal(err)
	}
	if thirdStats.ControlPacketsSent != 1 || len(sent) != 2 {
		t.Fatalf("expected keepalive control resend after suppression interval, stats=%+v sends=%d", thirdStats, len(sent))
	}
}

type recordingWriteCloser struct {
	writes [][]byte
	readTo int
}

func (w *recordingWriteCloser) Write(p []byte) (int, error) {
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *recordingWriteCloser) Close() error {
	return nil
}

func pumpWrites(t *testing.T, writer *recordingWriteCloser, channel *rrpitMaterializedTransferChannel.ChannelRx) {
	t.Helper()

	for writer.readTo < len(writer.writes) {
		if err := channel.OnNewMessageArrived(writer.writes[writer.readTo]); err != nil {
			t.Fatal(err)
		}
		writer.readTo += 1
	}
}

func materializedWire(seq uint64, payload []byte) []byte {
	wire := make([]byte, 8+len(payload))
	wire[0] = byte(seq >> 56)
	wire[1] = byte(seq >> 48)
	wire[2] = byte(seq >> 40)
	wire[3] = byte(seq >> 32)
	wire[4] = byte(seq >> 24)
	wire[5] = byte(seq >> 16)
	wire[6] = byte(seq >> 8)
	wire[7] = byte(seq)
	copy(wire[8:], payload)
	return wire
}

func marshalControlPacket(ctrl rriptMonoDirectionSession.ControlMessage) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if err := struc.Pack(buffer, &sessionControlPacketForTest{
		PacketKind: rriptMonoDirectionSession.PacketKind_CONTROL,
		Control:    ctrl,
	}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func mustNewBidirectionalSession(t *testing.T, config Config) *BidirectionalSession {
	t.Helper()

	session, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

type sessionControlPacketForTest struct {
	PacketKind uint8
	Control    rriptMonoDirectionSession.ControlMessage
}

var _ io.WriteCloser = (*recordingWriteCloser)(nil)
