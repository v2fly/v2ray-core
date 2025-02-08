package hysteria2

import (
	"context"
	"io"
	"time"

	hyProtocol "github.com/v2fly/hysteria/core/v2/international/protocol"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	hyTransport "github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}

// Server is an inbound connection handler that handles messages in protocol.
type Server struct {
	policyManager  policy.Manager
	packetEncoding packetaddr.PacketAddrType
}

// NewServer creates a new inbound handler.
func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {
	v := core.MustFromContext(ctx)
	server := &Server{
		policyManager:  v.GetFeature(policy.ManagerType()).(policy.Manager),
		packetEncoding: config.PacketEncoding,
	}
	return server, nil
}

// Network implements proxy.Inbound.Network().
func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

// Process implements proxy.Inbound.Process().
func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	sid := session.ExportIDToError(ctx)

	iConn := conn
	if statConn, ok := conn.(*internet.StatCouterConnection); ok {
		iConn = statConn.Connection // will not count the UDP traffic.
	}
	hyConn, IsHy2Transport := iConn.(*hyTransport.HyConn)

	if IsHy2Transport && hyConn.IsUDPExtension {
		network = net.Network_UDP
	}

	if !IsHy2Transport && network == net.Network_UDP {
		return newError(hyTransport.CanNotUseUDPExtension)
	}

	sessionPolicy := s.policyManager.ForLevel(0)
	if err := conn.SetReadDeadline(time.Now().Add(sessionPolicy.Timeouts.Handshake)); err != nil {
		return newError("unable to set read deadline").Base(err).AtWarning()
	}

	bufferedReader := &buf.BufferedReader{
		Reader: buf.NewReader(conn),
	}
	clientReader := &ConnReader{Reader: bufferedReader}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return newError("unable to set read deadline").Base(err).AtWarning()
	}

	if network == net.Network_UDP { // handle udp request
		return s.handleUDPPayload(ctx,
			&PacketReader{Reader: clientReader, HyConn: hyConn},
			&PacketWriter{Writer: conn, HyConn: hyConn}, dispatcher)
	}

	var reqAddr string
	var err error
	reqAddr, err = hyProtocol.ReadTCPRequest(conn)
	if err != nil {
		return newError("failed to parse header").Base(err)
	}
	err = hyProtocol.WriteTCPResponse(conn, true, "")
	if err != nil {
		return newError("failed to send response").Base(err)
	}

	address, stringPort, err := net.SplitHostPort(reqAddr)
	if err != nil {
		return err
	}
	port, err := net.PortFromString(stringPort)
	if err != nil {
		return err
	}
	destination := net.Destination{Network: network, Address: net.ParseAddress(address), Port: port}

	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		panic("no inbound metadata")
	}
	sessionPolicy = s.policyManager.ForLevel(0)

	ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   conn.RemoteAddr(),
		To:     destination,
		Status: log.AccessAccepted,
		Reason: "",
	})

	newError("received request for ", destination).WriteToLog(sid)
	return s.handleConnection(ctx, sessionPolicy, destination, clientReader, buf.NewWriter(conn), dispatcher)
}

func (s *Server) handleConnection(ctx context.Context, sessionPolicy policy.Session,
	destination net.Destination,
	clientReader buf.Reader,
	clientWriter buf.Writer, dispatcher routing.Dispatcher,
) error {
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)

	link, err := dispatcher.Dispatch(ctx, destination)
	if err != nil {
		return newError("failed to dispatch request to ", destination).Base(err)
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(clientReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transfer request").Base(err)
		}
		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		if err := buf.Copy(link.Reader, clientWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to write response").Base(err)
		}
		return nil
	}

	requestDonePost := task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDonePost, responseDone); err != nil {
		common.Must(common.Interrupt(link.Reader))
		common.Must(common.Interrupt(link.Writer))
		return newError("connection ends").Base(err)
	}

	return nil
}

func (s *Server) handleUDPPayload(ctx context.Context, clientReader *PacketReader, clientWriter *PacketWriter, dispatcher routing.Dispatcher) error {
	udpDispatcherConstructor := udp.NewSplitDispatcher
	switch s.packetEncoding {
	case packetaddr.PacketAddrType_None:
	case packetaddr.PacketAddrType_Packet:
		packetAddrDispatcherFactory := udp.NewPacketAddrDispatcherCreator(ctx)
		udpDispatcherConstructor = packetAddrDispatcherFactory.NewPacketAddrDispatcher
	}

	udpServer := udpDispatcherConstructor(dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
		if err := clientWriter.WriteMultiBufferWithMetadata(buf.MultiBuffer{packet.Payload}, packet.Source); err != nil {
			newError("failed to write response").Base(err).AtWarning().WriteToLog(session.ExportIDToError(ctx))
		}
	})

	inbound := session.InboundFromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			p, err := clientReader.ReadMultiBufferWithMetadata()
			if err != nil {
				if errors.Cause(err) != io.EOF {
					return newError("unexpected EOF").Base(err)
				}
				return nil
			}
			currentPacketCtx := ctx
			currentPacketCtx = log.ContextWithAccessMessage(currentPacketCtx, &log.AccessMessage{
				From:   inbound.Source,
				To:     p.Target,
				Status: log.AccessAccepted,
				Reason: "",
			})
			newError("tunnelling request to ", p.Target).WriteToLog(session.ExportIDToError(ctx))

			for _, b := range p.Buffer {
				udpServer.Dispatch(currentPacketCtx, p.Target, b)
			}
		}
	}
}
