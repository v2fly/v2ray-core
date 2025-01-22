package packetaddr

import (
	"bytes"
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

// AttachAddressToPacket
// relinquish ownership of data
// gain ownership of the returning value
func AttachAddressToPacket(data *buf.Buffer, address gonet.Addr) (*buf.Buffer, error) {
	packetBuf := buf.New()
	udpaddr := address.(*gonet.UDPAddr)
	port, err := net.PortFromInt(uint32(udpaddr.Port))
	if err != nil {
		return nil, err
	}
	err = addrParser.WriteAddressPort(packetBuf, net.IPAddress(udpaddr.IP), port)
	if err != nil {
		return nil, err
	}
	_, err = packetBuf.Write(data.Bytes())
	if err != nil {
		return nil, err
	}
	data.Release()
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
