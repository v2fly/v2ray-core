package packetToStream

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
	"github.com/xtaci/smux"
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

	smuxConfig := smux.DefaultConfig()
	smuxConfig.KeepAliveDisabled = true
	smuxConfig.MaxFrameSize = 256
	maxMessageSize, err := clientSession.MaxMessageSize()
	if err != nil {
		t.Fatal(err)
	}
	if maxMessageSize >= smuxConfig.MaxFrameSize {
		t.Fatalf("expected rrpit max message size %d to force fragmentation for smux frame size %d", maxMessageSize, smuxConfig.MaxFrameSize)
	}

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

func TestSessionPacketConnReordersPacketsBySequence(t *testing.T) {
	conn := newSessionPacketConn(nil)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	packet0 := marshalAdapterPacketForTest(0, []byte("hello "))
	packet1 := marshalAdapterPacketForTest(1, []byte("world"))

	if err := conn.OnMessage(packet1); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("out-of-order packet should not be released early")
	}
	conn.mu.Unlock()

	if err := conn.OnMessage(packet0); err != nil {
		t.Fatal(err)
	}

	got := make([]byte, len("hello world"))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Fatalf("unexpected reordered payload: %q", got)
	}

	if err := conn.OnMessage(packet1); err != nil {
		t.Fatal(err)
	}
	conn.mu.Lock()
	if conn.readBuf.Len() != 0 {
		conn.mu.Unlock()
		t.Fatal("duplicate packet should be ignored")
	}
	conn.mu.Unlock()
}

func TestSessionPacketConnCloseWaitsForBufferedPayloadDrain(t *testing.T) {
	conn := newSessionPacketConn(nil)

	if err := conn.OnMessage(marshalAdapterPacketForTest(0, []byte("hello"))); err != nil {
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

	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "hello" {
		t.Fatalf("unexpected payload: %q", buf)
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
	conn := newSessionPacketConn(nil)

	if err := conn.OnMessage(marshalAdapterPacketForTest(0, []byte("hello"))); err != nil {
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

	if _, err := conn.Write([]byte("blocked")); err != io.ErrClosedPipe {
		t.Fatalf("expected io.ErrClosedPipe while close is draining, got %v", err)
	}

	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}

	select {
	case <-closeDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("close did not finish after draining payload")
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

func marshalAdapterPacketForTest(seq uint64, payload []byte) []byte {
	packet := make([]byte, packetHeaderSize+len(payload))
	binary.BigEndian.PutUint64(packet[:packetSequenceFieldSize], seq)
	binary.BigEndian.PutUint32(packet[packetSequenceFieldSize:packetHeaderSize], uint32(len(payload)))
	copy(packet[packetHeaderSize:], payload)
	return packet
}
