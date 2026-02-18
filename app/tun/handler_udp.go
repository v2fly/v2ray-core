package tun

import (
	"context"

	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	gvisor_udp "gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"

	tun_net "github.com/v2fly/v2ray-core/v5/app/tun/net"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

type UDPHandler struct {
	ctx           context.Context
	dispatcher    routing.Dispatcher
	policyManager policy.Manager
	config        *Config
}

type udpConn struct {
	*gonet.UDPConn
	id stack.TransportEndpointID
}

func (c *udpConn) ID() *stack.TransportEndpointID {
	return &c.id
}

func SetUDPHandler(ctx context.Context, dispatcher routing.Dispatcher, policyManager policy.Manager, config *Config) StackOption {
	return func(s *stack.Stack) error {
		udpForwarder := gvisor_udp.NewForwarder(s, func(r *gvisor_udp.ForwarderRequest) bool {
			wg := new(waiter.Queue)
			linkedEndpoint, err := r.CreateEndpoint(wg)
			if err != nil {
				newError("failed to create endpoint: ", err).WriteToLog(session.ExportIDToError(ctx))
				return false
			}

			conn := &udpConn{
				UDPConn: gonet.NewUDPConn(wg, linkedEndpoint),
				id:      r.ID(),
			}

			handler := &UDPHandler{
				ctx:           ctx,
				dispatcher:    dispatcher,
				policyManager: policyManager,
				config:        config,
			}
			go handler.Handle(conn)
			return true
		})
		s.SetTransportProtocolHandler(gvisor_udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}

func (h *UDPHandler) Handle(conn tun_net.UDPConn) error {
	defer conn.Close()
	id := conn.ID()
	ctx := session.ContextWithInbound(h.ctx, &session.Inbound{Tag: h.config.Tag})
	content := new(session.Content)
	if h.config.SniffingSettings != nil {
		content.SniffingRequest.Enabled = h.config.SniffingSettings.Enabled
		content.SniffingRequest.OverrideDestinationForProtocol = h.config.SniffingSettings.DestinationOverride
		content.SniffingRequest.MetadataOnly = h.config.SniffingSettings.MetadataOnly
	}
	ctx = session.ContextWithContent(ctx, content)

	udpDispatcherConstructor := udp.NewSplitDispatcher

	dest := net.UDPDestination(tun_net.AddressFromTCPIPAddr(id.LocalAddress), net.Port(id.LocalPort))
	src := net.UDPDestination(tun_net.AddressFromTCPIPAddr(id.RemoteAddress), net.Port(id.RemotePort))

	udpServer := udpDispatcherConstructor(h.dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
		if _, err := conn.WriteTo(packet.Payload.Bytes(), &net.UDPAddr{
			IP:   src.Address.IP(),
			Port: int(src.Port),
		}); err != nil {
			newError("failed to write UDP packet").Base(err).WriteToLog()
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var buffer [2048]byte
			n, _, err := conn.ReadFrom(buffer[:])
			if err != nil {
				return newError("failed to read UDP packet").Base(err)
			}
			currentPacketCtx := ctx

			udpServer.Dispatch(currentPacketCtx, dest, buf.FromBytes(buffer[:n]))
		}
	}
}
