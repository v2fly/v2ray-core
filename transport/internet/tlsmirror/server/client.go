package server

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorbase"
)

func newPersistentMirrorTLSDialer(ctx context.Context) *persistentMirrorTLSDialer {
	return &persistentMirrorTLSDialer{}

}

type persistentMirrorTLSDialer struct {
	ctx context.Context

	config *Config

	requestNewConnection func(ctx context.Context) error
	incomingConnections  chan net.Conn

	listener *OutboundListener
	outbound *Outbound
}

func (d *persistentMirrorTLSDialer) init(ctx context.Context, config *Config) {
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

}

func (d *persistentMirrorTLSDialer) handleIncomingCarrierConnection(ctx context.Context, conn net.Conn) {
	transportEnvironment := envctx.EnvironmentFromContext(d.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	port, err := net.PortFromInt(d.config.ForwardPort)
	if err != nil {
		newError("failed to parse port").Base(err).AtWarning().WriteToLog()
		return
	}

	address := net.ParseAddress(d.config.ForwardAddress)

	dest := net.TCPDestination(address, port)

	forwardConn, err := dialer(d.ctx, dest, d.config.ForwardTag)
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

func (d *persistentMirrorTLSDialer) handleIncomingReadyConnection(conn internet.Connection) {
	d.incomingConnections <- conn
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
		dialer = newPersistentMirrorTLSDialer(ctx)
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
