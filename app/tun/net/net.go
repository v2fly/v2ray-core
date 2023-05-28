package net

import (
	"github.com/v2fly/v2ray-core/v5/common/net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type TCPConn interface {
	net.Conn

	ID() *stack.TransportEndpointID
}

type UDPConn interface {
	net.Conn
	net.PacketConn

	ID() *stack.TransportEndpointID
}

func AddressFromTCPIPAddr(addr tcpip.Address) net.Address {
	return net.IPAddress(addr.AsSlice())
}
