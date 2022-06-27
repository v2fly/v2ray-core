package websocket

import (
	"context"
	"io"
	"net"
	"time"
)

type connectionForwarder struct {
	io.ReadWriteCloser

	shouldWait        bool
	delayedDialFinish context.Context
	finishedDial      context.CancelFunc
	dialer            DelayedDialerForwarded
}

func (c *connectionForwarder) Read(p []byte) (n int, err error) {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.ReadWriteCloser == nil {
			return 0, newError("unable to read delayed dial websocket connection as it do not exist")
		}
	}
	return c.ReadWriteCloser.Read(p)
}

func (c *connectionForwarder) Write(p []byte) (n int, err error) {
	if c.shouldWait {
		var err error
		c.ReadWriteCloser, err = c.dialer.Dial(p)
		c.finishedDial()
		if err != nil {
			return 0, newError("Unable to proceed with delayed write").Base(err)
		}
		c.shouldWait = false
		return len(p), nil
	}
	return c.ReadWriteCloser.Write(p)
}

func (c *connectionForwarder) Close() error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.ReadWriteCloser == nil {
			return newError("unable to close delayed dial websocket connection as it do not exist")
		}
	}
	return c.ReadWriteCloser.Close()
}

func (c connectionForwarder) LocalAddr() net.Addr {
	return &net.UnixAddr{
		Name: "not available",
		Net:  "",
	}
}

func (c connectionForwarder) RemoteAddr() net.Addr {
	return &net.UnixAddr{
		Name: "not available",
		Net:  "",
	}
}

func (c connectionForwarder) SetDeadline(t time.Time) error {
	return nil
}

func (c connectionForwarder) SetReadDeadline(t time.Time) error {
	return nil
}

func (c connectionForwarder) SetWriteDeadline(t time.Time) error {
	return nil
}

type DelayedDialerForwarded interface {
	Dial(earlyData []byte) (io.ReadWriteCloser, error)
}
