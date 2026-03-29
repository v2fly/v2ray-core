package packetToStream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/xtaci/smux"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

func TestAdaptorCarriesSmuxStreamsOverRRpit(t *testing.T) {
	clientSession := mustNewPacketToStreamSession(t, true)
	serverSession := mustNewPacketToStreamSession(t, false)

	clientToServer := newPacketWire()
	serverToClient := newPacketWire()
	defer func() {
		_ = clientToServer.Close()
		_ = serverToClient.Close()
	}()

	clientRx, err := clientSession.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	serverRx, err := serverSession.AttachRxChannel()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := clientSession.AttachTxChannel(clientToServer); err != nil {
		t.Fatal(err)
	}
	if _, err := serverSession.AttachTxChannel(serverToClient); err != nil {
		t.Fatal(err)
	}

	pumpErrs := make(chan error, 2)
	go pumpPacketWire(clientToServer, serverRx, pumpErrs)
	go pumpPacketWire(serverToClient, clientRx, pumpErrs)

	smuxConfig := mustNewSmuxConfigForSession(t, clientSession, 2)

	clientAdaptor, err := NewClient(clientSession, smuxConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := clientAdaptor.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	serverAdaptor, err := NewServer(serverSession, smuxConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := serverAdaptor.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	serverAccepted := make(chan *smux.Stream, 2)
	serverAcceptErr := make(chan error, 1)
	go func() {
		defer close(serverAccepted)
		for i := 0; i < 2; i++ {
			stream, err := serverAdaptor.AcceptStream()
			if err != nil {
				serverAcceptErr <- err
				return
			}
			serverAccepted <- stream
		}
	}()

	clientStreamA, err := clientAdaptor.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = clientStreamA.Close() }()

	clientStreamB, err := clientAdaptor.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = clientStreamB.Close() }()

	serverStreamA := mustRecvSmuxStream(t, serverAccepted, serverAcceptErr)
	defer func() { _ = serverStreamA.Close() }()
	serverStreamB := mustRecvSmuxStream(t, serverAccepted, serverAcceptErr)
	defer func() { _ = serverStreamB.Close() }()

	payloadA := bytes.Repeat([]byte("a"), 1024)
	payloadB := bytes.Repeat([]byte("b"), 1536)

	if _, err := clientStreamA.Write(payloadA); err != nil {
		t.Fatal(err)
	}
	if _, err := clientStreamB.Write(payloadB); err != nil {
		t.Fatal(err)
	}

	serverReadA := make([]byte, len(payloadA))
	serverReadB := make([]byte, len(payloadB))
	if _, err := io.ReadFull(serverStreamA, serverReadA); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(serverStreamB, serverReadB); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(serverReadA, payloadA) {
		t.Fatal("payload A mismatch")
	}
	if !bytes.Equal(serverReadB, payloadB) {
		t.Fatal("payload B mismatch")
	}

	replyA := []byte("reply-a")
	replyB := []byte("reply-b-reply-b")
	if _, err := serverStreamA.Write(replyA); err != nil {
		t.Fatal(err)
	}
	if _, err := serverStreamB.Write(replyB); err != nil {
		t.Fatal(err)
	}

	clientReadA := make([]byte, len(replyA))
	clientReadB := make([]byte, len(replyB))
	if _, err := io.ReadFull(clientStreamA, clientReadA); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(clientStreamB, clientReadB); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(clientReadA, replyA) {
		t.Fatal("reply A mismatch")
	}
	if !bytes.Equal(clientReadB, replyB) {
		t.Fatal("reply B mismatch")
	}

	select {
	case err := <-pumpErrs:
		t.Fatal(err)
	default:
	}
}

func TestAdaptorRejectsSmuxFrameSizeAboveBudget(t *testing.T) {
	session := mustNewPacketToStreamSession(t, true)
	config := smux.DefaultConfig()
	config.KeepAliveDisabled = true
	config.MaxFrameSize = 256
	if _, err := NewClient(session, config); err == nil {
		t.Fatal("expected oversized smux frame config to be rejected")
	}
}

func TestSessionPacketConnPreservesPerStreamOrder(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	conn.locallyKnownStreams[1] = true

	frame0 := mustMarshalAdaptorFrameForTest(t, 0, 1, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("hello ")))
	frame1 := mustMarshalAdaptorFrameForTest(t, 1, 1, 1, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("world")))

	if err := conn.OnMessage(frame1); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("later same-stream frame should not be released before its predecessor")
	}
	conn.mu.Unlock()

	if err := conn.OnMessage(frame0); err != nil {
		t.Fatal(err)
	}

	want := append(decodePayloadForTest(t, frame0), decodePayloadForTest(t, frame1)...)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected reordered payload: %x", got)
	}
}

