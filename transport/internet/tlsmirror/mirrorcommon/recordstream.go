package mirrorcommon

import (
	"bufio"
	"errors"
	"io"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewTLSRecordStreamReader(reader *bufio.Reader) *TLSRecordStreamReader {
	return &TLSRecordStreamReader{bufferedReader: reader}
}

type TLSRecordStreamReader struct {
	bufferedReader *bufio.Reader
	consumedSize   int64
}

func (t *TLSRecordStreamReader) ReadNextRecord() (*tlsmirror.TLSRecord, error) {
	record, tryAgainLength, processedLength, err := PeekTLSRecord(t.bufferedReader, nil)
	if err == nil {
		_, err := t.bufferedReader.Discard(processedLength)
		t.consumedSize += int64(processedLength)
		if err != nil {
			return nil, err
		}
		return &record, nil
	}

	if errors.Is(err, io.EOF) {
		return nil, err
	} else { // nolint: gocritic
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

func (t *TLSRecordStreamReader) GetConsumedSize() int64 {
	return t.consumedSize
}

func NewTLSRecordStreamWriter(writer *bufio.Writer) *TLSRecordStreamWriter {
	return &TLSRecordStreamWriter{bufferedWriter: writer}
}

type TLSRecordStreamWriter struct {
	bufferedWriter *bufio.Writer
}

func (t *TLSRecordStreamWriter) WriteRecord(record *tlsmirror.TLSRecord, holdFlush bool) error {
	_, err := t.bufferedWriter.Write(PackTLSRecord(*record))
	if err != nil {
		return err
	}
	if holdFlush {
		return nil
	}
	return t.bufferedWriter.Flush()
}
