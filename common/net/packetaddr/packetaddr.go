package packetaddr

import (
	"bytes"
	"encoding/binary"
	"io"
	gonet "net"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

var addrParser = protocol.NewAddressParser(
	protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(0x02, net.AddressFamilyIPv6),
)

const (
	streamPacketLengthSize        = 2
	maxPacketAddrHeaderSize       = 1 + 16 + 2
	maxStreamPacketAddrSegmentLen = 0xffff
)

// AttachAddressToPacket
// relinquish ownership of data
// gain ownership of the returning value
func AttachAddressToPacket(data *buf.Buffer, address gonet.Addr) (*buf.Buffer, error) {
	udpaddr := address.(*gonet.UDPAddr)
	packetBuf := buf.NewWithSize(data.Len() + maxPacketAddrHeaderSize)
	port, err := net.PortFromInt(uint32(udpaddr.Port))
	if err != nil {
		packetBuf.Release()
		data.Release()
		return nil, err
	}
	err = addrParser.WriteAddressPort(packetBuf, net.IPAddress(udpaddr.IP), port)
	if err != nil {
		packetBuf.Release()
		data.Release()
		return nil, err
	}
	if n, err := packetBuf.Write(data.Bytes()); err != nil {
		packetBuf.Release()
		data.Release()
		return nil, err
	} else if n != int(data.Len()) {
		packetBuf.Release()
		data.Release()
		return nil, errors.New("failed to write full packet payload")
	}
	data.Release()
	return packetBuf, nil
}

// AttachLengthToPacket prefixes a packetaddr segment with its big-endian uint16 length.
// relinquish ownership of data
// gain ownership of the returning value
func AttachLengthToPacket(data *buf.Buffer) (*buf.Buffer, error) {
	if data.Len() > maxStreamPacketAddrSegmentLen {
		data.Release()
		return nil, errors.New("packetaddr segment too large")
	}
	packetBuf := buf.NewWithSize(data.Len() + streamPacketLengthSize)
	binary.BigEndian.PutUint16(packetBuf.Extend(streamPacketLengthSize), uint16(data.Len()))
	if n, err := packetBuf.Write(data.Bytes()); err != nil {
		packetBuf.Release()
		data.Release()
		return nil, err
	} else if n != int(data.Len()) {
		packetBuf.Release()
		data.Release()
		return nil, errors.New("failed to write full packetaddr segment")
	}
	data.Release()
	return packetBuf, nil
}

func ExtractPacketFromStream(reader io.Reader) (*buf.Buffer, error) {
	var lengthBuf [streamPacketLengthSize]byte
	if _, err := io.ReadFull(reader, lengthBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(lengthBuf[:])
	if length == 0 {
		return nil, errors.New("invalid empty packetaddr segment")
	}
	packetBuf := buf.NewWithSize(int32(length))
	if _, err := packetBuf.ReadFullFrom(reader, int32(length)); err != nil {
		packetBuf.Release()
		return nil, err
	}
	return packetBuf, nil
}

// ExtractAddressFromPacket
// relinquish ownership of data
// gain ownership of the returning value
func ExtractAddressFromPacket(data *buf.Buffer) (*buf.Buffer, gonet.Addr, error) {
	packetBuf := buf.StackNew()
	address, port, err := addrParser.ReadAddressPort(&packetBuf, bytes.NewReader(data.Bytes()))
	if err != nil {
		return nil, nil, err
	}
	if address.Family().IsDomain() {
		return nil, nil, errors.New("invalid address type")
	}
	addr := &gonet.UDPAddr{
		IP:   address.IP(),
		Port: int(port.Value()),
		Zone: "",
	}
	data.Advance(packetBuf.Len())
	packetBuf.Release()
	return data, addr, nil
}
