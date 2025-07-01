package mirrorbase

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
)

// NewMirroredTLSConn creates a new mirrored TLS connection.
// No stable interface
func NewMirroredTLSConn(ctx context.Context, clientConn net.Conn, serverConn net.Conn, onC2SMessage, onS2CMessage tlsmirror.MessageHook, closable common.Closable) tlsmirror.InsertableTLSConn {
	c := &conn{
		ctx:          ctx,
		clientConn:   clientConn,
		serverConn:   serverConn,
		c2sInsert:    make(chan *tlsmirror.TLSRecord, 100),
		s2cInsert:    make(chan *tlsmirror.TLSRecord, 100),
		OnC2SMessage: onC2SMessage,
		OnS2CMessage: onS2CMessage,
	}
	c.ctx, c.done = context.WithCancel(ctx)
	go c.c2sWorker()
	go c.s2cWorker()
	return c
}

type conn struct {
	ctx  context.Context
	done context.CancelFunc

	clientConn net.Conn
	serverConn net.Conn

	OnC2SMessage tlsmirror.MessageHook
	OnS2CMessage tlsmirror.MessageHook

	c2sInsert chan *tlsmirror.TLSRecord
	s2cInsert chan *tlsmirror.TLSRecord

	isClientRandomReady bool
	ClientRandom        [32]byte
	isServerRandomReady bool
	ServerRandom        [32]byte
}

func (c *conn) GetHandshakeRandom() ([]byte, []byte, error) {
	if !c.isClientRandomReady || !c.isServerRandomReady {
		return nil, nil, newError("client random or server random not ready")
	}
	return c.ClientRandom[:], c.ServerRandom[:], nil
}

func (c *conn) Close() error {
	c.done()
	return nil
}

func (c *conn) InsertC2SMessage(message *tlsmirror.TLSRecord) error {
	duplicatedRecord := mirrorcommon.DuplicateRecord(*message)
	c.c2sInsert <- &duplicatedRecord
	return nil
}

func (c *conn) InsertS2CMessage(message *tlsmirror.TLSRecord) error {
	duplicatedRecord := mirrorcommon.DuplicateRecord(*message)
	c.s2cInsert <- &duplicatedRecord
	return nil
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

func (r *readerWithInitialData) initialDataDrained() bool {
	return len(r.initialData) == 0
}

func (r *readerWithInitialData) Read(p []byte) (n int, err error) {
	if len(r.initialData) > 0 {
		n = copy(p, r.initialData)
		r.initialData = r.initialData[n:]
		return n, nil
	}
	return r.innerReader.Read(p)
}

func (c *conn) c2sWorker() {
	c2sHandshake, handshakeReminder, c2sReminderData, err := c.captureHandshake(c.clientConn, c.serverConn)
	if err != nil {
		c.done()
		return
	}
	_ = c2sHandshake
	_ = handshakeReminder
	_ = c2sReminderData
	serverConnectionWriter := bufio.NewWriter(c.serverConn)
	_, err = io.Copy(serverConnectionWriter, bytes.NewReader(handshakeReminder))
	if err != nil {
		c.done()
		newError("failed to copy handshake reminder").Base(err).AtWarning().WriteToLog()
		return
	}

	clientHello, err := mirrorcommon.UnpackTLSClientHello(c2sHandshake.Fragment)
	if err != nil {
		c.done()
		newError("failed to unpack client hello").Base(err).AtWarning().WriteToLog()
		return
	}
	c.ClientRandom = clientHello.ClientRandom
	c.isClientRandomReady = true

	clientSocketReader := &readerWithInitialData{initialData: c2sReminderData, innerReader: c.clientConn}
	clientSocket := bufio.NewReaderSize(clientSocketReader, 65536)

	recordReader := mirrorcommon.NewTLSRecordStreamReader(clientSocket)
	recordWriter := mirrorcommon.NewTLSRecordStreamWriter(serverConnectionWriter)
	if len(c2sReminderData) == 0 {
		err := serverConnectionWriter.Flush()
		if err != nil {
			c.done()
			newError("failed to flush server connection writer").Base(err).AtWarning().WriteToLog()
			return
		}
	}
	go func() {
		for c.ctx.Err() == nil {
			record := <-c.c2sInsert
			err := recordWriter.WriteRecord(record, false)
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

func (c *conn) s2cWorker() {
	// TODO: stick packets together, if they arrived so
	s2cHandshake, handshakeReminder, s2cReminderData, err := c.captureHandshake(c.serverConn, c.clientConn)
	if err != nil {
		c.done()
		return
	}
	_ = s2cHandshake
	_ = handshakeReminder
	_ = s2cReminderData

	clientConnectionWriter := bufio.NewWriter(c.clientConn)

	_, err = io.Copy(clientConnectionWriter, bytes.NewReader(handshakeReminder))
	if err != nil {
		c.done()
		newError("failed to copy handshake reminder").Base(err).AtWarning().WriteToLog()
		return
	}

	serverHello, err := mirrorcommon.UnpackTLSServerHello(s2cHandshake.Fragment)
	if err != nil {
		c.done()
		newError("failed to unpack server hello").Base(err).AtWarning().WriteToLog()
		return
	}
	c.ServerRandom = serverHello.ServerRandom
	c.isServerRandomReady = true

	serverSocketReader := &readerWithInitialData{initialData: s2cReminderData, innerReader: c.serverConn}
	serverSocket := bufio.NewReaderSize(serverSocketReader, 65536)
	recordReader := mirrorcommon.NewTLSRecordStreamReader(serverSocket)
	recordWriter := mirrorcommon.NewTLSRecordStreamWriter(clientConnectionWriter)

	if len(s2cReminderData) == 0 {
		err := clientConnectionWriter.Flush()
		if err != nil {
			c.done()
			newError("failed to flush client connection writer").Base(err).AtWarning().WriteToLog()
			return
		}
	}
	go func() {
		for c.ctx.Err() == nil {
			record := <-c.s2cInsert
			err := recordWriter.WriteRecord(record, false)
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

func (c *conn) captureHandshake(sourceConn, mirrorConn net.Conn) (handshake tlsmirror.TLSRecord, handshakeReminder, rest []byte, reterr error) {
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
