package rrpitTransferLane

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"bytes"
	stderrors "errors"
	"math"

	"github.com/lunixbochs/struc"
	"github.com/xssnick/raptorq"

	commonerrors "github.com/v2fly/v2ray-core/v5/common/errors"
)

const reconstructionLengthFieldSize = 2

var ErrNotEnoughSymbolsToReconstruct = stderrors.New("not enough symbols to reconstruct")

const (
	SeenChunksCompletionSentinel = math.MaxUint16
	maxReportedSeenChunks        = SeenChunksCompletionSentinel - 1
)

func IsNotEnoughSymbolsToReconstruct(err error) bool {
	return commonerrors.Cause(err) == ErrNotEnoughSymbolsToReconstruct
}

type TransferLaneRx struct {
	ShardSize           int
	RemoteMaxDataShards int

	TotalDataShards  uint32
	rxCodesState     *raptorq.Decoder
	seenDataShards   []ReconstructionData
	seenRepairShards map[uint32]struct{}
	seenShardCount   uint32
	completed        bool
}

func NewTransferLaneRx(shardSize int, remoteMaxDataShards int) (*TransferLaneRx, error) {
	if err := validateTransferLaneConfig(shardSize); err != nil {
		return nil, err
	}
	if remoteMaxDataShards < 0 {
		return nil, newError("invalid remote max data shards")
	}
	return &TransferLaneRx{
		ShardSize:           shardSize,
		RemoteMaxDataShards: remoteMaxDataShards,
		seenDataShards:      make([]ReconstructionData, 0),
		seenRepairShards:    make(map[uint32]struct{}),
	}, nil
}

func (lr *TransferLaneRx) AddTransferData(data TransferData) (done bool, err error) {
	if err := validateTransferLaneConfig(lr.ShardSize); err != nil {
		return false, err
	}
	if data.LengthOfData != uint16(len(data.Data)) {
		return false, newError("invalid transfer data length")
	}
	if data.TotalDataShards != 0 {
		if err := lr.ensureDecoder(data.TotalDataShards); err != nil {
			return false, err
		}
	}
	if lr.TotalDataShards != 0 && data.Seq >= lr.TotalDataShards {
		if len(data.Data) != lr.ShardSize {
			return false, newError("invalid reconstruction symbol size")
		}
		if lr.rxCodesState == nil {
			if hasAllShards(lr.seenDataShards, lr.TotalDataShards) {
				return true, nil
			}
			return false, newError("reconstruction decoder not initialized")
		}
		if lr.seenRepairShards == nil {
			lr.seenRepairShards = make(map[uint32]struct{})
		}
		if _, found := lr.seenRepairShards[data.Seq]; found {
			return lr.seenShardCount >= lr.TotalDataShards, nil
		}
		canTryDecode, err := lr.rxCodesState.AddSymbol(data.Seq, data.Data)
		if err != nil {
			return false, newError("failed to add reconstruction symbol: ", err)
		}
		lr.seenRepairShards[data.Seq] = struct{}{}
		lr.seenShardCount += 1
		return lr.seenShardCount >= lr.TotalDataShards || canTryDecode, nil
	}

	shard := ReconstructionData{
		LengthOfData: data.LengthOfData,
		Data:         append([]byte(nil), data.Data...),
	}
	if err := validateReconstructionData(lr.ShardSize, shard); err != nil {
		return false, err
	}
	if err := lr.storeDataShard(data.Seq, shard); err != nil {
		return false, err
	}
	if lr.TotalDataShards == 0 && lr.RemoteMaxDataShards > 0 && hasAllShards(lr.seenDataShards, uint32(lr.RemoteMaxDataShards)) {
		lr.TotalDataShards = uint32(lr.RemoteMaxDataShards)
		return true, nil
	}
	if lr.rxCodesState != nil {
		symbol, err := encodeReconstructionData(lr.ShardSize, shard)
		if err != nil {
			return false, err
		}
		canTryDecode, err := lr.rxCodesState.AddSymbol(data.Seq, symbol)
		if err != nil {
			return false, newError("failed to add data shard symbol: ", err)
		}
		return lr.seenShardCount == lr.TotalDataShards || canTryDecode, nil
	}

	return lr.TotalDataShards != 0 && lr.seenShardCount == lr.TotalDataShards, nil
}

