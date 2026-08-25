package wgcommon

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/conn"
)

type reconnectReadResult struct {
	packet []byte
	addr   net.Addr
	err    error
}

type reconnectTestPacketConn struct {
	reads       chan reconnectReadResult
	readStarted chan struct{}
	closed      chan struct{}
	closeOnce   sync.Once
	closes      atomic.Int32

	writesMu sync.Mutex
	writes   [][]byte
	write    func([]byte, net.Addr) (int, error)
	deadline func(time.Time) error
	local    net.Addr
}

func newReconnectTestPacketConn(port int) *reconnectTestPacketConn {
	return &reconnectTestPacketConn{
		reads:  make(chan reconnectReadResult, 8),
		closed: make(chan struct{}),
		local:  &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port},
	}
}

func (c *reconnectTestPacketConn) ReadFrom(packet []byte) (int, net.Addr, error) {
	if c.readStarted != nil {
		select {
		case c.readStarted <- struct{}{}:
		default:
		}
	}
	select {
	case result := <-c.reads:
		n := copy(packet, result.packet)
		if result.addr == nil {
			result.addr = &net.UDPAddr{IP: net.IPv4(192, 0, 2, 1), Port: 51820}
		}
		return n, result.addr, result.err
	case <-c.closed:
		return 0, nil, net.ErrClosed
	}
}

func (c *reconnectTestPacketConn) WriteTo(packet []byte, addr net.Addr) (int, error) {
	select {
	case <-c.closed:
		return 0, net.ErrClosed
	default:
	}
	if c.write != nil {
		n, err := c.write(packet, addr)
		if err != nil {
			return n, err
		}
	}
	copyOfPacket := append([]byte(nil), packet...)
	c.writesMu.Lock()
	c.writes = append(c.writes, copyOfPacket)
	c.writesMu.Unlock()
	return len(packet), nil
}

func (c *reconnectTestPacketConn) Close() error {
	c.closeOnce.Do(func() {
		c.closes.Add(1)
		close(c.closed)
	})
	return nil
}

func (c *reconnectTestPacketConn) LocalAddr() net.Addr { return c.local }

func (c *reconnectTestPacketConn) SetDeadline(time.Time) error { return nil }

func (c *reconnectTestPacketConn) SetReadDeadline(deadline time.Time) error {
	if c.deadline != nil {
		return c.deadline(deadline)
	}
	return nil
}

func (c *reconnectTestPacketConn) SetWriteDeadline(time.Time) error { return nil }

func (c *reconnectTestPacketConn) writeCount() int {
	c.writesMu.Lock()
	defer c.writesMu.Unlock()
	return len(c.writes)
}

type reconnectFactory struct {
	mu     sync.Mutex
	conns  []net.PacketConn
	errs   []error
	calls  int
	callCh chan int
}

func newReconnectFactory(conns ...net.PacketConn) *reconnectFactory {
	return &reconnectFactory{conns: conns, callCh: make(chan int, 16)}
}

func (f *reconnectFactory) create(context.Context) (net.PacketConn, error) {
	f.mu.Lock()
	index := f.calls
	f.calls++
	call := f.calls
	var packetConn net.PacketConn
	var err error
	if index < len(f.conns) {
		packetConn = f.conns[index]
	} else {
		err = errors.New("unexpected packet connection factory call")
	}
	if index < len(f.errs) && f.errs[index] != nil {
		err = f.errs[index]
	}
	f.mu.Unlock()
	f.callCh <- call
	return packetConn, err
}

