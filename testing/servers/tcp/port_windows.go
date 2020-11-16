package tcp

import (
	"fmt"
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/net"
)

// PickPort returns an unused TCP port in the system. The port returned is highly likely to be unused, but not guaranteed.
func PickPort() net.Port {
	var port uint16
	ports := []uint16{2479, 4491, 5356, 6044, 6275, 6490, 7002, 28384}
	for v := range ports {
		listener, err := net.Listen("tcp4", "127.0.0.1:" + v)
		if err != nil {
			fmt.Println(v, " port is not available")
			continue
		} else {
			addr := listener.Addr().(*net.TCPAddr)
			port = uint16(addr.Port)
			_ = conn.Close()
			break
		}
	}

	if port == 0 {
		common.Must(errors.New("Cannot get a port on windows"))
	}

	return net.Port(port)
}
