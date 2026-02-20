package hysteria2

import (
	"time"

	hyClient "github.com/v2fly/hysteria/core/v2/client"
	"github.com/v2fly/hysteria/core/v2/international/protocol"
	"github.com/v2fly/hysteria/core/v2/international/utils"
	hyServer "github.com/v2fly/hysteria/core/v2/server"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

const (
	CanNotUseUDPExtension = "Only hysteria2 proxy protocol can use udpExtension."
	Hy2MustNeedTLS        = "Hysteria2 based on QUIC that requires TLS."
)

type HyConn struct {
	IsUDPExtension   bool
	IsServer         bool
	ClientUDPSession hyClient.HyUDPConn
	ServerUDPSession *hyServer.UdpSessionEntry

	stream *utils.QStream
	local  net.Addr
	remote net.Addr
}

func (c *HyConn) Read(b []byte) (int, error) {
	if c.IsUDPExtension {
		n, data, _, err := c.ReadPacket()
		copy(b, data)
		return n, err
	}
	return c.stream.Read(b)
}

func (c *HyConn) Write(b []byte) (int, error) {
	if c.IsUDPExtension {
		dest, _ := net.ParseDestination("udp:v2fly.org:6666")
		return c.WritePacket(b, dest)
	}
	return c.stream.Write(b)
}

func (c *HyConn) WritePacket(b []byte, dest net.Destination) (int, error) {
	if !c.IsUDPExtension {
		return 0, newError(CanNotUseUDPExtension)
	}

	if c.IsServer {
		msg := &protocol.UDPMessage{
			SessionID: c.ServerUDPSession.ID,
			PacketID:  0,
			FragID:    0,
			FragCount: 1,
			Addr:      dest.NetAddr(),
			Data:      b,
		}
		c.ServerUDPSession.SendCh <- msg
		return len(b), nil
	}
	return len(b), c.ClientUDPSession.Send(b, dest.NetAddr())
}

func (c *HyConn) ReadPacket() (int, []byte, *net.Destination, error) {
	if !c.IsUDPExtension {
		return 0, nil, nil, newError(CanNotUseUDPExtension)
	}

	if c.IsServer {
		msg, ok := <-c.ServerUDPSession.ReceiveCh
		if !ok {
			return 0, nil, nil, newError("UDP session receive channel closed")
		}
		dest, err := net.ParseDestination("udp:" + msg.Addr)
		return len(msg.Data), msg.Data, &dest, err
	}
	data, address, err := c.ClientUDPSession.Receive()
	if err != nil {
		return 0, nil, nil, err
	}
	dest, err := net.ParseDestination("udp:" + address)
	if err != nil {
		return 0, nil, nil, err
	}
	return len(data), data, &dest, nil
}

func (c *HyConn) Close() error {
	if c.IsUDPExtension {
		if !c.IsServer && c.ClientUDPSession == nil || (c.IsServer && c.ServerUDPSession == nil) {
			return newError(CanNotUseUDPExtension)
		}
		if c.IsServer {
			c.ServerUDPSession.CloseWithErr(nil)
			return nil
		}
		return c.ClientUDPSession.Close()
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
	if c.IsUDPExtension {
		return nil
	}
	return c.stream.SetDeadline(t)
}

func (c *HyConn) SetReadDeadline(t time.Time) error {
	if c.IsUDPExtension {
		return nil
	}
	return c.stream.SetReadDeadline(t)
}

func (c *HyConn) SetWriteDeadline(t time.Time) error {
	if c.IsUDPExtension {
		return nil
	}
	return c.stream.SetWriteDeadline(t)
}
