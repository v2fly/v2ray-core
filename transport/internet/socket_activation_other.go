//go:build !unix
// +build !unix

package internet

import (
	"fmt"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

func activate_socket(address string) (net.Listener, error) {
	return nil, fmt.Errorf("socket activation is not supported on this platform")
}
