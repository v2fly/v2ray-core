//go:build !confonly
// +build !confonly

package encoding

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"time"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// GunService is the abstract interface of GunService_TunClient and GunService_TunServer
type GunService interface {
	Context() context.Context
	Send(*Hunk) error
	Recv() (*Hunk, error)
}

// GunConn implements net.Conn for gun tunnel
type GunConn struct {
	service GunService
	reader  io.Reader
	over    context.CancelFunc
	local   net.Addr
	remote  net.Addr
}

// Read implements net.Conn.Read()
func (c *GunConn) Read(b []byte) (n int, err error) {
	if c.reader == nil {
		h, err := c.service.Recv()
		if err != nil {
			return 0, newError("unable to read from gun tunnel").Base(err)
		}
		c.reader = bytes.NewReader(h.Data)
	}
	n, err = c.reader.Read(b)
	if err == io.EOF {
		c.reader = nil
		return n, nil
	}
	return n, err
}

// Write implements net.Conn.Write()
func (c *GunConn) Write(b []byte) (n int, err error) {
	err = c.service.Send(&Hunk{Data: b})
	if err != nil {
		return 0, newError("Unable to send data over gun").Base(err)
	}
	return len(b), nil
}

// Close implements net.Conn.Close()
func (c *GunConn) Close() error {
	if c.over != nil {
		c.over()
	}
	return nil
}

// LocalAddr implements net.Conn.LocalAddr()
func (c *GunConn) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr implements net.Conn.RemoteAddr()
func (c *GunConn) RemoteAddr() net.Addr {
	return c.remote
}

// SetDeadline implements net.Conn.SetDeadline()
func (*GunConn) SetDeadline(time.Time) error {
	return nil
}

// SetReadDeadline implements net.Conn.SetReadDeadline()
func (*GunConn) SetReadDeadline(time.Time) error {
	return nil
}

// SetWriteDeadline implements net.Conn.SetWriteDeadline()
func (*GunConn) SetWriteDeadline(time.Time) error {
	return nil
}

// NewGunConn creates GunConn which handles gun tunnel
func NewGunConn(service GunService, over context.CancelFunc, parseXForwardedFor bool) *GunConn {
	conn := &GunConn{
		service: service,
		reader:  nil,
		over:    over,
	}

	conn.local = &net.TCPAddr{
		IP:   []byte{0, 0, 0, 0},
		Port: 0,
	}
	pr, ok := peer.FromContext(service.Context())
	if ok {
		conn.remote = pr.Addr
	} else {
		conn.remote = &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		}
	}
	if parseXForwardedFor {
		if remoteAddr := remoteAddrFromXForwardedFor(service.Context()); remoteAddr != nil {
			conn.remote = remoteAddr
		}
	}

	return conn
}

func remoteAddrFromXForwardedFor(ctx context.Context) net.Addr {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	for _, value := range md.Get("x-forwarded-for") {
		for _, proxy := range strings.Split(value, ",") {
			addr := v2net.ParseAddress(strings.TrimSpace(proxy))
			if addr.Family().IsIP() {
				return &net.TCPAddr{
					IP:   addr.IP(),
					Port: 0,
				}
			}
		}
	}
	return nil
}
