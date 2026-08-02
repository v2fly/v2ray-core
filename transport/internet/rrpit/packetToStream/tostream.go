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
	adaptorStreamFrameSeqFieldSize = 8
	// SMUX already carries version, command, stream id, and payload length.
	// RRIPT only needs a per-stream sequence for independent reordering.
	adaptorHeaderSize = adaptorStreamFrameSeqFieldSize

	smuxVersionFieldOffset  = 0
	smuxCmdFieldOffset      = 1
	smuxLengthFieldOffset   = 2
	smuxStreamIDFieldOffset = 4
	smuxFrameHeaderSize     = 8

	adaptorBatchMarkerSize        = 8
	adaptorBatchCountSize         = 1
	adaptorBatchFrameLengthSize   = 2
	adaptorBatchMaximumFrameCount = 255
)

const adaptorBatchMarker = ^uint64(0)

var sessionPacketConnCloseDrainTimeout = 2 * time.Second

// Options configures same-session aggregation. Both endpoints receive the
// same values from the RRIPT transport configuration; no wire negotiation is
// performed.
type Options struct {
	AggregationMaxFrames      int
	AggregationMinQueueFrames int
	AggregationFlushDelay     time.Duration
	OnError                   func(error)
}

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
	session     *rrpitBidirectionalSession.BidirectionalSession
	sendMessage func([]byte) error

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
	maxMessageBytes         int

	aggregationMaxFrames      int
	aggregationMinQueueFrames int
	aggregationFlushDelay     time.Duration
	outgoingFrames            []*adaptorFrame
	sendDone                  chan struct{}
	onError                   func(error)

	closeOnce  sync.Once
	closing    bool
	closed     bool
	closeErr   error
	localAddr  net.Addr
	remoteAddr net.Addr
}

