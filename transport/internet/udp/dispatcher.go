package udp

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/features/routing"
	"github.com/v2fly/v2ray-core/v4/transport"
)

type ResponseCallback func(ctx context.Context, packet *udp.Packet)

type connEntry struct {
	link   *transport.Link
	ctx    context.Context
	cancel context.CancelFunc
}

type Dispatcher struct {
	sync.RWMutex
	conn       *connEntry
	dispatcher routing.Dispatcher
	callback   ResponseCallback
}

func NewDispatcher(dispatcher routing.Dispatcher, callback ResponseCallback) *Dispatcher {
	return &Dispatcher{
		dispatcher: dispatcher,
		callback:   callback,
	}
}

func (v *Dispatcher) getInboundRay(ctx context.Context, dest net.Destination) *connEntry {
	v.Lock()
	defer v.Unlock()

	if v.conn != nil {
		select {
		case <-v.conn.ctx.Done():
			v.conn = nil
		default:
			return v.conn
		}
	}

	newError("establishing new connection for ", dest).WriteToLog()

	ctx, cancel := context.WithCancel(ctx)
	link, _ := v.dispatcher.Dispatch(ctx, dest)
	entry := &connEntry{
		link:   link,
		ctx:    ctx,
		cancel: cancel,
	}
	v.conn = entry
	go handleInput(ctx, entry, dest, v.callback)
	return entry
}

func (v *Dispatcher) Dispatch(ctx context.Context, destination net.Destination, payload *buf.Buffer) {
	conn := v.getInboundRay(ctx, destination)
	outputStream := conn.link.Writer
	if outputStream != nil {
		if err := outputStream.WriteMultiBuffer(buf.MultiBuffer{payload}); err != nil {
			newError("write udp failed").Base(err).WriteToLog(session.ExportIDToError(ctx))
			conn.cancel()
			return
		}
	}
}

func handleInput(ctx context.Context, conn *connEntry, dest net.Destination, callback ResponseCallback) {
	defer conn.cancel()
	input := conn.link.Reader

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		mb, err := input.ReadMultiBuffer()
		if err != nil {
			buf.ReleaseMulti(mb)
			newError("udp connection closed").Base(err).WriteToLog(session.ExportIDToError(ctx))
			return
		}
		for _, b := range mb {
			packet := udp.Packet{
				Payload: b,
				Source:  dest,
			}
			if b.Endpoint == nil {
				packet.Source = dest
			} else {
				packet.Source = *b.Endpoint
			}
			callback(ctx, &packet)
		}
	}
}

type dispatcherConn struct {
	ctx        context.Context
	dispatcher *Dispatcher
	cache      chan *udp.Packet
	done       *done.Instance
}

func DialDispatcher(ctx context.Context, dispatcher routing.Dispatcher) (net.PacketConn, error) {
	c := &dispatcherConn{
		ctx:   ctx,
		cache: make(chan *udp.Packet, 16),
		done:  done.New(),
	}

	d := NewDispatcher(dispatcher, c.callback)
	c.dispatcher = d
	return c, nil
}

func (c *dispatcherConn) callback(ctx context.Context, packet *udp.Packet) {
	select {
	case <-c.done.Wait():
		packet.Payload.Release()
		return
	case c.cache <- packet:
	default:
		packet.Payload.Release()
		return
	}
}

func (c *dispatcherConn) ReadFrom(p []byte) (int, net.Addr, error) {
	select {
	case <-c.done.Wait():
		return 0, nil, io.EOF
	case packet := <-c.cache:
		n := copy(p, packet.Payload.Bytes())
		return n, &net.UDPAddr{
			IP:   packet.Source.Address.IP(),
			Port: int(packet.Source.Port),
		}, nil
	}
}

func (c *dispatcherConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	buffer := buf.New()
	raw := buffer.Extend(buf.Size)
	n := copy(raw, p)
	buffer.Resize(0, int32(n))

	endpoint := net.DestinationFromAddr(addr)
	buffer.Endpoint = &endpoint
	c.dispatcher.Dispatch(c.ctx, endpoint, buffer)
	return n, nil
}

func (c *dispatcherConn) Close() error {
	return c.done.Close()
}

func (c *dispatcherConn) LocalAddr() net.Addr {
	return &net.UDPAddr{
		IP:   []byte{0, 0, 0, 0},
		Port: 0,
	}
}

func (c *dispatcherConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *dispatcherConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *dispatcherConn) SetWriteDeadline(t time.Time) error {
	return nil
}
