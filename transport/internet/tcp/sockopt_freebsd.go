//go:build freebsd
// +build freebsd

package tcp

import (
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

// GetOriginalDestination from tcp conn
func GetOriginalDestination(conn internet.Connection) (net.Destination, error) {
	la := conn.LocalAddr()
	ra := conn.RemoteAddr()
	ip, port, err := internet.OriginalDst(la, ra)
	if err != nil {
		return net.Destination{}, newError("failed to get destination").Base(err)
	}
	dest := net.TCPDestination(net.IPAddress(ip), net.Port(port))
	if !dest.IsValid() {
		return net.Destination{}, newError("failed to parse destination.")
	}
	return dest, nil
}
