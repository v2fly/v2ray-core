package udp

import "github.com/v2fly/v2ray-core/v5/common/net"

// PickPort returns an unused UDP port of the system.
func PickPort() net.Port {
	conn := pickPort()
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return net.Port(addr.Port)
}

// PickPortRange returns the first port of a block of `count` consecutive unused
// UDP ports of the system. Unlike calling PickPort and assuming the ports that
// follow are free as well, every port of the block is verified to be available.
func PickPortRange(count int) net.Port {
	if count <= 0 {
		panic("udp: PickPortRange requires a positive count")
	}

	for attempt := 0; attempt < 100; attempt++ {
		conns := []*net.UDPConn{pickPort()}
		base := conns[0].LocalAddr().(*net.UDPAddr).Port

		complete := base+count-1 <= 65535
		for i := 1; complete && i < count; i++ {
			conn, err := net.ListenUDP("udp4", &net.UDPAddr{
				IP:   net.LocalHostIP.IP(),
				Port: base + i,
			})
			if err != nil {
				complete = false
				break
			}
			conns = append(conns, conn)
		}

		for _, conn := range conns {
			conn.Close()
		}
		if complete {
			return net.Port(base)
		}
	}

	panic("udp: unable to find a range of free ports")
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
