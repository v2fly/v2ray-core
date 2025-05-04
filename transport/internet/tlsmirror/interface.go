package tlsmirror

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
