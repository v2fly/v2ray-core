package tcp

import "github.com/v2fly/v2ray-core/v5/common/net"

// PickPort returns an unused TCP port of the system.
func PickPort() net.Port {
	listener := pickPort()
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return net.Port(addr.Port)
}

// PickPortRange returns the first port of a block of `count` consecutive unused
// TCP ports of the system. Unlike calling PickPort and assuming the ports that
// follow are free as well, every port of the block is verified to be available.
func PickPortRange(count int) net.Port {
	if count <= 0 {
		panic("tcp: PickPortRange requires a positive count")
	}

	for attempt := 0; attempt < 100; attempt++ {
		listeners := []net.Listener{pickPort()}
		base := listeners[0].Addr().(*net.TCPAddr).Port

		complete := base+count-1 <= 65535
		for i := 1; complete && i < count; i++ {
			listener, err := net.Listen("tcp4", (&net.TCPAddr{IP: net.LocalHostIP.IP(), Port: base + i}).String())
			if err != nil {
				complete = false
				break
			}
			listeners = append(listeners, listener)
		}

		for _, listener := range listeners {
			listener.Close()
		}
		if complete {
			return net.Port(base)
		}
	}

	panic("tcp: unable to find a range of free ports")
}

func pickPort() net.Listener {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		listener = pickPort()
	}
	return listener
}
