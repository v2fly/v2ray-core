package outbound

import (
	"context"
	"errors"
	"io"
	gonet "net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
)

type plainPacketConnTestDispatcher struct {
	dispatch func(context.Context, cnet.Destination) (*transport.Link, error)
}

func (d *plainPacketConnTestDispatcher) Dispatch(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
	return d.dispatch(ctx, destination)
}

func (*plainPacketConnTestDispatcher) Start() error      { return nil }
func (*plainPacketConnTestDispatcher) Close() error      { return nil }
func (*plainPacketConnTestDispatcher) Type() interface{} { return routing.DispatcherType() }

type plainPacketConnTestReadResult struct {
	buffer buf.MultiBuffer
	err    error
}

type plainPacketConnTestReader struct {
	results        chan plainPacketConnTestReadResult
	interrupted    chan struct{}
	secondRead     chan struct{}
	interruptOnce  sync.Once
	secondReadOnce sync.Once
	interruptCalls atomic.Int32
	readCalls      atomic.Int32
}

func newPlainPacketConnTestReader() *plainPacketConnTestReader {
	return &plainPacketConnTestReader{
		results:     make(chan plainPacketConnTestReadResult, 8),
		interrupted: make(chan struct{}),
		secondRead:  make(chan struct{}),
	}
}

func (r *plainPacketConnTestReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if r.readCalls.Add(1) == 2 {
		r.secondReadOnce.Do(func() { close(r.secondRead) })
	}
	select {
	case result := <-r.results:
		return result.buffer, result.err
	case <-r.interrupted:
		return nil, io.ErrClosedPipe
	}
}

func (r *plainPacketConnTestReader) Interrupt() {
	r.interruptCalls.Add(1)
	r.interruptOnce.Do(func() { close(r.interrupted) })
}

type plainPacketConnTestWriter struct {
	writes         chan []byte
	interrupted    chan struct{}
	interruptOnce  sync.Once
	interruptCalls atomic.Int32
	err            error
}

func newPlainPacketConnTestWriter() *plainPacketConnTestWriter {
	return &plainPacketConnTestWriter{
		writes:      make(chan []byte, 64),
		interrupted: make(chan struct{}),
	}
}

func (w *plainPacketConnTestWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	payload := make([]byte, mb.Len())
	mb.Copy(payload)
	buf.ReleaseMulti(mb)
	if w.err != nil {
		return w.err
	}
	select {
	case w.writes <- payload:
		return nil
	case <-w.interrupted:
		return gonet.ErrClosed
	}
}

func (w *plainPacketConnTestWriter) Interrupt() {
	w.interruptCalls.Add(1)
	w.interruptOnce.Do(func() { close(w.interrupted) })
}

type plainPacketConnTestRay struct {
	ctx    context.Context
	reader *plainPacketConnTestReader
	writer *plainPacketConnTestWriter
}

func newPlainPacketConnTestRay(ctx context.Context) *plainPacketConnTestRay {
	return &plainPacketConnTestRay{
		ctx:    ctx,
		reader: newPlainPacketConnTestReader(),
		writer: newPlainPacketConnTestWriter(),
	}
}

func (r *plainPacketConnTestRay) link() *transport.Link {
	return &transport.Link{Reader: r.reader, Writer: r.writer}
}

func plainPacketConnTestWait(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for " + description)
	}
}

func plainPacketConnTestWrite(t *testing.T, conn gonet.PacketConn, payload string, address gonet.Addr) {
	t.Helper()
	if n, err := conn.WriteTo([]byte(payload), address); err != nil || n != len(payload) {
		t.Fatalf("WriteTo(%q) = (%d, %v), want (%d, nil)", payload, n, err, len(payload))
	}
}

func plainPacketConnTestAwaitPayload(t *testing.T, writer *plainPacketConnTestWriter, want string) {
	t.Helper()
	select {
	case payload := <-writer.writes:
		if got := string(payload); got != want {
			t.Fatalf("written payload = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for payload " + want)
	}
}

func plainPacketConnTestBuffer(payload string) *buf.Buffer {
	buffer := buf.New()
	_, _ = buffer.WriteString(payload)
	return buffer
}

func TestWireguardPlainPacketConnReusesDestinationConnections(t *testing.T) {
	var access sync.Mutex
	rays := make(map[cnet.Destination]*plainPacketConnTestRay)
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
		ray := newPlainPacketConnTestRay(ctx)
		access.Lock()
		rays[destination] = ray
		access.Unlock()
		return ray.link(), nil
	}}

	conn, err := createLogicalPacketConn(context.Background(), dispatcher, packetaddr.PacketAddrType_None)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	addressA := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.1"), Port: 51820}
	addressB := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.2"), Port: 51821}
	plainPacketConnTestWrite(t, conn, "a1", addressA)
	plainPacketConnTestWrite(t, conn, "a2", addressA)
	plainPacketConnTestWrite(t, conn, "b1", addressB)

	access.Lock()
	if len(rays) != 2 {
		access.Unlock()
		t.Fatalf("dispatch count = %d, want one per destination (2)", len(rays))
	}
	rayA := rays[cnet.DestinationFromAddr(addressA)]
	rayB := rays[cnet.DestinationFromAddr(addressB)]
	access.Unlock()
	if rayA == nil || rayB == nil {
		t.Fatalf("dispatched destinations do not contain both peers: %v", rays)
	}
	plainPacketConnTestAwaitPayload(t, rayA.writer, "a1")
	plainPacketConnTestAwaitPayload(t, rayA.writer, "a2")
	plainPacketConnTestAwaitPayload(t, rayB.writer, "b1")

	rayB.reader.results <- plainPacketConnTestReadResult{buffer: buf.MultiBuffer{plainPacketConnTestBuffer("reply")}}
	readBuffer := make([]byte, 32)
	n, source, err := conn.ReadFrom(readBuffer)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(readBuffer[:n]); got != "reply" {
		t.Fatalf("read payload = %q, want reply", got)
	}
	if got, want := source.String(), addressB.String(); got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
}

