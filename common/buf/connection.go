//go:build !confonly
// +build !confonly

package buf

import (
	"io"
	"net"
	"time"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/errors"
	net2 "github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
)

type ConnectionOption func(*connection)

func ConnectionLocalAddr(a net.Addr) ConnectionOption {
	return func(c *connection) {
		c.local = a
	}
}

func ConnectionRemoteAddr(a net.Addr) ConnectionOption {
	return func(c *connection) {
		c.remote = a
	}
}

func ConnectionInput(writer io.Writer) ConnectionOption {
	return func(c *connection) {
		c.writer = NewWriter(writer)
	}
}

func ConnectionInputMulti(writer Writer) ConnectionOption {
	return func(c *connection) {
		c.writer = writer
	}
}

func ConnectionOutput(reader io.Reader) ConnectionOption {
	return func(c *connection) {
		c.reader = &BufferedReader{Reader: NewReader(reader)}
	}
}

func ConnectionOutputMulti(reader Reader) ConnectionOption {
	return func(c *connection) {
		c.reader = &BufferedReader{Reader: reader}
	}
}

func ConnectionOutputMultiUDP(reader Reader) ConnectionOption {
	return func(c *connection) {
		c.reader = &BufferedReader{
			Reader:  reader,
			Spliter: SplitFirstBytes,
		}
	}
}

func ConnectionOnClose(n io.Closer) ConnectionOption {
	return func(c *connection) {
		c.onClose = n
	}
}

func NewConnection(opts ...ConnectionOption) net.Conn {
	c := &connection{
		done: done.New(),
		local: &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		},
		remote: &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type connection struct {
	reader  *BufferedReader
	writer  Writer
	done    *done.Instance
	onClose io.Closer
	local   net2.Addr
	remote  net2.Addr
}

func (c *connection) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

// ReadMultiBuffer implements buf.Reader.
func (c *connection) ReadMultiBuffer() (MultiBuffer, error) {
	return c.reader.ReadMultiBuffer()
}

// Write implements net.Conn.Write().
func (c *connection) Write(b []byte) (int, error) {
	if c.done.Done() {
		return 0, io.ErrClosedPipe
	}

	if len(b)/Size+1 > 64*1024*1024 {
		return 0, errors.New("value too large")
	}
	l := len(b)
	sliceSize := l/Size + 1
	mb := make(MultiBuffer, 0, sliceSize)
	mb = MergeBytes(mb, b)
	return l, c.writer.WriteMultiBuffer(mb)
}

func (c *connection) WriteMultiBuffer(mb MultiBuffer) error {
	if c.done.Done() {
		ReleaseMulti(mb)
		return io.ErrClosedPipe
	}

	return c.writer.WriteMultiBuffer(mb)
}

// Close implements net.Conn.Close().
func (c *connection) Close() error {
	common.Must(c.done.Close())
	common.Interrupt(c.reader)
	common.Close(c.writer)
	if c.onClose != nil {
		return c.onClose.Close()
	}

	return nil
}

// LocalAddr implements net.Conn.LocalAddr().
func (c *connection) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr implements net.Conn.RemoteAddr().
func (c *connection) RemoteAddr() net.Addr {
	return c.remote
}

// SetDeadline implements net.Conn.SetDeadline().
func (c *connection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements net.Conn.SetReadDeadline().
func (c *connection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements net.Conn.SetWriteDeadline().
func (c *connection) SetWriteDeadline(t time.Time) error {
	return nil
}
