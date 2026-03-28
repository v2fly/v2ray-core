package rrpitTransferLane

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewTransferLaneConstructors(t *testing.T) {
	t.Run("rx validates shard size", func(t *testing.T) {
		if _, err := NewTransferLaneRx(reconstructionLengthFieldSize, 0); err == nil {
			t.Fatal("expected invalid shard size error")
		}

		rx, err := NewTransferLaneRx(16, 3)
		if err != nil {
			t.Fatal(err)
		}
		if rx.ShardSize != 16 {
			t.Fatalf("expected shard size 16, got %d", rx.ShardSize)
		}
		if rx.RemoteMaxDataShards != 3 {
			t.Fatalf("expected remote max data shards 3, got %d", rx.RemoteMaxDataShards)
		}
		if rx.seenDataShards == nil {
			t.Fatal("expected seenDataShards to be initialized")
		}
	})

	t.Run("tx validates config", func(t *testing.T) {
		if _, err := NewTransferLaneTx(reconstructionLengthFieldSize, 1); err == nil {
			t.Fatal("expected invalid shard size error")
		}
		if _, err := NewTransferLaneTx(16, -1); err == nil {
			t.Fatal("expected invalid max data shards error")
		}

		tx, err := NewTransferLaneTx(16, 3)
		if err != nil {
			t.Fatal(err)
		}
		if tx.ShardSize != 16 {
			t.Fatalf("expected shard size 16, got %d", tx.ShardSize)
		}
		if tx.MaxDataShards != 3 {
			t.Fatalf("expected max data shards 3, got %d", tx.MaxDataShards)
		}
		if cap(tx.seenShards) != 3 {
			t.Fatalf("expected preallocated seenShards capacity 3, got %d", cap(tx.seenShards))
		}
	})
}

func TestTransferLaneTxAddDataAndLimits(t *testing.T) {
	tx := mustNewTransferLaneTx(t, 8, 2)

	firstInput := []byte("hello")
	firstPacket := mustAddData(t, tx, firstInput)
	firstInput[0] = 'j'

	if firstPacket.Seq != 0 {
		t.Fatalf("expected first packet seq 0, got %d", firstPacket.Seq)
	}
	if firstPacket.TotalDataShards != 0 {
		t.Fatalf("expected data packet to omit total data shards, got %d", firstPacket.TotalDataShards)
	}
	if string(firstPacket.Data) != "hello" {
		t.Fatalf("expected packet data to be copied, got %q", string(firstPacket.Data))
	}

	secondPacket := mustAddData(t, tx, []byte("bye"))
	if secondPacket.Seq != 1 {
		t.Fatalf("expected second packet seq 1, got %d", secondPacket.Seq)
	}

	if _, err := tx.AddData([]byte("x")); err == nil {
		t.Fatal("expected max data shards reached error")
	}

	tx = mustNewTransferLaneTx(t, 8, 0)
	if _, err := tx.AddData(nil); err == nil {
		t.Fatal("expected invalid data length error for empty payload")
	}
	if _, err := tx.AddData([]byte("1234567")); err == nil {
		t.Fatal("expected data shard too large error")
	}

	tx = mustNewTransferLaneTx(t, 8, 0)
	mustAddData(t, tx, []byte("ok"))
	if _, err := tx.CreateReconstructionTransmissionData(); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.AddData([]byte("no")); err == nil {
		t.Fatal("expected finalized transfer lane error")
	}
}

