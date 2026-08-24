package udp

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/port"
)

// PickPort returns an unused UDP port of the system. The port stays reserved
// for the lifetime of the test binary, so it is never handed out twice.
func PickPort() net.Port {
	return port.Pick()
}

// PickPortRange returns the first port of a block of count consecutive unused
// UDP ports. The whole block stays reserved for the lifetime of the test
// binary.
func PickPortRange(count uint32) net.Port {
	return port.PickRange(count)
}
