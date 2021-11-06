package shadowsocks

import (
	"context"
	"io"
	"strconv"
	"time"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/proxyman"
	app_inbound "github.com/v2fly/v2ray-core/v4/app/proxyman/inbound"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/log"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	udp_proto "github.com/v2fly/v2ray-core/v4/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/common/signal"
	"github.com/v2fly/v2ray-core/v4/common/task"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
	"github.com/v2fly/v2ray-core/v4/features/inbound"
	"github.com/v2fly/v2ray-core/v4/features/policy"
	"github.com/v2fly/v2ray-core/v4/features/routing"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/udp"
)

type Server struct {
	config        *ServerConfig
	user          *protocol.MemoryUser
	policyManager policy.Manager
	tag           string
	pluginTag     string

	plugin         SIP003Plugin
	pluginOverride net.Destination
	receiverPort   int
}

func (s *Server) Initialize(self inbound.Handler) {
	s.tag = self.Tag()
}

func (s *Server) Close() error {
	if s.plugin != nil {
		return s.plugin.Close()
	}
	return nil
}

// NewServer create a new Shadowsocks server.
func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {
	if config.GetUser() == nil {
		return nil, newError("user is not specified")
	}

	mUser, err := config.User.ToMemoryUser()
	if err != nil {
		return nil, newError("failed to parse user account").Base(err)
	}

	v := core.MustFromContext(ctx)
	s := &Server{
		config:        config,
		user:          mUser,
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}

	if config.Plugin != "" {
		var plugin SIP003Plugin

		pc := plugins[config.Plugin]
		if pc != nil {
			plugin = pc()
		} else if pluginLoader == nil {
			return nil, newError("plugin loader not registered")
		} else {
			plugin = pluginLoader(config.Plugin)
		}

		port, err := net.GetFreePort()
		if err != nil {
			return nil, newError("failed to get free port for shadowsocks plugin").Base(err)
		}

		s.receiverPort, err = net.GetFreePort()
		if err != nil {
			return nil, newError("failed to get free port for shadowsocks plugin receiver").Base(err)
		}

		u := uuid.New()
		tag := "v2ray.system.shadowsocks-inbound-plugin-receiver." + u.String()
		s.pluginTag = tag

		handler, err := app_inbound.NewAlwaysOnInboundHandlerWithProxy(ctx, tag, &proxyman.ReceiverConfig{
			Listen:    net.NewIPOrDomain(net.LocalHostIP),
			PortRange: net.SinglePortRange(net.Port(s.receiverPort)),
		}, s, true)
		if err != nil {
			return nil, newError("failed to create shadowsocks plugin inbound").Base(err)
		}

		inboundManager := v.GetFeature(inbound.ManagerType()).(inbound.Manager)
		if err := inboundManager.AddHandler(ctx, handler); err != nil {
			return nil, newError("failed to add shadowsocks plugin inbound").Base(err)
		}

		s.pluginOverride = net.Destination{
			Network: net.Network_TCP,
			Address: net.LocalHostIP,
			Port:    net.Port(port),
		}

		if err := plugin.Init(net.LocalHostIP.String(), strconv.Itoa(s.receiverPort), net.LocalHostIP.String(), strconv.Itoa(port), config.PluginOpts, config.PluginArgs, mUser.Account.(*MemoryAccount)); err != nil {
			return nil, newError("failed to start plugin").Base(err)
		}

		s.plugin = plugin
	}

	return s, nil
}

func (s *Server) Network() []net.Network {
	list := s.config.Network
	if len(list) == 0 {
		list = append(list, net.Network_TCP)
	}
	if s.config.UdpEnabled {
		list = append(list, net.Network_UDP)
	}
	return list
}

func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	switch network {
	case net.Network_TCP:
		return s.handleConnection(ctx, conn, dispatcher)
	case net.Network_UDP:
		return s.handlerUDPPayload(ctx, conn, dispatcher)
	default:
		return newError("unknown network: ", network)
	}
}