func (f *reconnectFactory) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *reconnectFactory) awaitCall(t *testing.T, want int) {
	t.Helper()
	select {
	case got := <-f.callCh:
		if got != want {
			t.Fatalf("unexpected factory call: got %d, want %d", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for factory call %d", want)
	}
}

type reconnectClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *reconnectClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *reconnectClock) Advance(elapsed time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(elapsed)
	c.mu.Unlock()
}

func parseReconnectEndpoint(t *testing.T, bind conn.Bind) conn.Endpoint {
	t.Helper()
	endpoint, err := bind.ParseEndpoint("198.51.100.1:51820")
	if err != nil {
		t.Fatal(err)
	}
	return endpoint
}

func TestReconnectingBindOpenClosesConnReturnedWithError(t *testing.T) {
	errFactory := errors.New("factory failed")
	packetConn := newReconnectTestPacketConn(10001)
	factory := newReconnectFactory(packetConn)
	factory.errs = []error{errFactory}
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)

	if _, _, err := bind.Open(0); !errors.Is(err, errFactory) {
		t.Fatalf("Open error = %v, want %v", err, errFactory)
	}
	factory.awaitCall(t, 1)
	if got := packetConn.closes.Load(); got != 1 {
		t.Fatalf("factory connection close count = %d, want 1", got)
	}
}

func TestReconnectingBindWriteErrorReplacesWithoutReplay(t *testing.T) {
	errWrite := errors.New("write failed")
	first := newReconnectTestPacketConn(10001)
	first.write = func([]byte, net.Addr) (int, error) { return 0, errWrite }
	second := newReconnectTestPacketConn(10002)
	factory := newReconnectFactory(first, second)
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)
	t.Cleanup(func() { _ = bind.Close() })

	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	endpoint := parseReconnectEndpoint(t, bind)

	if err := bind.Send([][]byte{[]byte("failed")}, endpoint); !errors.Is(err, errWrite) {
		t.Fatalf("Send error = %v, want %v", err, errWrite)
	}
	factory.awaitCall(t, 2)
	if got := first.closes.Load(); got != 1 {
		t.Fatalf("first connection close count = %d, want 1", got)
	}
	if got := second.writeCount(); got != 0 {
		t.Fatalf("failed datagram was replayed on replacement: %d writes", got)
	}

	if err := bind.Send([][]byte{[]byte("next")}, endpoint); err != nil {
		t.Fatal(err)
	}
	if got := second.writeCount(); got != 1 {
		t.Fatalf("second connection write count = %d, want 1", got)
	}
}

func TestReconnectingBindReplacementClosesConnReturnedWithError(t *testing.T) {
	errWrite := errors.New("write failed")
	errFactory := errors.New("factory failed")
	first := newReconnectTestPacketConn(10001)
	first.write = func([]byte, net.Addr) (int, error) { return 0, errWrite }
	failedReplacement := newReconnectTestPacketConn(10002)
	third := newReconnectTestPacketConn(10003)
	factory := newReconnectFactory(first, failedReplacement, third)
	factory.errs = []error{nil, errFactory}
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)
	t.Cleanup(func() { _ = bind.Close() })

	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	endpoint := parseReconnectEndpoint(t, bind)
	if err := bind.Send([][]byte{[]byte("failed")}, endpoint); !errors.Is(err, errWrite) {
		t.Fatalf("Send error = %v, want %v", err, errWrite)
	}
	factory.awaitCall(t, 2)
	if got := failedReplacement.closes.Load(); got != 1 {
		t.Fatalf("failed replacement close count = %d, want 1", got)
	}

	if err := bind.Send([][]byte{[]byte("next")}, endpoint); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 3)
	if got := third.writeCount(); got != 1 {
		t.Fatalf("third connection write count = %d, want 1", got)
	}
}

func TestReconnectingBindReadErrorPhysicallyClosesAndContinues(t *testing.T) {
	errRead := errors.New("read failed")
	first := newReconnectTestPacketConn(10001)
	second := newReconnectTestPacketConn(10002)
	factory := newReconnectFactory(first, second)
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)
	t.Cleanup(func() { _ = bind.Close() })

	receivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)

	type receiveResult struct {
		n    int
		size int
		ep   conn.Endpoint
		err  error
	}
	resultCh := make(chan receiveResult, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		n, err := receivers[0]([][]byte{packet}, sizes, endpoints)
		resultCh <- receiveResult{n: n, size: sizes[0], ep: endpoints[0], err: err}
	}()

	first.reads <- reconnectReadResult{err: errRead}
	factory.awaitCall(t, 2)
	if got := first.closes.Load(); got != 1 {
		t.Fatalf("first connection close count = %d, want 1", got)
	}
	second.reads <- reconnectReadResult{packet: []byte("reply")}

	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.n != 1 || result.size != len("reply") || result.ep == nil {
			t.Fatalf("unexpected receive result: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for receive after replacement")
	}
}

