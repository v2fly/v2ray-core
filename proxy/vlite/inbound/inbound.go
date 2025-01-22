package inbound

import (
	"context"
	"io"
	gonet "net"
	"strconv"
	"sync"

	"github.com/mustafaturan/bus"
	"github.com/xiaokangwang/VLite/interfaces"
	"github.com/xiaokangwang/VLite/interfaces/ibus"
	"github.com/xiaokangwang/VLite/transport"
	udpsctpserver "github.com/xiaokangwang/VLite/transport/packetsctp/sctprelay"
	"github.com/xiaokangwang/VLite/transport/packetuni/puniServer"
	"github.com/xiaokangwang/VLite/transport/udp/udpServer"
	"github.com/xiaokangwang/VLite/transport/udp/udpuni/udpunis"
	"github.com/xiaokangwang/VLite/transport/uni/uniserver"
	"github.com/xiaokangwang/VLite/workers/server"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewUDPInboundHandler(ctx context.Context, config *UDPProtocolConfig) (*Handler, error) {
	proxyEnvironment := envctx.EnvironmentFromContext(ctx).(environment.ProxyEnvironment)
	statusInstance, err := createStatusFromConfig(config)
	if err != nil {
		return nil, newError("unable to initialize vlite").Base(err)
	}
	proxyEnvironment.TransientStorage().Put(ctx, "status", statusInstance)
	return &Handler{ctx: ctx}, nil
}

type Handler struct {
	ctx context.Context
}

func (h *Handler) Network() []net.Network {
	list := []net.Network{net.Network_UDP}
	return list
}

type status struct {
	config *UDPProtocolConfig

	password []byte
	msgbus   *bus.Bus

	ctx context.Context

	transport transport.UnderlayTransportListener

	access sync.Mutex
}

func (s *status) RelayStream(conn io.ReadWriteCloser, ctx context.Context) { //nolint:revive
}

func (s *status) Connection(conn gonet.Conn, connctx context.Context) context.Context { //nolint:revive,stylecheck
	S_S2CTraffic := make(chan server.UDPServerTxToClientTraffic, 8)         //nolint:revive,stylecheck
	S_S2CDataTraffic := make(chan server.UDPServerTxToClientDataTraffic, 8) //nolint:revive,stylecheck
	S_C2STraffic := make(chan server.UDPServerRxFromClientTraffic, 8)       //nolint:revive,stylecheck

	S_S2CTraffic2 := make(chan interfaces.TrafficWithChannelTag, 8)     //nolint:revive,stylecheck
	S_S2CDataTraffic2 := make(chan interfaces.TrafficWithChannelTag, 8) //nolint:revive,stylecheck
	S_C2STraffic2 := make(chan interfaces.TrafficWithChannelTag, 8)     //nolint:revive,stylecheck

	go func(ctx context.Context) {
		for {
			select {
			case data := <-S_S2CTraffic:
				S_S2CTraffic2 <- interfaces.TrafficWithChannelTag(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	go func(ctx context.Context) {
		for {
			select {
			case data := <-S_S2CDataTraffic:
				S_S2CDataTraffic2 <- interfaces.TrafficWithChannelTag(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	go func(ctx context.Context) {
		for {
			select {
			case data := <-S_C2STraffic2:
				S_C2STraffic <- server.UDPServerRxFromClientTraffic(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	if !s.config.EnableStabilization || !s.config.EnableRenegotiation {
		relay := udpsctpserver.NewPacketRelayServer(conn, S_S2CTraffic2, S_S2CDataTraffic2, S_C2STraffic2, s, s.password, connctx)
		udpserver := server.UDPServer(connctx, S_S2CTraffic, S_S2CDataTraffic, S_C2STraffic, relay)
		_ = udpserver
	} else {
		relay := puniServer.NewPacketUniServer(S_S2CTraffic2, S_S2CDataTraffic2, S_C2STraffic2, s, s.password, connctx)
		relay.OnAutoCarrier(conn, connctx)
		udpserver := server.UDPServer(connctx, S_S2CTraffic, S_S2CDataTraffic, S_C2STraffic, relay)
		_ = udpserver
	}
	return connctx
}

func createStatusFromConfig(config *UDPProtocolConfig) (*status, error) { //nolint:unparam
	s := &status{ctx: context.Background(), config: config}

	s.password = []byte(config.Password)

	s.msgbus = ibus.NewMessageBus()
	s.ctx = context.WithValue(s.ctx, interfaces.ExtraOptionsMessageBus, s.msgbus) //nolint:revive,staticcheck

	if config.ScramblePacket {
		s.ctx = context.WithValue(s.ctx, interfaces.ExtraOptionsUDPShouldMask, true) //nolint:revive,staticcheck
	}

	if s.config.EnableFec {
		s.ctx = context.WithValue(s.ctx, interfaces.ExtraOptionsUDPFECEnabled, true) //nolint:revive,staticcheck
	}

	s.ctx = context.WithValue(s.ctx, interfaces.ExtraOptionsUDPMask, string(s.password)) //nolint:revive,staticcheck

	if config.HandshakeMaskingPaddingSize != 0 {
		ctxv := &interfaces.ExtraOptionsUsePacketArmorValue{PacketArmorPaddingTo: int(config.HandshakeMaskingPaddingSize), UsePacketArmor: true}
		s.ctx = context.WithValue(s.ctx, interfaces.ExtraOptionsUsePacketArmor, ctxv) //nolint:revive,staticcheck
	}

	return s, nil
}

func enableInterface(s *status) error { //nolint: unparam
	s.transport = s
	if s.config.EnableStabilization {
		s.transport = uniserver.NewUnifiedConnectionTransportHub(s, s.ctx)
	}
	if s.config.EnableStabilization {
		s.transport = udpunis.NewUdpUniServer(string(s.password), s.ctx, s.transport)
	}
	return nil
}

func (h *Handler) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	proxyEnvironment := envctx.EnvironmentFromContext(h.ctx).(environment.ProxyEnvironment)
	statusInstanceIfce, err := proxyEnvironment.TransientStorage().Get(ctx, "status")
	if err != nil {
		return newError("uninitialized handler").Base(err)
	}
	statusInstance := statusInstanceIfce.(*status)
	err = h.ensureStarted(statusInstance)
	if err != nil {
		return newError("unable to initialize").Base(err)
	}
	finish := done.New()
	conn = newUDPConnAdaptor(conn, finish)
	var initialData [1600]byte
	c, err := conn.Read(initialData[:])
	if err != nil {
		return newError("unable to read initial data").Base(err)
	}
	connID := session.IDFromContext(ctx)
	vconn, connctx := udpServer.PrepareIncomingUDPConnection(conn, statusInstance.ctx, initialData[:c], strconv.FormatInt(int64(connID), 10))
	connctx = statusInstance.transport.Connection(vconn, connctx)
	if connctx == nil {
		return newError("invalid connection discarded")
	}
	<-finish.Wait()
	return nil
}

func (h *Handler) ensureStarted(s *status) error {
	s.access.Lock()
	defer s.access.Unlock()
	if s.transport == nil {
		err := enableInterface(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*UDPProtocolConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewUDPInboundHandler(ctx, config.(*UDPProtocolConfig))
	}))
}
