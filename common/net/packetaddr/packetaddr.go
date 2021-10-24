package packetaddr

import (
	"bytes"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	sysnet "net"
)

var addrParser = protocol.NewAddressParser(
	protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(0x02, net.AddressFamilyIPv6),
)

func AttachAddressToPacket(data []byte, address sysnet.Addr) []byte {
	packetBuf := buf.StackNew()
	udpaddr := address.(*sysnet.UDPAddr)
	port, err := net.PortFromInt(uint32(udpaddr.Port))
	if err != nil {
		panic(err)
	}
	err = addrParser.WriteAddressPort(&packetBuf, net.IPAddress(udpaddr.IP), port)
	if err != nil {
		panic(err)
	}
	data = append(packetBuf.Bytes(), data...)
	packetBuf.Release()
	return data
}

func ExtractAddressFromPacket(data []byte) ([]byte, sysnet.Addr) {
	packetBuf := buf.StackNew()
	address, port, err := addrParser.ReadAddressPort(&packetBuf, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	var addr = &sysnet.UDPAddr{
		IP:   address.IP(),
		Port: int(port.Value()),
		Zone: "",
	}
	payload := data[int(packetBuf.Len()):]
	return payload, addr
}
