package tun

import "github.com/v2fly/v2ray-core/v5/common/net"

type Handler interface {
	Handle(conn net.Conn) error
}
