package socks

import (
	"context"
	"io"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

// Server is a SOCKS 5 proxy server
type Server struct {
	config        *ServerConfig
	policyManager policy.Manager
}

// NewServer creates a new Server object.
func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {
	v := core.MustFromContext(ctx)
	s := &Server{
		config:        config,
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}
	return s, nil
}

func (s *Server) policy() policy.Session {
	config := s.config
	p := s.policyManager.ForLevel(config.UserLevel)
	if config.Timeout > 0 {
		features.PrintDeprecatedFeatureWarning("Socks timeout")
	}
	if config.Timeout > 0 && config.UserLevel == 0 {
		p.Timeouts.ConnectionIdle = time.Duration(config.Timeout) * time.Second
	}
	return p
}

// Network implements proxy.Inbound.
func (s *Server) Network() []net.Network {
	list := []net.Network{net.Network_TCP}
	if s.config.UdpEnabled {
		list = append(list, net.Network_UDP)
	}
	return list
}

// Process implements proxy.Inbound.
func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	if inbound := session.InboundFromContext(ctx); inbound != nil {
		inbound.User = &protocol.MemoryUser{
			Level: s.config.UserLevel,
		}
	}

	switch network {
	case net.Network_TCP:
		return s.processTCP(ctx, conn, dispatcher)
	case net.Network_UDP:
		return s.handleUDPPayload(ctx, conn, dispatcher)
	default:
		return newError("unknown network: ", network)
	}
}

func (s *Server) processTCP(ctx context.Context, conn internet.Connection, dispatcher routing.Dispatcher) error {
	plcy := s.policy()
	if err := conn.SetReadDeadline(time.Now().Add(plcy.Timeouts.Handshake)); err != nil {
		newError("failed to set deadline").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}

	inbound := session.InboundFromContext(ctx)
	if inbound == nil || !inbound.Gateway.IsValid() {
		return newError("inbound gateway not specified")
	}

	svrSession := &ServerSession{
		config:        s.config,
		address:       inbound.Gateway.Address,
		port:          inbound.Gateway.Port,
		clientAddress: inbound.Source.Address,
	}

	reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
	request, err := svrSession.Handshake(reader, conn)
	if err != nil {
		if inbound != nil && inbound.Source.IsValid() {
			log.Record(&log.AccessMessage{
				From:   inbound.Source,
				To:     "",
				Status: log.AccessRejected,
				Reason: err,
			})
		}
		return newError("failed to read request").Base(err)
	}
	if request.User != nil {
		inbound.User.Email = request.User.Email
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		newError("failed to clear deadline").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}

	if request.Command == protocol.RequestCommandTCP {
		dest := request.Destination()
		newError("TCP Connect request to ", dest).WriteToLog(session.ExportIDToError(ctx))
		if inbound != nil && inbound.Source.IsValid() {
			ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
				From:   inbound.Source,
				To:     dest,
				Status: log.AccessAccepted,
				Reason: "",
			})
		}

		dispatcher = &handshakeFinalizingDispatcher{Feature: dispatcher, delegate: dispatcher, serverSession: svrSession}
		return s.transport(ctx, reader, conn, dest, dispatcher)
	}

	err = svrSession.flushLastReply(true)
	if err != nil {
		return err
	}
	if request.Command == protocol.RequestCommandUDP {
		return s.handleUDP(conn)
	}

	return nil
}

// Wrapper to send final SOCKS reply only after Dispatch() returns
type handshakeFinalizingDispatcher struct {
	features.Feature // do not inherit Dispatch() in case its signature changes
	delegate         routing.Dispatcher
	serverSession    *ServerSession
}

func (d *handshakeFinalizingDispatcher) Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error) {
	link, err := d.delegate.Dispatch(ctx, dest)
	if err == nil {
		closeInDefer := true
		defer func() {
			if closeInDefer {
				common.Interrupt(link.Reader)
				common.Interrupt(link.Writer)
			}
		}()
		err = d.serverSession.flushLastReply(true)
		if err != nil {
			return nil, err
		}
		closeInDefer = false
	} else {
		d.serverSession.flushLastReply(false)
	}
	return link, err
}

