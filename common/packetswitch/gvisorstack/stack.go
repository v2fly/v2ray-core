package gvisorstack

import (
	"context"
	"fmt"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type WrappedStack struct {
	config *Config
	ctx    context.Context
	stack  *stack.Stack
}

func NewStack(ctx context.Context, config *Config) (*WrappedStack, error) {
	return &WrappedStack{
		config: config,
		ctx:    ctx,
	}, nil
}

func (w *WrappedStack) CreateStackFromNetworkLayerDevice(packetSwitchDevice packetswitch.NetworkLayerDevice) error {
	// Validate
	if w == nil || w.config == nil {
		return fmt.Errorf("no config")
	}

	// Determine MTU from config (0 means unspecified)
	mtu := int(w.config.GetMtu())

	// Create adaptor that implements stack.LinkEndpoint
	adaptor := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(w.ctx, mtu, packetSwitchDevice)

	// Create stack using adaptor as link endpoint
	s, err := w.createStack(adaptor)
	if err != nil {
		// cleanup adaptor on error
		adaptor.Close()
		return fmt.Errorf("failed to create gvisor stack: %v", err)
	}

	// When the adaptor is closed, close the stack as well.
	adaptor.SetOnCloseAction(func() {
		if s != nil {
			s.Close()
		}
	})

	w.stack = s
	return nil
}

func (w *WrappedStack) createStack(linkedEndpoint stack.LinkEndpoint) (*stack.Stack, error) {
	// Machine Generated
	if w == nil || w.config == nil {
		return nil, fmt.Errorf("no config")
	}

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

	// Create NIC
	if err := s.CreateNICWithOptions(nicID, linkedEndpoint, stack.NICOptions{Disabled: false, QDisc: nil}); err != nil {
		return nil, fmt.Errorf("failed to create NIC: %v", err)
	}

	// Add protocol addresses
	for _, ip := range w.config.Ips {
		tcpIPAddr := tcpip.AddrFromSlice(ip.Ip)
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
			return nil, fmt.Errorf("invalid IP address length: %d", tcpIPAddr.Len())
		}

		if err := s.AddProtocolAddress(nicID, protocolAddress, stack.AddressProperties{}); err != nil {
			return nil, fmt.Errorf("failed to add protocol address: %v", err)
		}
	}

	// Set route table
	s.SetRouteTable(func() (table []tcpip.Route) {
		for _, cidrs := range w.config.Routes {
			subnet := tcpip.AddressWithPrefix{
				Address:   tcpip.AddrFromSlice(cidrs.Ip),
				PrefixLen: int(cidrs.Prefix),
			}.Subnet()
			route := tcpip.Route{
				Destination: subnet,
				NIC:         nicID,
			}
			table = append(table, route)
		}
		return
	}())

	// Promiscuous & spoofing
	if err := s.SetPromiscuousMode(nicID, w.config.EnablePromiscuousMode); err != nil {
		return nil, fmt.Errorf("failed to set promiscuous mode: %v", err)
	}
	if err := s.SetSpoofing(nicID, w.config.EnableSpoofing); err != nil {
		return nil, fmt.Errorf("failed to set spoofing: %v", err)
	}

	// Apply socket buffer sizes if provided
	if w.config.SocketSettings != nil {
		if size := w.config.SocketSettings.TxBufSize; size != 0 {
			sendBufferSizeRangeOption := tcpip.TCPSendBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: int(size), Max: tcp.MaxBufferSize}
			if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &sendBufferSizeRangeOption); err != nil {
				return nil, fmt.Errorf("failed to set tcp send buffer size: %v", err)
			}
		}

		if size := w.config.SocketSettings.RxBufSize; size != 0 {
			receiveBufferSizeRangeOption := tcpip.TCPReceiveBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: int(size), Max: tcp.MaxBufferSize}
			if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &receiveBufferSizeRangeOption); err != nil {
				return nil, fmt.Errorf("failed to set tcp receive buffer size: %v", err)
			}
		}
	}

	return s, nil
}

func (w *WrappedStack) Close() error {
	if w == nil || w.stack == nil {
		return nil
	}
	w.stack.Close()
	w.stack = nil
	return nil
}
