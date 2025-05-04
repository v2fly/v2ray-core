package server

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
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

	OnC2SMessage func(message *tlsmirror.TLSRecord) (drop bool, ok error)
	OnS2CMessage func(message *tlsmirror.TLSRecord) (drop bool, ok error)

	c2sInsert chan *tlsmirror.TLSRecord
	s2cInsert chan *tlsmirror.TLSRecord
}

type InsertableTLSConn interface {
	InsertC2SMessage(message *tlsmirror.TLSRecord) error
	InsertS2CMessage(message *tlsmirror.TLSRecord) error
}

func (s *Server) accept(clientConn net.Conn, serverConn net.Conn) {
	ctx, done := context.WithCancel(context.Background())
	c := &Conn{
		ctx:        ctx,
		done:       done,
		clientConn: clientConn,
		serverConn: serverConn,
		c2sInsert:  make(chan *tlsmirror.TLSRecord),
		s2cInsert:  make(chan *tlsmirror.TLSRecord),
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

type readerWithInitialData struct {
	initialData []byte
	innerReader io.Reader
}

func (r *readerWithInitialData) Read(p []byte) (n int, err error) {
	if len(r.initialData) > 0 {
		n = copy(p, r.initialData)
		r.initialData = r.initialData[n:]
		return
	}
	return r.innerReader.Read(p)
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
	_, err = io.Copy(c.serverConn, bytes.NewReader(handshakeReminder))
	if err != nil {
		c.done()
		newError("failed to copy handshake reminder").Base(err).AtWarning().WriteToLog()
		return
	}

	clientSocketReader := readerWithInitialData{initialData: c2sReminderData, innerReader: c.clientConn}
	clientSocket := bufio.NewReader(&clientSocketReader)

	recordReader := mirrorcommon.NewTLSRecordStreamReader(clientSocket)
	recordWriter := mirrorcommon.NewTLSRecordStreamWriter(bufio.NewWriter(c.serverConn))
	go func() {
		for c.ctx.Err() == nil {
			record := <-c.c2sInsert
			err := recordWriter.WriteRecord(record)
			if err != nil {
				c.done()
				newError("failed to write C2S message").Base(err).AtWarning().WriteToLog()
				return
			}
		}
	}()
	for c.ctx.Err() == nil {
		record, err := recordReader.ReadNextRecord()
		if err != nil {
			drainCopy(c.clientConn, nil, c.serverConn)
			c.done()
			newError("failed to read TLS record").Base(err).AtWarning().WriteToLog()
			return
		}
		if c.OnC2SMessage != nil {
			drop, err := c.OnC2SMessage(record)
			if err != nil {
				c.done()
				newError("failed to process C2S message").Base(err).AtWarning().WriteToLog()
				return
			}
			if drop {
				continue
			}
		}
		duplicatedRecord := mirrorcommon.DuplicateRecord(*record)
		c.c2sInsert <- &duplicatedRecord
	}
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

	_, err = io.Copy(c.clientConn, bytes.NewReader(handshakeReminder))
	if err != nil {
		c.done()
		newError("failed to copy handshake reminder").Base(err).AtWarning().WriteToLog()
		return
	}

	serverSocketReader := readerWithInitialData{initialData: s2cReminderData, innerReader: c.serverConn}
	serverSocket := bufio.NewReader(&serverSocketReader)
	recordReader := mirrorcommon.NewTLSRecordStreamReader(serverSocket)
	recordWriter := mirrorcommon.NewTLSRecordStreamWriter(bufio.NewWriter(c.clientConn))
	go func() {
		for c.ctx.Err() == nil {
			record := <-c.s2cInsert
			err := recordWriter.WriteRecord(record)
			if err != nil {
				c.done()
				newError("failed to write S2C message").Base(err).AtWarning().WriteToLog()
				return
			}
		}
	}()
	for c.ctx.Err() == nil {
		record, err := recordReader.ReadNextRecord()
		if err != nil {
			drainCopy(c.clientConn, nil, c.serverConn)
			c.done()
			newError("failed to read TLS record").Base(err).AtWarning().WriteToLog()
			return
		}
		if c.OnS2CMessage != nil {
			drop, err := c.OnS2CMessage(record)
			if err != nil {
				c.done()
				newError("failed to process S2C message").Base(err).AtWarning().WriteToLog()
				return
			}
			if drop {
				continue
			}
		}
		duplicatedRecord := mirrorcommon.DuplicateRecord(*record)
		c.s2cInsert <- &duplicatedRecord
	}
}

func drainCopy(dst io.Writer, initData []byte, src io.Reader) {
	if initData != nil {
		_, err := io.Copy(dst, bytes.NewReader(initData))
		if err != nil {
			newError("failed to drain copy").Base(err).AtWarning().WriteToLog()
		}
		return
	}
	_, err := io.Copy(dst, src)
	if err != nil {
		newError("failed to drain copy").Base(err).AtWarning().WriteToLog()
	}
}

type rejectionDecisionMaker struct {
}

func (r *rejectionDecisionMaker) TestIfReject(record *tlsmirror.TLSRecord, readyFields int) error {
	if readyFields >= 1 {
		if record.RecordType != mirrorcommon.TLSRecord_RecordType_handshake {
			return newError("unexpected record type").AtWarning()
		}
	}
	if readyFields >= 2 {
		switch record.LegacyProtocolVersion[0] {
		case 0x01:
		case 0x02:
		case 0x03:
			if record.LegacyProtocolVersion[1] > 0x03 {
				return newError("unexpected minor protocol version").AtWarning()
			}
		default:
			return newError("unexpected major protocol version").AtWarning()
		}
	}
	return nil
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
		handshakeRejectionDecisionMaker := &rejectionDecisionMaker{}
		result, tryAgainLen, processed, err := mirrorcommon.PeekTLSRecord(&bufPeeker{buffer: readBuffer[:nextRead+n]}, handshakeRejectionDecisionMaker)
		if processed == 0 {
			if tryAgainLen == 0 {
				// TODO: directly copy
				drainCopy(mirrorConn, readBuffer[:nextRead+n], sourceConn)
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
			handshakeReminder = readBuffer[nextRead:processed]
			rest = readBuffer[processed : nextRead+n]
			reterr = nil
			return
		}
	}
	reterr = newError("context is done")
	return
}
