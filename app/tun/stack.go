package tun

import (
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type StackOption func(*stack.Stack) error

func (t *TUN) CreateStack(linkedEndpoint stack.LinkEndpoint) (*stack.Stack, error) {
	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
	})

	nicID := s.NextNICID()

	opts := []StackOption{
		SetTCPHandler(t.ctx, t.dispatcher, t.policyManager, t.config),
		SetUDPHandler(t.ctx, t.dispatcher, t.policyManager, t.config),

		CreateNIC(nicID, linkedEndpoint),
		AddProtocolAddress(nicID, t.config.Ips),
		SetRouteTable(nicID, t.config.Routes),
		SetPromiscuousMode(nicID, t.config.EnablePromiscuousMode),
		SetSpoofing(nicID, t.config.EnableSpoofing),
	}

	if t.config.SocketSettings != nil {
		if size := t.config.SocketSettings.TxBufSize; size != 0 {
			opts = append(opts, SetTCPSendBufferSize(int(size)))
		}

		if size := t.config.SocketSettings.RxBufSize; size != 0 {
			opts = append(opts, SetTCPReceiveBufferSize(int(size)))
		}
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}
