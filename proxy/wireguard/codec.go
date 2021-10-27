package wireguard

import (
	"context"
	"fmt"
	"golang.zx2c4.com/wireguard/tun"
	"net"
	"os"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type netTun struct {
	stack          *stack.Stack
	dispatcher     stack.NetworkDispatcher
	events         chan tun.Event
	incomingPacket chan buffer.VectorisedView
	mtu            int
	hasV4, hasV6   bool
}
type endpoint netTun

// WritePackets writes packets back into io.ReadWriter.
func (e *endpoint) WritePackets(_ stack.RouteInfo, pkts stack.PacketBufferList, _ tcpip.NetworkProtocolNumber) (int, tcpip.Error) {
	n := 0
	for pkt := pkts.Front(); pkt != nil; pkt = pkt.Next() {
		if err := e.WriteRawPacket(pkt); err != nil {
			break
		}
		n++
	}
	return n, nil
}

func (e *endpoint) WriteRawPacket(buffer *stack.PacketBuffer) tcpip.Error {
	data := buffer.Data().ExtractVV()
	_, err := (*netTun)(e).Write(data.ToView(), 0)
	if err != nil {
		return &tcpip.ErrAborted{}
	}
	return nil
}

type Net netTun

func (e *endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.dispatcher = dispatcher
}

func (e *endpoint) IsAttached() bool {
	return e.dispatcher != nil
}

func (e *endpoint) MTU() uint32 {
	mtu, err := (*netTun)(e).MTU()
	if err != nil {
		panic(err)
	}
	return uint32(mtu)
}

func (*endpoint) Capabilities() stack.LinkEndpointCapabilities {
	return stack.CapabilityNone
}

func (*endpoint) MaxHeaderLength() uint16 {
	return 0
}

func (*endpoint) LinkAddress() tcpip.LinkAddress {
	return ""
}

func (*endpoint) Wait() {}

func (e *endpoint) WritePacket(_ stack.RouteInfo, _ tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) tcpip.Error {
	e.incomingPacket <- buffer.NewVectorisedView(pkt.Size(), pkt.Views())
	return nil
}

func (*endpoint) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareNone
}

func (e *endpoint) AddHeader(tcpip.LinkAddress, tcpip.LinkAddress, tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {
}

func CreateNetTUN(localAddresses []net.IP, mtu int) (tun.Device, *Net, error) {
	opts := stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol, ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol, udp.NewProtocol},
		HandleLocal:        true,
	}
	dev := &netTun{
		stack:          stack.New(opts),
		events:         make(chan tun.Event, 10),
		incomingPacket: make(chan buffer.VectorisedView),
		mtu:            mtu,
	}
	tcpipErr := dev.stack.CreateNIC(1, (*endpoint)(dev))
	if tcpipErr != nil {
		return nil, nil, fmt.Errorf("CreateNIC: %v", tcpipErr)
	}

	for _, ip := range localAddresses {
		if ip4 := ip.To4(); ip4 != nil {
			protoAddr := tcpip.ProtocolAddress{
				Protocol:          ipv4.ProtocolNumber,
				AddressWithPrefix: tcpip.Address(ip4).WithPrefix(),
			}
			tcpipErr := dev.stack.AddProtocolAddress(1, protoAddr, stack.AddressProperties{})
			if tcpipErr != nil {
				return nil, nil, fmt.Errorf("AddProtocolAddress(%v): %v", ip4, tcpipErr)
			}
			dev.hasV4 = true
		} else {
			protoAddr := tcpip.ProtocolAddress{
				Protocol:          ipv6.ProtocolNumber,
				AddressWithPrefix: tcpip.Address(ip).WithPrefix(),
			}
			tcpipErr := dev.stack.AddProtocolAddress(1, protoAddr, stack.AddressProperties{})
			if tcpipErr != nil {
				return nil, nil, fmt.Errorf("AddProtocolAddress(%v): %v", ip, tcpipErr)
			}
			dev.hasV6 = true
		}
	}
	if dev.hasV4 {
		dev.stack.AddRoute(tcpip.Route{Destination: header.IPv4EmptySubnet, NIC: 1})
	}
	if dev.hasV6 {
		dev.stack.AddRoute(tcpip.Route{Destination: header.IPv6EmptySubnet, NIC: 1})
	}

	dev.events <- tun.EventUp
	return dev, (*Net)(dev), nil
}

