package webrtc

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"

	piontransport "github.com/pion/transport/v4"
	"github.com/pion/transport/v4/stdnet"
	pionwebrtc "github.com/pion/webrtc/v4"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/session"
	featuredns "github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

var fakeUDPPortCounter atomic.Uint32

type connectViaNet struct {
	base *stdnet.Net

	ctx            context.Context
	dispatcher     routing.Dispatcher
	dnsClient      featuredns.Client
	connectVia     string
	packetEncoding packetaddr.PacketAddrType
	tracked        trackedPacketConns
}

func newConnectViaNet(
	ctx context.Context,
	dispatcher routing.Dispatcher,
	dnsClient featuredns.Client,
	connectVia string,
	packetEncoding packetaddr.PacketAddrType,
) (*connectViaNet, error) {
	if packetEncoding != packetaddr.PacketAddrType_Packet {
		return nil, newError("active listener connect_via currently requires packet_encoding=Packet")
	}

	base, err := stdnet.NewNet()
	if err != nil {
		return nil, newError("failed to initialize stdnet").Base(err)
	}

	n := &connectViaNet{
		base:           base,
		ctx:            ctx,
		dispatcher:     dispatcher,
		dnsClient:      dnsClient,
		connectVia:     connectVia,
		packetEncoding: packetEncoding,
	}
	if fakeUDPPortCounter.Load() == 0 {
		fakeUDPPortCounter.Store(40000)
	}

	return n, nil
}

func (n *connectViaNet) NetworkTypes() []pionwebrtc.NetworkType {
	return []pionwebrtc.NetworkType{
		pionwebrtc.NetworkTypeUDP4,
		pionwebrtc.NetworkTypeUDP6,
	}
}

func (n *connectViaNet) ListenPacket(network string, _ string) (net.PacketConn, error) {
	if !strings.HasPrefix(network, "udp") {
		return nil, piontransport.ErrNotSupported
	}
	return n.ListenUDP(network, nil)
}

func (n *connectViaNet) ListenUDP(network string, _ *net.UDPAddr) (piontransport.UDPConn, error) {
	conn, localAddr, err := n.newPacketConn(network)
	if err != nil {
		return nil, err
	}
	return &connectViaPacketConn{
		PacketConn: conn,
		localAddr:  localAddr,
	}, nil
}

func (n *connectViaNet) ListenTCP(network string, laddr *net.TCPAddr) (piontransport.TCPListener, error) {
	_ = network
	_ = laddr
	return nil, newError("active listener connect_via rejects TCP listeners")
}

func (n *connectViaNet) Dial(network, address string) (net.Conn, error) {
	if strings.HasPrefix(network, "udp") {
		raddr, err := n.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return n.DialUDP(network, nil, raddr)
	}
	return nil, newError("active listener connect_via rejects non-UDP dials for network ", network)
}

func (n *connectViaNet) DialUDP(network string, _ *net.UDPAddr, raddr *net.UDPAddr) (piontransport.UDPConn, error) {
	conn, localAddr, err := n.newPacketConn(network)
	if err != nil {
		return nil, err
	}
	return &connectViaPacketConn{
		PacketConn: conn,
		localAddr:  localAddr,
		remoteAddr: raddr,
	}, nil
}

func (n *connectViaNet) DialTCP(network string, laddr, raddr *net.TCPAddr) (piontransport.TCPConn, error) {
	_ = network
	_ = laddr
	_ = raddr
	return nil, newError("active listener connect_via rejects TCP dials")
}

func (n *connectViaNet) ResolveIPAddr(network, address string) (*net.IPAddr, error) {
	ip, err := n.resolveIP(network, address)
	if err != nil {
		return nil, err
	}
	return &net.IPAddr{IP: ip}, nil
}

func (n *connectViaNet) ResolveUDPAddr(network, address string) (*net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ip, err := n.resolveIP(network, host)
	if err != nil {
		return nil, err
	}
	return net.ResolveUDPAddr(network, net.JoinHostPort(ip.String(), port))
}

func (n *connectViaNet) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) {
	_ = network
	_ = address
	return nil, newError("active listener connect_via rejects TCP resolution")
}

func (n *connectViaNet) Interfaces() ([]*piontransport.Interface, error) {
	iface := piontransport.NewInterface(net.Interface{
		Index: 1,
		MTU:   1500,
		Name:  "v2raywebrtc0",
		Flags: net.FlagUp,
	})
	iface.AddAddress(&net.IPNet{
		IP:   net.IPv4(192, 0, 2, 1),
		Mask: net.CIDRMask(32, 32),
	})
	iface.AddAddress(&net.IPNet{
		IP:   net.ParseIP("2001:db8::1"),
		Mask: net.CIDRMask(128, 128),
	})
	return []*piontransport.Interface{iface}, nil
}

