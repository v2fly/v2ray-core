package outbound

import (
	"context"
	"io"
	gonet "net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type dialerTestRecord struct {
	ctx         context.Context
	destination cnet.Destination
	connection  *dialerTestConnection
}

type recordingDialer struct {
	mu      sync.Mutex
	records []dialerTestRecord
	factory func(context.Context, cnet.Destination) (internet.Connection, *dialerTestConnection, error)
}

func (d *recordingDialer) Dial(ctx context.Context, destination cnet.Destination) (internet.Connection, error) {
	if d.factory != nil {
		connection, tracked, err := d.factory(ctx, destination)
		if err != nil {
			return nil, err
		}
		d.mu.Lock()
		d.records = append(d.records, dialerTestRecord{ctx: ctx, destination: destination, connection: tracked})
		d.mu.Unlock()
		return connection, nil
	}
	connection := newDialerTestConnection()
	d.mu.Lock()
	d.records = append(d.records, dialerTestRecord{ctx: ctx, destination: destination, connection: connection})
	d.mu.Unlock()
	return connection, nil
}

func (*recordingDialer) Address() cnet.Address { return nil }

func (d *recordingDialer) snapshot() []dialerTestRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]dialerTestRecord(nil), d.records...)
}

type dialerTestConnection struct {
	reads      chan []byte
	writes     chan []byte
	closed     chan struct{}
	closeOnce  sync.Once
	closeCalls atomic.Int32
}

func newDialerTestConnection() *dialerTestConnection {
	return &dialerTestConnection{
		reads:  make(chan []byte, 8),
		writes: make(chan []byte, 8),
		closed: make(chan struct{}),
	}
}

func (c *dialerTestConnection) Read(p []byte) (int, error) {
	select {
	case payload := <-c.reads:
		return copy(p, payload), nil
	case <-c.closed:
		return 0, gonet.ErrClosed
	}
}

func (c *dialerTestConnection) Write(p []byte) (int, error) {
	payload := append([]byte(nil), p...)
	select {
	case c.writes <- payload:
		return len(p), nil
	case <-c.closed:
		return 0, gonet.ErrClosed
	}
}

func (c *dialerTestConnection) Close() error {
	c.closeCalls.Add(1)
	c.closeOnce.Do(func() { close(c.closed) })
	return nil
}

func (*dialerTestConnection) LocalAddr() gonet.Addr {
	return &gonet.UDPAddr{IP: gonet.IPv4zero}
}

func (*dialerTestConnection) RemoteAddr() gonet.Addr {
	return &gonet.UDPAddr{IP: gonet.IPv4zero}
}

func (*dialerTestConnection) SetDeadline(time.Time) error      { return nil }
func (*dialerTestConnection) SetReadDeadline(time.Time) error  { return nil }
func (*dialerTestConnection) SetWriteDeadline(time.Time) error { return nil }

