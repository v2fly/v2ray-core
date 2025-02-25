package server

import (
	"bytes"
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

	c2sReminder []byte
}

func (s *Server) accept(clientConn net.Conn, serverConn net.Conn) {
	ctx, done := context.WithCancel(context.Background())
	c := &Conn{
		ctx:         ctx,
		done:        done,
		clientConn:  clientConn,
		serverConn:  serverConn,
		c2sReminder: nil,
	}
	c.c2sHandshake()
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

func (c *Conn) c2sHandshake() {
	var readBuffer [65536]byte
	var nextRead int
	var overallLength int
	for c.ctx.Err() == nil {
		n, err := c.clientConn.Read(readBuffer[nextRead:])
		if err != nil {
			c.done()
			return
		}
		result, tryAgainLen, processed, err := mirrorcommon.PeekTLSRecord(&bufPeeker{buffer: readBuffer[:nextRead+n]})
		if tryAgainLen == 0 && processed == 0 {
			panic("todo")
		}
		if err != nil {
			_, err := io.Copy(c.serverConn, bytes.NewReader(readBuffer[nextRead:n]))
			if err != nil {
				panic("todo")
				return
			}
			nextRead = nextRead + n
			continue
		}

		if result.RecordType == mirrorcommon.TLSRecord_RecordType_application_data {
			panic("todo")
		}
		overallLength += processed
	}
}
