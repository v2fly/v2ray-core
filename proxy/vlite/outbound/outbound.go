package outbound

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mustafaturan/bus"
	"github.com/xiaokangwang/VLite/ass/udpconn2tun"
	"github.com/xiaokangwang/VLite/interfaces"
	"github.com/xiaokangwang/VLite/interfaces/ibus"
	vltransport "github.com/xiaokangwang/VLite/transport"
	udpsctpserver "github.com/xiaokangwang/VLite/transport/packetsctp/sctprelay"
	"github.com/xiaokangwang/VLite/transport/packetuni/puniClient"
	"github.com/xiaokangwang/VLite/transport/udp/udpClient"
	"github.com/xiaokangwang/VLite/transport/udp/udpuni/udpunic"
	"github.com/xiaokangwang/VLite/transport/uni/uniclient"
	client2 "github.com/xiaokangwang/VLite/workers/client"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewUDPOutboundHandler(ctx context.Context, config *UDPProtocolConfig) (*Handler, error) {
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

// Process implements proxy.Outbound.Process().
func (h *Handler) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
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
	connid := session.IDFromContext(ctx)
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified")
	}
	destination := outbound.Target
	packetConnOut := statusInstance.connAdp.DialUDP(net.UDPAddr{Port: int(connid % 65535)})
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, time.Second*600)

	if packetConn, err := packetaddr.ToPacketAddrConn(link, destination); err == nil {
		requestDone := func() error {
			return udp.CopyPacketConn(packetConnOut, packetConn, udp.UpdateActivity(timer))
		}
		responseDone := func() error {
			return udp.CopyPacketConn(packetConn, packetConnOut, udp.UpdateActivity(timer))
		}
		responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
		if err := task.Run(ctx, requestDone, responseDoneAndCloseWriter); err != nil {
			return newError("connection ends").Base(err)
		}
	}
	return newError("unrecognized connection")
}

func (h *Handler) ensureStarted(s *status) error {
	s.access.Lock()
	defer s.access.Unlock()
	if s.TunnelRxFromTun == nil {
		err := enableInterface(s)
		if err != nil {
			return err
		}
	}
	return nil
}

type status struct {
	ctx      context.Context
	password []byte
	msgbus   *bus.Bus

	udpdialer vltransport.UnderlayTransportDialer
	puni      *puniClient.PacketUniClient
	udprelay  *udpsctpserver.PacketSCTPRelay
	udpserver *client2.UDPClientContext

	TunnelTxToTun   chan interfaces.UDPPacket
	TunnelRxFromTun chan interfaces.UDPPacket

	connAdp *udpconn2tun.UDPConn2Tun

	config UDPProtocolConfig

	access sync.Mutex
}

func createStatusFromConfig(config *UDPProtocolConfig) (*status, error) { //nolint:unparam
	s := &status{password: []byte(config.Password)}
	ctx := context.Background()

	s.msgbus = ibus.NewMessageBus()
	ctx = context.WithValue(ctx, interfaces.ExtraOptionsMessageBus, s.msgbus) //nolint:revive,staticcheck

	ctx = context.WithValue(ctx, interfaces.ExtraOptionsDisableAutoQuitForClient, true) //nolint:revive,staticcheck

	if config.EnableFec {
		ctx = context.WithValue(ctx, interfaces.ExtraOptionsUDPFECEnabled, true) //nolint:revive,staticcheck
	}

	if config.ScramblePacket {
		ctx = context.WithValue(ctx, interfaces.ExtraOptionsUDPShouldMask, true) //nolint:revive,staticcheck
	}

	ctx = context.WithValue(ctx, interfaces.ExtraOptionsUDPMask, string(s.password)) //nolint:revive,staticcheck

	if config.HandshakeMaskingPaddingSize != 0 {
		ctxv := &interfaces.ExtraOptionsUsePacketArmorValue{PacketArmorPaddingTo: int(config.HandshakeMaskingPaddingSize), UsePacketArmor: true}
		ctx = context.WithValue(ctx, interfaces.ExtraOptionsUsePacketArmor, ctxv) //nolint:revive,staticcheck
	}

	destinationString := fmt.Sprintf("%v:%v", config.Address.AsAddress().String(), config.Port)

	s.udpdialer = udpClient.NewUdpClient(destinationString, ctx)
	if config.EnableStabilization {
		s.udpdialer = udpunic.NewUdpUniClient(string(s.password), ctx, s.udpdialer)
		s.udpdialer = uniclient.NewUnifiedConnectionClient(s.udpdialer, ctx)
	}
	s.ctx = ctx
	return s, nil
}

func enableInterface(s *status) error {
	conn, err, connctx := s.udpdialer.Connect(s.ctx)
	if err != nil {
		return newError("unable to connect to remote").Base(err)
	}

	C_C2STraffic := make(chan client2.UDPClientTxToServerTraffic, 8)         //nolint:revive,stylecheck
	C_C2SDataTraffic := make(chan client2.UDPClientTxToServerDataTraffic, 8) //nolint:revive,stylecheck
	C_S2CTraffic := make(chan client2.UDPClientRxFromServerTraffic, 8)       //nolint:revive,stylecheck

	C_C2STraffic2 := make(chan interfaces.TrafficWithChannelTag, 8)     //nolint:revive,stylecheck
	C_C2SDataTraffic2 := make(chan interfaces.TrafficWithChannelTag, 8) //nolint:revive,stylecheck
	C_S2CTraffic2 := make(chan interfaces.TrafficWithChannelTag, 8)     //nolint:revive,stylecheck

	go func(ctx context.Context) {
		for {
			select {
			case data := <-C_C2STraffic:
				C_C2STraffic2 <- interfaces.TrafficWithChannelTag(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	go func(ctx context.Context) {
		for {
			select {
			case data := <-C_C2SDataTraffic:
				C_C2SDataTraffic2 <- interfaces.TrafficWithChannelTag(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	go func(ctx context.Context) {
		for {
			select {
			case data := <-C_S2CTraffic2:
				C_S2CTraffic <- client2.UDPClientRxFromServerTraffic(data)
			case <-ctx.Done():
				return
			}
		}
	}(connctx)

	TunnelTxToTun := make(chan interfaces.UDPPacket)
	TunnelRxFromTun := make(chan interfaces.UDPPacket)

	s.TunnelTxToTun = TunnelTxToTun
	s.TunnelRxFromTun = TunnelRxFromTun

	if s.config.EnableStabilization && s.config.EnableRenegotiation {
		s.puni = puniClient.NewPacketUniClient(C_C2STraffic2, C_C2SDataTraffic2, C_S2CTraffic2, s.password, connctx)
		s.puni.OnAutoCarrier(conn, connctx)
		s.udpserver = client2.UDPClient(connctx, C_C2STraffic, C_C2SDataTraffic, C_S2CTraffic, TunnelTxToTun, TunnelRxFromTun, s.puni)
	} else {
		s.udprelay = udpsctpserver.NewPacketRelayClient(conn, C_C2STraffic2, C_C2SDataTraffic2, C_S2CTraffic2, s.password, connctx)
		s.udpserver = client2.UDPClient(connctx, C_C2STraffic, C_C2SDataTraffic, C_S2CTraffic, TunnelTxToTun, TunnelRxFromTun, s.udprelay)
	}

	s.ctx = connctx

	s.connAdp = udpconn2tun.NewUDPConn2Tun(TunnelTxToTun, TunnelRxFromTun)
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*UDPProtocolConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewUDPOutboundHandler(ctx, config.(*UDPProtocolConfig))
	}))
}
