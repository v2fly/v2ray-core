package tlsmirror

import "github.com/v2fly/v2ray-core/v5/common"

type TLSRecord struct {
	RecordType            byte
	LegacyProtocolVersion [2]byte
	RecordLength          uint16
	Fragment              []byte
}

type RecordReader interface {
	ReadNextRecord(rejectProfile PartialTLSRecordRejectProfile) (*TLSRecord, error)
}

type RecordWriter interface {
	WriteRecord(record *TLSRecord) error
}

type Peeker interface {
	Peek(n int) ([]byte, error)
}

type PartialTLSRecordRejectProfile interface {
	TestIfReject(record *TLSRecord, readyFields int) error
}

type MessageHook func(message *TLSRecord) (drop bool, ok error)

type InsertableTLSConn interface {
	common.Closable
	InsertC2SMessage(message *TLSRecord) error
	InsertS2CMessage(message *TLSRecord) error
}
