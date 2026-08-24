package udp

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/port"
)

// PickPort returns an unused UDP port of the system.
func PickPort() net.Port {
	return port.Pick("udp4")
}