func TestReconnectingBindReceiveRetriesFactoryFailure(t *testing.T) {
	errRead := errors.New("read failed")
	errFactory := errors.New("factory failed")
	first := newReconnectTestPacketConn(10001)
	second := newReconnectTestPacketConn(10002)
	factory := newReconnectFactory(first, nil, second)
	factory.errs = []error{nil, errFactory}
	reconnectingBind := newReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute, time.Now).(*reconnectingNetPacketConnToWg)
	reconnectingBind.retryWait = func(context.Context, time.Duration) error { return nil }
	var bind conn.Bind = reconnectingBind
	t.Cleanup(func() { _ = bind.Close() })

	receivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	receiveErr := make(chan error, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		_, err := receivers[0]([][]byte{packet}, sizes, endpoints)
		receiveErr <- err
	}()

	first.reads <- reconnectReadResult{err: errRead}
	factory.awaitCall(t, 2)
	factory.awaitCall(t, 3)
	second.reads <- reconnectReadResult{packet: []byte("recovered")}
	select {
	case err := <-receiveErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("same receive function did not recover after factory failure")
	}
	if got := factory.callCount(); got != 3 {
		t.Fatalf("factory call count = %d, want 3", got)
	}
}

func TestReconnectingBindCloseCancelsFactoryRetryBackoff(t *testing.T) {
	errRead := errors.New("read failed")
	errFactory := errors.New("factory failed")
	first := newReconnectTestPacketConn(10001)
	factory := newReconnectFactory(first, nil)
	factory.errs = []error{nil, errFactory}
	reconnectingBind := newReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute, time.Now).(*reconnectingNetPacketConnToWg)
	retryStarted := make(chan struct{})
	reconnectingBind.retryWait = func(ctx context.Context, _ time.Duration) error {
		close(retryStarted)
		<-ctx.Done()
		return ctx.Err()
	}
	var bind conn.Bind = reconnectingBind

	receivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	receiveErr := make(chan error, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		_, err := receivers[0]([][]byte{packet}, sizes, endpoints)
		receiveErr <- err
	}()
	first.reads <- reconnectReadResult{err: errRead}
	factory.awaitCall(t, 2)
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("receive function did not enter retry backoff")
	}
	if err := bind.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-receiveErr:
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("receive error = %v, want net.ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel receive retry backoff")
	}
	if got := factory.callCount(); got != 2 {
		t.Fatalf("closed bind retried factory: %d calls", got)
	}
}

func TestReconnectingBindConcurrentFailuresCreateOneGeneration(t *testing.T) {
	errWrite := errors.New("write failed")
	first := newReconnectTestPacketConn(10001)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	first.write = func([]byte, net.Addr) (int, error) {
		started <- struct{}{}
		<-release
		return 0, errWrite
	}
	second := newReconnectTestPacketConn(10002)
	factory := newReconnectFactory(first, second)
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)
	t.Cleanup(func() { _ = bind.Close() })
	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	endpoint := parseReconnectEndpoint(t, bind)

	errorsCh := make(chan error, 2)
	for range 2 {
		go func() { errorsCh <- bind.Send([][]byte{[]byte("packet")}, endpoint) }()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for concurrent writes")
		}
	}
	close(release)
	for range 2 {
		if err := <-errorsCh; !errors.Is(err, errWrite) {
			t.Fatalf("Send error = %v, want %v", err, errWrite)
		}
	}
	factory.awaitCall(t, 2)
	if got := factory.callCount(); got != 2 {
		t.Fatalf("factory call count = %d, want 2", got)
	}
	if got := first.closes.Load(); got != 1 {
		t.Fatalf("first connection close count = %d, want 1", got)
	}
}

