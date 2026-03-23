package rrpitTransferChannel

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewChannelConstructors(t *testing.T) {
	rx := NewChannelRx(7)
	if rx.ChannelID != 7 {
		t.Fatalf("expected rx channel id 7, got %d", rx.ChannelID)
	}
	if rx.TotalPacketsReceived != 0 || rx.LastPacketSeqReceived != 0 {
		t.Fatal("expected rx counters to start at zero")
	}

	if _, err := NewChannelTx(9, -1, 1); err == nil {
		t.Fatal("expected invalid timestamp history size error")
	}
	if _, err := NewChannelTx(9, 1, -1); err == nil {
		t.Fatal("expected invalid control history size error")
	}

	tx, err := NewChannelTx(9, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	if tx.ChannelID != 9 {
		t.Fatalf("expected tx channel id 9, got %d", tx.ChannelID)
	}
	if tx.sentPacketHistory.Cap() != 3 {
		t.Fatalf("expected sent history capacity 3, got %d", tx.sentPacketHistory.Cap())
	}
	if tx.controlHistory.Cap() != 2 {
		t.Fatalf("expected control history capacity 2, got %d", tx.controlHistory.Cap())
	}
}

func TestChannelRxProcessMessageReceivedAndControl(t *testing.T) {
	rx := NewChannelRx(11)

	for _, msg := range []ChannelDataMessage{
		{ChannelSeq: 2, Data: []byte("c")},
		{ChannelSeq: 1, Data: []byte("b")},
		{ChannelSeq: 5, Data: []byte("f")},
	} {
		if err := rx.ProcessMessageReceived(msg); err != nil {
			t.Fatal(err)
		}
	}

	if rx.TotalPacketsReceived != 3 {
		t.Fatalf("expected total packets received 3, got %d", rx.TotalPacketsReceived)
	}
	if rx.LastPacketSeqReceived != 5 {
		t.Fatalf("expected last packet seq received 5, got %d", rx.LastPacketSeqReceived)
	}

	ctrl, err := rx.CreateControlMessage()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(&ChannelControlMessage{
		ChannelID:                  11,
		TotalPacketReceived:        3,
		LastSequenceNumberReceived: 5,
	}, ctrl); diff != "" {
		t.Fatalf("unexpected control message (-want +got):\n%s", diff)
	}
}

func TestChannelRxAssignChannelID(t *testing.T) {
	rx := NewChannelRx(0)
	if err := rx.AssignChannelID(0); err == nil {
		t.Fatal("expected invalid channel id error")
	}
	if err := rx.AssignChannelID(17); err != nil {
		t.Fatal(err)
	}
	if rx.ChannelID != 17 {
		t.Fatalf("expected learned channel id 17, got %d", rx.ChannelID)
	}
	if err := rx.AssignChannelID(17); err != nil {
		t.Fatal(err)
	}
	if err := rx.AssignChannelID(18); err == nil {
		t.Fatal("expected conflicting channel id assignment error")
	}
}

func TestChannelTxCreateDataMessageAndTimestampHistory(t *testing.T) {
	tx := mustNewChannelTx(t, 21, 2, 1)

	firstPayload := []byte("alpha")
	firstMessage := mustCreateDataMessage(t, tx, firstPayload, 100)
	firstPayload[0] = 'z'

	if firstMessage.ChannelSeq != 0 {
		t.Fatalf("expected first message seq 0, got %d", firstMessage.ChannelSeq)
	}
	if string(firstMessage.Data) != "alpha" {
		t.Fatalf("expected message payload copy, got %q", string(firstMessage.Data))
	}

	secondMessage := mustCreateDataMessage(t, tx, []byte("beta"), 200)
	thirdMessage := mustCreateDataMessage(t, tx, []byte("gamma"), 300)

	if secondMessage.ChannelSeq != 1 || thirdMessage.ChannelSeq != 2 {
		t.Fatalf("unexpected message sequences: second=%d third=%d", secondMessage.ChannelSeq, thirdMessage.ChannelSeq)
	}
	if tx.NextSeq != 3 {
		t.Fatalf("expected next seq to be 3, got %d", tx.NextSeq)
	}
	assertSentPacketHistory(t, tx.sentPacketHistory.Snapshot(), []sentPacketSnapshot{
		{seq: 1, timestamp: 200},
		{seq: 2, timestamp: 300},
	})
}

func assertSentPacketHistory(t *testing.T, got []sentPacketSnapshot, want []sentPacketSnapshot) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("unexpected sent packet history length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].seq != want[i].seq || got[i].timestamp != want[i].timestamp {
			t.Fatalf("unexpected sent packet history at %d: got %+v want %+v", i, got[i], want[i])
		}
	}
}

