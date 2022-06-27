package inbound

import (
	"io"
	"net"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
)

func newUDPConnAdaptor(conn net.Conn, done *done.Instance) net.Conn {
	return &udpConnAdp{
		Conn:              conn,
		reader:            buf.NewPacketReader(conn),
		cachedMultiBuffer: nil,
		finished:          done,
	}
}

type udpConnAdp struct {
	net.Conn
	reader buf.Reader

	cachedMultiBuffer buf.MultiBuffer

	finished *done.Instance
}

func (u *udpConnAdp) Read(p []byte) (n int, err error) {
	if u.cachedMultiBuffer.IsEmpty() {
		u.cachedMultiBuffer, err = u.reader.ReadMultiBuffer()
		if err != nil {
			return 0, newError("unable to read from connection").Base(err)
		}
	}
	var buffer *buf.Buffer
	u.cachedMultiBuffer, buffer = buf.SplitFirst(u.cachedMultiBuffer)
	defer buffer.Release()
	n = copy(p, buffer.Bytes())
	if n != int(buffer.Len()) {
		return 0, io.ErrShortBuffer
	}
	return n, nil
}

func (u *udpConnAdp) Close() error {
	u.finished.Close()
	return u.Conn.Close()
}
