package server

import (
	"bytes"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"io"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type Server struct {
	config *Config

	listener net.Listener
	handler  internet.ConnHandler

	ctx context.Context
}

func (s *Server) process(conn net.Conn) {
	transportEnvironment := envctx.EnvironmentFromContext(s.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	port, err := net.PortFromInt(s.config.ForwardPort)
	if err != nil {
		newError("failed to parse port").Base(err).AtWarning().WriteToLog()
		return
	}

	address := net.ParseAddress(s.config.ForwardAddress)

	dest := net.TCPDestination(address, port)

	forwardConn, err := dialer(s.ctx, dest, s.config.ForwardTag)
	if err != nil {
		newError("failed to dial to destination").Base(err).AtWarning().WriteToLog()
		return
	}

	s.accept(conn, forwardConn)
}

func (s *Server) accepts() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "closed") {
				break
			}
			newError("failed to accepted raw connections").Base(err).AtWarning().WriteToLog()
			if strings.Contains(errStr, "too many") {
				time.Sleep(time.Millisecond * 500)
			}
			continue
		}
		go s.process(conn)
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

type Conn struct {
	ctx  context.Context
	done context.CancelFunc

	clientConn net.Conn
	serverConn net.Conn
}

func (s *Server) accept(clientConn net.Conn, serverConn net.Conn) {
	ctx, done := context.WithCancel(context.Background())
	c := &Conn{
		ctx:        ctx,
		done:       done,
		clientConn: clientConn,
		serverConn: serverConn,
	}
	go c.c2sWorker()
	go c.s2cWorker()
}

type bufPeeker struct {
	buffer []byte
}

func (b *bufPeeker) Peek(n int) ([]byte, error) {
	if len(b.buffer) < n {
		return nil, newError("not enough data")
	}
	return b.buffer[:n], nil
}

func (c *Conn) c2sWorker() {
	c2sHandshake, handshakeReminder, c2sReminderData, err := c.captureHandshake(c.clientConn, c.serverConn)
	if err != nil {
		c.done()
		return
	}
	_ = c2sHandshake
	_ = handshakeReminder
	_ = c2sReminderData
}

func (c *Conn) s2cWorker() {
	s2cHandshake, handshakeReminder, s2cReminderData, err := c.captureHandshake(c.serverConn, c.clientConn)
	if err != nil {
		c.done()
		return
	}
	_ = s2cHandshake
	_ = handshakeReminder
	_ = s2cReminderData
}

func (c *Conn) captureHandshake(sourceConn, mirrorConn net.Conn) (handshake tlsmirror.TLSRecord, handshakeReminder, rest []byte, reterr error) {
	var readBuffer [65536]byte
	var nextRead int
	for c.ctx.Err() == nil {
		n, err := sourceConn.Read(readBuffer[nextRead:])
		if err != nil {
			c.done()
			return
		}
		result, tryAgainLen, processed, err := mirrorcommon.PeekTLSRecord(&bufPeeker{buffer: readBuffer[:nextRead+n]})
		if processed == 0 {
			if tryAgainLen == 0 {
				// TODO: directly copy
				c.done()
				err = newError("failed to peek tls record").Base(err).AtWarning()
				return
			}
			_, err = io.Copy(mirrorConn, bytes.NewReader(readBuffer[nextRead:nextRead+n]))
			if err != nil {
				c.done()
				newError("failed to copy to server connection").Base(err).AtWarning().WriteToLog()
				return
			}
			nextRead += n
		} else {
			// Parse the client hello
			if result.RecordType != mirrorcommon.TLSRecord_RecordType_handshake {
				c.done()
				err = newError("unexpected record type").AtWarning()
				return
			}
			handshake = result
			handshakeReminder = readBuffer[nextRead : nextRead+processed]
			rest = readBuffer[0:processed]
			reterr = nil
			return
		}
	}
	reterr = newError("context is done")
	return
}
