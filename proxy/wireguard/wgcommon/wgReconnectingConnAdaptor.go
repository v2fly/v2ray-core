package wgcommon

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/conn"
)

// PacketConnFactory creates a PacketConn for one logical WireGuard underlay
// generation. The supplied context is canceled when the bind is closed.
type PacketConnFactory func(context.Context) (net.PacketConn, error)

var (
	errNilPacketConnFactory  = errors.New("wireguard: packet connection factory is nil")
	errNilPacketConn         = errors.New("wireguard: packet connection factory returned nil")
	errPacketConnUnavailable = errors.New("wireguard: packet connection is unavailable")
)

// reconnectingNetPacketConnToWg owns the PacketConns returned by its factory.
// WireGuard's Bind Close/Open lifecycle is distinct from connection generation:
// a connection may be replaced multiple times while one Bind Open remains live.
type reconnectingNetPacketConnToWg struct {
	ctx              context.Context
	factory          PacketConnFactory
	noReceiveTimeout time.Duration
	now              func() time.Time
	retryWait        func(context.Context, time.Duration) error

	// replaceMu serializes factories and connection generation changes. Close
	// deliberately does not take it: it must be able to cancel an in-flight
	// factory and physically close a blocked read immediately.
	replaceMu   sync.Mutex
	mu          sync.Mutex
	open        bool
	lifecycle   uint64
	generation  uint64
	openCtx     context.Context
	cancel      context.CancelFunc
	conn        net.PacketConn
	actualPort  uint16
	lastReceive time.Time
}

// NewReconnectingNetPacketConnToWg constructs a WireGuard bind that owns and
// recreates factory-produced PacketConns. A zero or negative timeout disables
// replacement based solely on the absence of received packets; visible I/O
// errors still replace the connection.
func NewReconnectingNetPacketConnToWg(ctx context.Context, factory PacketConnFactory, noReceiveTimeout time.Duration) conn.Bind {
	return newReconnectingNetPacketConnToWg(ctx, factory, noReceiveTimeout, time.Now)
}

func newReconnectingNetPacketConnToWg(ctx context.Context, factory PacketConnFactory, noReceiveTimeout time.Duration, now func() time.Time) conn.Bind {
	if ctx == nil {
		ctx = context.Background()
	}
	if now == nil {
		now = time.Now
	}
	return &reconnectingNetPacketConnToWg{
		ctx:              ctx,
		factory:          factory,
		noReceiveTimeout: noReceiveTimeout,
		now:              now,
		retryWait:        waitForReconnectRetry,
	}
}

func waitForReconnectRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *reconnectingNetPacketConnToWg) Open(port uint16) ([]conn.ReceiveFunc, uint16, error) {
	b.replaceMu.Lock()
	defer b.replaceMu.Unlock()

	b.mu.Lock()
	if b.open {
		b.mu.Unlock()
		return nil, 0, conn.ErrBindAlreadyOpen
	}
	if b.factory == nil {
		b.mu.Unlock()
		return nil, 0, errNilPacketConnFactory
	}

	b.lifecycle++
	lifecycle := b.lifecycle
	b.openCtx, b.cancel = context.WithCancel(b.ctx)
	openCtx := b.openCtx
	b.open = true
	b.conn = nil
	b.actualPort = 0
	b.lastReceive = time.Time{}
	b.mu.Unlock()

	packetConn, err := b.factory(openCtx)
	if err == nil && packetConn == nil {
		err = errNilPacketConn
	}
	if err != nil && packetConn != nil {
		_ = packetConn.Close()
		packetConn = nil
	}
	if err == nil && openCtx.Err() != nil {
		_ = packetConn.Close()
		packetConn = nil
		err = openCtx.Err()
	}
	if err != nil {
		b.failOpen(lifecycle)
		return nil, 0, err
	}

	b.mu.Lock()
	if !b.open || b.lifecycle != lifecycle {
		b.mu.Unlock()
		_ = packetConn.Close()
		return nil, 0, net.ErrClosed
	}
	b.installLocked(packetConn)
	actualPort := b.actualPort
	b.mu.Unlock()

	receive := func(packets [][]byte, sizes []int, eps []conn.Endpoint) (int, error) {
		return b.receive(lifecycle, packets, sizes, eps)
	}
	return []conn.ReceiveFunc{receive}, actualPort, nil
}

func (b *reconnectingNetPacketConnToWg) failOpen(lifecycle uint64) {
	b.mu.Lock()
	if b.open && b.lifecycle == lifecycle {
		cancel := b.cancel
		b.open = false
		b.openCtx = nil
		b.cancel = nil
		b.lifecycle++
		b.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return
	}
	b.mu.Unlock()
}

func (b *reconnectingNetPacketConnToWg) installLocked(packetConn net.PacketConn) {
	b.conn = packetConn
	b.generation++
	b.lastReceive = b.now()
	b.actualPort = packetConnPort(packetConn)
}

