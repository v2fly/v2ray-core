package rrpitDebuger

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/lunixbochs/struc"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferLane"
)

const (
	materializedChannelSequenceFieldLength = 8
	defaultPacketLogBaseName               = "rrpit-packets"
	defaultPacketLogMaxFileSizeBytes       = 1 << 20
	defaultPacketLogMaxFiles               = 8
	packetLogScannerBufferSize             = 8 << 20
)

type PacketDirection string

const (
	PacketDirectionTx PacketDirection = "tx"
	PacketDirectionRx PacketDirection = "rx"
)

type DiagnoseOutput struct {
	GeneratedAt time.Time                    `json:"generated_at"`
	Session     BidirectionalSessionSnapshot `json:"session"`
	PacketLog   PacketLogManifest            `json:"packet_log"`
}

type BidirectionalSessionSnapshot struct {
	TimestampInterval string             `json:"timestamp_interval"`
	Tx                *SessionTxSnapshot `json:"tx,omitempty"`
	Rx                *SessionRxSnapshot `json:"rx,omitempty"`
}

type SessionTxSnapshot struct {
	OddChannelIDs        bool                `json:"odd_channel_ids"`
	CurrentTimestamp     uint64              `json:"current_timestamp"`
	NextChannelID        uint64              `json:"next_channel_id"`
	LaneShardSize        int                 `json:"lane_shard_size"`
	MaxDataShardsPerLane int                 `json:"max_data_shards_per_lane"`
	MaxBufferedLanes     int                 `json:"max_buffered_lanes"`
	FirstLaneID          int64               `json:"first_lane_id"`
	Lanes                []TxLaneSnapshot    `json:"lanes"`
	Channels             []TxChannelSnapshot `json:"channels"`
}

type TxLaneSnapshot struct {
	LaneID                        uint64                 `json:"lane_id"`
	DataShards                    uint32                 `json:"data_shards"`
	TotalDataShards               uint32                 `json:"total_data_shards"`
	Finalized                     bool                   `json:"finalized"`
	PeerSeenChunks                uint16                 `json:"peer_seen_chunks"`
	PeerReconstructed             bool                   `json:"peer_reconstructed"`
	RepairPackets                 uint32                 `json:"repair_packets"`
	InitialRepairPacketsPending   uint32                 `json:"initial_repair_packets_pending"`
	SecondaryRepairPacketsPending uint32                 `json:"secondary_repair_packets_pending"`
	NextSecondaryRepairTimestamp  uint64                 `json:"next_secondary_repair_timestamp"`
	CreatedAtTimestamp            uint64                 `json:"created_at_timestamp"`
	FinalizedAtTimestamp          uint64                 `json:"finalized_at_timestamp"`
	LastProgressTimestamp         uint64                 `json:"last_progress_timestamp"`
	TransferLane                  TransferLaneTxSnapshot `json:"transfer_lane"`
}

type TransferLaneTxSnapshot struct {
	ShardSize       int    `json:"shard_size"`
	MaxDataShards   int    `json:"max_data_shards"`
	TotalDataShards uint32 `json:"total_data_shards"`
	SeenShardCount  int    `json:"seen_shard_count"`
	PeerSeenChunks  uint16 `json:"peer_seen_chunks"`
	NextSeq         uint32 `json:"next_seq"`
	HasEncoder      bool   `json:"has_encoder"`
}

type TxChannelSnapshot struct {
	ChannelID                          uint64                                             `json:"channel_id"`
	NextSeq                            uint64                                             `json:"next_seq"`
	WriterType                         string                                             `json:"writer_type,omitempty"`
	Config                             rriptMonoDirectionSession.ChannelConfig            `json:"config"`
	Status                             rriptMonoDirectionSession.ChannelRateControlStatus `json:"status"`
	SentPacketHistory                  []SentPacketSnapshot                               `json:"sent_packet_history"`
	ControlHistory                     []rrpitTransferChannel.ChannelControlMessage       `json:"control_history"`
	RemoteLastSeenSenderTimestamp      *uint64                                            `json:"remote_last_seen_sender_timestamp,omitempty"`
	RemoteLastSeenSenderTimestampError string                                             `json:"remote_last_seen_sender_timestamp_error,omitempty"`
}

