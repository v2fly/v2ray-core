package tun

import (
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

func CreateNIC(id tcpip.NICID, linkEndpoint stack.LinkEndpoint) StackOption {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(id, linkEndpoint,
			stack.NICOptions{
				Disabled: false,
				QDisc:    nil,
			}); err != nil {
			return newError("failed to create NIC:", err)
		}
		return nil
	}
}

func SetPromiscuousMode(id tcpip.NICID, enable bool) StackOption {
	return func(s *stack.Stack) error {
		if err := s.SetPromiscuousMode(id, enable); err != nil {
			return newError("failed to set promiscuous mode:", err)
		}
		return nil
	}
}

func SetSpoofing(id tcpip.NICID, enable bool) StackOption {
	return func(s *stack.Stack) error {
		if err := s.SetSpoofing(id, enable); err != nil {
			return newError("failed to set spoofing:", err)
		}
		return nil
	}
}

func AddProtocolAddress(id tcpip.NICID, ips []*routercommon.CIDR) StackOption {
	return func(s *stack.Stack) error {
		for _, ip := range ips {
			tcpIPAddr := tcpip.AddrFrom4Slice(ip.Ip)
			protocolAddress := tcpip.ProtocolAddress{
				AddressWithPrefix: tcpip.AddressWithPrefix{
					Address:   tcpIPAddr,
					PrefixLen: int(ip.Prefix),
				},
			}

			switch tcpIPAddr.Len() {
			case 4:
				protocolAddress.Protocol = ipv4.ProtocolNumber
			case 16:
				protocolAddress.Protocol = ipv6.ProtocolNumber
			default:
				return newError("invalid IP address length:", tcpIPAddr.Len())
			}

			if err := s.AddProtocolAddress(id, protocolAddress, stack.AddressProperties{}); err != nil {
				return newError("failed to add protocol address:", err)
			}
		}

		return nil
	}
}

func SetRouteTable(id tcpip.NICID, routes []*routercommon.CIDR) StackOption {
	return func(s *stack.Stack) error {
		s.SetRouteTable(func() (table []tcpip.Route) {
			for _, cidrs := range routes {
				subnet := tcpip.AddressWithPrefix{
					Address:   tcpip.AddrFrom4Slice(cidrs.Ip),
					PrefixLen: int(cidrs.Prefix),
				}.Subnet()
				route := tcpip.Route{
					Destination: subnet,
					NIC:         id,
				}
				table = append(table, route)
			}
			return
		}())

		return nil
	}
}

func SetTCPSendBufferSize(size int) StackOption {
	return func(s *stack.Stack) error {
		sendBufferSizeRangeOption := tcpip.TCPSendBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: size, Max: tcp.MaxBufferSize}
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &sendBufferSizeRangeOption); err != nil {
			return newError("failed to set tcp send buffer size:", err)
		}
		return nil
	}
}

func SetTCPReceiveBufferSize(size int) StackOption {
	return func(s *stack.Stack) error {
		receiveBufferSizeRangeOption := tcpip.TCPReceiveBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: size, Max: tcp.MaxBufferSize}
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &receiveBufferSizeRangeOption); err != nil {
			return newError("failed to set tcp receive buffer size:", err)
		}
		return nil
	}
}
