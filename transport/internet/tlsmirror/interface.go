package tlsmirror

type TLSRecord struct {
	RecordType            byte
	LegacyProtocolVersion [2]byte
	RecordLength          uint16
	Fragment              []byte
}

type RecordReader interface {
	ReadNextRecord() (*TLSRecord, error)
}

type RecordWriter interface {
	WriteRecord(record *TLSRecord) error
}

type Peeker interface {
	Peek(n int) ([]byte, error)
}
