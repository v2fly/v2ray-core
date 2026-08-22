package udpbridge

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/app/tun/device"
)

type deliveredPacket struct {
	protocol tcpip.NetworkProtocolNumber
	payload  []byte
}

type recordingDispatcher struct {
	packets chan deliveredPacket
}

func (d *recordingDispatcher) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, packet *stack.PacketBuffer) {
	packetBuffer := packet.ToBuffer()
	d.packets <- deliveredPacket{
		protocol: protocol,
		payload:  packetBuffer.Flatten(),
	}
}

func (*recordingDispatcher) DeliverLinkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {}

func TestUDPBridgeBidirectionalIPv4(t *testing.T) {
	bridge, peer := newTestBridge(t)
	defer bridge.Close()
	defer peer.Close()

	dispatcher := &recordingDispatcher{packets: make(chan deliveredPacket, 1)}
	bridge.Attach(dispatcher)

	ipv4Packet := validIPv4Packet()
	if _, err := peer.WriteToUDP(ipv4Packet, bridge.conn.LocalAddr().(*net.UDPAddr)); err != nil {
		t.Fatalf("failed to send inbound packet: %v", err)
	}

	select {
	case delivered := <-dispatcher.packets:
		if delivered.protocol != header.IPv4ProtocolNumber {
			t.Fatalf("unexpected protocol: %v", delivered.protocol)
		}
		if !bytes.Equal(delivered.payload, ipv4Packet) {
			t.Fatalf("unexpected inbound payload: %x", delivered.payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for inbound packet")
	}

	packet := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer.MakeWithData(ipv4Packet),
	})
	list := stack.PacketBufferList{}
	list.PushBack(packet)
	written, writeErr := bridge.WritePackets(list)
	packet.DecRef()
	if writeErr != nil {
		t.Fatalf("WritePackets failed: %v", writeErr)
	}
	if written != 1 {
		t.Fatalf("unexpected write count: %d", written)
	}

	if err := peer.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	received := make([]byte, 2048)
	n, source, err := peer.ReadFromUDP(received)
	if err != nil {
		t.Fatalf("failed to receive outbound packet: %v", err)
	}
	if source.Port != bridge.conn.LocalAddr().(*net.UDPAddr).Port {
		t.Fatalf("outbound packet came from port %d, want %d", source.Port, bridge.conn.LocalAddr().(*net.UDPAddr).Port)
	}
	if !bytes.Equal(received[:n], ipv4Packet) {
		t.Fatalf("unexpected outbound payload: %x", received[:n])
	}
}

func TestUDPBridgeFiltersPeerAndMalformedPackets(t *testing.T) {
	bridge, peer := newTestBridge(t)
	defer bridge.Close()
	defer peer.Close()

	dispatcher := &recordingDispatcher{packets: make(chan deliveredPacket, 3)}
	bridge.Attach(dispatcher)
	destination := bridge.conn.LocalAddr().(*net.UDPAddr)

	unexpectedPeer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("failed to create unexpected peer: %v", err)
	}
	defer unexpectedPeer.Close()

	if _, err := unexpectedPeer.WriteToUDP(validIPv4Packet(), destination); err != nil {
		t.Fatalf("failed to send unexpected-peer packet: %v", err)
	}
	if _, err := peer.WriteToUDP([]byte{0x45, 0x00}, destination); err != nil {
		t.Fatalf("failed to send malformed packet: %v", err)
	}

	ipv6Packet := validIPv6Packet()
	if _, err := peer.WriteToUDP(ipv6Packet, destination); err != nil {
		t.Fatalf("failed to send IPv6 packet: %v", err)
	}

	select {
	case delivered := <-dispatcher.packets:
		if delivered.protocol != header.IPv6ProtocolNumber || !bytes.Equal(delivered.payload, ipv6Packet) {
			t.Fatalf("unexpected delivered packet: protocol=%v payload=%x", delivered.protocol, delivered.payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for valid packet")
	}

	select {
	case extra := <-dispatcher.packets:
		t.Fatalf("unexpected extra packet: protocol=%v payload=%x", extra.protocol, extra.payload)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestUDPBridgeValidationAndPortConflict(t *testing.T) {
	deviceOptions := device.Options{MTU: 1500}
	validOptions := Options{
		ListenAddress: "127.0.0.1",
		ListenPort:    19090,
		PeerAddress:   "127.0.0.1",
		PeerPort:      19091,
	}

	tests := []struct {
		name          string
		deviceOptions device.Options
		options       Options
	}{
		{name: "non-loopback", deviceOptions: deviceOptions, options: withListenAddress(validOptions, "192.0.2.1")},
		{name: "hostname", deviceOptions: deviceOptions, options: withListenAddress(validOptions, "localhost")},
		{name: "missing listen port", deviceOptions: deviceOptions, options: withListenPort(validOptions, 0)},
		{name: "same ports", deviceOptions: deviceOptions, options: withPeerPort(validOptions, validOptions.ListenPort)},
		{name: "mixed families", deviceOptions: deviceOptions, options: withPeerAddress(validOptions, "::1")},
		{name: "queue too large", deviceOptions: deviceOptions, options: withQueueSize(validOptions, maximumQueueSize+1)},
		{name: "MTU too small", deviceOptions: device.Options{MTU: minimumMTU - 1}, options: validOptions},
		{name: "preopened fd", deviceOptions: device.Options{MTU: 1500, PreopenedFDSet: true}, options: validOptions},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bridge, err := New(test.deviceOptions, test.options)
			if err == nil {
				bridge.Close()
				t.Fatal("expected validation error")
			}
		})
	}

	occupied, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("failed to reserve port: %v", err)
	}
	defer occupied.Close()
	options := validOptions
	options.ListenPort = uint32(occupied.LocalAddr().(*net.UDPAddr).Port)
	if bridge, err := New(deviceOptions, options); err == nil {
		bridge.Close()
		t.Fatal("expected occupied-port error")
	}
}

