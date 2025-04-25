package noop

import (
	"context"
	"net"

	"github.com/ghxhy/v2ray-core/v5/common"
)

type Header struct{}

func (Header) Size() int32 {
	return 0
}

// Serialize implements PacketHeader.
func (Header) Serialize([]byte) {}

func NewHeader(context.Context, interface{}) (interface{}, error) {
	return Header{}, nil
}

type ConnectionHeader struct{}

func (ConnectionHeader) Client(conn net.Conn) net.Conn {
	return conn
}

func (ConnectionHeader) Server(conn net.Conn) net.Conn {
	return conn
}

func NewConnectionHeader(context.Context, interface{}) (interface{}, error) {
	return ConnectionHeader{}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), NewHeader))
	common.Must(common.RegisterConfig((*ConnectionConfig)(nil), NewConnectionHeader))
}