func dialerTestAwaitWrite(t *testing.T, connection *dialerTestConnection, want string) {
	t.Helper()
	select {
	case got := <-connection.writes:
		if string(got) != want {
			t.Fatalf("write = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dialed connection write")
	}
}

func TestDialerDispatcherMarksContextAndOwnsConnection(t *testing.T) {
	type contextKey struct{}
	wireguard := new(WireguardOutbound)
	otherWireguard := new(WireguardOutbound)
	dialer := new(recordingDialer)
	dispatcher, err := newDialerDispatcher(wireguard, dialer)
	if err != nil {
		t.Fatal(err)
	}

	destination := cnet.UDPDestination(cnet.ParseAddress("192.0.2.1"), 51820)
	baseCtx := context.WithValue(context.Background(), contextKey{}, "preserved")
	link, err := dispatcher.Dispatch(baseCtx, destination)
	if err != nil {
		t.Fatal(err)
	}
	records := dialer.snapshot()
	if len(records) != 1 {
		t.Fatalf("dial count = %d, want 1", len(records))
	}
	if records[0].destination != destination {
		t.Fatalf("destination = %v, want %v", records[0].destination, destination)
	}
	if got := records[0].ctx.Value(contextKey{}); got != "preserved" {
		t.Fatalf("context value = %v, want preserved", got)
	}
	if !isWireguardUnderlayFor(records[0].ctx, wireguard) {
		t.Fatal("dial context is not marked for its WireGuard outbound")
	}
	if isWireguardUnderlayFor(records[0].ctx, otherWireguard) {
		t.Fatal("dial context was marked for an unrelated WireGuard outbound")
	}
	otherSessionCtx := preserveWireguardUnderlayMarkers(context.Background(), records[0].ctx)
	otherDialCtx := markWireguardUnderlay(otherSessionCtx, otherWireguard)
	if !isWireguardUnderlayFor(otherDialCtx, wireguard) || !isWireguardUnderlayFor(otherDialCtx, otherWireguard) {
		t.Fatal("WireGuard underlay marker chain was not preserved across shared-session context creation")
	}
	if err := wireguard.Process(records[0].ctx, nil, nil); err == nil || !strings.Contains(err.Error(), "cannot use itself") {
		t.Fatalf("recursive Process error = %v, want self-underlay rejection", err)
	}

	var closeGroup sync.WaitGroup
	for range 8 {
		closeGroup.Add(2)
		go func() {
			defer closeGroup.Done()
			_ = common.Interrupt(link.Reader)
		}()
		go func() {
			defer closeGroup.Done()
			_ = common.Interrupt(link.Writer)
		}()
	}
	closeGroup.Wait()
	if got := records[0].connection.closeCalls.Load(); got != 1 {
		t.Fatalf("underlying Close calls = %d, want 1", got)
	}
	select {
	case <-records[0].ctx.Done():
	default:
		t.Fatal("closing the link did not cancel the dial context")
	}
}

func TestDialerDispatcherPlainUDPReusesConnectionsByDestination(t *testing.T) {
	wireguard := new(WireguardOutbound)
	dialer := new(recordingDialer)
	dispatcher, err := newDialerDispatcher(wireguard, dialer)
	if err != nil {
		t.Fatal(err)
	}
	packetConn, err := createLogicalPacketConn(context.Background(), dispatcher, packetaddr.PacketAddrType_None)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = packetConn.Close() }()

	addressA := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.1"), Port: 51820}
	addressB := &gonet.UDPAddr{IP: gonet.ParseIP("192.0.2.2"), Port: 51821}
	for _, write := range []struct {
		payload string
		address *gonet.UDPAddr
	}{
		{payload: "a1", address: addressA},
		{payload: "a2", address: addressA},
		{payload: "b1", address: addressB},
	} {
		if n, err := packetConn.WriteTo([]byte(write.payload), write.address); err != nil || n != len(write.payload) {
			t.Fatalf("WriteTo(%q) = (%d, %v)", write.payload, n, err)
		}
	}

	records := dialer.snapshot()
	if len(records) != 2 {
		t.Fatalf("dial count = %d, want one per destination (2)", len(records))
	}
	if got, want := records[0].destination, cnet.DestinationFromAddr(addressA); got != want {
		t.Fatalf("first destination = %v, want %v", got, want)
	}
	if got, want := records[1].destination, cnet.DestinationFromAddr(addressB); got != want {
		t.Fatalf("second destination = %v, want %v", got, want)
	}
	dialerTestAwaitWrite(t, records[0].connection, "a1")
	dialerTestAwaitWrite(t, records[0].connection, "a2")
	dialerTestAwaitWrite(t, records[1].connection, "b1")

	records[1].connection.reads <- []byte("reply from b")
	readBuffer := make([]byte, 64)
	n, source, err := packetConn.ReadFrom(readBuffer)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(readBuffer[:n]); got != "reply from b" {
		t.Fatalf("reply = %q, want %q", got, "reply from b")
	}
	if got := source.String(); got != addressB.String() {
		t.Fatalf("reply source = %q, want %q", got, addressB)
	}

	if err := packetConn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := packetConn.Close(); err != nil {
		t.Fatal(err)
	}
	for i, record := range records {
		if got := record.connection.closeCalls.Load(); got != 1 {
			t.Fatalf("connection %d Close calls = %d, want 1", i, got)
		}
	}
}

