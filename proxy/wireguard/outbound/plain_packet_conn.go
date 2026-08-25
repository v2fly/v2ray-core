package outbound

import (
	"context"
	gonet "net"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
)

// wireguardPlainPacketConn presents the multi-destination PacketConn needed by
// a WireGuard bind while keeping each peer on its own dispatcher link. Keeping
// this adapter local is important: the dispatcher is backed by the outbound
// handler's dialer, so its links retain sender proxy settings and WireGuard's
// underlay-recursion markers without changing the shared UDP dispatcher.
type wireguardPlainPacketConn struct {
	ctx        context.Context
	cancel     context.CancelFunc
	dispatcher routing.Dispatcher

	mu       sync.Mutex
	closed   bool
	closeErr error
	entries  map[cnet.Destination]*wireguardPlainPacketEntry
	pending  map[cnet.Destination]*wireguardPlainPacketDial

	received  chan wireguardPlainPacket
	done      chan struct{}
	closeOnce sync.Once
}

type wireguardPlainPacket struct {
	payload []byte
	source  cnet.Destination
}

type wireguardPlainPacketDial struct {
	done  chan struct{}
	entry *wireguardPlainPacketEntry
	err   error
}

type wireguardPlainPacketEntry struct {
	destination cnet.Destination
	link        *transport.Link
	cancel      context.CancelFunc

	writeMu   sync.Mutex
	closeOnce sync.Once
}

