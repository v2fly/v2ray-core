// +build !confonly

package domainsocket

import (
	"v2ray.com/core/common"
	"v2ray.com/core/common/net"
	"v2ray.com/core/transport/internet"
)

const protocolName = "domainsocket"

func (c *Config) GetUnixAddr() (*net.UnixAddr, error) {
	path := c.Path
	if path == "" {
		return nil, newError("empty domain socket path")
	}

	if c.Abstract && path[0] != '@' {
		path = "@" + path
	}

	// Domain socket path exceeds the length limit will failed to bind.
	// Copy without checked may cause different domain socket setting
	// point to the same connection.
	sockaddrCap := UnixSockaddrCap()
	if len(path) > sockaddrCap {
		return nil, newError("domain socket path too long")
	}

	if c.Abstract && c.Padding {
		raw := []byte(path)
		addr := make([]byte, sockaddrCap)
		copy(addr, raw)
		path = string(addr)
	}

	return &net.UnixAddr{
		Name: path,
		Net:  "unix",
	}, nil
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