func TestWireguardPlainPacketConnDeduplicatesConcurrentDial(t *testing.T) {
	var dispatchCalls atomic.Int32
	dispatchStarted := make(chan struct{})
	releaseDispatch := make(chan struct{})
	var startOnce sync.Once
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
		dispatchCalls.Add(1)
		startOnce.Do(func() { close(dispatchStarted) })
		<-releaseDispatch
		return newPlainPacketConnTestRay(ctx).link(), nil
	}}
	conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	const writers = 24
	startWrites := make(chan struct{})
	writeResults := make(chan error, writers)
	address := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.10"), Port: 51820}
	for i := 0; i < writers; i++ {
		go func() {
			<-startWrites
			_, err := conn.WriteTo([]byte("packet"), address)
			writeResults <- err
		}()
	}
	close(startWrites)
	plainPacketConnTestWait(t, dispatchStarted, "first dispatch")

	// Keep the first dial pending long enough for all concurrent writers to
	// reach the per-destination in-flight entry.
	time.Sleep(100 * time.Millisecond)
	close(releaseDispatch)
	for i := 0; i < writers; i++ {
		select {
		case err := <-writeResults:
			if err != nil {
				t.Fatalf("concurrent WriteTo failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for concurrent WriteTo")
		}
	}
	if got := dispatchCalls.Load(); got != 1 {
		t.Fatalf("dispatch count = %d, want one concurrent dial", got)
	}
}

func TestWireguardPlainPacketConnDialErrorTerminatesGeneration(t *testing.T) {
	wantErr := errors.New("dial failed")
	var attempts atomic.Int32
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
		attempts.Add(1)
		return nil, wantErr
	}}
	conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	readResult := make(chan error, 1)
	go func() {
		_, _, err := conn.ReadFrom(make([]byte, 32))
		readResult <- err
	}()

	address := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.20"), Port: 51820}
	if n, err := conn.WriteTo([]byte("first"), address); n != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("first WriteTo = (%d, %v), want (0, %v)", n, err, wantErr)
	}
	select {
	case err := <-readResult:
		if !errors.Is(err, wantErr) {
			t.Fatalf("ReadFrom error = %v, want dial error", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ReadFrom remained blocked after dial failure")
	}
	if n, err := conn.WriteTo([]byte("second"), address); n != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("second WriteTo = (%d, %v), want terminal dial error", n, err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("dial attempts = %d, want terminal generation to avoid redial", got)
	}
}

func TestWireguardPlainPacketConnPropagatesReadError(t *testing.T) {
	wantErr := errors.New("underlay read failed")
	response := plainPacketConnTestBuffer("must be released")
	var ray *plainPacketConnTestRay
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
		ray = newPlainPacketConnTestRay(ctx)
		ray.reader.results <- plainPacketConnTestReadResult{
			buffer: buf.MultiBuffer{response},
			err:    wantErr,
		}
		return ray.link(), nil
	}}
	conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	plainPacketConnTestWrite(t, conn, "request", &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.30"), Port: 51820})
	plainPacketConnTestWait(t, ray.reader.interrupted, "failed ray reader interruption")
	plainPacketConnTestWait(t, ray.writer.interrupted, "failed ray writer interruption")
	if n, source, err := conn.ReadFrom(make([]byte, 32)); n != 0 || source != nil || !errors.Is(err, wantErr) {
		t.Fatalf("ReadFrom = (%d, %v, %v), want (0, nil, %v)", n, source, err, wantErr)
	}
	if response.Len() != 0 {
		t.Fatal("buffer returned alongside a read error was not released")
	}
}