func packetConnPort(packetConn net.PacketConn) uint16 {
	if packetConn == nil {
		return 0
	}
	if udpAddr, ok := packetConn.LocalAddr().(*net.UDPAddr); ok && udpAddr != nil {
		return uint16(udpAddr.Port)
	}
	return 0
}

func (b *reconnectingNetPacketConnToWg) Close() error {
	b.mu.Lock()
	if !b.open {
		b.mu.Unlock()
		return nil
	}
	b.open = false
	b.lifecycle++
	b.generation++
	packetConn := b.conn
	b.conn = nil
	b.actualPort = 0
	b.lastReceive = time.Time{}
	cancel := b.cancel
	b.cancel = nil
	b.openCtx = nil
	b.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if packetConn != nil {
		return packetConn.Close()
	}
	return nil
}

func (b *reconnectingNetPacketConnToWg) SetMark(mark uint32) error {
	return nil
}

func (b *reconnectingNetPacketConnToWg) Send(bufs [][]byte, ep conn.Endpoint) error {
	packetConn, generation, lifecycle, err := b.connectionForSend()
	if err != nil {
		return err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ep.DstToString())
	if err != nil {
		return err
	}
	for _, packet := range bufs {
		if _, err := packetConn.WriteTo(packet, udpAddr); err != nil {
			// The failed packet is deliberately not replayed. WireGuard handles
			// retransmission, while a synchronous replacement prepares the next
			// send to use a fresh logical underlay.
			_, _, _ = b.replace(lifecycle, generation, false)
			return err
		}
	}
	return nil
}

func (b *reconnectingNetPacketConnToWg) connectionForSend() (net.PacketConn, uint64, uint64, error) {
	b.mu.Lock()
	if !b.open {
		b.mu.Unlock()
		return nil, 0, 0, net.ErrClosed
	}
	lifecycle := b.lifecycle
	packetConn := b.conn
	generation := b.generation
	expired := packetConn != nil && b.expiredLocked()
	b.mu.Unlock()

	if packetConn == nil {
		return b.ensureConnection(lifecycle)
	}
	if expired {
		packetConn, generation, replaceErr := b.replace(lifecycle, generation, true)
		return packetConn, generation, lifecycle, replaceErr
	}
	return packetConn, generation, lifecycle, nil
}

func (b *reconnectingNetPacketConnToWg) expiredLocked() bool {
	return b.noReceiveTimeout > 0 &&
		!b.lastReceive.IsZero() &&
		b.now().Sub(b.lastReceive) >= b.noReceiveTimeout
}

func (b *reconnectingNetPacketConnToWg) ensureConnection(lifecycle uint64) (net.PacketConn, uint64, uint64, error) {
	b.replaceMu.Lock()
	defer b.replaceMu.Unlock()

	b.mu.Lock()
	if !b.open || b.lifecycle != lifecycle {
		b.mu.Unlock()
		return nil, 0, lifecycle, net.ErrClosed
	}
	if b.conn != nil {
		packetConn, generation := b.conn, b.generation
		b.mu.Unlock()
		return packetConn, generation, lifecycle, nil
	}
	openCtx := b.openCtx
	b.mu.Unlock()

	return b.createAndInstall(lifecycle, openCtx)
}

// replace invalidates expectedGeneration when it is still current. When
// onlyIfExpired is true it rechecks the receive timestamp while holding the
// replacement lock, preventing a concurrent successful receive from causing
// an unnecessary rotation.
func (b *reconnectingNetPacketConnToWg) replace(lifecycle, expectedGeneration uint64, onlyIfExpired bool) (net.PacketConn, uint64, error) {
	b.replaceMu.Lock()
	defer b.replaceMu.Unlock()

	b.mu.Lock()
	if !b.open || b.lifecycle != lifecycle {
		b.mu.Unlock()
		return nil, 0, net.ErrClosed
	}
	if b.generation != expectedGeneration {
		packetConn, generation := b.conn, b.generation
		b.mu.Unlock()
		if packetConn == nil {
			return nil, generation, errPacketConnUnavailable
		}
		return packetConn, generation, nil
	}
	if b.conn != nil && onlyIfExpired && !b.expiredLocked() {
		packetConn, generation := b.conn, b.generation
		b.mu.Unlock()
		return packetConn, generation, nil
	}

	packetConn := b.conn
	b.conn = nil
	b.generation++
	b.actualPort = 0
	b.lastReceive = time.Time{}
	openCtx := b.openCtx
	b.mu.Unlock()

	// Logical PacketConns may implement deadlines as no-ops. Physical Close is
	// therefore required to wake the receive function before installing the
	// next generation.
	if packetConn != nil {
		_ = packetConn.Close()
	}
	newConn, generation, _, err := b.createAndInstall(lifecycle, openCtx)
	return newConn, generation, err
}

