package hysteria2

import (
	"context"
	"time"

	"github.com/apernet/quic-go"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

type HyConn struct {
	UseUDPExtension bool
	quicConn        quic.Connection
	stream          quic.Stream
	local           net.Addr
	remote          net.Addr
}

func (c *HyConn) Read(b []byte) (int, error) {
	if c.UseUDPExtension {
		c.quicConn.ReceiveDatagram(context.Background())
	}
	return c.stream.Read(b)
}

func (c *HyConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	mb = buf.Compact(mb)
	mb, err := buf.WriteMultiBuffer(c, mb)
	buf.ReleaseMulti(mb)
	return err
}

func (c *HyConn) Write(b []byte) (int, error) {
	if c.UseUDPExtension {
		return len(b), c.quicConn.SendDatagram(b)
	}
	return c.stream.Write(b)
}

func (c *HyConn) Close() error {
	if c.UseUDPExtension {
		return nil
	}
	return c.stream.Close()
}

func (c *HyConn) LocalAddr() net.Addr {
	return c.local
}

func (c *HyConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *HyConn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *HyConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *HyConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}
