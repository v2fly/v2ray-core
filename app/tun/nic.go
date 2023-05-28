package tun

import (
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func CreateNIC(nicID tcpip.NICID, linkEndpoint stack.LinkEndpoint) StackOption {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(nicID, linkEndpoint,
			stack.NICOptions{
				Disabled: false,
				QDisc:    nil,
			}); err != nil {
			return newError("failed to create NIC:", err)
		}
		return nil
	}
}

func AddProtocolAddress(id tcpip.NICID, ips []*routercommon.CIDR) StackOption {
	return func(s *stack.Stack) error {
		for _, ip := range ips {
			tcpIpAddr := tcpip.AddrFrom4Slice(ip.Ip)
			protocolAddress := tcpip.ProtocolAddress{
				AddressWithPrefix: tcpip.AddressWithPrefix{
					Address:   tcpIpAddr,
					PrefixLen: int(ip.Prefix),
				},
			}

			switch tcpIpAddr.Len() {
			case 4:
				protocolAddress.Protocol = ipv4.ProtocolNumber
			case 16:
				protocolAddress.Protocol = ipv6.ProtocolNumber
			default:
				return newError("invalid IP address length:", tcpIpAddr.Len())
			}

			if err := s.AddProtocolAddress(id, protocolAddress, stack.AddressProperties{}); err != nil {
				return newError("failed to add protocol address:", err)
			}
		}

		return nil
	}
}
