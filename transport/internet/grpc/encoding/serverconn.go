// +build !confonly

package encoding

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"
)

type ServerConn struct {
	server GunService_TunServer
	reader io.Reader
	over   context.CancelFunc
}

func (s *ServerConn) Read(b []byte) (n int, err error) {
	if s.reader == nil {
		h, err := s.server.Recv()
		if err != nil {
			return 0, newError("unable to read from gun tunnel").Base(err)
		}
		s.reader = bytes.NewReader(h.Data)
	}
	n, err = s.reader.Read(b)
	if err == io.EOF {
		s.reader = nil
		return n, nil
	}
	return n, err
}

func (s *ServerConn) Write(b []byte) (n int, err error) {
	err = s.server.Send(&Hunk{Data: b})
	if err != nil {
		return 0, newError("Unable to send data over gun").Base(err)
	}
	return len(b), nil
}

func (s *ServerConn) Close() error {
	s.over()
	return nil
}

func (*ServerConn) LocalAddr() net.Addr {
	return nil
}

func (*ServerConn) RemoteAddr() net.Addr {
	newError("gun transport do not support get remote address").AtWarning().WriteToLog()
	return &net.UnixAddr{
		Name: "@placeholder",
		Net:  "unix",
	}
}

func (*ServerConn) SetDeadline(time.Time) error {
	return nil
}

func (*ServerConn) SetReadDeadline(time.Time) error {
	return nil
}

func (*ServerConn) SetWriteDeadline(time.Time) error {
	return nil
}

func NewServerConn(server GunService_TunServer, over context.CancelFunc) *ServerConn {
	return &ServerConn{
		server: server,
		reader: nil,
		over:   over,
	}
}
