package encoding

import (
	"context"
	"io"

	"google.golang.org/grpc/encoding"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/bytesgrp"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func init() {
	encoding.RegisterCodec(rawCodec{})
}

var _ encoding.Codec = (*rawCodec)(nil)

type rawMessage struct {
	data [][]byte
}

type rawCodec struct {
}

func (b rawCodec) Name() string {
	return "raw"
}

func (b rawCodec) Marshal(v interface{}) ([]byte, error) {
	return bytesgrp.Pack(v.(*rawMessage).data), nil
}

func (b rawCodec) Unmarshal(data []byte, v interface{}) error {
	v.(*rawMessage).data = bytesgrp.UnPack(data)
	return nil
}

type Stream interface {
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

type SendCloser interface {
	CloseSend() error
}

type RawConn struct {
	stream Stream
	legacy GunServiceClientX
	done   *done.Instance

	buf   [][]byte
	index int
}

func NewRawConn(stream Stream) (internet.Connection, <-chan struct{}) {
	c := &RawConn{stream: stream, done: done.New()}
	return net.NewConnection(net.ConnectionOutputMulti(c), net.ConnectionInputMulti(c), net.ConnectionOnClose(c)), c.done.Wait()
}

func (c *RawConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if c.done.Done() {
		return nil, io.EOF
	}
	message := new(rawMessage)
	err := c.stream.RecvMsg(message)
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, newError("failed to fetch data from gRPC tunnel").Base(err)
	}

	mb := buf.MultiBuffer{}
	for _, data := range message.data {
		if len(data) == 0 {
			continue
		}
		mb = buf.MergeBytes(mb, data)
	}
	return mb, nil
}

func (c *RawConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
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
	return c.stream.SendMsg(&rawMessage{hunks})
}

func (c *RawConn) Close() error {
	c.done.Close()

	if c, ok := c.stream.(SendCloser); ok {
		return c.CloseSend()
	}

	return nil
}

func (c *RawConn) Wait() <-chan struct{} {
	return c.Wait()
}
