package gvisorstack

import (
	"context"
	"fmt"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

// DialTCP will create a connection to the given destination, using the stack.
// Machine Generated
func (w *WrappedStack) DialTCP(ctx context.Context, remoteAddress net.Destination) (net.Conn, error) {
	if w == nil || w.stack == nil {
		return nil, fmt.Errorf("gvisor stack not initialized")
	}

	if remoteAddress.Network != net.Network_TCP {
		return nil, fmt.Errorf("destination is not tcp: %v", remoteAddress.Network)
	}

	// Resolve address to IP if necessary.
	var ipBytes []byte
	switch remoteAddress.Address.Family() {
	case net.AddressFamilyIPv4:
		ipBytes = remoteAddress.Address.IP().To4()
	case net.AddressFamilyIPv6:
		ipBytes = remoteAddress.Address.IP().To16()
	case net.AddressFamilyDomain:
		// Do not resolve domain names here. Return explicit error.
		return nil, fmt.Errorf("domain address not supported for gVisor dial: %s", remoteAddress.Address.String())
	default:
		return nil, fmt.Errorf("unsupported address family: %v", remoteAddress.Address.Family())
	}

	if ipBytes == nil {
		return nil, fmt.Errorf("failed to obtain IP bytes for %v", remoteAddress.Address)
	}

	// Choose network protocol number based on IP length.
	netProto := ipv4.ProtocolNumber
	if len(ipBytes) == 16 {
		netProto = ipv6.ProtocolNumber
	}

	remote := tcpip.FullAddress{
		Addr: tcpip.AddrFromSlice(ipBytes),
		Port: uint16(remoteAddress.Port),
	}

	// Use gonet dialer to create a TCP connection on the in-memory stack.
	conn, err := gonet.DialContextTCP(ctx, w.stack, remote, netProto)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ListenUDP will create a connection to the given destination, using the stack.
// Machine Generated
func (w *WrappedStack) ListenUDP(ctx context.Context, localAddress net.Destination) (net.PacketConn, error) {
	// allow ctx to be accepted by the function signature
	_ = ctx

	if w == nil || w.stack == nil {
		return nil, fmt.Errorf("gvisor stack not initialized")
	}

	if localAddress.Network != net.Network_UDP {
		return nil, fmt.Errorf("destination is not udp: %v", localAddress.Network)
	}

	// Determine local address bytes.
	var ipBytes []byte
	specified := false
	if localAddress.Address == nil {
		// If address is nil, treat as unspecified (zero) address.
		specified = false
	} else {
		switch localAddress.Address.Family() {
		case net.AddressFamilyIPv4:
			specified = true
			ipBytes = localAddress.Address.IP().To4()
		case net.AddressFamilyIPv6:
			specified = true
			ipBytes = localAddress.Address.IP().To16()
		case net.AddressFamilyDomain:
			// Listening on a domain name is not supported.
			return nil, fmt.Errorf("listening on domain address not supported: %s", localAddress.Address.String())
		default:
			// If unspecified (zero) address, allow kernel (stack) to choose.
			specified = false
		}
	}

	var laddr *tcpip.FullAddress
	if specified {
		if ipBytes == nil {
			return nil, fmt.Errorf("failed to obtain IP bytes for %v", localAddress.Address)
		}
		netProto := ipv4.ProtocolNumber
		if len(ipBytes) == 16 {
			netProto = ipv6.ProtocolNumber
		}
		l := tcpip.FullAddress{Addr: tcpip.AddrFromSlice(ipBytes), Port: uint16(localAddress.Port)}
		laddr = &l
		// Create UDP endpoint bound to local address.
		udpConn, err := gonet.DialUDP(w.stack, laddr, nil, netProto)
		if err != nil {
			return nil, err
		}
		return udpConn, nil
	}

	// If not specified, let the stack choose the local address (pass nil laddr).
	// Default network selection honors PreferIpv6ForUdp if configured.
	defaultNet := ipv4.ProtocolNumber
	if w.config != nil && w.config.GetPreferIpv6ForUdp() {
		defaultNet = ipv6.ProtocolNumber
	}
	udpConn, err := gonet.DialUDP(w.stack, nil, nil, defaultNet)
	if err != nil {
		return nil, err
	}
	return udpConn, nil
}
