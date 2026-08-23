package udpbridge

import (
	"context"
	"encoding/binary"
	"net"
	"sync"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/app/tun/device"
)

const (
	defaultMTU       = 1500
	minimumMTU       = 576
	maximumMTU       = 9000
	defaultQueueSize = 512
	maximumQueueSize = 4096
)

// Options describes the loopback UDP transport used to exchange raw IP
// packets with an external packet-tunnel provider.
type Options struct {
	ListenAddress string
	ListenPort    uint32
	PeerAddress   string
	PeerPort      uint32
	QueueSize     uint32
}

// UDPBridge is a gVisor link endpoint backed by a pair of loopback UDP ports.
// Each UDP datagram contains exactly one complete IPv4 or IPv6 packet.
type UDPBridge struct {
	*channel.Endpoint

	conn *net.UDPConn
	peer *net.UDPAddr
	mtu  uint32

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

var _ device.Device = (*UDPBridge)(nil)

// New creates and starts a UDP-backed link endpoint.
func New(deviceOptions device.Options, options Options) (*UDPBridge, error) {
	if deviceOptions.PreopenedFDSet {
		return nil, newError("preopened file descriptors are not supported").AtError()
	}

	mtu := deviceOptions.MTU
	if mtu == 0 {
		mtu = defaultMTU
	}
	if mtu < minimumMTU || mtu > maximumMTU {
		return nil, newError("MTU must be between ", minimumMTU, " and ", maximumMTU).AtError()
	}

	queueSize := options.QueueSize
	if queueSize == 0 {
		queueSize = defaultQueueSize
	}
	if queueSize > maximumQueueSize {
		return nil, newError("queue_size must not exceed ", maximumQueueSize).AtError()
	}

	if options.ListenPort == 0 || options.ListenPort > 65535 {
		return nil, newError("listen_port must be between 1 and 65535").AtError()
	}
	if options.PeerPort == 0 || options.PeerPort > 65535 {
		return nil, newError("peer_port must be between 1 and 65535").AtError()
	}
	if options.ListenPort == options.PeerPort {
		return nil, newError("listen_port and peer_port must be different").AtError()
	}

	listenIP, listenIsIPv4, err := parseLoopbackIP(options.ListenAddress)
	if err != nil {
		return nil, newError("invalid listen_address").Base(err).AtError()
	}
	peerIP, peerIsIPv4, err := parseLoopbackIP(options.PeerAddress)
	if err != nil {
		return nil, newError("invalid peer_address").Base(err).AtError()
	}
	if listenIsIPv4 != peerIsIPv4 {
		return nil, newError("listen_address and peer_address must use the same address family").AtError()
	}

	network := "udp6"
	if listenIsIPv4 {
		network = "udp4"
	}
	listenAddress := &net.UDPAddr{IP: listenIP, Port: int(options.ListenPort)}
	peerAddress := &net.UDPAddr{IP: peerIP, Port: int(options.PeerPort)}
	conn, err := net.ListenUDP(network, listenAddress)
	if err != nil {
		return nil, newError("failed to listen on ", listenAddress).Base(err).AtError()
	}

	ctx, cancel := context.WithCancel(context.Background())
	bridge := &UDPBridge{
		Endpoint: channel.New(int(queueSize), mtu, ""),
		conn:     conn,
		peer:     peerAddress,
		mtu:      mtu,
		ctx:      ctx,
		cancel:   cancel,
	}
	bridge.wg.Add(2)
	go bridge.readLoop()
	go bridge.writeLoop()
	return bridge, nil
}

func parseLoopbackIP(value string) (net.IP, bool, error) {
	ip := net.ParseIP(value)
	if ip == nil {
		return nil, false, newError("address must be an IP literal")
	}
	if !ip.IsLoopback() {
		return nil, false, newError("address must be loopback")
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4, true, nil
	}
	return ip.To16(), false, nil
}

func (b *UDPBridge) readLoop() {
	defer b.wg.Done()

	datagram := make([]byte, int(b.mtu)+1)
	for {
		n, source, err := b.conn.ReadFromUDP(datagram)
		if err != nil {
			if b.ctx.Err() != nil {
				return
			}
			newError("failed to read UDP packet").Base(err).AtWarning().WriteToLog()
			continue
		}
		if !sameUDPAddress(source, b.peer) {
			newError("discarding UDP packet from unexpected peer ", source).AtDebug().WriteToLog()
			continue
		}
		if n > int(b.mtu) {
			newError("discarding packet larger than MTU").AtDebug().WriteToLog()
			continue
		}

		protocol, ok := rawIPProtocol(datagram[:n])
		if !ok {
			newError("discarding malformed raw IP packet").AtDebug().WriteToLog()
			continue
		}

		packet := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: buffer.MakeWithData(datagram[:n]),
		})
		b.InjectInbound(protocol, packet)
		packet.DecRef()
	}
}

func (b *UDPBridge) writeLoop() {
	defer b.wg.Done()

	for {
		packet := b.ReadContext(b.ctx)
		if packet == nil {
			return
		}

		slices := packet.AsSlices()
		size := 0
		for _, part := range slices {
			size += len(part)
		}
		if size > int(b.mtu) {
			packet.DecRef()
			newError("discarding outbound packet larger than MTU").AtWarning().WriteToLog()
			continue
		}

		datagram := make([]byte, size)
		offset := 0
		for _, part := range slices {
			offset += copy(datagram[offset:], part)
		}
		packet.DecRef()

		if _, err := b.conn.WriteToUDP(datagram, b.peer); err != nil {
			if b.ctx.Err() != nil {
				return
			}
			newError("failed to write UDP packet").Base(err).AtWarning().WriteToLog()
		}
	}
}

func sameUDPAddress(left, right *net.UDPAddr) bool {
	return left != nil && right != nil && left.Port == right.Port && left.IP.Equal(right.IP)
}

func rawIPProtocol(packet []byte) (tcpip.NetworkProtocolNumber, bool) {
	if len(packet) == 0 {
		return 0, false
	}
	switch packet[0] >> 4 {
	case 4:
		if len(packet) < header.IPv4MinimumSize {
			return 0, false
		}
		headerLength := int(packet[0]&0x0f) * 4
		totalLength := int(binary.BigEndian.Uint16(packet[2:4]))
		if headerLength < header.IPv4MinimumSize || totalLength != len(packet) || totalLength < headerLength {
			return 0, false
		}
		return header.IPv4ProtocolNumber, true
	case 6:
		if len(packet) < header.IPv6MinimumSize {
			return 0, false
		}
		totalLength := header.IPv6MinimumSize + int(binary.BigEndian.Uint16(packet[4:6]))
		if totalLength != len(packet) {
			return 0, false
		}
		return header.IPv6ProtocolNumber, true
	default:
		return 0, false
	}
}

// Close stops both packet pumps and releases the UDP port. It is safe to call
// more than once.
func (b *UDPBridge) Close() {
	b.closeOnce.Do(func() {
		b.cancel()
		_ = b.conn.Close()
		b.Endpoint.Close()
		b.wg.Wait()
	})
}
