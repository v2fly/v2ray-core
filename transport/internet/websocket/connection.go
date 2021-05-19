// +build !confonly

package websocket

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/errors"
	"github.com/v2fly/v2ray-core/v4/common/serial"
)

var _ buf.Writer = (*connection)(nil)

// connection is a wrapper for net.Conn over WebSocket connection.
type connection struct {
	conn       *websocket.Conn
	reader     io.Reader
	remoteAddr net.Addr

	shouldWait        bool
	delayedDialFinish context.Context
	finishedDial      context.CancelFunc
	dialer            DelayedDialer
}

type DelayedDialer interface {
	Dial(earlyData []byte) (*websocket.Conn, error)
}

func newConnection(conn *websocket.Conn, remoteAddr net.Addr) *connection {
	return &connection{
		conn:       conn,
		remoteAddr: remoteAddr,
	}
}

func newConnectionWithEarlyData(conn *websocket.Conn, remoteAddr net.Addr, earlyData io.Reader) *connection {
	return &connection{
		conn:       conn,
		remoteAddr: remoteAddr,
		reader:     earlyData,
	}
}

func newConnectionWithDelayedDial(dialer DelayedDialer) *connection {
	delayedDialContext, CancellFunc := context.WithCancel(context.Background())
	return &connection{
		shouldWait:        true,
		delayedDialFinish: delayedDialContext,
		finishedDial:      CancellFunc,
		dialer:            dialer,
	}
}

func newRelayedConnectionWithDelayedDial(dialer DelayedDialerForwarded) *connectionForwarder {
	delayedDialContext, CancellFunc := context.WithCancel(context.Background())
	return &connectionForwarder{
		shouldWait:        true,
		delayedDialFinish: delayedDialContext,
		finishedDial:      CancellFunc,
		dialer:            dialer,
	}
}

func newRelayedConnection(conn io.ReadWriteCloser) *connectionForwarder {
	return &connectionForwarder{
		ReadWriteCloser: conn,
		shouldWait:      false,
	}
}

// Read implements net.Conn.Read()
func (c *connection) Read(b []byte) (int, error) {
	for {
		reader, err := c.getReader()
		if err != nil {
			return 0, err
		}

		nBytes, err := reader.Read(b)
		if errors.Cause(err) == io.EOF {
			c.reader = nil
			continue
		}
		return nBytes, err
	}
}

func (c *connection) getReader() (io.Reader, error) {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			return nil, newError("unable to read delayed dial websocket connection as it do not exist")
		}
	}
	if c.reader != nil {
		return c.reader, nil
	}

	_, reader, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	c.reader = reader
	return reader, nil
}

// Write implements io.Writer.
func (c *connection) Write(b []byte) (int, error) {
	if c.shouldWait {
		var err error
		c.conn, err = c.dialer.Dial(b)
		c.finishedDial()
		if err != nil {
			return 0, newError("Unable to proceed with delayed write").Base(err)
		}
		c.remoteAddr = c.conn.RemoteAddr()
		c.shouldWait = false
		return len(b), nil
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
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
	var errors []interface{}
	if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*5)); err != nil {
		errors = append(errors, err)
	}
	if err := c.conn.Close(); err != nil {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return newError("failed to close connection").Base(newError(serial.Concat(errors...)))
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
			newError("websocket transport is not materialized when SetReadDeadline() is called").AtWarning().WriteToLog()
			return nil
		}
	}
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			newError("websocket transport is not materialized when SetWriteDeadline() is called").AtWarning().WriteToLog()
			return nil
		}
	}
	return c.conn.SetWriteDeadline(t)
}