func (s *Server) handlerUDPPayload(ctx context.Context, conn internet.Connection, dispatcher routing.Dispatcher) error {
	udpServer := udp.NewDispatcher(dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
		request := protocol.RequestHeaderFromContext(ctx)
		if request == nil {
			return
		}

		payload := packet.Payload
		data, err := EncodeUDPPacket(request, payload.Bytes())
		payload.Release()
		if err != nil {
			newError("failed to encode UDP packet").Base(err).AtWarning().WriteToLog(session.ExportIDToError(ctx))
			return
		}
		defer data.Release()

		conn.Write(data.Bytes())
	})

	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		panic("no inbound metadata")
	}
	inbound.User = s.user

	reader := buf.NewPacketReader(conn)
	for {
		mpayload, err := reader.ReadMultiBuffer()
		if err != nil {
			break
		}

		for _, payload := range mpayload {
			request, data, err := DecodeUDPPacket(s.user, payload)
			if err != nil {
				if inbound := session.InboundFromContext(ctx); inbound != nil && inbound.Source.IsValid() {
					newError("dropping invalid UDP packet from: ", inbound.Source).Base(err).WriteToLog(session.ExportIDToError(ctx))
					log.Record(&log.AccessMessage{
						From:   inbound.Source,
						To:     "",
						Status: log.AccessRejected,
						Reason: err,
					})
				}
				payload.Release()
				continue
			}

			currentPacketCtx := ctx
			dest := request.Destination()
			if inbound.Source.IsValid() {
				currentPacketCtx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
					From:   inbound.Source,
					To:     dest,
					Status: log.AccessAccepted,
					Reason: "",
					Email:  request.User.Email,
				})
			}
			newError("tunnelling request to ", dest).WriteToLog(session.ExportIDToError(currentPacketCtx))

			currentPacketCtx = protocol.ContextWithRequestHeader(currentPacketCtx, request)
			udpServer.Dispatch(currentPacketCtx, dest, data)
		}
	}

	return nil
}

func (s *Server) handleConnection(ctx context.Context, conn internet.Connection, dispatcher routing.Dispatcher) error {
	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		panic("no inbound metadata")
	}
	if s.plugin != nil {
		if inbound.Tag != s.pluginTag {
			dest, err := internet.Dial(ctx, s.pluginOverride, nil)
			if err != nil {
				return newError("failed to handle request to shadowsocks SIP003 plugin").Base(err)
			}
			if err := task.Run(ctx, func() error {
				_, err := io.Copy(conn, dest)
				return err
			}, func() error {
				_, err := io.Copy(dest, conn)
				return err
			}); err != nil {
				return newError("connection ends").Base(err)
			}
			return nil
		}
		inbound.Tag = s.tag
	}

	sessionPolicy := s.policyManager.ForLevel(s.user.Level)
	conn.SetReadDeadline(time.Now().Add(sessionPolicy.Timeouts.Handshake))

	bufferedReader := buf.BufferedReader{Reader: buf.NewReader(conn)}
	request, bodyReader, err := ReadTCPSession(s.user, &bufferedReader)
	if err != nil {
		log.Record(&log.AccessMessage{
			From:   conn.RemoteAddr(),
			To:     "",
			Status: log.AccessRejected,
			Reason: err,
		})
		return newError("failed to create request from: ", conn.RemoteAddr()).Base(err)
	}
	conn.SetReadDeadline(time.Time{})

	inbound.User = s.user

	dest := request.Destination()
	ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   conn.RemoteAddr(),
		To:     dest,
		Status: log.AccessAccepted,
		Reason: "",
		Email:  request.User.Email,
	})
	newError("tunnelling request to ", dest).WriteToLog(session.ExportIDToError(ctx))

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)
	link, err := dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return err
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		responseWriter, err := WriteTCPResponse(request, bufferedWriter)
		if err != nil {
			return newError("failed to write response").Base(err)
		}

		{
			payload, err := link.Reader.ReadMultiBuffer()
			if err != nil {
				return err
			}
			if err := responseWriter.WriteMultiBuffer(payload); err != nil {
				return err
			}
		}

		if err := bufferedWriter.SetBuffered(false); err != nil {
			return err
		}

		if err := buf.Copy(link.Reader, responseWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP response").Base(err)
		}

		return nil
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(bodyReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP request").Base(err)
		}

		return nil
	}

	requestDoneAndCloseWriter := task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDoneAndCloseWriter, responseDone); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}
