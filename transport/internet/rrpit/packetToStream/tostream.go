package packetToStream

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xtaci/smux"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
)

const (
	adaptorFrameIDFieldSize        = 8
	adaptorStreamIDFieldSize       = 4
	adaptorStreamFrameSeqFieldSize = 8
	adaptorSmuxCmdFieldSize        = 1
	adaptorSmuxVersionFieldSize    = 1
	adaptorHeaderSize              = adaptorFrameIDFieldSize + adaptorStreamIDFieldSize + adaptorStreamFrameSeqFieldSize + adaptorSmuxCmdFieldSize + adaptorSmuxVersionFieldSize

	smuxVersionFieldOffset  = 0
	smuxCmdFieldOffset      = 1
	smuxLengthFieldOffset   = 2
	smuxStreamIDFieldOffset = 4
	smuxFrameHeaderSize     = 8
)

var sessionPacketConnCloseDrainTimeout = 2 * time.Second

func MaxSmuxFrameSizeForMessage(maxMessageSize int) int {
	maxFrameSize := maxMessageSize - adaptorHeaderSize - smuxFrameHeaderSize
	if maxFrameSize <= 0 {
		return 0
	}
	return maxFrameSize
}

// Adaptor consumes the rrpit receive callback and exposes an smux session on top.
type Adaptor struct {
	smux       *smux.Session
	session    *rrpitBidirectionalSession.BidirectionalSession
	packetConn *sessionPacketConn
}

type adaptorFrame struct {
	frameID        uint64
	streamID       uint32
	streamFrameSeq uint64
	smuxCmd        byte
	smuxVersion    byte
	payload        []byte
}

type sessionPacketConn struct {
	session *rrpitBidirectionalSession.BidirectionalSession

	mu sync.Mutex
	// cond protects all mutable fields below.
	cond *sync.Cond

	readBuf bytes.Buffer

	nextSendFrameID         uint64
	nextSendStreamFrameSeq  map[uint32]uint64
	nextExpectedStreamSeq   map[uint32]uint64
	readyFramesByStream     map[uint32]map[uint64]*adaptorFrame
	activeStreams           []uint32
	activeStreamSet         map[uint32]bool
	locallyKnownStreams     map[uint32]bool
	remoteSynEstablished    map[uint32]bool
	roundRobinIndex         int
	maxSerializedFrameBytes int

	closeOnce  sync.Once
	closing    bool
	closed     bool
	closeErr   error
	localAddr  net.Addr
	remoteAddr net.Addr
}