func (tun *netTun) Name() (string, error) {
	return "go", nil
}

func (tun *netTun) File() *os.File {
	return nil
}

func (tun *netTun) Events() chan tun.Event {
	return tun.events
}

func (tun *netTun) Read(buf []byte, offset int) (int, error) {
	view, ok := <-tun.incomingPacket
	if !ok {
		return 0, os.ErrClosed
	}
	return view.Read(buf[offset:])
}

func (tun *netTun) Write(buf []byte, offset int) (int, error) {
	packet := buf[offset:]
	if len(packet) == 0 {
		return 0, nil
	}

	pkb := stack.NewPacketBuffer(stack.PacketBufferOptions{Data: buffer.NewVectorisedView(len(packet), []buffer.View{buffer.NewViewFromBytes(packet)})})
	switch packet[0] >> 4 {
	case 4:
		tun.dispatcher.DeliverNetworkPacket("", "", ipv4.ProtocolNumber, pkb)
	case 6:
		tun.dispatcher.DeliverNetworkPacket("", "", ipv6.ProtocolNumber, pkb)
	}

	return len(buf), nil
}

func (tun *netTun) Flush() error {
	return nil
}

func (tun *netTun) Close() error {
	tun.stack.RemoveNIC(1)

	if tun.events != nil {
		close(tun.events)
	}
	if tun.incomingPacket != nil {
		close(tun.incomingPacket)
	}
	return nil
}

func (tun *netTun) MTU() (int, error) {
	return tun.mtu, nil
}

func convertToFullAddr(ip net.IP, port int) (tcpip.FullAddress, tcpip.NetworkProtocolNumber) {
	if ip4 := ip.To4(); ip4 != nil {
		return tcpip.FullAddress{
			NIC:  1,
			Addr: tcpip.Address(ip4),
			Port: uint16(port),
		}, ipv4.ProtocolNumber
	} else {
		return tcpip.FullAddress{
			NIC:  1,
			Addr: tcpip.Address(ip),
			Port: uint16(port),
		}, ipv6.ProtocolNumber
	}
}

func (net *Net) DialContextTCP(ctx context.Context, addr *net.TCPAddr) (*gonet.TCPConn, error) {
	if addr == nil {
		panic("todo: deal with auto addr semantics for nil addr")
	}
	fa, pn := convertToFullAddr(addr.IP, addr.Port)
	return gonet.DialContextTCP(ctx, net.stack, fa, pn)
}

func (net *Net) DialTCP(addr *net.TCPAddr) (*gonet.TCPConn, error) {
	if addr == nil {
		panic("todo: deal with auto addr semantics for nil addr")
	}
	fa, pn := convertToFullAddr(addr.IP, addr.Port)
	return gonet.DialTCP(net.stack, fa, pn)
}

func (net *Net) ListenTCP(addr *net.TCPAddr) (*gonet.TCPListener, error) {
	if addr == nil {
		panic("todo: deal with auto addr semantics for nil addr")
	}
	fa, pn := convertToFullAddr(addr.IP, addr.Port)
	return gonet.ListenTCP(net.stack, fa, pn)
}

func (net *Net) DialUDP(laddr, raddr *net.UDPAddr) (*gonet.UDPConn, error) {
	var lfa, rfa *tcpip.FullAddress
	var pn tcpip.NetworkProtocolNumber
	if laddr != nil {
		var addr tcpip.FullAddress
		addr, pn = convertToFullAddr(laddr.IP, laddr.Port)
		lfa = &addr
	}
	if raddr != nil {
		var addr tcpip.FullAddress
		addr, pn = convertToFullAddr(raddr.IP, raddr.Port)
		rfa = &addr
	}
	return gonet.DialUDP(net.stack, lfa, rfa, pn)
}
