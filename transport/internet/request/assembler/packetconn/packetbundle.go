package packetconn

import (
	"encoding/binary"
	"io"
)

func NewPacketBundle() PacketBundle {
	return &packetBundle{}
}

type packetBundle struct{}

func (p *packetBundle) Overhead() int {
	return 2
}

func (p *packetBundle) WriteToBundle(b []byte, writer io.Writer) (err error) {
	err = binary.Write(writer, binary.BigEndian, uint16(len(b)))
	if err != nil {
		return
	}
	_, err = writer.Write(b)
	return
}

func (p *packetBundle) ReadFromBundle(writer io.Reader) (b []byte, err error) {
	var length uint16
	err = binary.Read(writer, binary.BigEndian, &length)
	if err != nil {
		return
	}
	b = make([]byte, length)
	n, err := io.ReadFull(writer, b)
	if err != nil {
		return
	}
	if n != int(length) {
		return nil, io.ErrUnexpectedEOF
	}
	return
}

type PacketBundle interface {
	Overhead() int
	WriteToBundle(b []byte, writer io.Writer) (err error)
	ReadFromBundle(writer io.Reader) (b []byte, err error)
}