func TestTransferLaneRxSourceShardDedupAndControl(t *testing.T) {
	rx := mustNewTransferLaneRx(t, 16)

	done, err := rx.AddTransferData(TransferData{
		Seq:          1,
		LengthOfData: uint16(len("beta")),
		Data:         []byte("beta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect completion before total shard count is known")
	}

	done, err = rx.AddTransferData(TransferData{
		Seq:          0,
		LengthOfData: uint16(len("alpha")),
		Data:         []byte("alpha"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect completion before total shard count is known")
	}

	control, err := rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != 2 {
		t.Fatalf("expected SeenChunks 2, got %d", control.SeenChunks)
	}

	done, err = rx.AddTransferData(TransferData{
		Seq:          1,
		LengthOfData: uint16(len("beta")),
		Data:         []byte("beta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect duplicate packet to report completion")
	}
	if rx.seenShardCount != 2 {
		t.Fatalf("expected duplicate packet to preserve seen shard count, got %d", rx.seenShardCount)
	}

	if _, err := rx.Reconstruct(); err == nil {
		t.Fatal("expected reconstruct to fail before total shard count is known")
	}

	if _, err := rx.AddTransferData(TransferData{
		Seq:          1,
		LengthOfData: uint16(len("BETA")),
		Data:         []byte("BETA"),
	}); err == nil {
		t.Fatal("expected conflicting duplicate packet to fail")
	}
}

func TestTransferLaneRxGenerateControlCapsAtMaxReportedValue(t *testing.T) {
	rx := &TransferLaneRx{
		ShardSize:      16,
		seenDataShards: make([]ReconstructionData, 0),
		seenShardCount: uint32(math.MaxUint16) + 1,
	}

	control, err := rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != maxReportedSeenChunks {
		t.Fatalf("expected SeenChunks %d, got %d", maxReportedSeenChunks, control.SeenChunks)
	}
}

func TestTransferLaneDirectReconstructWithAllSourceShards(t *testing.T) {
	payloads := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
	}

	tx := mustNewTransferLaneTx(t, 24, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	rx := mustNewTransferLaneRx(t, 24)

	for _, packet := range sourcePackets {
		done, err := rx.AddTransferData(packet)
		if err != nil {
			t.Fatal(err)
		}
		if done {
			t.Fatal("did not expect completion before repair packet announces total shards")
		}
	}

	repair := mustCreateRepairPacket(t, tx)
	if repair.TotalDataShards != uint32(len(payloads)) {
		t.Fatalf("expected TotalDataShards %d, got %d", len(payloads), repair.TotalDataShards)
	}
	if repair.Seq != uint32(len(payloads)) {
		t.Fatalf("expected first repair seq %d, got %d", len(payloads), repair.Seq)
	}
	if len(repair.Data) != tx.ShardSize {
		t.Fatalf("expected repair symbol size %d, got %d", tx.ShardSize, len(repair.Data))
	}

	done, err := rx.AddTransferData(repair)
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("expected all source shards plus repair announcement to complete lane")
	}

	want := reconstructionDataFromPayloads(payloads)
	got, err := rx.Reconstruct()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected reconstruction (-want +got):\n%s", diff)
	}

	got[0].Data[0] = 'z'
	gotAgain, err := rx.Reconstruct()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, gotAgain); diff != "" {
		t.Fatalf("expected reconstruct to return cloned data (-want +got):\n%s", diff)
	}
}

func TestTransferLaneDirectReconstructWithAllSourceShardsRemoteMaxConfigured(t *testing.T) {
	payloads := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
	}

	tx := mustNewTransferLaneTx(t, 24, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	rx := mustNewTransferLaneRxWithRemoteMax(t, 24, len(payloads))

	for i, packet := range sourcePackets {
		done, err := rx.AddTransferData(packet)
		if err != nil {
			t.Fatal(err)
		}
		if i < len(sourcePackets)-1 && done {
			t.Fatal("did not expect early completion before final source shard")
		}
		if i == len(sourcePackets)-1 && !done {
			t.Fatal("expected completion after final source shard when remote max is configured")
		}
	}

	got, err := rx.Reconstruct()
	if err != nil {
		t.Fatal(err)
	}
	want := reconstructionDataFromPayloads(payloads)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected reconstruction (-want +got):\n%s", diff)
	}
}

func TestTransferLaneDirectReconstructWithAllSourceShardsOutOfOrderRemoteMaxConfigured(t *testing.T) {
	payloads := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
	}

	tx := mustNewTransferLaneTx(t, 24, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	rx := mustNewTransferLaneRxWithRemoteMax(t, 24, len(payloads))

	for _, index := range []int{2, 0} {
		done, err := rx.AddTransferData(sourcePackets[index])
		if err != nil {
			t.Fatal(err)
		}
		if done {
			t.Fatal("did not expect completion before all source shards exist")
		}
	}

	done, err := rx.AddTransferData(sourcePackets[1])
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("expected completion once the full contiguous source shard set exists")
	}
}

func TestTransferLaneIgnoresLateRepairAfterRemoteMaxSourceCompletion(t *testing.T) {
	payloads := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
	}

	tx := mustNewTransferLaneTx(t, 24, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	repair := mustCreateRepairPacket(t, tx)
	rx := mustNewTransferLaneRxWithRemoteMax(t, 24, len(payloads))

	for _, packet := range sourcePackets {
		done, err := rx.AddTransferData(packet)
		if err != nil {
			t.Fatal(err)
		}
		if packet.Seq == uint32(len(payloads)-1) && !done {
			t.Fatal("expected source-only completion before repair arrives")
		}
	}

	done, err := rx.AddTransferData(repair)
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("expected completed lane to stay complete when late repair arrives")
	}

	got, err := rx.Reconstruct()
	if err != nil {
		t.Fatal(err)
	}
	want := reconstructionDataFromPayloads(payloads)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected reconstruction (-want +got):\n%s", diff)
	}
}

