package mirrorcommon

import (
	"fmt"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

// PeekTLSRecord reads a TLS record from peeker.
// It returns the record, the number of bytes read from peeker, and any error encountered.
// It does not consume content from peeker, and the content peeked is borrowed from peeker.
func PeekTLSRecord(peeker tlsmirror.Peeker, rejectionProfile tlsmirror.PartialTLSRecordRejectProfile) (result tlsmirror.TLSRecord, tryAgainLength, processed int, err error) {
	var record tlsmirror.TLSRecord
	header, err := peeker.Peek(5)
	if err != nil {
		return record, 0, 0, err
	}
	if len(header) < 5 {
		return record, 5, 0, fmt.Errorf("tls: record header too short")
	}
	record.RecordType = header[0]
	record.LegacyProtocolVersion[0] = header[1]
	record.LegacyProtocolVersion[1] = header[2]
	record.RecordLength = uint16(header[3])<<8 | uint16(header[4])
	if record.RecordLength > 16384 {
		return record, 0, 0, fmt.Errorf("tls: record length %d is too large", record.RecordLength)
	}
	if rejectionProfile != nil {
		err = rejectionProfile.TestIfReject(&record, 2)
		if err != nil {
			return record, 0, 0, err
		}
	}
	fragment, err := peeker.Peek(int(5 + record.RecordLength))
	if err != nil {
		return record, int(5 + record.RecordLength), 0, err
	}
	if len(fragment) < 5+int(record.RecordLength) {
		return record, int(5 + record.RecordLength), 0, fmt.Errorf("tls: record fragment too short")
	}
	record.Fragment = fragment[5:]
	return record, 0, int(5 + record.RecordLength), nil
}

func PackTLSRecord(record tlsmirror.TLSRecord) []byte {
	buf := make([]byte, 5+len(record.Fragment))
	buf[0] = record.RecordType
	buf[1] = record.LegacyProtocolVersion[0]
	buf[2] = record.LegacyProtocolVersion[1]
	buf[3] = byte(record.RecordLength >> 8)
	buf[4] = byte(record.RecordLength)
	copy(buf[5:], record.Fragment)
	return buf
}

func DuplicateRecord(record tlsmirror.TLSRecord) tlsmirror.TLSRecord {
	newRecord := record
	newRecord.Fragment = make([]byte, len(record.Fragment))
	copy(newRecord.Fragment, record.Fragment)
	return newRecord
}