func TestChannelTxAcceptControlMessageAndRemoteLastSeenTimestamp(t *testing.T) {
	tx := mustNewChannelTx(t, 31, 4, 2)
	mustCreateDataMessage(t, tx, []byte("one"), 10)
	mustCreateDataMessage(t, tx, []byte("two"), 20)
	mustCreateDataMessage(t, tx, []byte("three"), 30)

	if _, err := tx.RemoteLastSeenMessageSenderTimestamp(); err == nil {
		t.Fatal("expected remote last seen timestamp query to fail without controls")
	}
	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  99,
		TotalPacketReceived:        1,
		LastSequenceNumberReceived: 0,
	}); err == nil {
		t.Fatal("expected channel id mismatch error")
	}

	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  31,
		TotalPacketReceived:        2,
		LastSequenceNumberReceived: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  31,
		TotalPacketReceived:        3,
		LastSequenceNumberReceived: 2,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  31,
		TotalPacketReceived:        1,
		LastSequenceNumberReceived: 0,
	}); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([]ChannelControlMessage{
		{
			ChannelID:                  31,
			TotalPacketReceived:        3,
			LastSequenceNumberReceived: 2,
		},
		{
			ChannelID:                  31,
			TotalPacketReceived:        1,
			LastSequenceNumberReceived: 0,
		},
	}, tx.controlHistory.Snapshot()); diff != "" {
		t.Fatalf("unexpected control history (-want +got):\n%s", diff)
	}

	timestamp, err := tx.RemoteLastSeenMessageSenderTimestamp()
	if err != nil {
		t.Fatal(err)
	}
	if timestamp != 30 {
		t.Fatalf("expected remote last seen timestamp 30, got %d", timestamp)
	}
}

func TestChannelTxHistoryUsesRingBufferOrdering(t *testing.T) {
	tx := mustNewChannelTx(t, 61, 2, 2)
	mustCreateDataMessage(t, tx, []byte("one"), 10)
	mustCreateDataMessage(t, tx, []byte("two"), 20)
	mustCreateDataMessage(t, tx, []byte("three"), 30)

	assertSentPacketHistory(t, tx.sentPacketHistory.Snapshot(), []sentPacketSnapshot{
		{seq: 1, timestamp: 20},
		{seq: 2, timestamp: 30},
	})

	for _, ctrl := range []ChannelControlMessage{
		{
			ChannelID:                  61,
			TotalPacketReceived:        1,
			LastSequenceNumberReceived: 0,
		},
		{
			ChannelID:                  61,
			TotalPacketReceived:        2,
			LastSequenceNumberReceived: 1,
		},
		{
			ChannelID:                  61,
			TotalPacketReceived:        3,
			LastSequenceNumberReceived: 2,
		},
	} {
		if err := tx.AcceptControlMessage(ctrl); err != nil {
			t.Fatal(err)
		}
	}

	if diff := cmp.Diff([]ChannelControlMessage{
		{
			ChannelID:                  61,
			TotalPacketReceived:        2,
			LastSequenceNumberReceived: 1,
		},
		{
			ChannelID:                  61,
			TotalPacketReceived:        3,
			LastSequenceNumberReceived: 2,
		},
	}, tx.controlHistory.Snapshot()); diff != "" {
		t.Fatalf("unexpected control ring contents (-want +got):\n%s", diff)
	}
}

