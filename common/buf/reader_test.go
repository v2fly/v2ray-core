package buf_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ghxhy/v2ray-core/v5/common"
	. "github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/transport/pipe"
)

func TestBytesReaderWriteTo(t *testing.T) {
	pReader, pWriter := pipe.New(pipe.WithSizeLimit(1024))
	reader := &BufferedReader{Reader: pReader}
	b1 := New()
	b1.WriteString("abc")
	b2 := New()
	b2.WriteString("efg")
	common.Must(pWriter.WriteMultiBuffer(MultiBuffer{b1, b2}))
	pWriter.Close()

	pReader2, pWriter2 := pipe.New(pipe.WithSizeLimit(1024))
	writer := NewBufferedWriter(pWriter2)
	writer.SetBuffered(false)

	nBytes, err := io.Copy(writer, reader)
	common.Must(err)
	if nBytes != 6 {
		t.Error("copy: ", nBytes)
	}

	mb, err := pReader2.ReadMultiBuffer()
	common.Must(err)
	if s := mb.String(); s != "abcefg" {
		t.Error("content: ", s)
	}
}

func TestBytesReaderMultiBuffer(t *testing.T) {
	pReader, pWriter := pipe.New(pipe.WithSizeLimit(1024))
	reader := &BufferedReader{Reader: pReader}
	b1 := New()
	b1.WriteString("abc")
	b2 := New()
	b2.WriteString("efg")
	common.Must(pWriter.WriteMultiBuffer(MultiBuffer{b1, b2}))
	pWriter.Close()

	mbReader := NewReader(reader)
	mb, err := mbReader.ReadMultiBuffer()
	common.Must(err)
	if s := mb.String(); s != "abcefg" {
		t.Error("content: ", s)
	}
}

func TestReadByte(t *testing.T) {
	sr := strings.NewReader("abcd")
	reader := &BufferedReader{
		Reader: NewReader(sr),
	}
	b, err := reader.ReadByte()
	common.Must(err)
	if b != 'a' {
		t.Error("unexpected byte: ", b, " want a")
	}
	if reader.BufferedBytes() != 3 { // 3 bytes left in buffer
		t.Error("unexpected buffered Bytes: ", reader.BufferedBytes())
	}

	nBytes, err := reader.WriteTo(DiscardBytes)
	common.Must(err)
	if nBytes != 3 {
		t.Error("unexpect bytes written: ", nBytes)
	}
}

func TestReadBuffer(t *testing.T) {
	{
		sr := strings.NewReader("abcd")
		buf, err := ReadBuffer(sr)
		common.Must(err)

		if s := buf.String(); s != "abcd" {
			t.Error("unexpected str: ", s, " want abcd")
		}
		buf.Release()
	}
}

func TestReadAtMost(t *testing.T) {
	sr := strings.NewReader("abcd")
	reader := &BufferedReader{
		Reader: NewReader(sr),
	}

	mb, err := reader.ReadAtMost(3)
	common.Must(err)
	if s := mb.String(); s != "abc" {
		t.Error("unexpected read result: ", s)
	}

	nBytes, err := reader.WriteTo(DiscardBytes)
	common.Must(err)
	if nBytes != 1 {
		t.Error("unexpect bytes written: ", nBytes)
	}
}

func TestPacketReader_ReadMultiBuffer(t *testing.T) {
	const alpha = "abcefg"
	buf := bytes.NewBufferString(alpha)
	reader := &PacketReader{buf}
	mb, err := reader.ReadMultiBuffer()
	common.Must(err)
	if s := mb.String(); s != alpha {
		t.Error("content: ", s)
	}
}

func TestReaderInterface(t *testing.T) {
	_ = (io.Reader)(new(ReadVReader))
	_ = (Reader)(new(ReadVReader))

	_ = (Reader)(new(BufferedReader))
	_ = (io.Reader)(new(BufferedReader))
	_ = (io.ByteReader)(new(BufferedReader))
	_ = (io.WriterTo)(new(BufferedReader))
}
