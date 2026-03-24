package packetToStream

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/xtaci/smux"
)

const packetSequenceFieldSize = 8
const packetLengthFieldSize = 4
const packetHeaderSize = packetSequenceFieldSize + packetLengthFieldSize

var sessionPacketConnCloseDrainTimeout = 2 * time.Second

// Adaptor consumes the rrpit receive callback and exposes an smux session on top
// of a length-prefixed byte stream carried by rrpit packets.
type Adaptor struct {
	smux       *smux.Session
	session    *rrpitBidirectionalSession.BidirectionalSession
	packetConn *sessionPacketConn
}

type sessionPacketConn struct {
	session *rrpitBidirectionalSession.BidirectionalSession

	mu          sync.Mutex
	cond        *sync.Cond
	readBuf     bytes.Buffer
	frameBuf    []byte
	pending     map[uint64][]byte
	nextSendSeq uint64
	nextRecvSeq uint64
	closeOnce   sync.Once
	closing     bool
	closed      bool
	closeErr    error
	localAddr   net.Addr
	remoteAddr  net.Addr
}

func New(session *rrpitBidirectionalSession.BidirectionalSession, client bool, config *smux.Config) (*Adaptor, error) {
	if session == nil {
		return nil, fmt.Errorf("nil bidirectional session")
	}
	packetConn := newSessionPacketConn(session)
	if session.Rx() == nil {
		return nil, fmt.Errorf("nil rx session")
	}
	session.Rx().OnMessage = packetConn.OnMessage

	var (
		smuxSession *smux.Session
		err         error
	)
	if client {
		smuxSession, err = smux.Client(packetConn, config)
	} else {
		smuxSession, err = smux.Server(packetConn, config)
	}
	if err != nil {
		_ = packetConn.Close()
		return nil, err
	}

	return &Adaptor{
		smux:       smuxSession,
		session:    session,
		packetConn: packetConn,
	}, nil
}

func NewClient(session *rrpitBidirectionalSession.BidirectionalSession, config *smux.Config) (*Adaptor, error) {
	return New(session, true, config)
}

func NewServer(session *rrpitBidirectionalSession.BidirectionalSession, config *smux.Config) (*Adaptor, error) {
	return New(session, false, config)
}

func (a *Adaptor) Session() *smux.Session {
	if a == nil {
		return nil
	}
	return a.smux
}

func (a *Adaptor) OpenStream() (*smux.Stream, error) {
	if a == nil || a.smux == nil {
		return nil, io.ErrClosedPipe
	}
	return a.smux.OpenStream()
}

func (a *Adaptor) AcceptStream() (*smux.Stream, error) {
	if a == nil || a.smux == nil {
		return nil, io.ErrClosedPipe
	}
	return a.smux.AcceptStream()
}