func TestReconnectingBindConcurrentFailuresDoNotStormAfterFactoryError(t *testing.T) {
	errWrite := errors.New("write failed")
	errFactory := errors.New("factory failed")
	first := newReconnectTestPacketConn(10001)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	first.write = func([]byte, net.Addr) (int, error) {
		started <- struct{}{}
		<-release
		return 0, errWrite
	}
	third := newReconnectTestPacketConn(10003)
	factory := newReconnectFactory(first, nil, third)
	factory.errs = []error{nil, errFactory}
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)
	t.Cleanup(func() { _ = bind.Close() })
	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	endpoint := parseReconnectEndpoint(t, bind)

	errorsCh := make(chan error, 2)
	for range 2 {
		go func() { errorsCh <- bind.Send([][]byte{[]byte("packet")}, endpoint) }()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for concurrent writes")
		}
	}
	close(release)
	for range 2 {
		if err := <-errorsCh; !errors.Is(err, errWrite) {
			t.Fatalf("Send error = %v, want %v", err, errWrite)
		}
	}
	factory.awaitCall(t, 2)
	if got := factory.callCount(); got != 2 {
		t.Fatalf("concurrent stale failures caused %d factory calls, want 2", got)
	}

	if err := bind.Send([][]byte{[]byte("recovered")}, endpoint); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 3)
	if got := third.writeCount(); got != 1 {
		t.Fatalf("recovered connection write count = %d, want 1", got)
	}
}

func TestReconnectingBindTimeoutIsLazyAndReceiveResetsIt(t *testing.T) {
	clock := &reconnectClock{now: time.Unix(1_000, 0)}
	first := newReconnectTestPacketConn(10001)
	first.readStarted = make(chan struct{}, 1)
	second := newReconnectTestPacketConn(10002)
	third := newReconnectTestPacketConn(10003)
	factory := newReconnectFactory(first, second, third)
	bind := newReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute, clock.Now)
	t.Cleanup(func() { _ = bind.Close() })
	receivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	endpoint := parseReconnectEndpoint(t, bind)

	clock.Advance(10 * time.Minute)
	if got := factory.callCount(); got != 1 {
		t.Fatalf("idle bind churned connections: %d factory calls", got)
	}

	resultCh := make(chan error, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		_, err := receivers[0]([][]byte{packet}, sizes, endpoints)
		resultCh <- err
	}()
	select {
	case <-first.readStarted:
	case <-time.After(time.Second):
		t.Fatal("receive function did not block on the first generation")
	}

	if err := bind.Send([][]byte{[]byte("wake")}, endpoint); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 2)
	if got := first.closes.Load(); got != 1 {
		t.Fatalf("timed-out connection close count = %d, want 1", got)
	}
	second.reads <- reconnectReadResult{packet: []byte("received")}
	select {
	case err := <-resultCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for receive on replacement")
	}
	if got := factory.callCount(); got != 2 {
		t.Fatalf("stale read error replaced the new generation: %d calls", got)
	}

	clock.Advance(59 * time.Second)
	if err := bind.Send([][]byte{[]byte("before threshold")}, endpoint); err != nil {
		t.Fatal(err)
	}
	if got := factory.callCount(); got != 2 {
		t.Fatalf("connection replaced before threshold: %d calls", got)
	}
	clock.Advance(time.Second)
	if err := bind.Send([][]byte{[]byte("at threshold")}, endpoint); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 3)
	if got := second.closes.Load(); got != 1 {
		t.Fatalf("second connection close count = %d, want 1", got)
	}
}

func TestReconnectingBindZeroTimeoutDisablesSilentReplacement(t *testing.T) {
	clock := &reconnectClock{now: time.Unix(1_000, 0)}
	packetConn := newReconnectTestPacketConn(10001)
	factory := newReconnectFactory(packetConn)
	bind := newReconnectingNetPacketConnToWg(context.Background(), factory.create, 0, clock.Now)
	t.Cleanup(func() { _ = bind.Close() })
	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	clock.Advance(24 * time.Hour)
	if err := bind.Send([][]byte{[]byte("still live")}, parseReconnectEndpoint(t, bind)); err != nil {
		t.Fatal(err)
	}
	if got := factory.callCount(); got != 1 {
		t.Fatalf("zero timeout created %d connections, want 1", got)
	}
}

