package httpupgrade

import (
	"bytes"
	"context"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"io"
	"time"
)

type connection struct {
	conn       net.Conn
	reader     io.Reader
	remoteAddr net.Addr

	shouldWait        bool
	delayedDialFinish context.Context
	finishedDial      context.CancelFunc
	dialer            DelayedDialer
}

type DelayedDialer interface {
	Dial(earlyData []byte) (conn net.Conn, earlyReply []byte, err error)
}

func newConnectionWithEarlyReply(conn net.Conn, remoteAddr net.Addr, earlyReplyReader io.Reader) *connection {
	return &connection{
		conn:       conn,
		remoteAddr: remoteAddr,
		reader:     earlyReplyReader,
	}
}

// Read implements net.Conn.Read()
func (c *connection) Read(b []byte) (int, error) {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			return 0, newError("unable to read delayed dial websocket connection as it do not exist")
		}
	}

	if c.reader != nil {
		n, err := c.reader.Read(b)
		if err == io.EOF {
			c.reader = nil
			return c.conn.Read(b)
		}
		return n, err
	}
	return c.conn.Read(b)
}

// Write implements io.Writer.
func (c *connection) Write(b []byte) (int, error) {
	if c.shouldWait {
		var err error
		var earlyReply []byte
		c.conn, earlyReply, err = c.dialer.Dial(b)
		if earlyReply != nil {
			c.reader = bytes.NewReader(earlyReply)
		}
		c.finishedDial()
		if err != nil {
			return 0, newError("Unable to proceed with delayed write").Base(err)
		}
		c.remoteAddr = c.conn.RemoteAddr()
		c.shouldWait = false
		return len(b), nil
	}
	return c.conn.Write(b)
}

func (c *connection) WriteMultiBuffer(mb buf.MultiBuffer) error {
	mb = buf.Compact(mb)
	mb, err := buf.WriteMultiBuffer(c, mb)
	buf.ReleaseMulti(mb)
	return err
}

func (c *connection) Close() error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			return newError("unable to close delayed dial websocket connection as it do not exist")
		}
	}
	var closeErrors []interface{}
	if err := c.conn.Close(); err != nil {
		closeErrors = append(closeErrors, err)
	}
	if len(closeErrors) > 0 {
		return newError("failed to close connection").Base(newError(serial.Concat(closeErrors...)))
	}
	return nil
}

func (c *connection) LocalAddr() net.Addr {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			newError("websocket transport is not materialized when LocalAddr() is called").AtWarning().WriteToLog()
			return &net.UnixAddr{
				Name: "@placeholder",
				Net:  "unix",
			}
		}
	}
	return c.conn.LocalAddr()
}

func (c *connection) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *connection) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			newError("httpupgrade transport is not materialized when SetReadDeadline() is called").AtWarning().WriteToLog()
			return nil
		}
	}
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			newError("httpupgrade transport is not materialized when SetWriteDeadline() is called").AtWarning().WriteToLog()
			return nil
		}
	}
	return c.conn.SetWriteDeadline(t)
}
