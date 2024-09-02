//go:build !unix
// +build !unix

package internet

import (
	"fmt"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

func activateSocket(address string, f func(network, address string, fd uintptr)) (net.Listener, error) {
	return nil, fmt.Errorf("socket activation is not supported on this platform")
}
