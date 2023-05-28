package tun

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var (
	tcpQueue = make(chan TCPConn)
	udpQueue = make(chan UDPConn)
)

type TCPConn interface {
	net.Conn

	ID() *stack.TransportEndpointID
}

type UDPConn interface {
	net.Conn

	ID() *stack.TransportEndpointID
}

func handleTCP(conn TCPConn) {
	tcpQueue <- conn
}

func handleUDP(conn UDPConn) {
	udpQueue <- conn
}
