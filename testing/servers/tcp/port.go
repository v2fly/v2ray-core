package tcp

import "github.com/v2fly/v2ray-core/v4/common/net"

// PickPort returns an unused TCP port of the system.
func PickPort() net.Port {
	listener := pickPort()
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return net.Port(addr.Port)
}

func pickPort() net.Listener {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		listener = pickPort()
	}
	return listener
}