func (n *connectViaNet) InterfaceByIndex(index int) (*piontransport.Interface, error) {
	ifaces, err := n.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Index == index {
			return iface, nil
		}
	}
	return nil, piontransport.ErrInterfaceNotFound
}

func (n *connectViaNet) InterfaceByName(name string) (*piontransport.Interface, error) {
	ifaces, err := n.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Name == name {
			return iface, nil
		}
	}
	return nil, piontransport.ErrInterfaceNotFound
}

func (n *connectViaNet) CreateDialer(d *net.Dialer) piontransport.Dialer {
	return n.base.CreateDialer(d)
}

func (n *connectViaNet) CreateListenConfig(c *net.ListenConfig) piontransport.ListenConfig {
	return n.base.CreateListenConfig(c)
}

func (n *connectViaNet) newPacketConn(network string) (v2net.PacketConn, *net.UDPAddr, error) {
	ctx := n.ctx
	ctx = session.SetForcedOutboundTagToContext(ctx, n.connectVia)

	conn, err := packetaddr.CreatePacketAddrConn(ctx, n.dispatcher, false)
	if err != nil {
		return nil, nil, newError("failed to create packetaddr connection for active listener").Base(err)
	}
	n.tracked.track(conn)
	return conn, fakeLocalUDPAddr(network), nil
}

func (n *connectViaNet) BlastPorts(ip net.IP) error {
	return n.tracked.blast(ip)
}

func (n *connectViaNet) resolveIP(network, address string) (net.IP, error) {
	if ip := net.ParseIP(address); ip != nil {
		return ip, nil
	}
	if n.dnsClient == nil {
		return nil, newError("active listener connect_via requires built-in dns client to resolve ", address)
	}

	ips, err := featuredns.LookupIPWithOption(n.dnsClient, address, featuredns.IPOption{
		IPv4Enable: wantsIPv4(network),
		IPv6Enable: wantsIPv6(network),
		FakeEnable: false,
	})
	if err != nil {
		return nil, newError("failed to resolve ", address, " via built-in dns").Base(err)
	}
	if len(ips) == 0 {
		return nil, newError("empty built-in dns response for ", address)
	}

	return ips[0], nil
}

func wantsIPv4(network string) bool {
	switch {
	case strings.HasSuffix(network, "4"):
		return true
	case strings.HasSuffix(network, "6"):
		return false
	default:
		return true
	}
}

func wantsIPv6(network string) bool {
	switch {
	case strings.HasSuffix(network, "6"):
		return true
	case strings.HasSuffix(network, "4"):
		return false
	default:
		return true
	}
}

func fakeLocalUDPAddr(network string) *net.UDPAddr {
	port := int(fakeUDPPortCounter.Add(1))
	if strings.HasSuffix(network, "6") {
		return &net.UDPAddr{
			IP:   net.ParseIP("2001:db8::1"),
			Port: port,
		}
	}
	return &net.UDPAddr{
		IP:   net.IPv4(192, 0, 2, 1),
		Port: port,
	}
}

type connectViaPacketConn struct {
	net.PacketConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
}

func (c *connectViaPacketConn) LocalAddr() net.Addr {
	if c.localAddr != nil {
		return c.localAddr
	}
	return c.PacketConn.LocalAddr()
}

func (c *connectViaPacketConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *connectViaPacketConn) SetReadBuffer(bytes int) error {
	if setter, ok := c.PacketConn.(interface{ SetReadBuffer(int) error }); ok {
		return setter.SetReadBuffer(bytes)
	}
	return nil
}

func (c *connectViaPacketConn) SetWriteBuffer(bytes int) error {
	if setter, ok := c.PacketConn.(interface{ SetWriteBuffer(int) error }); ok {
		return setter.SetWriteBuffer(bytes)
	}
	return nil
}

func (c *connectViaPacketConn) Read(b []byte) (int, error) {
	n, _, err := c.ReadFrom(b)
	return n, err
}

func (c *connectViaPacketConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, addr, err := c.ReadFrom(b)
	if addr == nil {
		return n, nil, err
	}
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return n, nil, piontransport.ErrNotUDPAddress
	}
	return n, udpAddr, err
}

func (c *connectViaPacketConn) ReadMsgUDP(b, _ []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	n, addr, err = c.ReadFromUDP(b)
	return n, 0, 0, addr, err
}

func (c *connectViaPacketConn) Write(b []byte) (int, error) {
	if c.remoteAddr == nil {
		return 0, errors.New("missing remote UDP address")
	}
	return c.WriteTo(b, c.remoteAddr)
}

func (c *connectViaPacketConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return c.WriteTo(b, addr)
}

func (c *connectViaPacketConn) WriteMsgUDP(b, _ []byte, addr *net.UDPAddr) (n, oobn int, err error) {
	n, err = c.WriteToUDP(b, addr)
	return n, 0, err
}