func TestUDPBridgeCloseIsIdempotent(t *testing.T) {
	bridge, peer := newTestBridge(t)
	defer peer.Close()

	done := make(chan struct{})
	go func() {
		bridge.Close()
		bridge.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not unblock packet pumps")
	}
}

func TestRawIPProtocol(t *testing.T) {
	tests := []struct {
		name     string
		packet   []byte
		protocol tcpip.NetworkProtocolNumber
		valid    bool
	}{
		{name: "IPv4", packet: validIPv4Packet(), protocol: header.IPv4ProtocolNumber, valid: true},
		{name: "IPv6", packet: validIPv6Packet(), protocol: header.IPv6ProtocolNumber, valid: true},
		{name: "empty"},
		{name: "unknown version", packet: []byte{0x70}},
		{name: "short IPv4", packet: []byte{0x45}},
		{name: "invalid IPv4 length", packet: append(validIPv4Packet(), 0)},
		{name: "short IPv6", packet: []byte{0x60}},
		{name: "invalid IPv6 length", packet: append(validIPv6Packet(), 0)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			protocol, valid := rawIPProtocol(test.packet)
			if valid != test.valid || protocol != test.protocol {
				t.Fatalf("rawIPProtocol() = (%v, %v), want (%v, %v)", protocol, valid, test.protocol, test.valid)
			}
		})
	}
}

func newTestBridge(t *testing.T) (*UDPBridge, *net.UDPConn) {
	t.Helper()

	peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("failed to bind peer: %v", err)
	}
	listenPort := unusedUDPPort(t)
	peerPort := peer.LocalAddr().(*net.UDPAddr).Port
	if listenPort == peerPort {
		listenPort = unusedUDPPort(t)
	}

	bridge, err := New(device.Options{MTU: 1500}, Options{
		ListenAddress: "127.0.0.1",
		ListenPort:    uint32(listenPort),
		PeerAddress:   "127.0.0.1",
		PeerPort:      uint32(peerPort),
		QueueSize:     8,
	})
	if err != nil {
		peer.Close()
		t.Fatalf("failed to create bridge: %v", err)
	}
	return bridge, peer
}

func unusedUDPPort(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("failed to select UDP port: %v", err)
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	if err := conn.Close(); err != nil {
		t.Fatalf("failed to release UDP port: %v", err)
	}
	return port
}

func validIPv4Packet() []byte {
	packet := make([]byte, header.IPv4MinimumSize)
	packet[0] = 0x45
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))
	return packet
}

func validIPv6Packet() []byte {
	packet := make([]byte, header.IPv6MinimumSize)
	packet[0] = 0x60
	binary.BigEndian.PutUint16(packet[4:6], uint16(len(packet)-header.IPv6MinimumSize))
	return packet
}

func withListenAddress(options Options, value string) Options {
	options.ListenAddress = value
	return options
}

func withListenPort(options Options, value uint32) Options {
	options.ListenPort = value
	return options
}

func withPeerAddress(options Options, value string) Options {
	options.PeerAddress = value
	return options
}

func withPeerPort(options Options, value uint32) Options {
	options.PeerPort = value
	return options
}

func withQueueSize(options Options, value uint32) Options {
	options.QueueSize = value
	return options
}
