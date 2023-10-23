package tunsorter

import (
	"context"
	"io"
	"sync"

	"github.com/v2fly/v2ray-core/v5/app/tun/packetparse"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	vudp "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewTunSorter(tunWriter io.Writer, dispatcher routing.Dispatcher, packetAddrType packetaddr.PacketAddrType, ctx context.Context) *TunSorter {
	return &TunSorter{
		tunWriter:      tunWriter,
		dispatcher:     dispatcher,
		packetAddrType: packetAddrType,
		ctx:            ctx,
	}
}

type TunSorter struct {
	tunWriter      io.Writer
	dispatcher     routing.Dispatcher
	packetAddrType packetaddr.PacketAddrType

	trackedConnections sync.Map
	ctx                context.Context
}

func (t *TunSorter) OnPacketReceived(b []byte) (n int, err error) {
	src, dst, data, err := packetparse.TryParseAsUDPPacket(b)
	if err != nil {
		return 0, err
	}
	conn := newTrackedUDPConnection(src, t)
	trackedConnection, loaded := t.trackedConnections.LoadOrStore(src.String(), conn)
	conn = trackedConnection.(*trackedUDPConnection)
	if !loaded {
		t.onNewConnection(conn)
	}
	conn.onNewPacket(dst, data)
	return len(b), nil
}

func (t *TunSorter) onNewConnection(connection *trackedUDPConnection) {
	udpDispatcherConstructor := udp.NewSplitDispatcher
	switch t.packetAddrType { // nolint: gocritic
	case packetaddr.PacketAddrType_Packet:
		ctx := context.WithValue(t.ctx, udp.DispatcherConnectionTerminationSignalReceiverMark, connection) // nolint:staticcheck
		packetAddrDispatcherFactory := udp.NewPacketAddrDispatcherCreator(ctx)
		udpDispatcherConstructor = packetAddrDispatcherFactory.NewPacketAddrDispatcher
	}
	udpDispatcher := udpDispatcherConstructor(t.dispatcher, func(ctx context.Context, packet *vudp.Packet) {
		connection.onWritePacket(packet.Source, packet.Payload.Bytes())
	})
	connection.packetDispatcher = udpDispatcher
}

func (t *TunSorter) onWritePacket(src net.Destination, dest net.Destination, data []byte) {
	data, err := packetparse.TryConstructUDPPacket(src, dest, data)
	if err != nil {
		newError("failed to construct udp packet").Base(err).WriteToLog()
		return
	}
	_, err = t.tunWriter.Write(data)
	if err != nil {
		newError("failed to write udp packet to tun").Base(err).WriteToLog()
		return
	}
}

func newTrackedUDPConnection(src net.Destination, tunSorter *TunSorter) *trackedUDPConnection {
	return &trackedUDPConnection{
		tunSorter: tunSorter,
		src:       src,
	}
}

type trackedUDPConnection struct {
	packetDispatcher udp.DispatcherI
	tunSorter        *TunSorter
	src              net.Destination
}

func (t *trackedUDPConnection) onNewPacket(dst net.Destination, data []byte) {
	t.packetDispatcher.Dispatch(context.Background(), dst, buf.FromBytes(data))
}

func (t *trackedUDPConnection) onWritePacket(src net.Destination, data []byte) {
	t.tunSorter.onWritePacket(src, t.src, data)
}

func (t *trackedUDPConnection) Close() error {
	t.tunSorter.trackedConnections.Delete(t.src.String())
	return nil
}
