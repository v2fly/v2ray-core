package tunsorter

import (
	"context"
	"io"
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/tun/packetparse"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

type recordingDispatcher struct {
	destination net.Destination
}

func (d *recordingDispatcher) Dispatch(_ context.Context, destination net.Destination) (*transport.Link, error) {
	d.destination = destination
	_, uplinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	downlinkReader, _ := pipe.New(pipe.WithSizeLimit(1024))
	return &transport.Link{Reader: downlinkReader, Writer: uplinkWriter}, nil
}

func (*recordingDispatcher) Start() error      { return nil }
func (*recordingDispatcher) Close() error      { return nil }
func (*recordingDispatcher) Type() interface{} { return routing.DispatcherType() }

func TestPacketEncodingBypassPort(t *testing.T) {
	sorter := NewTunSorter(
		io.Discard,
		nil,
		packetaddr.PacketAddrType_Stream,
		context.Background(),
		[]net.Port{53},
	)

	src := net.UDPDestination(net.ParseAddress("198.18.0.2"), 49152)
	dst := net.UDPDestination(net.ParseAddress("1.1.1.1"), 53)
	packet, err := packetparse.TryConstructUDPPacket(src, dst, []byte("dns query"))
	if err != nil {
		t.Fatalf("failed to construct UDP packet: %v", err)
	}

	handled, err := sorter.OnPacketReceived(packet)
	if err != nil {
		t.Fatalf("OnPacketReceived returned an error: %v", err)
	}
	if handled {
		t.Fatal("DNS packet was handled by packetaddr instead of being bypassed")
	}
}

func TestUnlistedPortStillUsesStreamPacketEncoding(t *testing.T) {
	dispatcher := new(recordingDispatcher)
	sorter := NewTunSorter(io.Discard, dispatcher, packetaddr.PacketAddrType_Stream, context.Background(), []net.Port{53})

	src := net.UDPDestination(net.ParseAddress("198.18.0.2"), 49152)
	dst := net.UDPDestination(net.ParseAddress("1.1.1.1"), 443)
	packet, err := packetparse.TryConstructUDPPacket(src, dst, []byte("quic packet"))
	if err != nil {
		t.Fatalf("failed to construct UDP packet: %v", err)
	}

	handled, err := sorter.OnPacketReceived(packet)
	if err != nil {
		t.Fatalf("OnPacketReceived returned an error: %v", err)
	}
	if !handled {
		t.Fatal("unlisted UDP port bypassed packet encoding")
	}
	if dispatcher.destination.Network != net.Network_TCP || dispatcher.destination.Address.Domain() != "st.packet-addr.v2fly.arpa" || dispatcher.destination.Port != 0 {
		t.Fatalf("packetaddr destination = %v, want tcp:st.packet-addr.v2fly.arpa:0", dispatcher.destination)
	}

	if tracked, found := sorter.trackedConnections.Load(src.String()); found {
		_ = tracked.(*trackedUDPConnection).packetDispatcher.Close()
	}
}