func TestDialerDispatcherPacketAddrMagicDestinations(t *testing.T) {
	tests := []struct {
		name        string
		encoding    packetaddr.PacketAddrType
		wantNetwork cnet.Network
		wantAddress string
	}{
		{
			name:        "packet",
			encoding:    packetaddr.PacketAddrType_Packet,
			wantNetwork: cnet.Network_UDP,
			wantAddress: "sp.packet-addr.v2fly.arpa",
		},
		{
			name:        "stream",
			encoding:    packetaddr.PacketAddrType_Stream,
			wantNetwork: cnet.Network_TCP,
			wantAddress: "st.packet-addr.v2fly.arpa",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialer := new(recordingDialer)
			dispatcher, err := newDialerDispatcher(new(WireguardOutbound), dialer)
			if err != nil {
				t.Fatal(err)
			}
			packetConn, err := createLogicalPacketConn(context.Background(), dispatcher, test.encoding)
			if err != nil {
				t.Fatal(err)
			}

			records := dialer.snapshot()
			if len(records) != 1 {
				t.Fatalf("dial count = %d, want 1", len(records))
			}
			destination := records[0].destination
			if destination.Network != test.wantNetwork {
				t.Fatalf("network = %v, want %v", destination.Network, test.wantNetwork)
			}
			if got := destination.Address.String(); got != test.wantAddress {
				t.Fatalf("address = %q, want %q", got, test.wantAddress)
			}
			if destination.Port != 0 {
				t.Fatalf("port = %d, want 0", destination.Port)
			}

			if err := packetConn.Close(); err != nil {
				t.Fatal(err)
			}
			if got := records[0].connection.closeCalls.Load(); got != 1 {
				t.Fatalf("underlying Close calls = %d, want 1", got)
			}
		})
	}
}

type multiBufferDialerTestConnection struct {
	*dialerTestConnection
	readMulti chan buf.MultiBuffer
}

func newMultiBufferDialerTestConnection() *multiBufferDialerTestConnection {
	return &multiBufferDialerTestConnection{
		dialerTestConnection: newDialerTestConnection(),
		readMulti:            make(chan buf.MultiBuffer, 1),
	}
}

func (c *multiBufferDialerTestConnection) ReadMultiBuffer() (buf.MultiBuffer, error) {
	select {
	case mb := <-c.readMulti:
		return mb, nil
	case <-c.closed:
		return nil, io.ErrClosedPipe
	}
}

type dialerTestCounter struct {
	value atomic.Int64
}

func (c *dialerTestCounter) Value() int64 { return c.value.Load() }
func (c *dialerTestCounter) Set(value int64) int64 {
	return c.value.Swap(value)
}

func (c *dialerTestCounter) Add(delta int64) int64 {
	return c.value.Add(delta) - delta
}

func TestDialerDispatcherPreservesUDPBoundariesThroughStatConnection(t *testing.T) {
	underlying := newMultiBufferDialerTestConnection()
	readCounter := new(dialerTestCounter)
	statConnection := &internet.StatCouterConnection{
		Connection:  underlying,
		ReadCounter: readCounter,
	}
	dialer := &recordingDialer{
		factory: func(context.Context, cnet.Destination) (internet.Connection, *dialerTestConnection, error) {
			return statConnection, underlying.dialerTestConnection, nil
		},
	}
	dispatcher, err := newDialerDispatcher(new(WireguardOutbound), dialer)
	if err != nil {
		t.Fatal(err)
	}
	link, err := dispatcher.Dispatch(context.Background(), cnet.UDPDestination(cnet.ParseAddress("192.0.2.1"), 51820))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = common.Interrupt(link.Reader) }()

	underlying.readMulti <- buf.MultiBuffer{buf.FromBytes([]byte("first")), buf.FromBytes([]byte("second"))}
	mb, err := link.Reader.ReadMultiBuffer()
	if err != nil {
		t.Fatal(err)
	}
	defer buf.ReleaseMulti(mb)
	if len(mb) != 2 {
		t.Fatalf("packet buffer count = %d, want 2", len(mb))
	}
	if got := string(mb[0].Bytes()); got != "first" {
		t.Fatalf("first packet = %q, want first", got)
	}
	if got := string(mb[1].Bytes()); got != "second" {
		t.Fatalf("second packet = %q, want second", got)
	}
	if got := readCounter.Value(); got != int64(len("first")+len("second")) {
		t.Fatalf("read counter = %d, want %d", got, len("first")+len("second"))
	}
}