func (lr *TransferLaneRx) GenerateControl() (TransferControl, error) {
	if lr.completed {
		return TransferControl{SeenChunks: SeenChunksCompletionSentinel}, nil
	}
	seenChunks := lr.seenShardCount
	if seenChunks > uint32(maxReportedSeenChunks) {
		seenChunks = uint32(maxReportedSeenChunks)
	}
	return TransferControl{SeenChunks: uint16(seenChunks)}, nil
}

func (lr *TransferLaneRx) Reconstruct() ([]ReconstructionData, error) {
	if lr.TotalDataShards == 0 {
		return nil, newError("total data shards unknown")
	}
	if hasAllShards(lr.seenDataShards, lr.TotalDataShards) {
		lr.completed = true
		return cloneReconstructionDataSlice(lr.seenDataShards[:int(lr.TotalDataShards)]), nil
	}
	if lr.rxCodesState == nil {
		return nil, newError("reconstruction decoder not initialized")
	}

	ok, payload, err := lr.rxCodesState.Decode()
	if err != nil {
		return nil, newError("failed to decode reconstruction data: ", err)
	}
	if !ok {
		return nil, newError("not enough symbols to reconstruct").Base(ErrNotEnoughSymbolsToReconstruct)
	}

	reconstructed, err := decodeReconstructionPayload(lr.ShardSize, lr.TotalDataShards, payload)
	if err != nil {
		return nil, err
	}
	lr.seenDataShards = reconstructed
	lr.seenShardCount = lr.TotalDataShards
	lr.completed = true
	return cloneReconstructionDataSlice(reconstructed), nil
}

type TransferLaneTx struct {
	ShardSize     int
	MaxDataShards int

	TotalDataShards uint32
	txCodesState    *raptorq.Encoder
	seenShards      []ReconstructionData
	peerSeenChunks  uint16
	nextSeq         uint32
}

func NewTransferLaneTx(shardSize int, maxDataShards int) (*TransferLaneTx, error) {
	if err := validateTransferLaneConfig(shardSize); err != nil {
		return nil, err
	}
	if maxDataShards < 0 {
		return nil, newError("invalid max data shards")
	}

	lane := &TransferLaneTx{
		ShardSize:     shardSize,
		MaxDataShards: maxDataShards,
	}
	if maxDataShards > 0 {
		lane.seenShards = make([]ReconstructionData, 0, maxDataShards)
	}
	return lane, nil
}

func (lr *TransferLaneTx) AddData(data []byte) (transferData *TransferData, err error) {
	if err := validateTransferLaneConfig(lr.ShardSize); err != nil {
		return nil, err
	}
	if lr.TotalDataShards != 0 {
		return nil, newError("finalized transfer lane")
	}
	if lr.MaxDataShards > 0 && lr.nextSeq >= uint32(lr.MaxDataShards) {
		return nil, newError("max data shards reached")
	}
	if len(data) == 0 || len(data) > int(^uint16(0)) {
		return nil, newError("invalid data length")
	}
	if len(data)+reconstructionLengthFieldSize > lr.ShardSize {
		return nil, newError("data shard too large for configured shard size")
	}
	copied := append([]byte(nil), data...)
	lr.seenShards = append(lr.seenShards, ReconstructionData{
		Data:         copied,
		LengthOfData: uint16(len(copied)),
	})
	thisSeq := lr.nextSeq
	lr.nextSeq += 1
	return &TransferData{Seq: thisSeq, Data: copied, LengthOfData: uint16(len(copied))}, nil
}