func TestTransferLaneReconstructWithRepairSymbolsAfterLoss(t *testing.T) {
	payloads := [][]byte{
		[]byte("first"),
		[]byte("second"),
		[]byte("third"),
		[]byte("fourth"),
		[]byte("fifth"),
	}

	tx := mustNewTransferLaneTx(t, 32, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	repairPackets := make([]TransferData, 0, 12)
	for i := 0; i < 12; i++ {
		repairPackets = append(repairPackets, mustCreateRepairPacket(t, tx))
	}

	rx := mustNewTransferLaneRx(t, 32)
	for i, packet := range sourcePackets {
		if i == 1 || i == 4 {
			continue
		}
		done, err := rx.AddTransferData(packet)
		if err != nil {
			t.Fatal(err)
		}
		if done {
			t.Fatal("did not expect completion from partial source data")
		}
	}

	want := reconstructionDataFromPayloads(payloads)
	var (
		got     []ReconstructionData
		lastErr error
	)
	for _, packet := range repairPackets {
		done, err := rx.AddTransferData(packet)
		if err != nil {
			t.Fatal(err)
		}
		if !done {
			continue
		}

		got, err = rx.Reconstruct()
		if err == nil {
			break
		}
		lastErr = err
	}

	if got == nil {
		t.Fatalf("failed to reconstruct after repair symbols: %v", lastErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected reconstruction (-want +got):\n%s", diff)
	}

	control, err := rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != SeenChunksCompletionSentinel {
		t.Fatalf("expected completion sentinel %d after reconstruction, got %d", SeenChunksCompletionSentinel, control.SeenChunks)
	}
}

func TestTransferLaneRxGenerateControlUsesCompletionSentinelAfterDirectReconstruct(t *testing.T) {
	rx := mustNewTransferLaneRx(t, 16)
	if _, err := rx.AddTransferData(TransferData{
		TotalDataShards: 1,
		Seq:             0,
		LengthOfData:    uint16(len("done")),
		Data:            []byte("done"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := rx.Reconstruct(); err != nil {
		t.Fatal(err)
	}

	control, err := rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != SeenChunksCompletionSentinel {
		t.Fatalf("expected completion sentinel %d, got %d", SeenChunksCompletionSentinel, control.SeenChunks)
	}
}

func TestTransferLaneRxDoesNotEarlyCompleteShortLaneWithRemoteMaxConfigured(t *testing.T) {
	tx := mustNewTransferLaneTx(t, 16, 4)
	packet := mustAddData(t, tx, []byte("done"))
	rx := mustNewTransferLaneRxWithRemoteMax(t, 16, 4)

	done, err := rx.AddTransferData(packet)
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect short lane to complete before total shard count is known")
	}
}

func TestTransferLaneRxGenerateControlCountsUniqueRepairSymbols(t *testing.T) {
	payloads := [][]byte{
		[]byte("first"),
		[]byte("second"),
		[]byte("third"),
	}

	tx := mustNewTransferLaneTx(t, 32, len(payloads))
	sourcePackets := addPayloads(t, tx, payloads)
	repair := mustCreateRepairPacket(t, tx)

	rx := mustNewTransferLaneRx(t, 32)

	done, err := rx.AddTransferData(sourcePackets[0])
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect completion from one source shard")
	}

	done, err = rx.AddTransferData(repair)
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect completion from one repair shard")
	}

	control, err := rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != 2 {
		t.Fatalf("expected SeenChunks 2 after one data + one repair, got %d", control.SeenChunks)
	}

	done, err = rx.AddTransferData(repair)
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("did not expect duplicate repair shard to complete lane")
	}

	control, err = rx.GenerateControl()
	if err != nil {
		t.Fatal(err)
	}
	if control.SeenChunks != 2 {
		t.Fatalf("expected duplicate repair shard not to change SeenChunks, got %d", control.SeenChunks)
	}
}

func TestTransferLaneTxAcceptControlData(t *testing.T) {
	tx := mustNewTransferLaneTx(t, 16, 2)
	addPayloads(t, tx, [][]byte{
		[]byte("one"),
		[]byte("two"),
	})

	firstRepair := mustCreateRepairPacket(t, tx)
	if firstRepair.Seq != 2 {
		t.Fatalf("expected first repair seq 2, got %d", firstRepair.Seq)
	}

	if err := tx.AcceptControlData(TransferControl{SeenChunks: 3}); err != nil {
		t.Fatal(err)
	}
	if tx.peerSeenChunks != 3 {
		t.Fatalf("expected sender to store peer seen chunks 3, got %d", tx.peerSeenChunks)
	}
	secondRepair := mustCreateRepairPacket(t, tx)
	if secondRepair.Seq != 3 {
		t.Fatalf("expected second repair seq 3 after oversized control report, got %d", secondRepair.Seq)
	}

	if err := tx.AcceptControlData(TransferControl{SeenChunks: 1}); err != nil {
		t.Fatal(err)
	}
	if tx.peerSeenChunks != 3 {
		t.Fatalf("expected peer seen chunk count to remain at max value 3, got %d", tx.peerSeenChunks)
	}

	if err := tx.AcceptControlData(TransferControl{SeenChunks: 2}); err != nil {
		t.Fatal(err)
	}
	if tx.peerSeenChunks != 3 {
		t.Fatalf("expected peer seen chunk count to remain at max value 3, got %d", tx.peerSeenChunks)
	}

	thirdRepair := mustCreateRepairPacket(t, tx)
	if thirdRepair.Seq != 4 {
		t.Fatalf("expected third repair seq 4 after control updates, got %d", thirdRepair.Seq)
	}

	tx2 := mustNewTransferLaneTx(t, 16, 2)
	if err := tx2.AcceptControlData(TransferControl{SeenChunks: 4}); err != nil {
		t.Fatal(err)
	}
	if tx2.peerSeenChunks != 4 {
		t.Fatalf("expected peer seen chunk count 4 before finalization, got %d", tx2.peerSeenChunks)
	}
}

func TestTransferLaneRxValidationErrors(t *testing.T) {
	t.Run("invalid transfer data length", func(t *testing.T) {
		rx := mustNewTransferLaneRx(t, 16)
		if _, err := rx.AddTransferData(TransferData{
			Seq:          0,
			LengthOfData: 4,
			Data:         []byte("abc"),
		}); err == nil {
			t.Fatal("expected invalid transfer data length error")
		}
	})

	t.Run("invalid repair symbol size", func(t *testing.T) {
		rx := mustNewTransferLaneRx(t, 16)
		if _, err := rx.AddTransferData(TransferData{
			TotalDataShards: 1,
			Seq:             1,
			LengthOfData:    3,
			Data:            []byte("abc"),
		}); err == nil {
			t.Fatal("expected invalid reconstruction symbol size error")
		}
	})

	t.Run("mismatched total shard count", func(t *testing.T) {
		rx := mustNewTransferLaneRx(t, 16)
		if _, err := rx.AddTransferData(TransferData{
			TotalDataShards: 2,
			Seq:             2,
			LengthOfData:    16,
			Data:            make([]byte, 16),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := rx.AddTransferData(TransferData{
			TotalDataShards: 3,
			Seq:             3,
			LengthOfData:    16,
			Data:            make([]byte, 16),
		}); err == nil {
			t.Fatal("expected mismatched total data shards error")
		}
	})
}

func mustNewTransferLaneRx(t *testing.T, shardSize int) *TransferLaneRx {
	return mustNewTransferLaneRxWithRemoteMax(t, shardSize, 0)
}

func mustNewTransferLaneRxWithRemoteMax(t *testing.T, shardSize int, remoteMaxDataShards int) *TransferLaneRx {
	t.Helper()

	rx, err := NewTransferLaneRx(shardSize, remoteMaxDataShards)
	if err != nil {
		t.Fatal(err)
	}
	return rx
}

func mustNewTransferLaneTx(t *testing.T, shardSize int, maxDataShards int) *TransferLaneTx {
	t.Helper()

	tx, err := NewTransferLaneTx(shardSize, maxDataShards)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func mustAddData(t *testing.T, tx *TransferLaneTx, payload []byte) TransferData {
	t.Helper()

	packet, err := tx.AddData(payload)
	if err != nil {
		t.Fatal(err)
	}
	return *packet
}

func addPayloads(t *testing.T, tx *TransferLaneTx, payloads [][]byte) []TransferData {
	t.Helper()

	packets := make([]TransferData, 0, len(payloads))
	for _, payload := range payloads {
		packets = append(packets, mustAddData(t, tx, payload))
	}
	return packets
}

func mustCreateRepairPacket(t *testing.T, tx *TransferLaneTx) TransferData {
	t.Helper()

	packet, err := tx.CreateReconstructionTransmissionData()
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

func reconstructionDataFromPayloads(payloads [][]byte) []ReconstructionData {
	reconstructed := make([]ReconstructionData, 0, len(payloads))
	for _, payload := range payloads {
		cp := append([]byte(nil), payload...)
		reconstructed = append(reconstructed, ReconstructionData{
			LengthOfData: uint16(len(cp)),
			Data:         cp,
		})
	}
	return reconstructed
}
