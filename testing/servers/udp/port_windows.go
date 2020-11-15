// +build windows

package udp

import (
	"fmt"
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/net"
)

// PickPort returns an unused UDP port in the system. The port returned is highly likely to be unused, but not guaranteed.
func PickPort() net.Port {
	var port uint16
	ports := []uint16{2479, 4491, 5356, 6044, 6275, 6490, 7002, 28384}
	for v := range ports {
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{
			IP:   net.LocalHostIP.IP(),
			Port: v,
		})
		if err != nil {
			fmt.Println(v, " port is not available")
			continue
		} else {
			addr := conn.LocalAddr().(*net.UDPAddr)
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
