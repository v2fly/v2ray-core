package internet

import (
	"net"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/features/stats"
)

type Connection interface {
	net.Conn
}

type AbstractPacketConnReader interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
}

type AbstractPacketConnWriter interface {
	WriteTo(p []byte, addr net.Addr) (n int, err error)
}

type AbstractPacketConn interface {
	AbstractPacketConnReader
	AbstractPacketConnWriter
	common.Closable
}

type PacketConn interface {
	AbstractPacketConn
	net.PacketConn
}

type StatCouterConnection struct {
	Connection
	ReadCounter  stats.Counter
	WriteCounter stats.Counter
}

func (c *StatCouterConnection) Read(b []byte) (int, error) {
	nBytes, err := c.Connection.Read(b)
	if c.ReadCounter != nil {
		c.ReadCounter.Add(int64(nBytes))
	}

	return nBytes, err
}

func (c *StatCouterConnection) Write(b []byte) (int, error) {
	nBytes, err := c.Connection.Write(b)
	if c.WriteCounter != nil {
		c.WriteCounter.Add(int64(nBytes))
	}
	return nBytes, err
}
