package mirrorenrollment

import (
	"context"
	"net"
)

func NewCancelContextOnCloseConn(conn net.Conn, done context.CancelFunc) net.Conn {
	return &cancelContextOnCloseConn{
		Conn: conn,
		done: done,
	}
}

type cancelContextOnCloseConn struct {
	net.Conn
	done context.CancelFunc
}

func (c *cancelContextOnCloseConn) Close() error {
	c.done()
	return c.Conn.Close()
}