func (b *reconnectingNetPacketConnToWg) createAndInstall(lifecycle uint64, openCtx context.Context) (net.PacketConn, uint64, uint64, error) {
	packetConn, err := b.factory(openCtx)
	if err == nil && packetConn == nil {
		err = errNilPacketConn
	}
	if err != nil && packetConn != nil {
		_ = packetConn.Close()
		packetConn = nil
	}
	if err == nil && openCtx.Err() != nil {
		_ = packetConn.Close()
		packetConn = nil
		err = openCtx.Err()
	}
	if err != nil {
		b.mu.Lock()
		closed := !b.open || b.lifecycle != lifecycle
		b.mu.Unlock()
		if closed {
			return nil, 0, lifecycle, net.ErrClosed
		}
		return nil, 0, lifecycle, err
	}

	b.mu.Lock()
	if !b.open || b.lifecycle != lifecycle || openCtx != b.openCtx {
		b.mu.Unlock()
		_ = packetConn.Close()
		return nil, 0, lifecycle, net.ErrClosed
	}
	if b.conn != nil {
		// A defensive guard for future callers that might create outside
		// replaceMu. Never overwrite a newer generation.
		current, generation := b.conn, b.generation
		b.mu.Unlock()
		_ = packetConn.Close()
		return current, generation, lifecycle, nil
	}
	b.installLocked(packetConn)
	generation := b.generation
	b.mu.Unlock()
	return packetConn, generation, lifecycle, nil
}

func (b *reconnectingNetPacketConnToWg) receive(lifecycle uint64, packets [][]byte, sizes []int, eps []conn.Endpoint) (int, error) {
	retryAttempt := 0
	for i := 0; i < len(packets); i++ {
		for {
			packetConn, generation, _, err := b.ensureConnection(lifecycle)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return 0, net.ErrClosed
				}
				if err := b.waitForReceiveRetry(lifecycle, retryAttempt); err != nil {
					return 0, err
				}
				retryAttempt++
				continue
			}
			retryAttempt = 0

			n, addr, err := packetConn.ReadFrom(packets[i])
			if err != nil {
				if _, _, replaceErr := b.replace(lifecycle, generation, false); replaceErr != nil {
					if errors.Is(replaceErr, net.ErrClosed) {
						return 0, net.ErrClosed
					}
					if err := b.waitForReceiveRetry(lifecycle, retryAttempt); err != nil {
						return 0, err
					}
					retryAttempt++
					continue
				}
				retryAttempt = 0
				continue
			}

			b.mu.Lock()
			if !b.open || b.lifecycle != lifecycle {
				b.mu.Unlock()
				return 0, net.ErrClosed
			}
			if b.generation != generation || b.conn != packetConn {
				b.mu.Unlock()
				// A packet racing with replacement belongs to a physically closed
				// generation. Discard it and continue on the current connection.
				continue
			}
			b.lastReceive = b.now()
			b.mu.Unlock()

			sizes[i] = n
			eps[i] = endpointFromAddr(addr)
			break
		}
	}
	return len(packets), nil
}

func (b *reconnectingNetPacketConnToWg) waitForReceiveRetry(lifecycle uint64, attempt int) error {
	b.mu.Lock()
	if !b.open || b.lifecycle != lifecycle {
		b.mu.Unlock()
		return net.ErrClosed
	}
	openCtx := b.openCtx
	wait := b.retryWait
	b.mu.Unlock()

	// Start at 100ms and cap at one second. The cap lets a recovered logical
	// route become useful promptly without allowing a failed factory to spin.
	delay := 100 * time.Millisecond
	for i := 0; i < attempt && delay < time.Second; i++ {
		delay *= 2
		if delay > time.Second {
			delay = time.Second
		}
	}
	if wait == nil {
		wait = waitForReconnectRetry
	}
	if err := wait(openCtx, delay); err != nil {
		b.mu.Lock()
		closed := !b.open || b.lifecycle != lifecycle
		b.mu.Unlock()
		if closed {
			return net.ErrClosed
		}
		return err
	}
	return nil
}

func endpointFromAddr(addr net.Addr) conn.Endpoint {
	if udpAddr, ok := addr.(*net.UDPAddr); ok && udpAddr != nil {
		ip, _ := netip.AddrFromSlice(udpAddr.IP)
		return &wgEndpoint{ap: netip.AddrPortFrom(ip, uint16(udpAddr.Port))}
	}
	if addr != nil {
		if addrPort, err := netip.ParseAddrPort(addr.String()); err == nil {
			return &wgEndpoint{ap: addrPort}
		}
	}
	return &wgEndpoint{}
}

func (b *reconnectingNetPacketConnToWg) ParseEndpoint(s string) (conn.Endpoint, error) {
	addrPort, err := netip.ParseAddrPort(s)
	if err != nil {
		return nil, err
	}
	return &wgEndpoint{ap: addrPort}, nil
}

func (b *reconnectingNetPacketConnToWg) BatchSize() int {
	return 1
}
