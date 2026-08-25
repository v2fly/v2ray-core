package outbound

import (
	"context"
	"errors"
	gonet "net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

type recordingDispatcher struct {
	mu          sync.Mutex
	destination cnet.Destination
}

func (d *recordingDispatcher) Dispatch(_ context.Context, destination cnet.Destination) (*transport.Link, error) {
	d.mu.Lock()
	d.destination = destination
	d.mu.Unlock()

	_, uplinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	downlinkReader, _ := pipe.New(pipe.WithSizeLimit(1024))
	return &transport.Link{Reader: downlinkReader, Writer: uplinkWriter}, nil
}

func (*recordingDispatcher) Start() error      { return nil }
func (*recordingDispatcher) Close() error      { return nil }
func (*recordingDispatcher) Type() interface{} { return routing.DispatcherType() }

func (d *recordingDispatcher) lastDestination() cnet.Destination {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.destination
}

func TestValidateWireguardConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{name: "nil config", wantErr: true},
		{
			name: "restricted system network",
			config: &Config{
				Restricted:            true,
				ListenOnSystemNetwork: true,
			},
			wantErr: true,
		},
		{
			name:   "restricted logical network",
			config: &Config{Restricted: true},
		},
		{
			name:   "unrestricted system network",
			config: &Config{ListenOnSystemNetwork: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateWireguardConfig(test.config)
			if got := err != nil; got != test.wantErr {
				t.Fatalf("validateWireguardConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestNewWireguardOutboundValidatesBeforeInitialization(t *testing.T) {
	_, err := NewWireguardOutbound(context.Background(), &Config{
		Restricted:            true,
		ListenOnSystemNetwork: true,
	})
	if err == nil {
		t.Fatal("NewWireguardOutbound() accepted restricted system network config")
	}
	if !strings.Contains(err.Error(), "restricted WireGuard outbound cannot listen on the system network") {
		t.Fatalf("NewWireguardOutbound() error = %q, want restricted system-network validation error", err)
	}
}

func TestCreateLogicalPacketConn(t *testing.T) {
	tests := []struct {
		name        string
		encoding    packetaddr.PacketAddrType
		wantNetwork cnet.Network
		wantAddress string
		writeFirst  bool
	}{
		{
			name:        "plain UDP",
			encoding:    packetaddr.PacketAddrType_None,
			wantNetwork: cnet.Network_UDP,
			wantAddress: "192.0.2.1",
			writeFirst:  true,
		},
		{
			name:        "packet packetaddr",
			encoding:    packetaddr.PacketAddrType_Packet,
			wantNetwork: cnet.Network_UDP,
			wantAddress: "sp.packet-addr.v2fly.arpa",
		},
		{
			name:        "stream packetaddr",
			encoding:    packetaddr.PacketAddrType_Stream,
			wantNetwork: cnet.Network_TCP,
			wantAddress: "st.packet-addr.v2fly.arpa",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dispatcher := new(recordingDispatcher)
			conn, err := createLogicalPacketConn(context.Background(), dispatcher, test.encoding)
			if err != nil {
				t.Fatalf("createLogicalPacketConn() failed: %v", err)
			}
			defer func() { _ = conn.Close() }()

			if test.writeFirst {
				if _, err := conn.WriteTo([]byte("wireguard"), &gonet.UDPAddr{IP: gonet.ParseIP(test.wantAddress), Port: 51820}); err != nil {
					t.Fatalf("WriteTo() failed: %v", err)
				}
			}

			destination := dispatcher.lastDestination()
			if destination.Network != test.wantNetwork {
				t.Fatalf("network = %v, want %v", destination.Network, test.wantNetwork)
			}
			if got := destination.Address.String(); got != test.wantAddress {
				t.Fatalf("address = %q, want %q", got, test.wantAddress)
			}
		})
	}
}

func TestCreateLogicalPacketConnRejectsUnsupportedEncoding(t *testing.T) {
	if _, err := createLogicalPacketConn(context.Background(), new(recordingDispatcher), packetaddr.PacketAddrType(99)); err == nil {
		t.Fatal("createLogicalPacketConn() accepted an unsupported encoding")
	}
}

func TestOutboundNoReceiveTimeout(t *testing.T) {
	disabled := uint32(0)
	custom := uint32(7)
	tests := []struct {
		name   string
		config *Config
		want   time.Duration
	}{
		{name: "nil config", config: nil, want: time.Minute},
		{name: "omitted", config: new(Config), want: time.Minute},
		{name: "disabled", config: &Config{OutboundNoReceiveTimeoutSec: &disabled}, want: 0},
		{name: "custom", config: &Config{OutboundNoReceiveTimeoutSec: &custom}, want: 7 * time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := outboundNoReceiveTimeout(test.config); got != test.want {
				t.Fatalf("outboundNoReceiveTimeout() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestClientConnStateRetriesFailedInitialization(t *testing.T) {
	state, err := NewClientConnState()
	if err != nil {
		t.Fatal(err)
	}
	attempts := 0
	want := new(WireguardOutboundSession)
	create := func() (*WireguardOutboundSession, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("temporary failure")
		}
		return want, nil
	}

	if _, err := state.GetOrCreateSession(create); err == nil {
		t.Fatal("first initialization unexpectedly succeeded")
	}
	got, err := state.GetOrCreateSession(create)
	if err != nil {
		t.Fatalf("second initialization failed: %v", err)
	}
	if got != want {
		t.Fatalf("session = %p, want %p", got, want)
	}
	got, err = state.GetOrCreateSession(create)
	if err != nil || got != want {
		t.Fatalf("cached session = (%p, %v), want (%p, nil)", got, err, want)
	}
	if attempts != 2 {
		t.Fatalf("initialization attempts = %d, want 2", attempts)
	}
}

func TestClientConnStateConcurrentInitialization(t *testing.T) {
	state, err := NewClientConnState()
	if err != nil {
		t.Fatal(err)
	}
	want := new(WireguardOutboundSession)
	var attempts atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	create := func() (*WireguardOutboundSession, error) {
		if attempt := attempts.Add(1); attempt != 1 {
			return nil, errors.New("duplicate concurrent initialization")
		}
		close(started)
		<-release
		return want, nil
	}

	const callers = 16
	results := make(chan *WireguardOutboundSession, callers)
	errs := make(chan error, callers)
	var workers sync.WaitGroup
	workers.Add(callers)
	for range callers {
		go func() {
			defer workers.Done()
			session, err := state.GetOrCreateSession(create)
			results <- session
			errs <- err
		}()
	}
	<-started
	close(release)
	workers.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("GetOrCreateSession() failed: %v", err)
		}
	}
	for got := range results {
		if got != want {
			t.Fatalf("session = %p, want %p", got, want)
		}
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("initialization attempts = %d, want 1", got)
	}
}

func TestClientConnStateClosePreventsInitialization(t *testing.T) {
	state, err := NewClientConnState()
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
	if _, err := state.GetOrCreateSession(func() (*WireguardOutboundSession, error) {
		return new(WireguardOutboundSession), nil
	}); err == nil {
		t.Fatal("closed state allowed initialization")
	}
}

func TestClientConnStateCloseCancelsInitialization(t *testing.T) {
	state, err := NewClientConnState()
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	createDone := make(chan error, 1)
	go func() {
		_, err := state.GetOrCreateSessionWithContext(context.Background(), func(ctx context.Context) (*WireguardOutboundSession, error) {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		})
		createDone <- err
	}()
	<-started

	closeDone := make(chan error, 1)
	go func() { closeDone <- state.Close() }()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() did not cancel in-flight initialization")
	}
	select {
	case err := <-createDone:
		if err == nil {
			t.Fatal("canceled initialization unexpectedly succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("initialization did not return after cancellation")
	}
}

func TestClientConnStateCloseWaitsForActiveSession(t *testing.T) {
	state, err := NewClientConnState()
	if err != nil {
		t.Fatal(err)
	}
	sess, release, err := state.AcquireOrCreateSessionWithContext(context.Background(), func(ctx context.Context) (*WireguardOutboundSession, error) {
		return &WireguardOutboundSession{ctx: ctx}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	closeDone := make(chan error, 1)
	go func() { closeDone <- state.Close() }()
	select {
	case <-sess.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("Close() did not cancel the shared session context")
	}
	select {
	case err := <-closeDone:
		t.Fatalf("Close() returned before the active session was released: %v", err)
	default:
	}

	release()
	release()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() did not finish after the active session was released")
	}
}
