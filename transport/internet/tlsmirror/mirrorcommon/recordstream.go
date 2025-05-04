package mirrorcommon

import (
	"bufio"
	"errors"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"io"
)

func NewTLSRecordStreamReader(reader *bufio.Reader) *TLSRecordStreamReader {
	return &TLSRecordStreamReader{bufferedReader: reader}
}

type TLSRecordStreamReader struct {
	bufferedReader *bufio.Reader
}

func (t *TLSRecordStreamReader) ReadNextRecord() (*tlsmirror.TLSRecord, error) {
	record, tryAgainLength, processedLength, err := PeekTLSRecord(t.bufferedReader)
	if err == nil {
		_, err := t.bufferedReader.Discard(processedLength)
		if err != nil {
			return nil, err
		}
		return &record, nil
	}

	if errors.Is(err, io.EOF) {
		return nil, err
	} else {
		if tryAgainLength == 0 {
			return nil, err
		}
	}

	if tryAgainLength > 0 {
		_, err := t.bufferedReader.Read(make([]byte, tryAgainLength))
		if err != nil {
			return nil, err
		}
		err = t.bufferedReader.UnreadByte()
		if err != nil {
			return nil, err
		}
	}
	return t.ReadNextRecord()
}

func NewTLSRecordStreamWriter(writer *bufio.Writer) *TLSRecordStreamWriter {
	return &TLSRecordStreamWriter{bufferedWriter: writer}
}

type TLSRecordStreamWriter struct {
	bufferedWriter *bufio.Writer
}

func (t *TLSRecordStreamWriter) WriteRecord(record *tlsmirror.TLSRecord) error {
	_, err := t.bufferedWriter.Write(PackTLSRecord(*record))
	if err != nil {
		return err
	}

	return t.bufferedWriter.Flush()
}
