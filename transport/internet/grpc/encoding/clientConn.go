// +build !confonly

package encoding

import (
	"bytes"
	"io"
	"net"
	"time"
)

type ClientConn struct {
	client GunService_TunClient
	reader io.Reader
}

func (*ClientConn) LocalAddr() net.Addr {
	return nil
}

func (*ClientConn) RemoteAddr() net.Addr {
	return nil
}

func (*ClientConn) SetDeadline(time.Time) error {
	return nil
}

func (*ClientConn) SetReadDeadline(time.Time) error {
	return nil
}

func (*ClientConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (s *ClientConn) Read(b []byte) (n int, err error) {
	if s.reader == nil {
		h, err := s.client.Recv()
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

func (s *ClientConn) Write(b []byte) (n int, err error) {
	err = s.client.Send(&Hunk{Data: b})
	if err != nil {
		return 0, newError("Unable to send data over gun").Base(err)
	}
	return len(b), nil
}

func (s *ClientConn) Close() error {
	return s.client.CloseSend()
}

func NewClientConn(client GunService_TunClient) *ClientConn {
	return &ClientConn{
		client: client,
		reader: nil,
	}
}