func (a *Adaptor) Close() error {
	if a == nil {
		return nil
	}

	var firstErr error
	if a.smux != nil {
		if err := a.smux.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.packetConn != nil {
		if err := a.packetConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.session != nil {
		if err := a.session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func newSessionPacketConn(session *rrpitBidirectionalSession.BidirectionalSession) *sessionPacketConn {
	conn := &sessionPacketConn{
		session:    session,
		localAddr:  adaptorAddr("rrpit-local"),
		remoteAddr: adaptorAddr("rrpit-remote"),
		pending:    make(map[uint64][]byte),
	}
	conn.cond = sync.NewCond(&conn.mu)
	return conn
}

func (c *sessionPacketConn) OnMessage(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.frameBuf = append(c.frameBuf, data...)
	for {
		if len(c.frameBuf) < packetHeaderSize {
			break
		}

		seq := binary.BigEndian.Uint64(c.frameBuf[:packetSequenceFieldSize])
		payloadLen := binary.BigEndian.Uint32(c.frameBuf[packetSequenceFieldSize:packetHeaderSize])
		frameSize := packetHeaderSize + int(payloadLen)
		if len(c.frameBuf) < frameSize {
			break
		}

		payload := append([]byte(nil), c.frameBuf[packetHeaderSize:frameSize]...)
		c.acceptPacketLocked(seq, payload)
		c.frameBuf = append(c.frameBuf[:0], c.frameBuf[frameSize:]...)
	}
	c.cond.Broadcast()
	return nil
}

func (c *sessionPacketConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for c.readBuf.Len() == 0 && !c.closed {
		c.cond.Wait()
	}
	if c.readBuf.Len() == 0 && c.closed {
		if c.closeErr != nil {
			return 0, c.closeErr
		}
		return 0, io.EOF
	}
	n, err := c.readBuf.Read(p)
	if c.readBuf.Len() == 0 {
		c.cond.Broadcast()
	}
	return n, err
}

func (c *sessionPacketConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	closed := c.closed
	closing := c.closing
	closeErr := c.closeErr
	c.mu.Unlock()
	if closed || closing {
		if closeErr != nil {
			return 0, closeErr
		}
		return 0, io.ErrClosedPipe
	}
	if len(p) > int(^uint32(0)) {
		return 0, fmt.Errorf("packet too large")
	}
	if len(p) == 0 {
		return 0, nil
	}
	maxFragmentPayload, err := c.maxFragmentPayload()
	if err != nil {
		c.fail(err)
		return 0, err
	}
	if c.session == nil {
		return 0, io.ErrClosedPipe
	}

	written := 0
	for offset := 0; offset < len(p); offset += maxFragmentPayload {
		end := offset + maxFragmentPayload
		if end > len(p) {
			end = len(p)
		}

		c.mu.Lock()
		seq := c.nextSendSeq
		c.nextSendSeq += 1
		c.mu.Unlock()

		frame := make([]byte, packetHeaderSize+(end-offset))
		binary.BigEndian.PutUint64(frame[:packetSequenceFieldSize], seq)
		binary.BigEndian.PutUint32(frame[packetSequenceFieldSize:packetHeaderSize], uint32(end-offset))
		copy(frame[packetHeaderSize:], p[offset:end])

		if err := c.session.SendMessage(frame); err != nil {
			c.fail(err)
			return written, err
		}
		written = end
	}
	return written, nil
}

func (c *sessionPacketConn) acceptPacketLocked(seq uint64, payload []byte) {
	if seq < c.nextRecvSeq {
		return
	}
	if _, found := c.pending[seq]; found {
		return
	}
	c.pending[seq] = payload
	for {
		current, found := c.pending[c.nextRecvSeq]
		if !found {
			return
		}
		delete(c.pending, c.nextRecvSeq)
		_, _ = c.readBuf.Write(current)
		c.nextRecvSeq += 1
	}
}

func (c *sessionPacketConn) Close() error {
	c.waitForBufferedPayloadBeforeClose()
	c.closeWithError(nil)
	return nil
}

func (c *sessionPacketConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *sessionPacketConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *sessionPacketConn) SetDeadline(time.Time) error {
	return nil
}

func (c *sessionPacketConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *sessionPacketConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (c *sessionPacketConn) fail(err error) {
	if err == nil {
		err = io.ErrClosedPipe
	}
	c.closeWithError(err)
}

func (c *sessionPacketConn) maxFragmentPayload() (int, error) {
	if c.session == nil {
		return 0, io.ErrClosedPipe
	}
	maxMessageSize, err := c.session.MaxMessageSize()
	if err != nil {
		return 0, err
	}
	maxFragmentPayload := maxMessageSize - packetHeaderSize
	if maxFragmentPayload <= 0 {
		return 0, fmt.Errorf("rrpit max message size %d is too small for adaptor header", maxMessageSize)
	}
	return maxFragmentPayload, nil
}

func (c *sessionPacketConn) closeWithError(err error) {
	c.closeOnce.Do(func() {
		if err != nil {
			c.closeErr = err
		}
		c.mu.Lock()
		c.closing = true
		c.closed = true
		c.cond.Broadcast()
		c.mu.Unlock()
	})
}

func (c *sessionPacketConn) hasBufferedPayloadLocked() bool {
	return c.readBuf.Len() > 0 || len(c.pending) > 0 || len(c.frameBuf) > 0
}

func (c *sessionPacketConn) waitForBufferedPayloadBeforeClose() {
	if c == nil || sessionPacketConnCloseDrainTimeout <= 0 {
		return
	}

	c.mu.Lock()
	if c.closed || c.closing || !c.hasBufferedPayloadLocked() {
		if !c.closed {
			c.closing = true
		}
		c.mu.Unlock()
		return
	}
	c.closing = true

	timedOut := false
	timer := time.AfterFunc(sessionPacketConnCloseDrainTimeout, func() {
		c.mu.Lock()
		timedOut = true
		c.cond.Broadcast()
		c.mu.Unlock()
	})
	for !c.closed && c.hasBufferedPayloadLocked() && !timedOut {
		c.cond.Wait()
	}
	if !timer.Stop() && !timedOut {
		timedOut = true
	}
	c.mu.Unlock()
}

type adaptorAddr string

func (a adaptorAddr) Network() string { return "rrpit" }
func (a adaptorAddr) String() string  { return string(a) }
