package encoding

import (
	"io"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

type MultiConn struct {
	stream Stream
	legacy GunServiceClientX
	done   *done.Instance

	buf   [][]byte
	index int
}

func NewMultiConn(stream Stream) (internet.Connection, <-chan struct{}) {
	c := &MultiConn{stream: stream, done: done.New()}
	return net.NewConnection(net.ConnectionOutputMulti(c), net.ConnectionInputMulti(c), net.ConnectionOnClose(c)), c.done.Wait()
}

func (c *MultiConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if c.done.Done() {
		return nil, io.EOF
	}
	message := new(MultiHunk)
	err := c.stream.RecvMsg(message)
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, newError("failed to fetch data from gRPC tunnel").Base(err)
	}

	mb := buf.MultiBuffer{}
	for _, data := range message.Data {
		if len(data) == 0 {
			continue
		}
		mb = buf.MergeBytes(mb, data)
	}
	return mb, nil
}

func (c *MultiConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)
	if c.done.Done() {
		return io.ErrClosedPipe
	}

	hunks := make([][]byte, 0, len(mb))
	for _, b := range mb {
		if b.Len() > 0 {
			hunks = append(hunks, b.Bytes())
		}
	}
	return c.stream.SendMsg(&MultiHunk{Data: hunks})
}

func (c *MultiConn) Close() error {
	c.done.Close()

	if c, ok := c.stream.(SendCloser); ok {
		return c.CloseSend()
	}

	return nil
}