type SentPacketSnapshot struct {
	Seq       uint64 `json:"seq"`
	Timestamp uint64 `json:"timestamp"`
}

type SessionRxSnapshot struct {
	LaneShardSize    int                 `json:"lane_shard_size"`
	MaxBufferedLanes int                 `json:"max_buffered_lanes"`
	FirstLaneID      int64               `json:"first_lane_id"`
	Lanes            []RxLaneSnapshot    `json:"lanes"`
	Channels         []RxChannelSnapshot `json:"channels"`
}

type RxLaneSnapshot struct {
	LaneID             uint64                 `json:"lane_id"`
	Ready              bool                   `json:"ready"`
	NextDeliverIndex   int                    `json:"next_deliver_index"`
	ReconstructedCount int                    `json:"reconstructed_count"`
	ReconstructedLens  []int                  `json:"reconstructed_lengths"`
	TransferLane       TransferLaneRxSnapshot `json:"transfer_lane"`
}

type TransferLaneRxSnapshot struct {
	ShardSize          int    `json:"shard_size"`
	TotalDataShards    uint32 `json:"total_data_shards"`
	SeenShardCount     uint32 `json:"seen_shard_count"`
	SeenDataShardCount int    `json:"seen_data_shard_count"`
	HasDecoder         bool   `json:"has_decoder"`
	Completed          bool   `json:"completed"`
}

type RxChannelSnapshot struct {
	ChannelID             uint64 `json:"channel_id"`
	TotalPacketsReceived  uint64 `json:"total_packets_received"`
	LastPacketSeqReceived uint64 `json:"last_packet_seq_received"`
}

type PacketLog struct {
	Index          int                                       `json:"index"`
	ObservedAt     time.Time                                 `json:"observed_at"`
	Peer           string                                    `json:"peer"`
	Direction      PacketDirection                           `json:"direction"`
	ChannelIndex   int                                       `json:"channel_index"`
	WireSequence   uint64                                    `json:"wire_sequence"`
	WireSize       int                                       `json:"wire_size"`
	PayloadSize    int                                       `json:"payload_size"`
	PacketKind     uint8                                     `json:"packet_kind"`
	PacketKindName string                                    `json:"packet_kind_name"`
	LaneID         *uint64                                   `json:"lane_id,omitempty"`
	Transfer       *rrpitTransferLane.TransferData           `json:"transfer,omitempty"`
	Control        *rriptMonoDirectionSession.ControlMessage `json:"control,omitempty"`
	RawWireHex     string                                    `json:"raw_wire_hex"`
	RawPayloadHex  string                                    `json:"raw_payload_hex,omitempty"`
	DecodeError    string                                    `json:"decode_error,omitempty"`
	TransportError string                                    `json:"transport_error,omitempty"`
}

type PacketRecorderConfig struct {
	Directory        string
	BaseName         string
	MaxFileSizeBytes int64
	MaxFiles         int
}

type PacketLogManifest struct {
	Directory          string   `json:"directory"`
	Files              []string `json:"files"`
	MaxFileSizeBytes   int64    `json:"max_file_size_bytes"`
	MaxFiles           int      `json:"max_files"`
	TotalPacketsLogged int      `json:"total_packets_logged"`
	DroppedLogFiles    int      `json:"dropped_log_files"`
	WriteErrors        []string `json:"write_errors,omitempty"`
}

type PacketRecorder struct {
	mu               sync.Mutex
	config           PacketRecorderConfig
	runID            string
	nextIdx          int
	currentFileIndex int
	currentFileSize  int64
	currentFile      *os.File
	files            []string
	droppedLogFiles  int
	writeErrors      []string
}

