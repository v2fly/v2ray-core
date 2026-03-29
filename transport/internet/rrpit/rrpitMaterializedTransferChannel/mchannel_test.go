package rrpitMaterializedTransferChannel

import (
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
)

func TestNewChannelConstructors(t *testing.T) {
	if _, err := NewChannelRx(1, nil); err == nil {
		t.Fatal("expected nil callback error")
	}

	rx, err := NewChannelRx(2, func([]byte) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if rx.ChannelID != 2 {
		t.Fatalf("expected rx channel id 2, got %d", rx.ChannelID)
	}

	if _, err := NewChannelTx(3, nil, 1, 1); err == nil {
		t.Fatal("expected nil writer error")
	}
	if _, err := NewChannelTx(3, &recordingWriteCloser{}, -1, 1); err == nil {
		t.Fatal("expected invalid timestamp history size error")
	}

	tx, err := NewChannelTx(4, &recordingWriteCloser{}, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if tx.ChannelID != 4 {
		t.Fatalf("expected tx channel id 4, got %d", tx.ChannelID)
	}
}

func TestChannelTxSendDataMessageWritesWireFormatAndClose(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewChannelTx(t, 10, writer, 4, 4)

	payload := []byte("hello")
	if err := tx.SendDataMessage(payload); err != nil {
		t.Fatal(err)
	}
	payload[0] = 'j'
	if err := tx.SendDataMessage([]byte("world")); err != nil {
		t.Fatal(err)
	}

	if tx.NextSeq != 2 {
		t.Fatalf("expected next seq 2, got %d", tx.NextSeq)
	}
	if len(writer.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(writer.writes))
	}

	firstSeq, firstPayload := parseWireMessage(t, writer.writes[0])
	if firstSeq != 0 {
		t.Fatalf("expected first wire seq 0, got %d", firstSeq)
	}
	if string(firstPayload) != "hello" {
		t.Fatalf("expected first payload hello, got %q", string(firstPayload))
	}

	secondSeq, secondPayload := parseWireMessage(t, writer.writes[1])
	if secondSeq != 1 {
		t.Fatalf("expected second wire seq 1, got %d", secondSeq)
	}
	if string(secondPayload) != "world" {
		t.Fatalf("expected second payload world, got %q", string(secondPayload))
	}

	if err := tx.Close(); err != nil {
		t.Fatal(err)
	}
	if writer.closeCalls != 1 {
		t.Fatalf("expected one close call, got %d", writer.closeCalls)
	}
}

func TestChannelTxSendDataMessageWriteFailures(t *testing.T) {
	t.Run("writer error", func(t *testing.T) {
		writer := &recordingWriteCloser{writeErr: errors.New("write failed")}
		tx := mustNewChannelTx(t, 20, writer, 2, 2)

		if err := tx.SendDataMessage([]byte("payload")); err == nil || err.Error() != "write failed" {
			t.Fatalf("expected write failed error, got %v", err)
		}
		if tx.NextSeq != 0 {
			t.Fatalf("expected next seq to remain 0 after write failure, got %d", tx.NextSeq)
		}
	})

	t.Run("short write", func(t *testing.T) {
		writer := &recordingWriteCloser{shortWrite: 3}
		tx := mustNewChannelTx(t, 21, writer, 2, 2)

		if err := tx.SendDataMessage([]byte("payload")); !errors.Is(err, io.ErrShortWrite) {
			t.Fatalf("expected io.ErrShortWrite, got %v", err)
		}
		if tx.NextSeq != 0 {
			t.Fatalf("expected next seq to remain 0 after short write, got %d", tx.NextSeq)
		}
	})
}

func TestChannelRxOnNewMessageArrivedAndControl(t *testing.T) {
	var received [][]byte
	rx := mustNewChannelRx(t, 30, func(data []byte) error {
		received = append(received, append([]byte(nil), data...))
		return nil
	})

	if err := rx.OnNewMessageArrived(materializedMessage(2, []byte("gamma"))); err != nil {
		t.Fatal(err)
	}
	if err := rx.OnNewMessageArrived(materializedMessage(1, []byte("beta"))); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([][]byte{[]byte("gamma"), []byte("beta")}, received); diff != "" {
		t.Fatalf("unexpected received payloads (-want +got):\n%s", diff)
	}
	if rx.TotalPacketsReceived != 2 {
		t.Fatalf("expected total packets received 2, got %d", rx.TotalPacketsReceived)
	}
	if rx.LastPacketSeqReceived != 2 {
		t.Fatalf("expected last packet seq received 2, got %d", rx.LastPacketSeqReceived)
	}

	control, err := rx.CreateControlMessage()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(&rrpitTransferChannel.ChannelControlMessage{
		ChannelID:                  30,
		TotalPacketReceived:        2,
		LastSequenceNumberReceived: 2,
	}, control); diff != "" {
		t.Fatalf("unexpected control message (-want +got):\n%s", diff)
	}

	if err := rx.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rx.OnNewMessageArrived(materializedMessage(3, []byte("late"))); err == nil {
		t.Fatal("expected closed receiver error")
	}
}

func TestChannelRxOnNewMessageArrivedErrors(t *testing.T) {
	t.Run("short message", func(t *testing.T) {
		rx := mustNewChannelRx(t, 40, func([]byte) error { return nil })
		if err := rx.OnNewMessageArrived([]byte{1, 2, 3}); err == nil {
			t.Fatal("expected short message error")
		}
	})

	t.Run("callback error", func(t *testing.T) {
		callbackErr := errors.New("callback failed")
		rx := mustNewChannelRx(t, 41, func([]byte) error { return callbackErr })
		err := rx.OnNewMessageArrived(materializedMessage(5, []byte("boom")))
		if !errors.Is(err, callbackErr) {
			t.Fatalf("expected callback error, got %v", err)
		}
		if rx.TotalPacketsReceived != 1 || rx.LastPacketSeqReceived != 5 {
			t.Fatalf("expected receiver stats to reflect delivered frame, got total=%d last=%d", rx.TotalPacketsReceived, rx.LastPacketSeqReceived)
		}
	})
}

func TestMaterializedChannelRoundTrip(t *testing.T) {
	writer := &recordingWriteCloser{}
	tx := mustNewChannelTx(t, 50, writer, 4, 4)

	var received [][]byte
	rx := mustNewChannelRx(t, 50, func(data []byte) error {
		received = append(received, append([]byte(nil), data...))
		return nil
	})

	if err := tx.SendDataMessage([]byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := tx.SendDataMessage(nil); err != nil {
		t.Fatal(err)
	}

	for _, wire := range writer.writes {
		if err := rx.OnNewMessageArrived(wire); err != nil {
			t.Fatal(err)
		}
	}

	if diff := cmp.Diff([][]byte{[]byte("one"), []byte(nil)}, received); diff != "" {
		t.Fatalf("unexpected round-trip payloads (-want +got):\n%s", diff)
	}
	control, err := rx.CreateControlMessage()
	if err != nil {
		t.Fatal(err)
	}
	if control.TotalPacketReceived != 2 || control.LastSequenceNumberReceived != 1 {
		t.Fatalf("unexpected round-trip control message: %+v", control)
	}
}

type recordingWriteCloser struct {
	writes     [][]byte
	writeErr   error
	shortWrite int
	closeErr   error
	closeCalls int
}

func (w *recordingWriteCloser) Write(p []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	if w.shortWrite > 0 && w.shortWrite < len(p) {
		w.writes = append(w.writes, append([]byte(nil), p[:w.shortWrite]...))
		return w.shortWrite, nil
	}
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *recordingWriteCloser) Close() error {
	w.closeCalls++
	return w.closeErr
}

func materializedMessage(seq uint64, payload []byte) []byte {
	wire := make([]byte, channelSequenceFieldLength+len(payload))
	binary.BigEndian.PutUint64(wire[:channelSequenceFieldLength], seq)
	copy(wire[channelSequenceFieldLength:], payload)
	return wire
}

func parseWireMessage(t *testing.T, wire []byte) (uint64, []byte) {
	t.Helper()

	if len(wire) < channelSequenceFieldLength {
		t.Fatalf("wire message too short: %d", len(wire))
	}
	return binary.BigEndian.Uint64(wire[:channelSequenceFieldLength]), append([]byte(nil), wire[channelSequenceFieldLength:]...)
}

func mustNewChannelRx(t *testing.T, channelID uint64, onNewDataMessage func([]byte) error) *ChannelRx {
	t.Helper()

	rx, err := NewChannelRx(channelID, onNewDataMessage)
	if err != nil {
		t.Fatal(err)
	}
	return rx
}

func mustNewChannelTx(t *testing.T, channelID uint64, writer io.WriteCloser, maxTimestampHistory int, maxControlHistory int) *ChannelTx {
	t.Helper()

	tx, err := NewChannelTx(channelID, writer, maxTimestampHistory, maxControlHistory)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}
