package tun

import (
	tun_net "github.com/v2fly/v2ray-core/v5/app/tun/net"
)

var (
	tcpQueue = make(chan tun_net.TCPConn)
	udpQueue = make(chan tun_net.UDPConn)
)

func handleTCP(conn tun_net.TCPConn) {
	tcpQueue <- conn
}

func handleUDP(conn tun_net.UDPConn) {
	udpQueue <- conn
}
