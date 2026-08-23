package tunsorter

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/app/tun/packetparse"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

type recordingDispatcher struct {
	destination    net.Destination
	downlinkWriter *pipe.Writer
	dispatchCount  int
}

func (d *recordingDispatcher) Dispatch(_ context.Context, destination net.Destination) (*transport.Link, error) {
	d.destination = destination
	d.dispatchCount++
	_, uplinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	downlinkReader, downlinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	d.downlinkWriter = downlinkWriter
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

	tracked, found := sorter.trackedConnections.Load(src.String())
	if !found {
		t.Fatal("stream packet encoding did not track the UDP source")
	}
	if err := tracked.(*trackedUDPConnection).packetDispatcher.Close(); err != nil {
		t.Fatalf("failed to close packet dispatcher: %v", err)
	}
	if _, found := sorter.trackedConnections.Load(src.String()); found {
		t.Fatal("explicit dispatcher close did not remove the tracked UDP source")
	}
	if err := tracked.(*trackedUDPConnection).packetDispatcher.Close(); err != nil {
		t.Fatalf("second dispatcher close returned an error: %v", err)
	}
}

func TestStreamPacketEncodingRemovesTrackedConnectionWhenStreamEnds(t *testing.T) {
	dispatcher := new(recordingDispatcher)
	sorter := NewTunSorter(io.Discard, dispatcher, packetaddr.PacketAddrType_Stream, context.Background(), nil)

	src := net.UDPDestination(net.ParseAddress("198.18.0.2"), 49153)
	dst := net.UDPDestination(net.ParseAddress("9.9.9.9"), 443)
	packet, err := packetparse.TryConstructUDPPacket(src, dst, []byte("quic packet"))
	if err != nil {
		t.Fatalf("failed to construct UDP packet: %v", err)
	}

	handled, err := sorter.OnPacketReceived(packet)
	if err != nil {
		t.Fatalf("OnPacketReceived returned an error: %v", err)
	}
	if !handled {
		t.Fatal("UDP packet was not handled by stream packet encoding")
	}
	if _, found := sorter.trackedConnections.Load(src.String()); !found {
		t.Fatal("stream packet encoding did not track the UDP source")
	}
	if dispatcher.downlinkWriter == nil {
		t.Fatal("dispatcher did not create a downlink stream")
	}
	if err := dispatcher.downlinkWriter.Close(); err != nil {
		t.Fatalf("failed to end downlink stream: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, found := sorter.trackedConnections.Load(src.String()); !found {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if _, found := sorter.trackedConnections.Load(src.String()); found {
		t.Fatal("tracked UDP source was not removed after the stream ended")
	}

	handled, err = sorter.OnPacketReceived(packet)
	if err != nil {
		t.Fatalf("OnPacketReceived after cleanup returned an error: %v", err)
	}
	if !handled {
		t.Fatal("UDP packet was not handled after tracked connection cleanup")
	}
	if dispatcher.dispatchCount != 2 {
		t.Fatalf("dispatch count = %d, want 2 after recreating the stream", dispatcher.dispatchCount)
	}
	if tracked, found := sorter.trackedConnections.Load(src.String()); found {
		_ = tracked.(*trackedUDPConnection).packetDispatcher.Close()
	}
}
