package outbound

import (
	"context"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type wireguardUnderlayContextKey struct{}

type wireguardUnderlayMarker struct {
	outbound *WireguardOutbound
	parent   *wireguardUnderlayMarker
}

func markWireguardUnderlay(ctx context.Context, outbound *WireguardOutbound) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	parent, _ := ctx.Value(wireguardUnderlayContextKey{}).(*wireguardUnderlayMarker)
	return context.WithValue(ctx, wireguardUnderlayContextKey{}, &wireguardUnderlayMarker{
		outbound: outbound,
		parent:   parent,
	})
}

func isWireguardUnderlayFor(ctx context.Context, outbound *WireguardOutbound) bool {
	if ctx == nil || outbound == nil {
		return false
	}
	marker, _ := ctx.Value(wireguardUnderlayContextKey{}).(*wireguardUnderlayMarker)
	for marker != nil {
		if marker.outbound == outbound {
			return true
		}
		marker = marker.parent
	}
	return false
}

// preserveWireguardUnderlayMarkers copies only the immutable recursion marker
// chain to a long-lived context. Request cancellation and other request-scoped
// values must not become parents of the shared WireGuard session.
func preserveWireguardUnderlayMarkers(base, source context.Context) context.Context {
	if base == nil {
		base = context.Background()
	}
	if source == nil {
		return base
	}
	marker, _ := source.Value(wireguardUnderlayContextKey{}).(*wireguardUnderlayMarker)
	if marker == nil {
		return base
	}
	return context.WithValue(base, wireguardUnderlayContextKey{}, marker)
}

// dialerDispatcher adapts the outbound handler's internet.Dialer to the
// routing.Dispatcher expected by the logical packet-connection implementations.
// The dialer remains responsible for applying sender proxy settings.
type dialerDispatcher struct {
	outbound *WireguardOutbound
	dialer   internet.Dialer
}

func newDialerDispatcher(outbound *WireguardOutbound, dialer internet.Dialer) (routing.Dispatcher, error) {
	if dialer == nil {
		return nil, newError("internet dialer is not configured for wireguard outbound")
	}
	return &dialerDispatcher{
		outbound: outbound,
		dialer:   dialer,
	}, nil
}

func (d *dialerDispatcher) Dispatch(ctx context.Context, destination cnet.Destination) (*transport.Link, error) {
	if d == nil || d.dialer == nil {
		return nil, newError("internet dialer is not configured for wireguard outbound")
	}

	dialCtx, cancel := context.WithCancel(markWireguardUnderlay(ctx, d.outbound))
	conn, err := d.dialer.Dial(dialCtx, destination)
	if err != nil {
		cancel()
		return nil, newError("failed to dial wireguard underlay to ", destination).Base(err)
	}
	if conn == nil {
		cancel()
		return nil, newError("internet dialer returned a nil wireguard underlay connection")
	}
	return newDialerLink(conn, destination.Network, cancel), nil
}

func (*dialerDispatcher) Start() error { return nil }

func (*dialerDispatcher) Close() error { return nil }

func (*dialerDispatcher) Type() interface{} { return routing.DispatcherType() }

type dialedConnectionOwner struct {
	connection internet.Connection
	cancel     context.CancelFunc
	closeOnce  sync.Once
	closeErr   error
}

func (o *dialedConnectionOwner) Close() error {
	if o == nil {
		return nil
	}
	o.closeOnce.Do(func() {
		if o.cancel != nil {
			o.cancel()
		}
		if o.connection != nil {
			o.closeErr = o.connection.Close()
		}
	})
	return o.closeErr
}

type ownedLinkReader struct {
	buf.Reader
	owner *dialedConnectionOwner
}

func (r *ownedLinkReader) Interrupt() {
	_ = r.owner.Close()
}

func (r *ownedLinkReader) Close() error {
	return r.owner.Close()
}

type ownedLinkWriter struct {
	buf.Writer
	owner *dialedConnectionOwner
}

func (w *ownedLinkWriter) Interrupt() {
	_ = w.owner.Close()
}

func (w *ownedLinkWriter) Close() error {
	return w.owner.Close()
}

// statCounterPacketReader preserves packet boundaries by reading from the
// wrapped packet connection's MultiBuffer/packet reader directly. Calling the
// generic StatCouterConnection.Read method could merge adjacent UDP buffers.
type statCounterPacketReader struct {
	reader     buf.Reader
	connection *internet.StatCouterConnection
}

func (r *statCounterPacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := r.reader.ReadMultiBuffer()
	if r.connection.ReadCounter != nil {
		r.connection.ReadCounter.Add(int64(mb.Len()))
	}
	return mb, err
}

func newPacketReader(conn internet.Connection) buf.Reader {
	if statConn, ok := conn.(*internet.StatCouterConnection); ok && statConn.Connection != nil {
		return &statCounterPacketReader{
			reader:     newPacketReader(statConn.Connection),
			connection: statConn,
		}
	}
	return buf.NewPacketReader(conn)
}

func newDialerLink(conn internet.Connection, network cnet.Network, cancel context.CancelFunc) *transport.Link {
	owner := &dialedConnectionOwner{connection: conn, cancel: cancel}

	var reader buf.Reader
	if network == cnet.Network_UDP {
		reader = newPacketReader(conn)
	} else {
		reader = buf.NewReader(conn)
	}

	return &transport.Link{
		Reader: &ownedLinkReader{
			Reader: reader,
			owner:  owner,
		},
		Writer: &ownedLinkWriter{
			Writer: buf.NewWriter(conn),
			owner:  owner,
		},
	}
}
