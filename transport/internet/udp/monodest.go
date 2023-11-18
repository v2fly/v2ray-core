package udp

import (
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func NewMonoDestUDPConn(conn internet.AbstractPacketConn, addr net.Addr) *MonoDestUDPConn {
	return &MonoDestUDPConn{
		AbstractPacketConn: conn,
		dest:               addr,
	}
}

type MonoDestUDPConn struct {
	internet.AbstractPacketConn
	dest net.Addr
}

func (m *MonoDestUDPConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	buffer := buf.New()
	buffer.Extend(2048)
	nBytes, _, err := m.AbstractPacketConn.ReadFrom(buffer.Bytes())
	if err != nil {
		buffer.Release()
		return nil, err
	}
	buffer.Resize(0, int32(nBytes))
	return buf.MultiBuffer{buffer}, nil
}

func (m *MonoDestUDPConn) WriteMultiBuffer(buffer buf.MultiBuffer) error {
	for _, b := range buffer {
		_, err := m.AbstractPacketConn.WriteTo(b.Bytes(), m.dest)
		if err != nil {
			return err
		}
	}
	buf.ReleaseMulti(buffer)
	return nil
}

func (m *MonoDestUDPConn) Read(p []byte) (n int, err error) {
	n, _, err = m.AbstractPacketConn.ReadFrom(p)
	return
}

func (m *MonoDestUDPConn) Write(p []byte) (n int, err error) {
	return m.AbstractPacketConn.WriteTo(p, m.dest)
}