func New(session *rrpitBidirectionalSession.BidirectionalSession, client bool, config *smux.Config) (*Adaptor, error) {
	if session == nil {
		return nil, fmt.Errorf("nil bidirectional session")
	}
	if session.Rx() == nil {
		return nil, fmt.Errorf("nil rx session")
	}
	if config == nil {
		config = smux.DefaultConfig()
	}
	maxSerializedFrameBytes, err := validateSmuxFrameSize(session, config)
	if err != nil {
		return nil, err
	}

	packetConn := newSessionPacketConn(session, maxSerializedFrameBytes)
	session.Rx().OnMessage = packetConn.OnMessage

	var smuxSession *smux.Session
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

func newSessionPacketConn(session *rrpitBidirectionalSession.BidirectionalSession, maxSerializedFrameBytes int) *sessionPacketConn {
	conn := &sessionPacketConn{
		session:                 session,
		nextSendStreamFrameSeq:  make(map[uint32]uint64),
		nextExpectedStreamSeq:   make(map[uint32]uint64),
		readyFramesByStream:     make(map[uint32]map[uint64]*adaptorFrame),
		activeStreamSet:         make(map[uint32]bool),
		locallyKnownStreams:     make(map[uint32]bool),
		remoteSynEstablished:    make(map[uint32]bool),
		maxSerializedFrameBytes: maxSerializedFrameBytes,
		localAddr:               adaptorAddr("rrpit-local"),
		remoteAddr:              adaptorAddr("rrpit-remote"),
	}
	conn.cond = sync.NewCond(&conn.mu)
	return conn
}

func (c *sessionPacketConn) OnMessage(data []byte) error {
	frame, err := decodeAdaptorFrame(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	if frame.streamFrameSeq < c.nextExpectedStreamSeq[frame.streamID] {
		return nil
	}
	streamFrames := c.readyFramesByStream[frame.streamID]
	if streamFrames == nil {
		streamFrames = make(map[uint64]*adaptorFrame)
		c.readyFramesByStream[frame.streamID] = streamFrames
	}
	if _, found := streamFrames[frame.streamFrameSeq]; found {
		return nil
	}
	streamFrames[frame.streamFrameSeq] = frame
	c.activateStreamLocked(frame.streamID)
	c.deliverReadyFramesLocked()
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
	if len(p) == 0 {
		return 0, nil
	}
	if len(p) > c.maxSerializedFrameBytes {
		err := fmt.Errorf("serialized smux frame size %d exceeds rrpit adaptor budget %d", len(p), c.maxSerializedFrameBytes)
		c.fail(err)
		return 0, err
	}
	if c.session == nil {
		return 0, io.ErrClosedPipe
	}

	smuxFrame, err := parseSmuxFrame(p)
	if err != nil {
		c.fail(err)
		return 0, err
	}

	c.mu.Lock()
	frameID := c.nextSendFrameID
	c.nextSendFrameID++
	streamFrameSeq := c.nextSendStreamFrameSeq[smuxFrame.streamID]
	c.nextSendStreamFrameSeq[smuxFrame.streamID] = streamFrameSeq + 1
	if smuxFrame.streamID != 0 {
		c.locallyKnownStreams[smuxFrame.streamID] = true
	}
	c.mu.Unlock()

	wire := encodeAdaptorFrame(&adaptorFrame{
		frameID:        frameID,
		streamID:       smuxFrame.streamID,
		streamFrameSeq: streamFrameSeq,
		smuxCmd:        smuxFrame.cmd,
		smuxVersion:    smuxFrame.version,
		payload:        append([]byte(nil), p...),
	})
	if err := c.session.SendMessage(wire); err != nil {
		c.fail(err)
		return 0, err
	}
	return len(p), nil
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
	return c.readBuf.Len() > 0 || len(c.activeStreams) > 0
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

func (c *sessionPacketConn) activateStreamLocked(streamID uint32) {
	if c.activeStreamSet[streamID] {
		return
	}
	c.activeStreams = append(c.activeStreams, streamID)
	c.activeStreamSet[streamID] = true
}

func (c *sessionPacketConn) deactivateStreamAtLocked(index int) {
	if index < 0 || index >= len(c.activeStreams) {
		return
	}
	streamID := c.activeStreams[index]
	c.activeStreams = append(c.activeStreams[:index], c.activeStreams[index+1:]...)
	delete(c.activeStreamSet, streamID)
	if c.roundRobinIndex > index {
		c.roundRobinIndex--
	}
	if c.roundRobinIndex >= len(c.activeStreams) {
		c.roundRobinIndex = 0
	}
}

func (c *sessionPacketConn) deliverReadyFramesLocked() {
	for {
		if len(c.activeStreams) == 0 {
			return
		}

		progressed := false
		scans := len(c.activeStreams)
		for i := 0; i < scans && len(c.activeStreams) > 0; i++ {
			if c.roundRobinIndex >= len(c.activeStreams) {
				c.roundRobinIndex = 0
			}
			streamID := c.activeStreams[c.roundRobinIndex]
			frame, ok := c.nextDeliverableFrameLocked(streamID)
			if !ok {
				c.roundRobinIndex++
				continue
			}

			_, _ = c.readBuf.Write(frame.payload)
			delete(c.readyFramesByStream[streamID], frame.streamFrameSeq)
			c.nextExpectedStreamSeq[streamID] = frame.streamFrameSeq + 1
			if frame.smuxCmd == smuxCmdSYN {
				c.remoteSynEstablished[streamID] = true
			}
			if len(c.readyFramesByStream[streamID]) == 0 {
				delete(c.readyFramesByStream, streamID)
				c.deactivateStreamAtLocked(c.roundRobinIndex)
			} else {
				c.roundRobinIndex++
			}
			progressed = true
		}
		if !progressed {
			return
		}
	}
}

func (c *sessionPacketConn) nextDeliverableFrameLocked(streamID uint32) (*adaptorFrame, bool) {
	streamFrames := c.readyFramesByStream[streamID]
	if len(streamFrames) == 0 {
		return nil, false
	}
	seq := c.nextExpectedStreamSeq[streamID]
	frame, found := streamFrames[seq]
	if !found {
		return nil, false
	}
	if streamID == 0 {
		return frame, true
	}
	if frame.smuxCmd == smuxCmdSYN {
		return frame, true
	}
	if c.locallyKnownStreams[streamID] || c.remoteSynEstablished[streamID] {
		return frame, true
	}
	return nil, false
}

func validateSmuxFrameSize(session *rrpitBidirectionalSession.BidirectionalSession, config *smux.Config) (int, error) {
	if session == nil {
		return 0, io.ErrClosedPipe
	}
	maxMessageSize, err := session.MaxMessageSize()
	if err != nil {
		return 0, err
	}
	maxSerializedFrameBytes := maxMessageSize - adaptorHeaderSize
	if maxSerializedFrameBytes <= smuxFrameHeaderSize {
		return 0, fmt.Errorf("rrpit max message size %d is too small for adaptor header %d and smux header %d", maxMessageSize, adaptorHeaderSize, smuxFrameHeaderSize)
	}
	if config.Version == 2 && maxSerializedFrameBytes < smuxFrameHeaderSize+smuxCommandUPDLength {
		return 0, fmt.Errorf("rrpit max message size %d is too small for adaptor and smux control frames", maxMessageSize)
	}
	if config.MaxFrameSize+smuxFrameHeaderSize > maxSerializedFrameBytes {
		return 0, fmt.Errorf("smux max frame size %d exceeds rrpit adaptor budget %d", config.MaxFrameSize, maxSerializedFrameBytes-smuxFrameHeaderSize)
	}
	return maxSerializedFrameBytes, nil
}

func encodeAdaptorFrame(frame *adaptorFrame) []byte {
	wire := make([]byte, adaptorHeaderSize+len(frame.payload))
	binary.BigEndian.PutUint64(wire[:adaptorFrameIDFieldSize], frame.frameID)
	binary.BigEndian.PutUint32(wire[adaptorFrameIDFieldSize:adaptorFrameIDFieldSize+adaptorStreamIDFieldSize], frame.streamID)
	binary.BigEndian.PutUint64(wire[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize:adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize], frame.streamFrameSeq)
	wire[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize] = frame.smuxCmd
	wire[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize+adaptorSmuxCmdFieldSize] = frame.smuxVersion
	copy(wire[adaptorHeaderSize:], frame.payload)
	return wire
}

func decodeAdaptorFrame(data []byte) (*adaptorFrame, error) {
	if len(data) < adaptorHeaderSize+smuxFrameHeaderSize {
		return nil, fmt.Errorf("rrpit adaptor frame too short: %d", len(data))
	}
	frame := &adaptorFrame{
		frameID:        binary.BigEndian.Uint64(data[:adaptorFrameIDFieldSize]),
		streamID:       binary.BigEndian.Uint32(data[adaptorFrameIDFieldSize : adaptorFrameIDFieldSize+adaptorStreamIDFieldSize]),
		streamFrameSeq: binary.BigEndian.Uint64(data[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize : adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize]),
		smuxCmd:        data[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize],
		smuxVersion:    data[adaptorFrameIDFieldSize+adaptorStreamIDFieldSize+adaptorStreamFrameSeqFieldSize+adaptorSmuxCmdFieldSize],
		payload:        append([]byte(nil), data[adaptorHeaderSize:]...),
	}
	smuxFrame, err := parseSmuxFrame(frame.payload)
	if err != nil {
		return nil, err
	}
	if smuxFrame.streamID != frame.streamID || smuxFrame.cmd != frame.smuxCmd || smuxFrame.version != frame.smuxVersion {
		return nil, fmt.Errorf("rrpit adaptor metadata does not match smux frame header")
	}
	return frame, nil
}

type smuxFrameMetadata struct {
	version  byte
	cmd      byte
	streamID uint32
}

func parseSmuxFrame(data []byte) (*smuxFrameMetadata, error) {
	if len(data) < smuxFrameHeaderSize {
		return nil, fmt.Errorf("smux frame too short: %d", len(data))
	}
	version := data[smuxVersionFieldOffset]
	cmd := data[smuxCmdFieldOffset]
	length := int(binary.LittleEndian.Uint16(data[smuxLengthFieldOffset : smuxLengthFieldOffset+2]))
	streamID := binary.LittleEndian.Uint32(data[smuxStreamIDFieldOffset : smuxStreamIDFieldOffset+4])
	if len(data) != smuxFrameHeaderSize+length {
		return nil, fmt.Errorf("smux frame length mismatch: header=%d payload=%d", length, len(data)-smuxFrameHeaderSize)
	}
	if err := validateSmuxCommand(version, cmd, length, streamID); err != nil {
		return nil, err
	}
	return &smuxFrameMetadata{
		version:  version,
		cmd:      cmd,
		streamID: streamID,
	}, nil
}

func validateSmuxCommand(version byte, cmd byte, length int, streamID uint32) error {
	switch cmd {
	case smuxCmdSYN, smuxCmdFIN:
		if length != 0 {
			return fmt.Errorf("smux control frame %d must have empty payload", cmd)
		}
		if streamID == 0 {
			return fmt.Errorf("smux control frame %d requires non-zero stream id", cmd)
		}
	case smuxCmdPSH:
		if streamID == 0 {
			return fmt.Errorf("smux psh frame requires non-zero stream id")
		}
	case smuxCmdNOP:
		if length != 0 || streamID != 0 {
			return fmt.Errorf("smux nop frame must have zero stream id and empty payload")
		}
	case smuxCmdUPD:
		if version != 2 {
			return fmt.Errorf("smux upd frame requires version 2")
		}
		if length != smuxCommandUPDLength {
			return fmt.Errorf("smux upd frame must have %d-byte payload", smuxCommandUPDLength)
		}
		if streamID == 0 {
			return fmt.Errorf("smux upd frame requires non-zero stream id")
		}
	default:
		return fmt.Errorf("unsupported smux command %d", cmd)
	}
	return nil
}

type adaptorAddr string

func (a adaptorAddr) Network() string { return "rrpit" }
func (a adaptorAddr) String() string  { return string(a) }

const (
	smuxCmdSYN = byte(iota)
	smuxCmdFIN
	smuxCmdPSH
	smuxCmdNOP
	smuxCmdUPD
)

const smuxCommandUPDLength = 8
