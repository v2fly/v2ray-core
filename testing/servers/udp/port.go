package udp

import "v2ray.com/core/common/net"

// PickPort returns an unused UDP port of the system.
func PickPort() net.Port {
	conn := pickPort()
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return net.Port(addr.Port)
}

func pickPort() *net.UDPConn {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.LocalHostIP.IP(),
		Port: 0,
	})
	if err != nil {
		conn = pickPort()
	}
	return conn
}
