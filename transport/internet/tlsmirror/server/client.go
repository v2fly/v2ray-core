package server

import (
	"context"
	"github.com/v2fly/v2ray-core/v5/common/serial"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorbase"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/tlstrafficgen"
)

func newPersistentMirrorTLSDialer(ctx context.Context, serverAddress net.Destination, overrideSecuritySetting proto.Message) *persistentMirrorTLSDialer {
	return &persistentMirrorTLSDialer{
		ctx:                        ctx,
		serverAddress:              serverAddress,
		overridingSecuritySettings: overrideSecuritySetting,
	}

}

type persistentMirrorTLSDialer struct {
	ctx context.Context

	config *Config

	requestNewConnection func(ctx context.Context) error
	incomingConnections  chan net.Conn

	listener *OutboundListener
	outbound *Outbound

	serverAddress              net.Destination
	overridingSecuritySettings proto.Message

	trafficGenerator *tlstrafficgen.TrafficGenerator
}

func (d *persistentMirrorTLSDialer) init(ctx context.Context, config *Config) {
	d.requestNewConnection = func(ctx context.Context) error {
		return nil
	}

	d.ctx = ctx
	d.config = config

	d.incomingConnections = make(chan net.Conn, 4)
	d.listener = NewOutboundListener()
	d.outbound = NewOutbound(d.config.CarrierConnectionTag, d.listener)

	go func() {
		for {
			conn, err := d.listener.Accept()
			if err != nil {
				break
			}
			d.handleIncomingCarrierConnection(ctx, conn)
		}
	}()

	if d.config.EmbeddedTrafficGenerator != nil {
		if d.overridingSecuritySettings != nil && d.config.EmbeddedTrafficGenerator.SecuritySettings == nil {
			d.config.EmbeddedTrafficGenerator.SecuritySettings = serial.ToTypedMessage(d.overridingSecuritySettings)
		}
		d.trafficGenerator = tlstrafficgen.NewTrafficGenerator(d.ctx, d.config.EmbeddedTrafficGenerator,
			d.serverAddress, d.config.CarrierConnectionTag)

		d.requestNewConnection = func(ctx context.Context) error {
			go func() {
				err := d.trafficGenerator.GenerateNextTraffic(d.ctx)
				if err != nil {
					newError("failed to generate next traffic").Base(err).AtWarning().WriteToLog()
				} else {
					newError("traffic generation request sent").AtDebug().WriteToLog()
				}
			}()
			return nil
		}
	}
}

func (d *persistentMirrorTLSDialer) handleIncomingCarrierConnection(ctx context.Context, conn net.Conn) {
	transportEnvironment := envctx.EnvironmentFromContext(d.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	forwardConn, err := dialer(d.ctx, d.serverAddress, d.config.ForwardTag)
	if err != nil {
		newError("failed to dial to destination").Base(err).AtWarning().WriteToLog()
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	cconnState := &clientConnState{
		ctx:        ctx,
		done:       cancel,
		localAddr:  conn.LocalAddr(),
		remoteAddr: conn.RemoteAddr(),
		handler:    d.handleIncomingReadyConnection,
	}

	cconnState.mirrorConn = mirrorbase.NewMirroredTLSConn(ctx, conn, forwardConn, cconnState.onC2SMessage, cconnState.onS2CMessage, conn)
}

type connectionContextGetter interface {
	GetConnectionContext() context.Context
}

func (d *persistentMirrorTLSDialer) handleIncomingReadyConnection(conn internet.Connection) {
	go func() {
		if getter, ok := conn.(connectionContextGetter); ok {
			ctx := getter.GetConnectionContext()

			if managedConnectionController := ctx.Value(tlsmirror.TrafficGeneratorManagedConnectionContextKey); managedConnectionController != nil {
				if controller, ok := managedConnectionController.(tlsmirror.TrafficGeneratorManagedConnection); ok {
					<-controller.WaitConnectionReady().Done()
				}
			}
		}
		d.incomingConnections <- conn
	}()
}

func (d *persistentMirrorTLSDialer) Dial(ctx context.Context,
	dest net.Destination, settings *internet.MemoryStreamConfig) (internet.Connection, error) {
	var recvConn net.Conn
	select {
	case conn := <-d.incomingConnections:
		recvConn = conn
	default:
		err := d.requestNewConnection(ctx)
		if err != nil {
			return nil, newError("failed to request new connection").Base(err)
		}
		select {
		case conn := <-d.incomingConnections:
			recvConn = conn
		}
	}

	if recvConn == nil {
		return nil, newError("failed to receive connection")
	}

	return recvConn, nil

}

func Dial(ctx context.Context, dest net.Destination, settings *internet.MemoryStreamConfig) (internet.Connection, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer, err := transportEnvironment.TransientStorage().Get(ctx, "persistentDialer")
	if err != nil {
		var securitySetting proto.Message
		if settings.SecurityType != "" && settings.SecurityType != "none" {
			securitySetting = settings.SecuritySettings.(proto.Message)
		}
		dialer = newPersistentMirrorTLSDialer(ctx, dest, securitySetting)
		err = transportEnvironment.TransientStorage().Put(ctx, "persistentDialer", dialer)
		if err != nil {
			return nil, newError("failed to put persistent dialer").Base(err)
		}
	}
	conn, err := dialer.(*persistentMirrorTLSDialer).Dial(ctx, dest, settings)
	if err != nil {
		return nil, newError("failed to dial").Base(err)
	}
	return conn, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