func TestSessionPacketConnAllowsIndependentStreamProgress(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	conn.locallyKnownStreams[1] = true
	conn.locallyKnownStreams[3] = true

	blocked := mustMarshalAdaptorFrameForTest(t, 1, 1, 1, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("stream-a-late")))
	unaffected := mustMarshalAdaptorFrameForTest(t, 2, 3, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 3, []byte("stream-b-now")))

	if err := conn.OnMessage(blocked); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("blocked frame should not become readable early")
	}
	conn.mu.Unlock()

	if err := conn.OnMessage(unaffected); err != nil {
		t.Fatal(err)
	}

	want := decodePayloadForTest(t, unaffected)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected unrelated stream payload: %x", got)
	}
}

func TestSessionPacketConnRemoteStreamWaitsForSyn(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	psh := mustMarshalAdaptorFrameForTest(t, 1, 11, 1, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 11, []byte("payload")))
	if err := conn.OnMessage(psh); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("remote payload should wait for stream syn")
	}
	conn.mu.Unlock()

	syn := mustMarshalAdaptorFrameForTest(t, 0, 11, 0, mustMarshalSmuxFrameForTest(t, smuxCmdSYN, 11, nil))
	if err := conn.OnMessage(syn); err != nil {
		t.Fatal(err)
	}

	want := append(decodePayloadForTest(t, syn), decodePayloadForTest(t, psh)...)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected remote stream sequence: %x", got)
	}
}

func TestSessionPacketConnAllowsLocallyKnownStreamWithoutInboundSyn(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	conn.locallyKnownStreams[9] = true

	frame := mustMarshalAdaptorFrameForTest(t, 0, 9, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 9, []byte("payload")))
	if err := conn.OnMessage(frame); err != nil {
		t.Fatal(err)
	}

	want := decodePayloadForTest(t, frame)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected locally-known stream payload: %x", got)
	}
}

func TestSessionPacketConnDeliversNOPControlFrames(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	nop := mustMarshalAdaptorFrameForTest(t, 0, 0, 0, mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil))
	if err := conn.OnMessage(nop); err != nil {
		t.Fatal(err)
	}

	want := decodePayloadForTest(t, nop)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected nop payload: %x", got)
	}
}

func TestSessionPacketConnIgnoresDuplicateAndOldFrames(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	conn.locallyKnownStreams[1] = true

	frame0 := mustMarshalAdaptorFrameForTest(t, 0, 1, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("one")))
	if err := conn.OnMessage(frame0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(decodePayloadForTest(t, frame0)))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}

	if err := conn.OnMessage(frame0); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("duplicate old frame should be ignored")
	}
	conn.mu.Unlock()

	frame1 := mustMarshalAdaptorFrameForTest(t, 1, 1, 1, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("two")))
	if err := conn.OnMessage(frame1); err != nil {
		t.Fatal(err)
	}
	if err := conn.OnMessage(frame1); err != nil {
		t.Fatal(err)
	}
	want := decodePayloadForTest(t, frame1)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected duplicate handling payload: %x", got)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("duplicate current frame should not be released twice")
	}
	conn.mu.Unlock()
}

func TestSessionPacketConnRoundRobinDeliveryAcrossStreams(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	conn.locallyKnownStreams[1] = true
	conn.locallyKnownStreams[3] = true

	frame10 := &adaptorFrame{frameID: 0, streamID: 1, streamFrameSeq: 0, smuxCmd: smuxCmdPSH, smuxVersion: 2, payload: mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("a0"))}
	frame11 := &adaptorFrame{frameID: 1, streamID: 1, streamFrameSeq: 1, smuxCmd: smuxCmdPSH, smuxVersion: 2, payload: mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("a1"))}
	frame30 := &adaptorFrame{frameID: 2, streamID: 3, streamFrameSeq: 0, smuxCmd: smuxCmdPSH, smuxVersion: 2, payload: mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 3, []byte("b0"))}
	frame31 := &adaptorFrame{frameID: 3, streamID: 3, streamFrameSeq: 1, smuxCmd: smuxCmdPSH, smuxVersion: 2, payload: mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 3, []byte("b1"))}

	conn.mu.Lock()
	conn.readyFramesByStream[1] = map[uint64]*adaptorFrame{0: frame10, 1: frame11}
	conn.readyFramesByStream[3] = map[uint64]*adaptorFrame{0: frame30, 1: frame31}
	conn.activateStreamLocked(1)
	conn.activateStreamLocked(3)
	conn.deliverReadyFramesLocked()
	conn.mu.Unlock()

	want := append([]byte{}, frame10.payload...)
	want = append(want, frame30.payload...)
	want = append(want, frame11.payload...)
	want = append(want, frame31.payload...)
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected round-robin order: %x", got)
	}
}

