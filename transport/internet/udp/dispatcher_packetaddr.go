package udp

import (
	"context"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

type PacketAddrDispatcher struct {
	conn      net.PacketConn
	callback  ResponseCallback
	ctx       context.Context
	closeOnce sync.Once
	closeErr  error
}

func (p *PacketAddrDispatcher) Close() error {
	p.closeOnce.Do(func() {
		if receiver := p.ctx.Value(DispatcherConnectionTerminationSignalReceiverMark); receiver != nil {
			_ = receiver.(DispatcherConnectionTerminationSignalReceiver).Close()
		}
		p.closeErr = p.conn.Close()
	})
	return p.closeErr
}

func (p *PacketAddrDispatcher) Dispatch(ctx context.Context, destination net.Destination, payload *buf.Buffer) {
	if destination.Network != net.Network_UDP {
		return
	}

	// Processing of domain address is unsupported as it adds unpredictable overhead, it will be dropped.
	if destination.Address.Family().IsDomain() {
		return
	}

	p.conn.WriteTo(payload.Bytes(), &net.UDPAddr{IP: destination.Address.IP(), Port: int(destination.Port.Value())})
}

func (p *PacketAddrDispatcher) readWorker() {
	defer p.Close()
	for {
		readBuf := buf.New()
		n, addr, err := p.conn.ReadFrom(readBuf.Extend(2048))
		if err != nil {
			readBuf.Release()
			return
		}
		readBuf.Resize(0, int32(n))
		p.callback(p.ctx, &udp.Packet{Payload: readBuf, Source: net.DestinationFromAddr(addr)})
	}
}

type PacketAddrDispatcherCreator struct {
	ctx      context.Context
	isStream bool
}

func NewPacketAddrDispatcherCreator(ctx context.Context) PacketAddrDispatcherCreator {
	return PacketAddrDispatcherCreator{ctx: ctx}
}

func NewStreamPacketAddrDispatcherCreator(ctx context.Context) PacketAddrDispatcherCreator {
	return PacketAddrDispatcherCreator{ctx: ctx, isStream: true}
}

func (pdc *PacketAddrDispatcherCreator) NewPacketAddrDispatcher(
	dispatcher routing.Dispatcher, callback ResponseCallback,
) DispatcherI {
	packetConn, _ := packetaddr.CreatePacketAddrConn(pdc.ctx, dispatcher, pdc.isStream)
	pd := &PacketAddrDispatcher{conn: packetConn, callback: callback, ctx: pdc.ctx}
	go pd.readWorker()
	return pd
}
