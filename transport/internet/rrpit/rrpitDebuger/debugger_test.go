package rrpitDebuger

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

func TestDiagnoseBidirectionalSessionCapturesPacketsAndState(t *testing.T) {
	logDir := t.TempDir()
	recorder, err := NewPacketRecorder(PacketRecorderConfig{
		Directory:        logDir,
		MaxFileSizeBytes: 128,
		MaxFiles:         16,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := recorder.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	var receivedA [][]byte
	var receivedB [][]byte

	a := mustNewDebugBidirectionalSession(t, rrpitBidirectionalSession.Config{
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
			MaxBufferedLanes:               4,
			MaxRewindableTimestampNum:      8,
			MaxRewindableControlMessageNum: 8,
			OddChannelIDs:                  true,
		},
	})
	b := mustNewDebugBidirectionalSession(t, rrpitBidirectionalSession.Config{
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
			MaxBufferedLanes:               4,
			MaxRewindableTimestampNum:      8,
			MaxRewindableControlMessageNum: 8,
		},
	})

	aToB := &debugWriteCloser{}
	bToA := &debugWriteCloser{}

	aTxChannelID, err := a.AttachTxChannel(recorder.WrapWriter("a", 0, aToB))
	if err != nil {
		t.Fatal(err)
	}
	aRxChannel, err := a.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	bTxChannelID, err := b.AttachTxChannel(recorder.WrapWriter("b", 0, bToA))
	if err != nil {
		t.Fatal(err)
	}
	bRxChannel, err := b.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}

	if err := a.SendMessage([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := a.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	pumpDebugTraffic(t, recorder, "b", 0, aToB, bRxChannel)

	if len(receivedB) != 1 || string(receivedB[0]) != "hello" {
		t.Fatalf("unexpected payloads delivered to B: %q", receivedB)
	}

	if err := b.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	pumpDebugTraffic(t, recorder, "a", 0, bToA, aRxChannel)

	output, err := DiagnoseBidirectionalSession(a, recorder)
	if err != nil {
		t.Fatal(err)
	}
	if output.Session.Tx == nil || output.Session.Rx == nil {
		t.Fatal("expected tx and rx snapshots to be present")
	}
	if output.PacketLog.Directory != logDir {
		t.Fatalf("expected packet log directory %q, got %q", logDir, output.PacketLog.Directory)
	}
	if len(output.PacketLog.Files) == 0 {
		t.Fatal("expected packet log manifest to include files")
	}
	if len(output.PacketLog.Files) < 2 {
		t.Fatalf("expected log rotation to create multiple files, got %d", len(output.PacketLog.Files))
	}
	if output.Session.Tx.Channels[0].ChannelID != aTxChannelID {
		t.Fatalf("expected tx channel id %d, got %d", aTxChannelID, output.Session.Tx.Channels[0].ChannelID)
	}
	if output.Session.Rx.Channels[0].ChannelID != bTxChannelID {
		t.Fatalf("expected rx channel to learn remote id %d, got %d", bTxChannelID, output.Session.Rx.Channels[0].ChannelID)
	}

	packets, err := recorder.ReadAllPackets()
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != output.PacketLog.TotalPacketsLogged {
		t.Fatalf("expected %d retained packets, got %d", output.PacketLog.TotalPacketsLogged, len(packets))
	}

	var sawTx bool
	var sawRx bool
	for _, packet := range packets {
		if packet.Direction == PacketDirectionTx {
			sawTx = true
		}
		if packet.Direction == PacketDirectionRx {
			sawRx = true
		}
	}
	if !sawTx || !sawRx {
		t.Fatalf("expected both tx and rx packet logs, got tx=%v rx=%v", sawTx, sawRx)
	}

	if !strings.Contains(output.String(), "\"packet_log\"") {
		t.Fatal("expected formatted output to include packet_log section")
	}
	if len(receivedA) != 0 {
		t.Fatalf("expected no payloads delivered to A in this flow, got %d", len(receivedA))
	}
}

func TestPacketRecorderRotatesAndDropsOldFiles(t *testing.T) {
	logDir := t.TempDir()
	recorder, err := NewPacketRecorder(PacketRecorderConfig{
		Directory:        logDir,
		MaxFileSizeBytes: 128,
		MaxFiles:         2,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := recorder.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for i := 0; i < 32; i++ {
		wire := make([]byte, 40)
		binary.BigEndian.PutUint64(wire[:8], uint64(i))
		for j := 8; j < len(wire); j++ {
			wire[j] = byte(i + j)
		}
		recorder.RecordInbound("peer", i%2, wire)
	}

	manifest := recorder.Manifest()
	if manifest.MaxFiles != 2 {
		t.Fatalf("expected max files 2, got %d", manifest.MaxFiles)
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("expected exactly 2 retained files, got %d", len(manifest.Files))
	}
	if manifest.DroppedLogFiles == 0 {
		t.Fatal("expected old log files to be dropped after rotation")
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected exactly 2 log files on disk, got %d", len(entries))
	}

	packets, err := recorder.ReadAllPackets()
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) == 0 {
		t.Fatal("expected retained packet logs to be readable")
	}
	if len(packets) >= manifest.TotalPacketsLogged {
		t.Fatalf("expected some packets to be trimmed after rotation, got %d retained of %d total", len(packets), manifest.TotalPacketsLogged)
	}
}

type debugWriteCloser struct {
	writes [][]byte
	readTo int
}

func (w *debugWriteCloser) Write(p []byte) (int, error) {
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *debugWriteCloser) Close() error {
	return nil
}

func pumpDebugTraffic(
	t *testing.T,
	recorder *PacketRecorder,
	peer string,
	channelIndex int,
	writer *debugWriteCloser,
	channel *rrpitMaterializedTransferChannel.ChannelRx,
) {
	t.Helper()

	for writer.readTo < len(writer.writes) {
		wire := writer.writes[writer.readTo]
		recorder.RecordInbound(peer, channelIndex, wire)
		if err := channel.OnNewMessageArrived(wire); err != nil {
			t.Fatal(err)
		}
		writer.readTo += 1
	}
}

func mustNewDebugBidirectionalSession(
	t *testing.T,
	config rrpitBidirectionalSession.Config,
) *rrpitBidirectionalSession.BidirectionalSession {
	t.Helper()

	session, err := rrpitBidirectionalSession.New(config)
	if err != nil {
		t.Fatal(err)
	}
	return session
}

var _ io.WriteCloser = (*debugWriteCloser)(nil)