func TestSessionPacketConnOnMessageIgnoresFramesAfterClose(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	frame := mustMarshalAdaptorFrameForTest(t, 0, 0, 0, mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil))
	if err := conn.OnMessage(frame); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.readBuf.Len() != 0 {
		t.Fatal("closed conn should ignore later frames")
	}
}

func TestSessionPacketConnOnMessageRejectsMetadataMismatch(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	smuxFrame := mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("payload"))
	badWire := encodeAdaptorFrame(&adaptorFrame{
		frameID:        0,
		streamID:       2,
		streamFrameSeq: 0,
		smuxCmd:        smuxCmdPSH,
		smuxVersion:    2,
		payload:        smuxFrame,
	})
	if err := conn.OnMessage(badWire); err == nil {
		t.Fatal("expected metadata mismatch to be rejected")
	}
}

func TestSessionPacketConnCloseWaitsForBufferedPayloadDrain(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)

	conn.locallyKnownStreams[1] = true
	frame := mustMarshalAdaptorFrameForTest(t, 0, 1, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("hello")))
	if err := conn.OnMessage(frame); err != nil {
		t.Fatal(err)
	}

	closeDone := make(chan struct{})
	go func() {
		_ = conn.Close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
		t.Fatal("close returned before buffered payload drained")
	case <-time.After(50 * time.Millisecond):
	}

	buf := make([]byte, len(decodePayloadForTest(t, frame)))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, decodePayloadForTest(t, frame)) {
		t.Fatalf("unexpected payload: %x", buf)
	}

	select {
	case <-closeDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("close did not finish after buffered payload drained")
	}

	if _, err := conn.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("expected EOF after graceful close, got %v", err)
	}
}

func TestSessionPacketConnCloseRejectsWritesWhileDraining(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	conn.locallyKnownStreams[1] = true
	frame := mustMarshalAdaptorFrameForTest(t, 0, 1, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("hello")))
	if err := conn.OnMessage(frame); err != nil {
		t.Fatal(err)
	}

	closeStarted := make(chan struct{})
	closeDone := make(chan struct{})
	go func() {
		close(closeStarted)
		_ = conn.Close()
		close(closeDone)
	}()

	<-closeStarted
	time.Sleep(20 * time.Millisecond)

	if _, err := conn.Write(mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil)); err != io.ErrClosedPipe {
		t.Fatalf("expected io.ErrClosedPipe while close is draining, got %v", err)
	}

	buf := make([]byte, len(decodePayloadForTest(t, frame)))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}

	select {
	case <-closeDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("close did not finish after draining payload")
	}
}

func TestSessionPacketConnCloseTimesOutWithUndeliverableFrames(t *testing.T) {
	oldTimeout := sessionPacketConnCloseDrainTimeout
	sessionPacketConnCloseDrainTimeout = 50 * time.Millisecond
	defer func() {
		sessionPacketConnCloseDrainTimeout = oldTimeout
	}()

	conn := newSessionPacketConn(nil, 1024)
	frame := mustMarshalAdaptorFrameForTest(t, 0, 9, 0, mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 9, []byte("blocked")))
	if err := conn.OnMessage(frame); err != nil {
		t.Fatal(err)
	}

	closeDone := make(chan struct{})
	go func() {
		_ = conn.Close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
		t.Fatal("close returned before drain timeout for undeliverable frame")
	case <-time.After(10 * time.Millisecond):
	}

	select {
	case <-closeDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("close did not time out for undeliverable frame")
	}

	if _, err := conn.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("expected EOF after close timeout, got %v", err)
	}
}