func TestReconnectingBindCloseStopsOldReceiversAndAllowsReopen(t *testing.T) {
	first := newReconnectTestPacketConn(10001)
	second := newReconnectTestPacketConn(10002)
	factory := newReconnectFactory(first, second)
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory.create, time.Minute)

	oldReceivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 1)
	receiveErr := make(chan error, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		_, err := oldReceivers[0]([][]byte{packet}, sizes, endpoints)
		receiveErr <- err
	}()

	if err := bind.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-receiveErr:
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("receive error = %v, want net.ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not unblock receive")
	}
	if err := bind.Send([][]byte{[]byte("closed")}, parseReconnectEndpoint(t, bind)); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Send after Close error = %v, want net.ErrClosed", err)
	}
	if got := factory.callCount(); got != 1 {
		t.Fatalf("closed bind recreated a connection: %d calls", got)
	}

	newReceivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	factory.awaitCall(t, 2)
	defer func() { _ = bind.Close() }()
	if len(newReceivers) != 1 {
		t.Fatalf("new receive function count = %d, want 1", len(newReceivers))
	}

	packet := make([]byte, 64)
	sizes := make([]int, 1)
	endpoints := make([]conn.Endpoint, 1)
	if _, err := oldReceivers[0]([][]byte{packet}, sizes, endpoints); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("old receiver after reopen error = %v, want net.ErrClosed", err)
	}
}

func TestReconnectingBindCloseCancelsInFlightReplacement(t *testing.T) {
	errWrite := errors.New("write failed")
	first := newReconnectTestPacketConn(10001)
	first.write = func([]byte, net.Addr) (int, error) { return 0, errWrite }
	late := newReconnectTestPacketConn(10002)
	third := newReconnectTestPacketConn(10003)
	replacementStarted := make(chan struct{})
	var calls atomic.Int32
	factory := func(ctx context.Context) (net.PacketConn, error) {
		switch calls.Add(1) {
		case 1:
			return first, nil
		case 2:
			close(replacementStarted)
			<-ctx.Done()
			return late, nil
		case 3:
			return third, nil
		default:
			return nil, errors.New("unexpected factory call")
		}
	}
	bind := NewReconnectingNetPacketConnToWg(context.Background(), factory, time.Minute)
	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	endpoint := parseReconnectEndpoint(t, bind)

	sendDone := make(chan error, 1)
	go func() { sendDone <- bind.Send([][]byte{[]byte("failed")}, endpoint) }()
	select {
	case <-replacementStarted:
	case <-time.After(time.Second):
		t.Fatal("replacement factory did not start")
	}
	if err := bind.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-sendDone:
		if !errors.Is(err, errWrite) {
			t.Fatalf("Send error = %v, want %v", err, errWrite)
		}
	case <-time.After(time.Second):
		t.Fatal("in-flight replacement did not stop after Close")
	}
	if got := late.closes.Load(); got != 1 {
		t.Fatalf("late factory connection close count = %d, want 1", got)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("factory call count after Close = %d, want 2", got)
	}

	if _, _, err := bind.Open(0); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bind.Close() }()
	if err := bind.Send([][]byte{[]byte("reopened")}, endpoint); err != nil {
		t.Fatal(err)
	}
	if got := third.writeCount(); got != 1 {
		t.Fatalf("reopened connection write count = %d, want 1", got)
	}
}

func TestStaticBindCloseMapsReadWakeupToNetErrClosed(t *testing.T) {
	packetConn := newReconnectTestPacketConn(10001)
	packetConn.deadline = func(deadline time.Time) error {
		if !deadline.IsZero() {
			packetConn.reads <- reconnectReadResult{err: errors.New("deadline wakeup")}
		}
		return nil
	}
	bind := NewNetPacketConnToWg(packetConn)
	receivers, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	receiveErr := make(chan error, 1)
	go func() {
		packet := make([]byte, 64)
		sizes := make([]int, 1)
		endpoints := make([]conn.Endpoint, 1)
		_, err := receivers[0]([][]byte{packet}, sizes, endpoints)
		receiveErr <- err
	}()
	if err := bind.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-receiveErr:
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("receive error = %v, want net.ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("static bind Close did not unblock receive")
	}
}