func TestWireguardPlainPacketConnCloseUnblocksRead(t *testing.T) {
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(context.Context, cnet.Destination) (*transport.Link, error) {
		t.Fatal("unexpected dispatch")
		return nil, nil
	}}
	conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
	if err != nil {
		t.Fatal(err)
	}

	readResult := make(chan error, 1)
	go func() {
		_, _, err := conn.ReadFrom(make([]byte, 32))
		readResult <- err
	}()
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-readResult:
		if !errors.Is(err, gonet.ErrClosed) {
			t.Fatalf("ReadFrom error = %v, want net.ErrClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ReadFrom remained blocked after Close")
	}
	if n, err := conn.WriteTo([]byte("rejected"), &gonet.UDPAddr{IP: gonet.IPv4zero, Port: 51820}); n != 0 || !errors.Is(err, gonet.ErrClosed) {
		t.Fatalf("WriteTo after Close = (%d, %v), want (0, net.ErrClosed)", n, err)
	}
}

func TestWireguardPlainPacketConnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dispatchCtx := make(chan context.Context, 1)
	var ray *plainPacketConnTestRay
	dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
		dispatchCtx <- ctx
		ray = newPlainPacketConnTestRay(ctx)
		return ray.link(), nil
	}}
	conn, err := newWireguardPlainPacketConn(ctx, dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	plainPacketConnTestWrite(t, conn, "request", &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.40"), Port: 51820})
	rayCtx := <-dispatchCtx

	readResult := make(chan error, 1)
	go func() {
		_, _, err := conn.ReadFrom(make([]byte, 32))
		readResult <- err
	}()
	cancel()
	select {
	case err := <-readResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ReadFrom error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ReadFrom remained blocked after context cancellation")
	}
	plainPacketConnTestWait(t, rayCtx.Done(), "ray context cancellation")
	plainPacketConnTestWait(t, ray.reader.interrupted, "ray reader interruption")
	plainPacketConnTestWait(t, ray.writer.interrupted, "ray writer interruption")
}

func TestWireguardPlainPacketConnOwnsBuffersAndConnections(t *testing.T) {
	t.Run("read releases delivered buffer", func(t *testing.T) {
		response := plainPacketConnTestBuffer("response")
		var ray *plainPacketConnTestRay
		dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
			ray = newPlainPacketConnTestRay(ctx)
			ray.reader.results <- plainPacketConnTestReadResult{buffer: buf.MultiBuffer{response}}
			return ray.link(), nil
		}}
		conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = conn.Close() }()
		plainPacketConnTestWrite(t, conn, "request", &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.50"), Port: 51820})
		plainPacketConnTestWait(t, ray.reader.secondRead, "response to enter read queue")

		readBuffer := make([]byte, 32)
		n, _, err := conn.ReadFrom(readBuffer)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(readBuffer[:n]); got != "response" {
			t.Fatalf("response = %q, want response", got)
		}
		if response.Len() != 0 {
			t.Fatal("delivered response buffer was not released")
		}
	})

	t.Run("close releases queue and owns rays exactly once", func(t *testing.T) {
		queued := plainPacketConnTestBuffer("queued response")
		var access sync.Mutex
		var rays []*plainPacketConnTestRay
		dispatcher := &plainPacketConnTestDispatcher{dispatch: func(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
			ray := newPlainPacketConnTestRay(ctx)
			access.Lock()
			if len(rays) == 0 {
				ray.reader.results <- plainPacketConnTestReadResult{buffer: buf.MultiBuffer{queued}}
			}
			rays = append(rays, ray)
			access.Unlock()
			return ray.link(), nil
		}}
		conn, err := newWireguardPlainPacketConn(context.Background(), dispatcher)
		if err != nil {
			t.Fatal(err)
		}
		plainPacketConnTestWrite(t, conn, "one", &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.60"), Port: 51820})
		plainPacketConnTestWrite(t, conn, "two", &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.61"), Port: 51820})
		access.Lock()
		firstRay := rays[0]
		access.Unlock()
		plainPacketConnTestWait(t, firstRay.reader.secondRead, "response to enter queue")

		var closeGroup sync.WaitGroup
		for i := 0; i < 12; i++ {
			closeGroup.Add(1)
			go func() {
				defer closeGroup.Done()
				if err := conn.Close(); err != nil {
					t.Errorf("Close failed: %v", err)
				}
			}()
		}
		closeGroup.Wait()
		if queued.Len() != 0 {
			t.Fatal("queued response buffer was not released by Close")
		}

		access.Lock()
		defer access.Unlock()
		if len(rays) != 2 {
			t.Fatalf("ray count = %d, want 2", len(rays))
		}
		for i, ray := range rays {
			plainPacketConnTestWait(t, ray.ctx.Done(), "ray context cancellation")
			if got := ray.reader.interruptCalls.Load(); got != 1 {
				t.Errorf("ray %d reader Interrupt calls = %d, want 1", i, got)
			}
			if got := ray.writer.interruptCalls.Load(); got != 1 {
				t.Errorf("ray %d writer Interrupt calls = %d, want 1", i, got)
			}
		}
	})
}
