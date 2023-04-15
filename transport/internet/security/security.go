package security

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
)

type Engine interface {
	Client(conn net.Conn, opts ...Option) (Conn, error)
}

type Conn interface {
	net.Conn
}

type Option interface {
	isSecurityOption()
}

type OptionWithALPN struct {
	ALPNs []string
}

func (a OptionWithALPN) isSecurityOption() {
}

type OptionWithDestination struct {
	Dest net.Destination
}

func (a OptionWithDestination) isSecurityOption() {
}