func newWireguardPlainPacketConn(ctx context.Context, dispatcher routing.Dispatcher) (cnet.PacketConn, error) {
	if dispatcher == nil {
		return nil, newError("routing dispatcher is not configured for plain WireGuard UDP")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithCancel(ctx)
	conn := &wireguardPlainPacketConn{
		ctx:        connCtx,
		cancel:     cancel,
		dispatcher: dispatcher,
		entries:    make(map[cnet.Destination]*wireguardPlainPacketEntry),
		pending:    make(map[cnet.Destination]*wireguardPlainPacketDial),
		received:   make(chan wireguardPlainPacket, 64),
		done:       make(chan struct{}),
	}
	go conn.watchContext()
	return conn, nil
}

func (c *wireguardPlainPacketConn) watchContext() {
	<-c.ctx.Done()
	c.shutdown(c.ctx.Err())
}

func (c *wireguardPlainPacketConn) getOrCreateEntry(destination cnet.Destination) (*wireguardPlainPacketEntry, error) {
	c.mu.Lock()
	if c.closed {
		err := c.closedErrorLocked()
		c.mu.Unlock()
		return nil, err
	}
	if entry := c.entries[destination]; entry != nil {
		c.mu.Unlock()
		return entry, nil
	}
	if dial := c.pending[destination]; dial != nil {
		c.mu.Unlock()
		select {
		case <-dial.done:
			if dial.err != nil {
				return nil, dial.err
			}
			if dial.entry == nil {
				return nil, newError("plain WireGuard UDP dial completed without a link")
			}
			return dial.entry, nil
		case <-c.done:
			return nil, c.closedError()
		}
	}

	dial := &wireguardPlainPacketDial{done: make(chan struct{})}
	c.pending[destination] = dial
	c.mu.Unlock()

	entry, err := c.dialEntry(destination)

	c.mu.Lock()
	delete(c.pending, destination)
	if c.closed {
		if err == nil {
			err = c.closedErrorLocked()
		}
	} else if err == nil {
		c.entries[destination] = entry
	}
	dial.entry = entry
	dial.err = err
	close(dial.done)
	c.mu.Unlock()

	if err != nil {
		if entry != nil {
			entry.close()
		}
		c.shutdown(err)
		return nil, err
	}

	go c.readEntry(entry)
	return entry, nil
}

func (c *wireguardPlainPacketConn) dialEntry(destination cnet.Destination) (*wireguardPlainPacketEntry, error) {
	entryCtx, cancel := context.WithCancel(c.ctx)
	link, err := c.dispatcher.Dispatch(entryCtx, destination)
	if err != nil {
		entry := &wireguardPlainPacketEntry{link: link, cancel: cancel}
		entry.close()
		return nil, newError("failed to dial plain WireGuard UDP to ", destination).Base(err)
	}

	entry := &wireguardPlainPacketEntry{
		destination: destination,
		link:        link,
		cancel:      cancel,
	}
	if link == nil {
		entry.close()
		return nil, newError("dispatcher returned a nil link for plain WireGuard UDP to ", destination)
	}
	if link.Reader == nil || link.Writer == nil {
		entry.close()
		return nil, newError("dispatcher returned an incomplete link for plain WireGuard UDP to ", destination)
	}
	return entry, nil
}

func (c *wireguardPlainPacketConn) readEntry(entry *wireguardPlainPacketEntry) {
	for {
		mb, err := entry.link.Reader.ReadMultiBuffer()
		if err != nil {
			buf.ReleaseMulti(mb)
			c.shutdown(newError("failed to read plain WireGuard UDP from ", entry.destination).Base(err))
			return
		}
		for i, packet := range mb {
			if packet == nil {
				continue
			}
			payload := append([]byte(nil), packet.Bytes()...)
			packet.Release()
			mb[i] = nil

			select {
			case c.received <- wireguardPlainPacket{payload: payload, source: entry.destination}:
			case <-c.done:
				buf.ReleaseMulti(mb[i+1:])
				return
			}
		}
		buf.ReleaseMulti(mb)
	}
}

func (c *wireguardPlainPacketConn) ReadFrom(p []byte) (int, gonet.Addr, error) {
	select {
	case packet := <-c.received:
		select {
		case <-c.done:
			return 0, nil, c.closedError()
		default:
		}
		n := copy(p, packet.payload)
		return n, &gonet.UDPAddr{
			IP:   append(gonet.IP(nil), packet.source.Address.IP()...),
			Port: int(packet.source.Port),
		}, nil
	case <-c.done:
		return 0, nil, c.closedError()
	}
}

func (c *wireguardPlainPacketConn) WriteTo(p []byte, addr gonet.Addr) (int, error) {
	udpAddr, ok := addr.(*gonet.UDPAddr)
	if !ok || udpAddr == nil {
		return 0, newError("plain WireGuard UDP requires a UDP destination")
	}
	if udpAddr.Port < 0 || udpAddr.Port > 65535 {
		return 0, newError("invalid plain WireGuard UDP destination port: ", udpAddr.Port)
	}
	address := cnet.IPAddress(udpAddr.IP)
	if address == nil {
		return 0, newError("invalid plain WireGuard UDP destination address: ", udpAddr.IP)
	}
	destination := cnet.UDPDestination(address, cnet.Port(udpAddr.Port))

	entry, err := c.getOrCreateEntry(destination)
	if err != nil {
		return 0, err
	}

	packet := buf.NewWithSize(int32(len(p)))
	if _, err := packet.Write(p); err != nil {
		packet.Release()
		return 0, newError("failed to buffer plain WireGuard UDP packet").Base(err)
	}

	entry.writeMu.Lock()
	select {
	case <-c.done:
		entry.writeMu.Unlock()
		packet.Release()
		return 0, c.closedError()
	default:
	}
	err = entry.link.Writer.WriteMultiBuffer(buf.MultiBuffer{packet})
	entry.writeMu.Unlock()
	if err != nil {
		err = newError("failed to write plain WireGuard UDP to ", destination).Base(err)
		c.shutdown(err)
		return 0, err
	}
	return len(p), nil
}

func (c *wireguardPlainPacketConn) Close() error {
	c.shutdown(gonet.ErrClosed)
	return nil
}

func (c *wireguardPlainPacketConn) shutdown(err error) {
	if err == nil {
		err = gonet.ErrClosed
	}
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.closeErr = err
		entries := make([]*wireguardPlainPacketEntry, 0, len(c.entries))
		for destination, entry := range c.entries {
			entries = append(entries, entry)
			delete(c.entries, destination)
		}
		close(c.done)
		c.mu.Unlock()

		c.cancel()
		for _, entry := range entries {
			entry.close()
		}
	})
}

func (c *wireguardPlainPacketConn) closedError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closedErrorLocked()
}

func (c *wireguardPlainPacketConn) closedErrorLocked() error {
	if c.closeErr != nil {
		return c.closeErr
	}
	return gonet.ErrClosed
}

func (c *wireguardPlainPacketConn) LocalAddr() gonet.Addr {
	return &gonet.UDPAddr{IP: gonet.IPv4zero}
}

func (*wireguardPlainPacketConn) SetDeadline(time.Time) error { return nil }

func (*wireguardPlainPacketConn) SetReadDeadline(time.Time) error { return nil }

func (*wireguardPlainPacketConn) SetWriteDeadline(time.Time) error { return nil }

func (e *wireguardPlainPacketEntry) close() {
	if e == nil {
		return
	}
	e.closeOnce.Do(func() {
		if e.cancel != nil {
			e.cancel()
		}
		if e.link != nil {
			_ = common.Interrupt(e.link.Reader)
			_ = common.Interrupt(e.link.Writer)
		}
	})
}
