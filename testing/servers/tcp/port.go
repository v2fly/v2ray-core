package tcp

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/port"
)

// PickPort returns an unused TCP port of the system.
func PickPort() net.Port {
	return port.Pick("tcp4")
}