func TestSessionPacketConnReadAndWriteReturnCloseError(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	wantErr := errors.New("boom")
	conn.fail(wantErr)

	if _, err := conn.Read(make([]byte, 1)); !errors.Is(err, wantErr) {
		t.Fatalf("expected read error %v, got %v", wantErr, err)
	}
	if _, err := conn.Write(mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil)); !errors.Is(err, wantErr) {
		t.Fatalf("expected write error %v, got %v", wantErr, err)
	}
}

func TestSessionPacketConnWriteRejectsMalformedSmuxFrame(t *testing.T) {
	conn, cleanup := newWritableSessionPacketConn(t)
	defer cleanup()

	if _, err := conn.Write([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected malformed smux frame write to fail")
	}
	if _, err := conn.Read(make([]byte, 1)); err == nil {
		t.Fatal("expected malformed write to close conn with error")
	}
}

func TestSessionPacketConnWriteFailsWithoutTransferChannel(t *testing.T) {
	session := mustNewPacketToStreamSession(t, true)
	defer func() { _ = session.Close() }()

	conn := newSessionPacketConn(session, 1024)
	frame := mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil)
	firstErr := error(nil)
	if _, err := conn.Write(frame); err == nil {
		t.Fatal("expected missing transfer channel to fail")
	} else {
		firstErr = err
	}
	if _, err := conn.Read(make([]byte, 1)); !errors.Is(err, firstErr) && err.Error() != firstErr.Error() {
		t.Fatalf("expected read to return write failure %v, got %v", firstErr, err)
	}
}

func TestSessionPacketConnWriteTracksFrameAndStreamSequence(t *testing.T) {
	conn, cleanup := newWritableSessionPacketConn(t)
	defer cleanup()

	frameA := mustMarshalSmuxFrameForTest(t, smuxCmdPSH, 1, []byte("a"))
	frameB := mustMarshalSmuxFrameForTest(t, smuxCmdFIN, 1, nil)
	frameNOP := mustMarshalSmuxFrameForTest(t, smuxCmdNOP, 0, nil)

	if n, err := conn.Write(frameA); err != nil || n != len(frameA) {
		t.Fatalf("unexpected first write result n=%d err=%v", n, err)
	}
	if n, err := conn.Write(frameB); err != nil || n != len(frameB) {
		t.Fatalf("unexpected second write result n=%d err=%v", n, err)
	}
	if n, err := conn.Write(frameNOP); err != nil || n != len(frameNOP) {
		t.Fatalf("unexpected nop write result n=%d err=%v", n, err)
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.nextSendFrameID != 3 {
		t.Fatalf("unexpected nextSendFrameID %d", conn.nextSendFrameID)
	}
	if conn.nextSendStreamFrameSeq[1] != 2 {
		t.Fatalf("unexpected stream frame seq for stream 1: %d", conn.nextSendStreamFrameSeq[1])
	}
	if conn.nextSendStreamFrameSeq[0] != 1 {
		t.Fatalf("unexpected stream frame seq for control stream: %d", conn.nextSendStreamFrameSeq[0])
	}
	if !conn.locallyKnownStreams[1] {
		t.Fatal("locally known stream should be recorded after write")
	}
	if conn.locallyKnownStreams[0] {
		t.Fatal("control stream should not be tracked as locally known stream")
	}
}

func TestSessionPacketConnLocalAddrAndDeadlines(t *testing.T) {
	conn := newSessionPacketConn(nil, 1024)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if got := conn.LocalAddr(); got == nil || got.Network() != "rrpit" || got.String() != "rrpit-local" {
		t.Fatalf("unexpected local addr: %#v", got)
	}
	if got := conn.RemoteAddr(); got == nil || got.Network() != "rrpit" || got.String() != "rrpit-remote" {
		t.Fatalf("unexpected remote addr: %#v", got)
	}
	if err := conn.SetDeadline(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetReadDeadline(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetWriteDeadline(time.Now()); err != nil {
		t.Fatal(err)
	}
}

type packetWire struct {
	mu     sync.Mutex
	closed bool
	ch     chan []byte
}

func newPacketWire() *packetWire {
	return &packetWire{
		ch: make(chan []byte, 256),
	}
}

func (w *packetWire) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, io.ErrClosedPipe
	}
	wire := append([]byte(nil), p...)
	w.ch <- wire
	return len(p), nil
}

func (w *packetWire) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true
	close(w.ch)
	return nil
}

func pumpPacketWire(writer *packetWire, channel *rrpitMaterializedTransferChannel.ChannelRx, errs chan<- error) {
	for wire := range writer.ch {
		if err := channel.OnNewMessageArrived(wire); err != nil {
			errs <- err
			return
		}
	}
}