func New(session *rrpitBidirectionalSession.BidirectionalSession, client bool, config *smux.Config, options ...Options) (*Adaptor, error) {
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

	var adaptorOptions Options
	if len(options) > 0 {
		adaptorOptions = options[0]
	}
	packetConn, err := newSessionPacketConnWithOptions(session, maxSerializedFrameBytes, adaptorOptions)
	if err != nil {
		return nil, err
	}
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

func NewClient(session *rrpitBidirectionalSession.BidirectionalSession, config *smux.Config, options ...Options) (*Adaptor, error) {
	return New(session, true, config, options...)
}

func NewServer(session *rrpitBidirectionalSession.BidirectionalSession, config *smux.Config, options ...Options) (*Adaptor, error) {
	return New(session, false, config, options...)
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
	conn, _ := newSessionPacketConnWithOptions(session, maxSerializedFrameBytes, Options{})
	return conn
}

func newSessionPacketConnWithOptions(session *rrpitBidirectionalSession.BidirectionalSession, maxSerializedFrameBytes int, options Options) (*sessionPacketConn, error) {
	if options.AggregationMaxFrames < 0 || options.AggregationMaxFrames > adaptorBatchMaximumFrameCount {
		return nil, fmt.Errorf("rrpit adaptor aggregation max frames must be between 0 and %d", adaptorBatchMaximumFrameCount)
	}
	if options.AggregationMinQueueFrames < 0 {
		return nil, fmt.Errorf("rrpit adaptor aggregation minimum queue frames must not be negative")
	}
	if options.AggregationFlushDelay < 0 {
		return nil, fmt.Errorf("rrpit adaptor aggregation flush delay must not be negative")
	}
	if options.AggregationMaxFrames > 1 && options.AggregationMinQueueFrames < 2 {
		return nil, fmt.Errorf("rrpit adaptor aggregation minimum queue frames must be at least 2 when aggregation is enabled")
	}
	conn := &sessionPacketConn{
		session:                   session,
		nextSendStreamFrameSeq:    make(map[uint32]uint64),
		nextExpectedStreamSeq:     make(map[uint32]uint64),
		readyFramesByStream:       make(map[uint32]map[uint64]*adaptorFrame),
		activeStreamSet:           make(map[uint32]bool),
		locallyKnownStreams:       make(map[uint32]bool),
		remoteSynEstablished:      make(map[uint32]bool),
		maxSerializedFrameBytes:   maxSerializedFrameBytes,
		maxMessageBytes:           maxSerializedFrameBytes + adaptorHeaderSize,
		aggregationMaxFrames:      options.AggregationMaxFrames,
		aggregationMinQueueFrames: options.AggregationMinQueueFrames,
		aggregationFlushDelay:     options.AggregationFlushDelay,
		onError:                   options.OnError,
		localAddr:                 adaptorAddr("rrpit-local"),
		remoteAddr:                adaptorAddr("rrpit-remote"),
	}
	conn.cond = sync.NewCond(&conn.mu)
	if session != nil {
		conn.sendMessage = session.SendMessage
	}
	if conn.aggregationEnabled() {
		conn.sendDone = make(chan struct{})
		go conn.runSender()
	}
	return conn, nil
}

func (c *sessionPacketConn) OnMessage(data []byte) error {
	frames, err := decodeAdaptorMessage(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	for _, frame := range frames {
		if frame.streamFrameSeq < c.nextExpectedStreamSeq[frame.streamID] {
			continue
		}
		streamFrames := c.readyFramesByStream[frame.streamID]
		if streamFrames == nil {
			streamFrames = make(map[uint64]*adaptorFrame)
			c.readyFramesByStream[frame.streamID] = streamFrames
		}
		if _, found := streamFrames[frame.streamFrameSeq]; found {
			continue
		}
		streamFrames[frame.streamFrameSeq] = frame
		c.activateStreamLocked(frame.streamID)
	}
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
	if c.sendMessage == nil {
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
	if streamFrameSeq == adaptorBatchMarker {
		c.mu.Unlock()
		err := fmt.Errorf("rrpit adaptor stream sequence exhausted")
		c.fail(err)
		return 0, err
	}
	c.nextSendStreamFrameSeq[smuxFrame.streamID] = streamFrameSeq + 1
	if smuxFrame.streamID != 0 {
		c.locallyKnownStreams[smuxFrame.streamID] = true
	}
	c.mu.Unlock()

	frame := &adaptorFrame{
		frameID:        frameID,
		streamID:       smuxFrame.streamID,
		streamFrameSeq: streamFrameSeq,
		smuxCmd:        smuxFrame.cmd,
		smuxVersion:    smuxFrame.version,
		payload:        append([]byte(nil), p...),
	}
	if c.aggregationEnabled() {
		if err := c.enqueueFrame(frame); err != nil {
			return 0, err
		}
		return len(p), nil
	}

	wire := encodeAdaptorFrame(frame)
	if err := c.sendMessage(wire); err != nil {
		c.fail(err)
		return 0, err
	}
	return len(p), nil
}

func (c *sessionPacketConn) Close() error {
	c.waitForBufferedPayloadBeforeClose()
	c.waitForSendDrain()
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
	if c.onError != nil {
		c.onError(err)
	}
	c.closeWithError(err)
}

func (c *sessionPacketConn) aggregationEnabled() bool {
	return c != nil && c.aggregationMaxFrames > 1
}

func (c *sessionPacketConn) enqueueFrame(frame *adaptorFrame) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.closing {
		if c.closeErr != nil {
			return c.closeErr
		}
		return io.ErrClosedPipe
	}
	c.outgoingFrames = append(c.outgoingFrames, frame)
	c.cond.Broadcast()
	return nil
}

func (c *sessionPacketConn) runSender() {
	defer close(c.sendDone)
	for {
		wire, ok := c.nextOutgoingMessage()
		if !ok {
			return
		}
		if c.sendMessage == nil {
			c.fail(io.ErrClosedPipe)
			return
		}
		if err := c.sendMessage(wire); err != nil {
			c.fail(err)
			return
		}
	}
}

func (c *sessionPacketConn) nextOutgoingMessage() ([]byte, bool) {
	c.mu.Lock()
	for len(c.outgoingFrames) == 0 && !c.closed && !c.closing {
		c.cond.Wait()
	}
	if c.closed || len(c.outgoingFrames) == 0 && c.closing {
		c.mu.Unlock()
		return nil, false
	}
	shouldWaitForBurst := len(c.outgoingFrames) < c.aggregationMinQueueFrames && !c.closing && c.aggregationFlushDelay > 0
	delay := c.aggregationFlushDelay
	c.mu.Unlock()

	if shouldWaitForBurst {
		timer := time.NewTimer(delay)
		<-timer.C
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || len(c.outgoingFrames) == 0 {
		return nil, false
	}

	count := 1
	if len(c.outgoingFrames) >= c.aggregationMinQueueFrames {
		count = c.batchableFrameCountLocked()
	}
	frames := append([]*adaptorFrame(nil), c.outgoingFrames[:count]...)
	c.outgoingFrames = c.outgoingFrames[count:]
	if len(c.outgoingFrames) == 0 {
		c.outgoingFrames = nil
	}
	if count == 1 {
		return encodeAdaptorFrame(frames[0]), true
	}
	wire, err := encodeAdaptorBatch(frames)
	if err != nil {
		go c.fail(err)
		return nil, false
	}
	return wire, true
}

func (c *sessionPacketConn) batchableFrameCountLocked() int {
	limit := min(c.aggregationMaxFrames, len(c.outgoingFrames))
	messageBytes := adaptorBatchMarkerSize + adaptorBatchCountSize
	count := 0
	for count < limit {
		frameBytes := adaptorHeaderSize + len(c.outgoingFrames[count].payload)
		if frameBytes > int(^uint16(0)) || messageBytes+adaptorBatchFrameLengthSize+frameBytes > c.maxMessageBytes {
			break
		}
		messageBytes += adaptorBatchFrameLengthSize + frameBytes
		count++
	}
	if count < 2 {
		return 1
	}
	return count
}

func (c *sessionPacketConn) waitForSendDrain() {
	if c == nil || c.sendDone == nil {
		return
	}
	c.mu.Lock()
	c.closing = true
	c.cond.Broadcast()
	c.mu.Unlock()

	if sessionPacketConnCloseDrainTimeout <= 0 {
		<-c.sendDone
		return
	}
	timer := time.NewTimer(sessionPacketConnCloseDrainTimeout)
	defer timer.Stop()
	select {
	case <-c.sendDone:
	case <-timer.C:
		c.fail(io.ErrClosedPipe)
	}
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
	binary.BigEndian.PutUint64(wire[:adaptorStreamFrameSeqFieldSize], frame.streamFrameSeq)
	copy(wire[adaptorHeaderSize:], frame.payload)
	return wire
}

func encodeAdaptorBatch(frames []*adaptorFrame) ([]byte, error) {
	if len(frames) < 2 || len(frames) > adaptorBatchMaximumFrameCount {
		return nil, fmt.Errorf("rrpit adaptor batch frame count %d is outside 2..%d", len(frames), adaptorBatchMaximumFrameCount)
	}
	total := adaptorBatchMarkerSize + adaptorBatchCountSize
	encoded := make([][]byte, len(frames))
	for index, frame := range frames {
		encoded[index] = encodeAdaptorFrame(frame)
		if len(encoded[index]) > int(^uint16(0)) {
			return nil, fmt.Errorf("rrpit adaptor batch child %d is too large: %d", index, len(encoded[index]))
		}
		total += adaptorBatchFrameLengthSize + len(encoded[index])
	}
	wire := make([]byte, total)
	binary.BigEndian.PutUint64(wire[:adaptorBatchMarkerSize], adaptorBatchMarker)
	wire[adaptorBatchMarkerSize] = byte(len(frames))
	offset := adaptorBatchMarkerSize + adaptorBatchCountSize
	for _, child := range encoded {
		binary.BigEndian.PutUint16(wire[offset:offset+adaptorBatchFrameLengthSize], uint16(len(child)))
		offset += adaptorBatchFrameLengthSize
		copy(wire[offset:], child)
		offset += len(child)
	}
	return wire, nil
}

func decodeAdaptorMessage(data []byte) ([]*adaptorFrame, error) {
	if len(data) < adaptorBatchMarkerSize || binary.BigEndian.Uint64(data[:adaptorBatchMarkerSize]) != adaptorBatchMarker {
		frame, err := decodeAdaptorFrame(data)
		if err != nil {
			return nil, err
		}
		return []*adaptorFrame{frame}, nil
	}
	if len(data) < adaptorBatchMarkerSize+adaptorBatchCountSize {
		return nil, fmt.Errorf("rrpit adaptor batch header too short: %d", len(data))
	}
	count := int(data[adaptorBatchMarkerSize])
	if count < 2 {
		return nil, fmt.Errorf("rrpit adaptor batch frame count must be at least 2: %d", count)
	}
	frames := make([]*adaptorFrame, 0, count)
	offset := adaptorBatchMarkerSize + adaptorBatchCountSize
	for index := 0; index < count; index++ {
		if len(data)-offset < adaptorBatchFrameLengthSize {
			return nil, fmt.Errorf("rrpit adaptor batch child %d length is truncated", index)
		}
		length := int(binary.BigEndian.Uint16(data[offset : offset+adaptorBatchFrameLengthSize]))
		offset += adaptorBatchFrameLengthSize
		if length == 0 || len(data)-offset < length {
			return nil, fmt.Errorf("rrpit adaptor batch child %d length %d exceeds remaining %d", index, length, len(data)-offset)
		}
		frame, err := decodeAdaptorFrame(data[offset : offset+length])
		if err != nil {
			return nil, fmt.Errorf("rrpit adaptor batch child %d: %w", index, err)
		}
		frames = append(frames, frame)
		offset += length
	}
	if offset != len(data) {
		return nil, fmt.Errorf("rrpit adaptor batch has %d trailing bytes", len(data)-offset)
	}
	return frames, nil
}

func decodeAdaptorFrame(data []byte) (*adaptorFrame, error) {
	if len(data) < adaptorHeaderSize+smuxFrameHeaderSize {
		return nil, fmt.Errorf("rrpit adaptor frame too short: %d", len(data))
	}
	frame := &adaptorFrame{
		streamFrameSeq: binary.BigEndian.Uint64(data[:adaptorStreamFrameSeqFieldSize]),
		payload:        append([]byte(nil), data[adaptorHeaderSize:]...),
	}
	smuxFrame, err := parseSmuxFrame(frame.payload)
	if err != nil {
		return nil, err
	}
	frame.streamID = smuxFrame.streamID
	frame.smuxCmd = smuxFrame.cmd
	frame.smuxVersion = smuxFrame.version
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
