// +build freebsd
// +build !confonly

package tcp

import (
	"github.com/v2fly/v2ray-core/common/net"
	"github.com/v2fly/v2ray-core/transport/internet"
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