func (lr *TransferLaneTx) CreateReconstructionTransmissionData() (TransferData, error) {
	if err := validateTransferLaneConfig(lr.ShardSize); err != nil {
		return TransferData{}, err
	}
	if lr.nextSeq == 0 {
		return TransferData{}, newError("no data shards available for reconstruction")
	}
	if lr.txCodesState == nil {
		totalDataShards := lr.TotalDataShards
		if totalDataShards == 0 {
			totalDataShards = lr.nextSeq
		}
		payloadSize, err := reconstructionPayloadSize(totalDataShards, lr.ShardSize)
		if err != nil {
			return TransferData{}, err
		}
		bufferBuilder := bytes.NewBuffer(make([]byte, 0, payloadSize))
		for _, shard := range lr.seenShards {
			symbol, err := encodeReconstructionData(lr.ShardSize, shard)
			if err != nil {
				return TransferData{}, err
			}
			bufferBuilder.Write(symbol)
		}
		rpQ := raptorq.NewRaptorQ(uint32(lr.ShardSize))
		encoder, err := rpQ.CreateEncoder(bufferBuilder.Bytes())
		if err != nil {
			return TransferData{}, newError("failed to create reconstruction encoder: ", err)
		}
		if encoder.BaseSymbolsNum() != totalDataShards {
			return TransferData{}, newError("unexpected number of base symbols")
		}
		lr.TotalDataShards = totalDataShards
		lr.txCodesState = encoder
		if lr.nextSeq < lr.TotalDataShards {
			lr.nextSeq = lr.TotalDataShards
		}
	}

	seq := lr.nextSeq
	symbol := lr.txCodesState.GenSymbol(seq)
	lr.nextSeq += 1
	return TransferData{
		TotalDataShards: lr.TotalDataShards,
		Seq:             seq,
		LengthOfData:    uint16(len(symbol)),
		Data:            symbol,
	}, nil
}

func (lr *TransferLaneTx) AcceptControlData(control TransferControl) error {
	if control.SeenChunks > lr.peerSeenChunks {
		lr.peerSeenChunks = control.SeenChunks
	}
	return nil
}

func (lr *TransferLaneRx) ensureDecoder(totalDataShards uint32) error {
	if err := validateTransferLaneConfig(lr.ShardSize); err != nil {
		return err
	}
	if totalDataShards == 0 {
		return newError("invalid total data shards")
	}
	if lr.TotalDataShards != 0 {
		if lr.TotalDataShards != totalDataShards {
			return newError("mismatched total data shards")
		}
		return nil
	}
	if exceedsMaxInt(totalDataShards) {
		return newError("total data shards exceed supported slice size")
	}
	if uint32(len(lr.seenDataShards)) > totalDataShards {
		return newError("received source shard exceeds announced total data shards")
	}
	payloadSize, err := reconstructionPayloadSize(totalDataShards, lr.ShardSize)
	if err != nil {
		return err
	}
	decoder, err := raptorq.NewRaptorQ(uint32(lr.ShardSize)).CreateDecoder(uint32(payloadSize))
	if err != nil {
		return newError("failed to create reconstruction decoder: ", err)
	}
	lr.TotalDataShards = totalDataShards
	lr.rxCodesState = decoder
	for i, shard := range lr.seenDataShards {
		if shard.Data == nil {
			continue
		}
		symbol, err := encodeReconstructionData(lr.ShardSize, shard)
		if err != nil {
			return err
		}
		if _, err := lr.rxCodesState.AddSymbol(uint32(i), symbol); err != nil {
			return newError("failed to add cached data shard symbol: ", err)
		}
	}
	return nil
}

