package stun

import (
	"github.com/pion/stun/v3"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type STUNMessageCallback func(b []byte, addr net.Addr)

func NewFilteredConnection(inner net.PacketConn, callback STUNMessageCallback) (*FilteredConnection, error) {
	return &FilteredConnection{
		PacketConn:      inner,
		stunMsgCallback: callback,
	}, nil
}

type FilteredConnection struct {
	net.PacketConn
	stunMsgCallback STUNMessageCallback
}

func (f *FilteredConnection) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := f.PacketConn.ReadFrom(b)
	if stun.IsMessage(b[:n]) {
		f.stunMsgCallback(b[:n], addr)
	}
	return n, addr, err
}