type debugSessionDataPacket struct {
	PacketKind uint8
	LaneID     uint64
	Transfer   rrpitTransferLane.TransferData
}

type debugSessionControlPacket struct {
	PacketKind uint8
	Control    rriptMonoDirectionSession.ControlMessage
}

type packetRecordingWriteCloser struct {
	recorder     *PacketRecorder
	peer         string
	channelIndex int
	io.WriteCloser
}

type FailureLogger interface {
	Helper()
	Failed() bool
	Logf(format string, args ...any)
}

func NewPacketRecorder(config PacketRecorderConfig) (*PacketRecorder, error) {
	if config.Directory == "" {
		return nil, fmt.Errorf("packet recorder directory is required")
	}
	if config.BaseName == "" {
		config.BaseName = defaultPacketLogBaseName
	}
	if config.MaxFileSizeBytes <= 0 {
		config.MaxFileSizeBytes = defaultPacketLogMaxFileSizeBytes
	}
	if config.MaxFiles <= 0 {
		config.MaxFiles = defaultPacketLogMaxFiles
	}

	absDirectory, err := filepath.Abs(config.Directory)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDirectory, 0o755); err != nil {
		return nil, err
	}
	config.Directory = absDirectory

	recorder := &PacketRecorder{
		config: config,
		runID:  time.Now().UTC().Format("20060102T150405.000000000Z07"),
	}
	if err := recorder.rotateLocked(); err != nil {
		return nil, err
	}
	return recorder, nil
}

func (r *PacketRecorder) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closeCurrentFileLocked()
}

func (r *PacketRecorder) WrapWriter(peer string, channelIndex int, writer io.WriteCloser) io.WriteCloser {
	if r == nil || writer == nil {
		return writer
	}
	return &packetRecordingWriteCloser{
		recorder:     r,
		peer:         peer,
		channelIndex: channelIndex,
		WriteCloser:  writer,
	}
}

func (r *PacketRecorder) RecordInbound(peer string, channelIndex int, wire []byte) {
	if r == nil {
		return
	}
	r.recordPacket(peer, PacketDirectionRx, channelIndex, wire, nil)
}