func (lr *TransferLaneRx) storeDataShard(seq uint32, shard ReconstructionData) error {
	if lr.TotalDataShards != 0 && seq >= lr.TotalDataShards {
		return newError("source shard sequence out of range")
	}
	if exceedsMaxInt(seq) {
		return newError("source shard sequence exceeds supported slice size")
	}

	index := int(seq)
	if index >= len(lr.seenDataShards) {
		extended := make([]ReconstructionData, index+1)
		copy(extended, lr.seenDataShards)
		lr.seenDataShards = extended
	}
	if existing := lr.seenDataShards[index]; existing.Data != nil {
		if existing.LengthOfData != shard.LengthOfData || !bytes.Equal(existing.Data, shard.Data) {
			return newError("conflicting source shard data")
		}
		return nil
	}

	lr.seenDataShards[index] = shard
	lr.seenShardCount += 1
	return nil
}

func validateTransferLaneConfig(shardSize int) error {
	if shardSize <= reconstructionLengthFieldSize {
		return newError("invalid shard size")
	}
	if shardSize > int(^uint16(0)) {
		return newError("shard size exceeds transfer packet capacity")
	}
	return nil
}

func validateReconstructionData(shardSize int, shard ReconstructionData) error {
	if shard.LengthOfData != uint16(len(shard.Data)) {
		return newError("invalid reconstruction data length")
	}
	if shard.LengthOfData == 0 {
		return newError("invalid reconstruction data length")
	}
	if int(shard.LengthOfData)+reconstructionLengthFieldSize > shardSize {
		return newError("reconstruction data shard too large")
	}
	return nil
}

func reconstructionPayloadSize(totalDataShards uint32, shardSize int) (int, error) {
	if totalDataShards == 0 {
		return 0, newError("invalid total data shards")
	}
	if err := validateTransferLaneConfig(shardSize); err != nil {
		return 0, err
	}
	payloadSize := uint64(totalDataShards) * uint64(shardSize)
	if payloadSize > uint64(^uint32(0)) || payloadSize > uint64(int(^uint(0)>>1)) {
		return 0, newError("reconstruction payload too large")
	}
	return int(payloadSize), nil
}

func encodeReconstructionData(shardSize int, shard ReconstructionData) ([]byte, error) {
	if err := validateReconstructionData(shardSize, shard); err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(make([]byte, 0, shardSize))
	if err := struc.Pack(buffer, &shard); err != nil {
		return nil, newError("failed to pack reconstruction data: ", err)
	}
	if buffer.Len() > shardSize {
		return nil, newError("packed reconstruction data exceeds shard size")
	}
	symbol := make([]byte, shardSize)
	copy(symbol, buffer.Bytes())
	return symbol, nil
}

func decodeReconstructionPayload(shardSize int, totalDataShards uint32, payload []byte) ([]ReconstructionData, error) {
	expectedSize, err := reconstructionPayloadSize(totalDataShards, shardSize)
	if err != nil {
		return nil, err
	}
	if len(payload) != expectedSize {
		return nil, newError("invalid reconstructed payload size")
	}

	reconstructed := make([]ReconstructionData, 0, totalDataShards)
	for offset := 0; offset < len(payload); offset += shardSize {
		var shard ReconstructionData
		if err := struc.Unpack(bytes.NewReader(payload[offset:offset+shardSize]), &shard); err != nil {
			return nil, newError("failed to unpack reconstructed shard: ", err)
		}
		if err := validateReconstructionData(shardSize, shard); err != nil {
			return nil, err
		}
		reconstructed = append(reconstructed, shard)
	}
	return reconstructed, nil
}

func hasAllShards(shards []ReconstructionData, totalDataShards uint32) bool {
	if uint32(len(shards)) < totalDataShards {
		return false
	}
	for i := 0; i < int(totalDataShards); i++ {
		if shards[i].Data == nil {
			return false
		}
	}
	return true
}

func cloneReconstructionDataSlice(shards []ReconstructionData) []ReconstructionData {
	cloned := make([]ReconstructionData, len(shards))
	for i, shard := range shards {
		cloned[i] = ReconstructionData{
			LengthOfData: shard.LengthOfData,
			Data:         append([]byte(nil), shard.Data...),
		}
	}
	return cloned
}

func exceedsMaxInt(v uint32) bool {
	return uint64(v) > uint64(^uint(0)>>1)
}
