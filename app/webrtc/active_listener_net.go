package webrtc

import (
	"net"
	"sync"

	piontransport "github.com/pion/transport/v4"
	"github.com/pion/transport/v4/stdnet"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

type trackedPacketConns struct {
	mu          sync.RWMutex
	packetConns []v2net.PacketConn
}

func (t *trackedPacketConns) track(packetConn net.PacketConn) {
	if packetConn == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, existing := range t.packetConns {
		if existing == packetConn {
			return
		}
	}
	t.packetConns = append(t.packetConns, packetConn)
}

func (t *trackedPacketConns) blast(ip net.IP) error {
	t.mu.RLock()
	packetConns := append([]v2net.PacketConn(nil), t.packetConns...)
	t.mu.RUnlock()
	return blossomUDPPorts(packetConns, ip)
}

type trackingNet struct {
	*stdnet.Net
	tracked trackedPacketConns
}

func newTrackingNet() (*trackingNet, error) {
	base, err := stdnet.NewNet()
	if err != nil {
		return nil, newError("failed to initialize stdnet").Base(err)
	}

	return &trackingNet{Net: base}, nil
}

func (n *trackingNet) ListenPacket(network, address string) (net.PacketConn, error) {
	packetConn, err := n.Net.ListenPacket(network, address)
	if err == nil {
		n.tracked.track(packetConn)
	}
	return packetConn, err
}

func (n *trackingNet) ListenUDP(network string, laddr *net.UDPAddr) (piontransport.UDPConn, error) {
	packetConn, err := n.Net.ListenUDP(network, laddr)
	if err == nil {
		n.tracked.track(packetConn)
	}
	return packetConn, err
}

func (n *trackingNet) DialUDP(network string, laddr, raddr *net.UDPAddr) (piontransport.UDPConn, error) {
	packetConn, err := n.Net.DialUDP(network, laddr, raddr)
	if err == nil {
		n.tracked.track(packetConn)
	}
	return packetConn, err
}

func (n *trackingNet) BlastPorts(ip net.IP) error {
	return n.tracked.blast(ip)
}
