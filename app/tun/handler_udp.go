package tun

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	gvisor_udp "gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

type UDPHandler struct {
	ctx           context.Context
	dispatcher    routing.Dispatcher
	policyManager policy.Manager
	config        *Config

	stack *stack.Stack
}

func SetUDPHandler(ctx context.Context, dispatcher routing.Dispatcher, policyManager policy.Manager, config *Config) StackOption {
	return func(s *stack.Stack) error {
		udpForwarder := gvisor_udp.NewForwarder(s, func(r *gvisor_udp.ForwarderRequest) {
			wg := new(waiter.Queue)
			linkedEndpoint, err := r.CreateEndpoint(wg)
			if err != nil {
				// TODO: log
				return
			}

			udpConn := gonet.NewUDPConn(s, wg, linkedEndpoint)
			udpHandler := &UDPHandler{
				ctx:           ctx,
				dispatcher:    dispatcher,
				policyManager: policyManager,
				config:        config,
				stack:         s,
			}
			udpHandler.Handle(udpConn)
		})
		s.SetTransportProtocolHandler(gvisor_udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}
func (h *UDPHandler) Handle(conn net.Conn) error {
	ctx := session.ContextWithInbound(h.ctx, &session.Inbound{Tag: h.config.Tag})
	packetConn := conn.(net.PacketConn)

	udpDispatcherConstructor := udp.NewSplitDispatcher
	switch h.config.PacketEncoding {
	case packetaddr.PacketAddrType_None:
		break
	case packetaddr.PacketAddrType_Packet:
		packetAddrDispatcherFactory := udp.NewPacketAddrDispatcherCreator(ctx)
		udpDispatcherConstructor = packetAddrDispatcherFactory.NewPacketAddrDispatcher
	}

	udpServer := udpDispatcherConstructor(h.dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
		if _, err := packetConn.WriteTo(packet.Payload.Bytes(), &net.UDPAddr{
			IP:   packet.Source.Address.IP(),
			Port: int(packet.Source.Port),
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
			n, addr, err := packetConn.ReadFrom(buffer[:])
			if err != nil {
				return newError("failed to read UDP packet").Base(err)
			}
			currentPacketCtx := ctx

			udpServer.Dispatch(currentPacketCtx, net.DestinationFromAddr(addr), buf.FromBytes(buffer[:n]))
		}
	}
}