func (r *PacketRecorder) Manifest() PacketLogManifest {
	if r == nil {
		return PacketLogManifest{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	manifest := PacketLogManifest{
		Directory:          r.config.Directory,
		Files:              append([]string(nil), r.files...),
		MaxFileSizeBytes:   r.config.MaxFileSizeBytes,
		MaxFiles:           r.config.MaxFiles,
		TotalPacketsLogged: r.nextIdx,
		DroppedLogFiles:    r.droppedLogFiles,
	}
	if len(r.writeErrors) > 0 {
		manifest.WriteErrors = append([]string(nil), r.writeErrors...)
	}
	return manifest
}

func (r *PacketRecorder) ReadAllPackets() ([]PacketLog, error) {
	if r == nil {
		return nil, nil
	}

	manifest := r.Manifest()
	packets := make([]PacketLog, 0, manifest.TotalPacketsLogged)
	for _, path := range manifest.Files {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), packetLogScannerBufferSize)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber += 1
			var packet PacketLog
			if err := json.Unmarshal(scanner.Bytes(), &packet); err != nil {
				_ = file.Close()
				return nil, fmt.Errorf("decode packet log %s line %d: %w", path, lineNumber, err)
			}
			packets = append(packets, packet)
		}
		if err := scanner.Err(); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("read packet log %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}
	return packets, nil
}

func DiagnoseBidirectionalSession(
	session *rrpitBidirectionalSession.BidirectionalSession,
	recorder *PacketRecorder,
) (*DiagnoseOutput, error) {
	if session == nil {
		return nil, fmt.Errorf("nil bidirectional session")
	}

	output := &DiagnoseOutput{
		GeneratedAt: time.Now().UTC(),
		Session: BidirectionalSessionSnapshot{
			TimestampInterval: session.TimestampInterval.String(),
		},
	}

	output.Session.Tx = snapshotSessionTx(session.Tx())
	output.Session.Rx = snapshotSessionRx(session.Rx())

	if recorder != nil {
		output.PacketLog = recorder.Manifest()
	}
	return output, nil
}

func (o *DiagnoseOutput) MarshalIndented() ([]byte, error) {
	if o == nil {
		return []byte("null"), nil
	}
	return json.MarshalIndent(o, "", "  ")
}

func (o *DiagnoseOutput) String() string {
	data, err := o.MarshalIndented()
	if err != nil {
		return fmt.Sprintf("failed to marshal diagnose output: %v", err)
	}
	return string(data)
}

func WriteDiagnoseOutput(
	writer io.Writer,
	session *rrpitBidirectionalSession.BidirectionalSession,
	recorder *PacketRecorder,
) error {
	if writer == nil {
		return fmt.Errorf("nil writer")
	}
	output, err := DiagnoseBidirectionalSession(session, recorder)
	if err != nil {
		return err
	}
	data, err := output.MarshalIndented()
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func LogBidirectionalSessionOnFailure(
	logger FailureLogger,
	label string,
	session *rrpitBidirectionalSession.BidirectionalSession,
	recorder *PacketRecorder,
) {
	if logger == nil {
		return
	}
	logger.Helper()
	if !logger.Failed() {
		return
	}

	output, err := DiagnoseBidirectionalSession(session, recorder)
	if err != nil {
		logger.Logf("%s: failed to capture rrpit diagnose output: %v", label, err)
		return
	}
	logger.Logf("%s:\n%s", label, output.String())
}

func (w *packetRecordingWriteCloser) Write(p []byte) (int, error) {
	written, err := w.WriteCloser.Write(p)
	if written > 0 {
		var transportErr error
		switch {
		case err != nil:
			transportErr = err
		case written != len(p):
			transportErr = io.ErrShortWrite
		}
		w.recorder.recordPacket(w.peer, PacketDirectionTx, w.channelIndex, p[:written], transportErr)
	}
	return written, err
}

func (r *PacketRecorder) recordPacket(
	peer string,
	direction PacketDirection,
	channelIndex int,
	wire []byte,
	transportErr error,
) {
	clonedWire := append([]byte(nil), wire...)
	packet := PacketLog{
		ObservedAt:   time.Now().UTC(),
		Peer:         peer,
		Direction:    direction,
		ChannelIndex: channelIndex,
		WireSize:     len(clonedWire),
		RawWireHex:   hex.EncodeToString(clonedWire),
	}
	if transportErr != nil {
		packet.TransportError = transportErr.Error()
	}

	if len(clonedWire) < materializedChannelSequenceFieldLength {
		packet.DecodeError = "materialized channel message too short"
		r.writePacket(packet)
		return
	}

	packet.WireSequence = binary.BigEndian.Uint64(clonedWire[:materializedChannelSequenceFieldLength])
	payload := append([]byte(nil), clonedWire[materializedChannelSequenceFieldLength:]...)
	packet.PayloadSize = len(payload)
	packet.RawPayloadHex = hex.EncodeToString(payload)
	if len(payload) == 0 {
		packet.DecodeError = "empty session payload"
		r.writePacket(packet)
		return
	}

	packet.PacketKind = payload[0]
	packet.PacketKindName = packetKindName(payload[0])
	switch payload[0] {
	case rriptMonoDirectionSession.PacketKind_DATA:
		var decoded debugSessionDataPacket
		if err := struc.Unpack(bytes.NewReader(payload), &decoded); err != nil {
			packet.DecodeError = err.Error()
			break
		}
		laneID := decoded.LaneID
		packet.LaneID = &laneID
		transfer := decoded.Transfer
		transfer.Data = append([]byte(nil), transfer.Data...)
		packet.Transfer = &transfer
	case rriptMonoDirectionSession.PacketKind_CONTROL:
		var decoded debugSessionControlPacket
		if err := struc.Unpack(bytes.NewReader(payload), &decoded); err != nil {
			packet.DecodeError = err.Error()
			break
		}
		packet.Control = cloneControlMessage(decoded.Control)
	default:
		packet.DecodeError = fmt.Sprintf("unknown session packet kind %d", payload[0])
	}

	r.writePacket(packet)
}

func (r *PacketRecorder) writePacket(packet PacketLog) {
	r.mu.Lock()
	defer r.mu.Unlock()

	packet.Index = r.nextIdx
	r.nextIdx += 1

	if r.currentFile == nil {
		if err := r.rotateLocked(); err != nil {
			r.writeErrors = append(r.writeErrors, err.Error())
			return
		}
	}

	line, err := json.Marshal(packet)
	if err != nil {
		r.writeErrors = append(r.writeErrors, err.Error())
		return
	}
	line = append(line, '\n')

	if r.currentFileSize > 0 && r.currentFileSize+int64(len(line)) > r.config.MaxFileSizeBytes {
		if err := r.rotateLocked(); err != nil {
			r.writeErrors = append(r.writeErrors, err.Error())
			return
		}
	}

	written, err := r.currentFile.Write(line)
	r.currentFileSize += int64(written)
	if err != nil {
		r.writeErrors = append(r.writeErrors, err.Error())
		return
	}
	if written != len(line) {
		r.writeErrors = append(r.writeErrors, io.ErrShortWrite.Error())
	}
}

func (r *PacketRecorder) rotateLocked() error {
	if r == nil {
		return nil
	}
	if err := r.closeCurrentFileLocked(); err != nil {
		return err
	}

	filePath := filepath.Join(
		r.config.Directory,
		fmt.Sprintf("%s-%s-%06d.jsonl", r.config.BaseName, r.runID, r.currentFileIndex),
	)
	r.currentFileIndex += 1

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	r.currentFile = file
	r.currentFileSize = 0
	r.files = append(r.files, filePath)
	for len(r.files) > r.config.MaxFiles {
		oldest := r.files[0]
		r.files = r.files[1:]
		if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
			r.writeErrors = append(r.writeErrors, err.Error())
			continue
		}
		r.droppedLogFiles += 1
	}
	return nil
}

func (r *PacketRecorder) closeCurrentFileLocked() error {
	if r == nil || r.currentFile == nil {
		return nil
	}

	err := r.currentFile.Close()
	r.currentFile = nil
	r.currentFileSize = 0
	return err
}

func clonePacketLog(packet PacketLog) PacketLog {
	cloned := packet
	if packet.Transfer != nil {
		transfer := *packet.Transfer
		transfer.Data = append([]byte(nil), transfer.Data...)
		cloned.Transfer = &transfer
	}
	if packet.Control != nil {
		cloned.Control = cloneControlMessage(*packet.Control)
	}
	if packet.LaneID != nil {
		laneID := *packet.LaneID
		cloned.LaneID = &laneID
	}
	return cloned
}

func cloneControlMessage(ctrl rriptMonoDirectionSession.ControlMessage) *rriptMonoDirectionSession.ControlMessage {
	cloned := ctrl
	cloned.Lane.LaneControl = append([]rrpitTransferLane.TransferControl(nil), ctrl.Lane.LaneControl...)
	cloned.Channel.ChannelControl = append([]rrpitTransferChannel.ChannelControlMessage(nil), ctrl.Channel.ChannelControl...)
	return &cloned
}

func packetKindName(kind uint8) string {
	return rriptMonoDirectionSession.PacketKindName(kind)
}

func snapshotSessionTx(tx *rriptMonoDirectionSession.SessionTx) *SessionTxSnapshot {
	if tx == nil {
		return nil
	}

	root := forceValue(reflect.ValueOf(tx).Elem())
	txLanes := forceValue(root.FieldByName("txLanes"))
	txChannels := forceValue(root.FieldByName("txChannels"))
	channelsConfig := forceValue(root.FieldByName("txChannelsConfig"))

	snapshot := &SessionTxSnapshot{
		OddChannelIDs:        root.FieldByName("OddChannelIDs").Bool(),
		CurrentTimestamp:     root.FieldByName("currentTimestamp").Uint(),
		NextChannelID:        txChannels.FieldByName("nextChannelID").Uint(),
		LaneShardSize:        int(txLanes.FieldByName("laneShardSize").Int()),
		MaxDataShardsPerLane: int(txLanes.FieldByName("maxDataShardsPerLane").Int()),
		MaxBufferedLanes:     int(txLanes.FieldByName("maxBufferedLanes").Int()),
		FirstLaneID:          txLanes.FieldByName("firstLaneID").Int(),
		Lanes:                make([]TxLaneSnapshot, 0, forceValue(txLanes.FieldByName("lanes")).Len()),
		Channels:             make([]TxChannelSnapshot, 0, channelsConfig.Len()),
	}

	lanes := forceValue(txLanes.FieldByName("lanes"))
	for i := 0; i < lanes.Len(); i++ {
		laneValue := forceValue(lanes.Index(i))
		if laneValue.IsNil() {
			continue
		}
		snapshot.Lanes = append(snapshot.Lanes, snapshotTxLane(laneValue))
	}

	for i := 0; i < channelsConfig.Len(); i++ {
		snapshot.Channels = append(snapshot.Channels, snapshotTxChannel(forceValue(channelsConfig.Index(i))))
	}

	return snapshot
}

func snapshotSessionRx(rx *rriptMonoDirectionSession.SessionRx) *SessionRxSnapshot {
	if rx == nil {
		return nil
	}

	root := forceValue(reflect.ValueOf(rx).Elem())
	rxLanes := forceValue(root.FieldByName("rxLanes"))
	rxChannels := forceValue(root.FieldByName("rxChannels"))

	snapshot := &SessionRxSnapshot{
		LaneShardSize:    int(rxLanes.FieldByName("laneShardSize").Int()),
		MaxBufferedLanes: int(rxLanes.FieldByName("maxBufferedLanes").Int()),
		FirstLaneID:      rxLanes.FieldByName("firstLaneID").Int(),
		Lanes:            make([]RxLaneSnapshot, 0, forceValue(rxLanes.FieldByName("lanes")).Len()),
		Channels:         make([]RxChannelSnapshot, 0, forceValue(rxChannels.FieldByName("channels")).Len()),
	}

	lanes := forceValue(rxLanes.FieldByName("lanes"))
	for i := 0; i < lanes.Len(); i++ {
		laneValue := forceValue(lanes.Index(i))
		if laneValue.IsNil() {
			continue
		}
		snapshot.Lanes = append(snapshot.Lanes, snapshotRxLane(laneValue))
	}

	channels := forceValue(rxChannels.FieldByName("channels"))
	for i := 0; i < channels.Len(); i++ {
		channelValue := forceValue(channels.Index(i))
		if channelValue.IsNil() {
			continue
		}
		snapshot.Channels = append(snapshot.Channels, snapshotRxChannel(channelValue))
	}

	return snapshot
}

func snapshotTxLane(lanePtr reflect.Value) TxLaneSnapshot {
	lane := forceValue(lanePtr.Elem())
	transferLane := forceValue(lane.FieldByName("TransferLane"))

	snapshot := TxLaneSnapshot{
		LaneID:                        lane.FieldByName("LaneID").Uint(),
		DataShards:                    uint32(lane.FieldByName("DataShards").Uint()),
		TotalDataShards:               uint32(lane.FieldByName("TotalDataShards").Uint()),
		Finalized:                     lane.FieldByName("Finalized").Bool(),
		PeerSeenChunks:                uint16(lane.FieldByName("PeerSeenChunks").Uint()),
		PeerReconstructed:             lane.FieldByName("PeerReconstructed").Bool(),
		RepairPackets:                 uint32(lane.FieldByName("RepairPackets").Uint()),
		InitialRepairPacketsPending:   uint32(lane.FieldByName("InitialRepairPacketsPending").Uint()),
		SecondaryRepairPacketsPending: uint32(lane.FieldByName("SecondaryRepairPacketsPending").Uint()),
		NextSecondaryRepairTimestamp:  lane.FieldByName("NextSecondaryRepairTimestamp").Uint(),
		CreatedAtTimestamp:            lane.FieldByName("CreatedAtTimestamp").Uint(),
		FinalizedAtTimestamp:          lane.FieldByName("FinalizedAtTimestamp").Uint(),
		LastProgressTimestamp:         lane.FieldByName("LastProgressTimestamp").Uint(),
	}
	if !transferLane.IsNil() {
		snapshot.TransferLane = snapshotTransferLaneTx(transferLane)
	}
	return snapshot
}

func snapshotTransferLaneTx(transferLanePtr reflect.Value) TransferLaneTxSnapshot {
	transferLane := forceValue(transferLanePtr.Elem())
	seenShards := forceValue(transferLane.FieldByName("seenShards"))

	return TransferLaneTxSnapshot{
		ShardSize:       int(transferLane.FieldByName("ShardSize").Int()),
		MaxDataShards:   int(transferLane.FieldByName("MaxDataShards").Int()),
		TotalDataShards: uint32(transferLane.FieldByName("TotalDataShards").Uint()),
		SeenShardCount:  seenShards.Len(),
		PeerSeenChunks:  uint16(transferLane.FieldByName("peerSeenChunks").Uint()),
		NextSeq:         uint32(transferLane.FieldByName("nextSeq").Uint()),
		HasEncoder:      !transferLane.FieldByName("txCodesState").IsNil(),
	}
}

func snapshotTxChannel(channelStatus reflect.Value) TxChannelSnapshot {
	config := forceInterface[rriptMonoDirectionSession.ChannelConfig](channelStatus.FieldByName("Config"))
	status := forceInterface[rriptMonoDirectionSession.ChannelRateControlStatus](channelStatus.FieldByName("Status"))

	snapshot := TxChannelSnapshot{
		Config: config,
		Status: status,
	}

	materializedValue := forceValue(channelStatus.FieldByName("MaterializeChannel"))
	if materializedValue.IsNil() {
		return snapshot
	}
	materialized := forceInterface[*rrpitMaterializedTransferChannel.ChannelTx](materializedValue)
	snapshot.ChannelID = materialized.ChannelID
	snapshot.NextSeq = materialized.NextSeq
	snapshot.WriterType = fmt.Sprintf("%T", materialized.WriteCloser)

	channelTx := forceValue(materializedValue.Elem().FieldByName("ChannelTx"))
	snapshot.SentPacketHistory = snapshotSentPacketHistory(channelTx.FieldByName("sentPacketHistory"))
	snapshot.ControlHistory = snapshotControlHistory(channelTx.FieldByName("controlHistory"))
	if timestamp, err := materialized.RemoteLastSeenMessageSenderTimestamp(); err == nil {
		timestampCopy := timestamp
		snapshot.RemoteLastSeenSenderTimestamp = &timestampCopy
	} else {
		snapshot.RemoteLastSeenSenderTimestampError = err.Error()
	}

	return snapshot
}

func snapshotSentPacketHistory(history reflect.Value) []SentPacketSnapshot {
	items := snapshotRingBuffer(history)
	result := make([]SentPacketSnapshot, 0, len(items))
	for _, item := range items {
		item = forceValue(item)
		result = append(result, SentPacketSnapshot{
			Seq:       item.FieldByName("seq").Uint(),
			Timestamp: item.FieldByName("timestamp").Uint(),
		})
	}
	return result
}

func snapshotControlHistory(history reflect.Value) []rrpitTransferChannel.ChannelControlMessage {
	items := snapshotRingBuffer(history)
	result := make([]rrpitTransferChannel.ChannelControlMessage, 0, len(items))
	for _, item := range items {
		result = append(result, forceInterface[rrpitTransferChannel.ChannelControlMessage](item))
	}
	return result
}

func snapshotRxLane(lanePtr reflect.Value) RxLaneSnapshot {
	lane := forceValue(lanePtr.Elem())
	transferLane := forceValue(lane.FieldByName("TransferLane"))
	reconstructed := forceValue(lane.FieldByName("Reconstructed"))

	snapshot := RxLaneSnapshot{
		LaneID:             lane.FieldByName("LaneID").Uint(),
		Ready:              lane.FieldByName("Ready").Bool(),
		NextDeliverIndex:   int(lane.FieldByName("NextDeliverIndex").Int()),
		ReconstructedCount: reconstructed.Len(),
		ReconstructedLens:  make([]int, 0, reconstructed.Len()),
	}
	for i := 0; i < reconstructed.Len(); i++ {
		reconstruction := forceInterface[rrpitTransferLane.ReconstructionData](reconstructed.Index(i))
		snapshot.ReconstructedLens = append(snapshot.ReconstructedLens, len(reconstruction.Data))
	}
	if !transferLane.IsNil() {
		snapshot.TransferLane = snapshotTransferLaneRx(transferLane)
	}
	return snapshot
}

func snapshotTransferLaneRx(transferLanePtr reflect.Value) TransferLaneRxSnapshot {
	transferLane := forceValue(transferLanePtr.Elem())
	seenDataShards := forceValue(transferLane.FieldByName("seenDataShards"))

	return TransferLaneRxSnapshot{
		ShardSize:          int(transferLane.FieldByName("ShardSize").Int()),
		TotalDataShards:    uint32(transferLane.FieldByName("TotalDataShards").Uint()),
		SeenShardCount:     uint32(transferLane.FieldByName("seenShardCount").Uint()),
		SeenDataShardCount: seenDataShards.Len(),
		HasDecoder:         !transferLane.FieldByName("rxCodesState").IsNil(),
		Completed:          transferLane.FieldByName("completed").Bool(),
	}
}

func snapshotRxChannel(channelPtr reflect.Value) RxChannelSnapshot {
	channel := forceValue(channelPtr.Elem())
	materializedValue := forceValue(channel.FieldByName("MaterializeChannel"))
	if materializedValue.IsNil() {
		return RxChannelSnapshot{}
	}
	materialized := forceInterface[*rrpitMaterializedTransferChannel.ChannelRx](materializedValue)
	return RxChannelSnapshot{
		ChannelID:             materialized.ChannelID,
		TotalPacketsReceived:  materialized.TotalPacketsReceived,
		LastPacketSeqReceived: materialized.LastPacketSeqReceived,
	}
}

func snapshotRingBuffer(history reflect.Value) []reflect.Value {
	history = forceValue(history)
	values := forceValue(history.FieldByName("values"))
	if values.Len() == 0 {
		return nil
	}

	start := int(history.FieldByName("start").Int())
	size := int(history.FieldByName("size").Int())
	items := make([]reflect.Value, 0, size)
	for i := 0; i < size; i++ {
		index := (start + i) % values.Len()
		items = append(items, forceValue(values.Index(index)))
	}
	return items
}

func forceValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	if value.CanInterface() {
		return value
	}
	if value.CanAddr() {
		return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
	}
	return value
}

func forceInterface[T any](value reflect.Value) T {
	return forceValue(value).Interface().(T)
}