func TestChannelTxRemotePacketLossSinceTimestamp(t *testing.T) {
	tx := mustNewChannelTx(t, 41, 8, 4)
	for i, timestamp := range []uint64{10, 20, 30, 40, 50} {
		payload := []byte{byte('a' + i)}
		mustCreateDataMessage(t, tx, payload, timestamp)
	}

	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  41,
		TotalPacketReceived:        2,
		LastSequenceNumberReceived: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.AcceptControlMessage(ChannelControlMessage{
		ChannelID:                  41,
		TotalPacketReceived:        4,
		LastSequenceNumberReceived: 4,
	}); err != nil {
		t.Fatal(err)
	}

	sent, lost, err := tx.RemotePacketLossSinceTimestamp(30)
	if err != nil {
		t.Fatal(err)
	}
	if sent != 3 || lost != 1 {
		t.Fatalf("expected sent=3 lost=1 since timestamp 30, got sent=%d lost=%d", sent, lost)
	}

	sent, lost, err = tx.RemotePacketLossSinceTimestamp(10)
	if err != nil {
		t.Fatal(err)
	}
	if sent != 5 || lost != 1 {
		t.Fatalf("expected sent=5 lost=1 since timestamp 10, got sent=%d lost=%d", sent, lost)
	}

	sent, lost, err = tx.RemotePacketLossSinceTimestamp(60)
	if err != nil {
		t.Fatal(err)
	}
	if sent != 0 || lost != 0 {
		t.Fatalf("expected empty window since timestamp 60, got sent=%d lost=%d", sent, lost)
	}
}

func TestChannelTxRemotePacketLossErrors(t *testing.T) {
	t.Run("no rewind history", func(t *testing.T) {
		tx := mustNewChannelTx(t, 51, 0, 2)
		mustCreateDataMessage(t, tx, []byte("one"), 10)

		if _, _, err := tx.RemotePacketLossSinceTimestamp(10); err == nil {
			t.Fatal("expected missing rewind history error")
		}
	})

	t.Run("no control history", func(t *testing.T) {
		tx := mustNewChannelTx(t, 52, 2, 2)
		mustCreateDataMessage(t, tx, []byte("one"), 10)

		if _, _, err := tx.RemotePacketLossSinceTimestamp(10); err == nil {
			t.Fatal("expected missing control history error")
		}
	})

	t.Run("timestamp outside rewind window", func(t *testing.T) {
		tx := mustNewChannelTx(t, 53, 2, 2)
		mustCreateDataMessage(t, tx, []byte("one"), 10)
		mustCreateDataMessage(t, tx, []byte("two"), 20)
		mustCreateDataMessage(t, tx, []byte("three"), 30)
		if err := tx.AcceptControlMessage(ChannelControlMessage{
			ChannelID:                  53,
			TotalPacketReceived:        3,
			LastSequenceNumberReceived: 2,
		}); err != nil {
			t.Fatal(err)
		}

		if _, _, err := tx.RemotePacketLossSinceTimestamp(5); err == nil {
			t.Fatal("expected rewind window error")
		}
	})
}

func mustNewChannelTx(t *testing.T, channelID uint64, maxTimestampHistory int, maxControlHistory int) *ChannelTx {
	t.Helper()

	tx, err := NewChannelTx(channelID, maxTimestampHistory, maxControlHistory)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func mustCreateDataMessage(t *testing.T, tx *ChannelTx, data []byte, timestamp uint64) *ChannelDataMessage {
	t.Helper()

	msg, err := tx.CreateDataMessage(data, timestamp)
	if err != nil {
		t.Fatal(err)
	}
	return msg
}