func mustNewPacketToStreamSession(t *testing.T, oddChannelIDs bool) *rrpitBidirectionalSession.BidirectionalSession {
	t.Helper()

	session, err := rrpitBidirectionalSession.New(rrpitBidirectionalSession.Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:    128,
			MaxBufferedLanes: 16,
			OnMessage: func([]byte) error {
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  128,
			MaxDataShardsPerLane:           4,
			MaxBufferedLanes:               16,
			MaxRewindableTimestampNum:      32,
			MaxRewindableControlMessageNum: 32,
			OddChannelIDs:                  oddChannelIDs,
		},
		TimestampInterval: 2 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func mustRecvSmuxStream(t *testing.T, streams <-chan *smux.Stream, errs <-chan error) *smux.Stream {
	t.Helper()

	select {
	case stream := <-streams:
		if stream == nil {
			t.Fatal("nil smux stream")
		}
		return stream
	case err := <-errs:
		t.Fatal(err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for smux stream")
	}
	return nil
}

func newWritableSessionPacketConn(t *testing.T) (*sessionPacketConn, func()) {
	t.Helper()

	session := mustNewPacketToStreamSession(t, true)
	wire := newPacketWire()
	if _, err := session.AttachTxChannel(wire); err != nil {
		t.Fatal(err)
	}
	conn := newSessionPacketConn(session, 1024)
	maxMessageSize, err := session.MaxMessageSize()
	if err != nil {
		t.Fatal(err)
	}
	conn.maxSerializedFrameBytes = maxMessageSize - adaptorHeaderSize
	if conn.maxSerializedFrameBytes <= 0 {
		t.Fatalf("unexpected max serialized frame bytes: %d", conn.maxSerializedFrameBytes)
	}
	return conn, func() {
		_ = wire.Close()
		_ = session.Close()
	}
}

func mustNewSmuxConfigForSession(t *testing.T, session *rrpitBidirectionalSession.BidirectionalSession, version int) *smux.Config {
	t.Helper()

	maxMessageSize, err := session.MaxMessageSize()
	if err != nil {
		t.Fatal(err)
	}
	maxFrameSize := maxMessageSize - adaptorHeaderSize - smuxFrameHeaderSize
	if maxFrameSize <= 0 {
		t.Fatalf("rrpit message budget too small for smux frame: %d", maxMessageSize)
	}
	if maxFrameSize > 64 {
		maxFrameSize = 64
	}
	config := smux.DefaultConfig()
	config.Version = version
	config.KeepAliveDisabled = true
	config.MaxFrameSize = maxFrameSize
	return config
}

func mustMarshalSmuxFrameForTest(t *testing.T, cmd byte, streamID uint32, payload []byte) []byte {
	t.Helper()

	if err := validateSmuxCommand(2, cmd, len(payload), streamID); err != nil {
		t.Fatal(err)
	}
	frame := make([]byte, smuxFrameHeaderSize+len(payload))
	frame[smuxVersionFieldOffset] = 2
	frame[smuxCmdFieldOffset] = cmd
	binary.LittleEndian.PutUint16(frame[smuxLengthFieldOffset:smuxLengthFieldOffset+2], uint16(len(payload)))
	binary.LittleEndian.PutUint32(frame[smuxStreamIDFieldOffset:smuxStreamIDFieldOffset+4], streamID)
	copy(frame[smuxFrameHeaderSize:], payload)
	return frame
}

func mustMarshalAdaptorFrameForTest(t *testing.T, frameID uint64, streamID uint32, streamSeq uint64, smuxFrame []byte) []byte {
	t.Helper()

	meta, err := parseSmuxFrame(smuxFrame)
	if err != nil {
		t.Fatal(err)
	}
	if meta.streamID != streamID {
		t.Fatalf("smux frame stream id %d does not match adaptor header %d", meta.streamID, streamID)
	}
	return encodeAdaptorFrame(&adaptorFrame{
		frameID:        frameID,
		streamID:       streamID,
		streamFrameSeq: streamSeq,
		smuxCmd:        meta.cmd,
		smuxVersion:    meta.version,
		payload:        append([]byte(nil), smuxFrame...),
	})
}

func decodePayloadForTest(t *testing.T, wire []byte) []byte {
	t.Helper()

	frame, err := decodeAdaptorFrame(wire)
	if err != nil {
		t.Fatal(err)
	}
	return frame.payload
}
