package tun

import (
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
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
