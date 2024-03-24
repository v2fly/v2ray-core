package packetstream

import (
	"encoding/binary"
	"io"

	"github.com/v2fly/v2ray-core/v5/common/buf"
)

// PacketStreamWriter is a Writer that writes one packet with length info as header into a byte stream.
type PacketStreamWriter struct {
	io.Writer
}

// WriteMultiBuffer implements buf.Writer
func (w *PacketStreamWriter) WriteMultiBuffer(multiBuffer buf.MultiBuffer) error {
	defer buf.ReleaseMulti(multiBuffer)
	for mb := multiBuffer; mb.Len() > 0; {
		var payload *buf.Buffer
		mb, payload = buf.SplitFirst(mb)
		if payload.IsEmpty() {
			continue
		}

		var lengthBuf [2]byte
		binary.BigEndian.PutUint16(lengthBuf[:], uint16(payload.Len()))

		buffer := buf.NewWithSize(2 + payload.Len())
		defer buffer.Release()
		if _, err := buffer.Write(lengthBuf[:]); err != nil {
			return err
		}
		if _, err := buffer.Write(payload.Bytes()); err != nil {
			return err
		}
		if _, err := w.Writer.Write(buffer.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// PacketStreamReader is a Reader that reads one complete packet every time from a byte stream.
// It first reads the packet length header, then read payload of exact length from the reader.
type PacketStreamReader struct {
	io.Reader
}

// ReadMultiBuffer implements buf.Reader.
func (r *PacketStreamReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	var lengthBuf [2]byte
	if _, err := io.ReadFull(r, lengthBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(lengthBuf[:])

	payload := buf.NewWithSize(int32(length))
	if _, err := payload.ReadFullFrom(r, int32(length)); err != nil {
		return nil, err
	}
	return buf.MultiBuffer{payload}, nil
}