func (*Server) handleUDP(c io.Reader) error {
	// The TCP connection closes after this method returns. We need to wait until
	// the client closes it.
	return common.Error2(io.Copy(buf.DiscardBytes, c))
}

func (s *Server) transport(ctx context.Context, reader io.Reader, writer io.Writer, dest net.Destination, dispatcher routing.Dispatcher) error {
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, s.policy().Timeouts.ConnectionIdle)

	plcy := s.policy()
	ctx = policy.ContextWithBufferPolicy(ctx, plcy.Buffer)
	link, err := dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return err
	}

	requestDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.DownlinkOnly)
		if err := buf.Copy(buf.NewReader(reader), link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP request").Base(err)
		}

		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.UplinkOnly)

		v2writer := buf.NewWriter(writer)
		if err := buf.Copy(link.Reader, v2writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP response").Base(err)
		}

		return nil
	}

	requestDonePost := task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDonePost, responseDone); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}

	return nil
}

func (s *Server) handleUDPPayload(ctx context.Context, conn internet.Connection, dispatcher routing.Dispatcher) error {
	udpDispatcherConstructor := udp.NewSplitDispatcher
	switch s.config.PacketEncoding {
	case packetaddr.PacketAddrType_None:
		break
	case packetaddr.PacketAddrType_Packet:
		packetAddrDispatcherFactory := udp.NewPacketAddrDispatcherCreator(ctx)
		udpDispatcherConstructor = packetAddrDispatcherFactory.NewPacketAddrDispatcher
	}
	udpServer := udpDispatcherConstructor(dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
		payload := packet.Payload
		newError("writing back UDP response with ", payload.Len(), " bytes").AtDebug().WriteToLog(session.ExportIDToError(ctx))

		request := protocol.RequestHeaderFromContext(ctx)
		var packetSource net.Destination
		if request == nil {
			packetSource = packet.Source
		} else {
			packetSource = net.UDPDestination(request.Address, request.Port)
		}
		udpMessage, err := EncodeUDPPacketFromAddress(packetSource, payload.Bytes())
		payload.Release()

		defer udpMessage.Release()
		if err != nil {
			newError("failed to write UDP response").AtWarning().Base(err).WriteToLog(session.ExportIDToError(ctx))
		}

		conn.Write(udpMessage.Bytes())
	})

	if inbound := session.InboundFromContext(ctx); inbound != nil && inbound.Source.IsValid() {
		newError("client UDP connection from ", inbound.Source).WriteToLog(session.ExportIDToError(ctx))
	}

	reader := buf.NewPacketReader(conn)
	for {
		mpayload, err := reader.ReadMultiBuffer()
		if err != nil {
			return err
		}

		for _, payload := range mpayload {
			request, err := DecodeUDPPacket(payload)
			if err != nil {
				newError("failed to parse UDP request").Base(err).WriteToLog(session.ExportIDToError(ctx))
				payload.Release()
				continue
			}

			if payload.IsEmpty() {
				payload.Release()
				continue
			}
			currentPacketCtx := ctx
			newError("send packet to ", request.Destination(), " with ", payload.Len(), " bytes").AtDebug().WriteToLog(session.ExportIDToError(ctx))
			if inbound := session.InboundFromContext(ctx); inbound != nil && inbound.Source.IsValid() {
				currentPacketCtx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
					From:   inbound.Source,
					To:     request.Destination(),
					Status: log.AccessAccepted,
					Reason: "",
				})
			}

			currentPacketCtx = protocol.ContextWithRequestHeader(currentPacketCtx, request)
			udpServer.Dispatch(currentPacketCtx, request.Destination(), payload)
		}
	}
}

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}
