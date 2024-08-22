package packetconn

import (
	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type wrappedTransportEnvironment struct {
	environment.TransportEnvironment
	client *requestToPacketConnClient
	server *requestToPacketConnServer
}

func (w *wrappedTransportEnvironment) Listen(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.Listener, error) {
	return nil, newError("not implemented")
}

func (w *wrappedTransportEnvironment) ListenPacket(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.PacketConn, error) {
	packetConn := newWrappedPacketConn(ctx)
	w.server.onSessionReceiverReady(packetConn)
	return packetConn, nil
}

func (w *wrappedTransportEnvironment) Dial(ctx context.Context, source net.Address, destination net.Destination, sockopt *internet.SocketConfig) (net.Conn, error) {
	session, err := w.client.Dial()
	if err != nil {
		return nil, err
	}
	return newWrappedConn(session), nil
}

func (w *wrappedTransportEnvironment) Dialer() internet.SystemDialer {
	return w
}

func (w *wrappedTransportEnvironment) Listener() internet.SystemListener {
	return w
}

func newUDPAssemblerServerFromConfig(ctx context.Context, config *ServerConfig) (*udpAssemblerServer, error) {
	instance, err := serial.GetInstanceOf(config.UnderlyingTransportSetting)
	if err != nil {
		return nil, newError("failed to get instance of underlying transport").Base(err).AtError()
	}
	memcfg := &internet.MemoryStreamConfig{ProtocolName: config.UnderlyingTransportName, ProtocolSettings: instance}
	return newUDPAssemblerServer(ctx, config, memcfg), nil
}

func newUDPAssemblerClientFromConfig(ctx context.Context, config *ClientConfig) (*udpAssemblerClient, error) {
	instance, err := serial.GetInstanceOf(config.UnderlyingTransportSetting)
	if err != nil {
		return nil, newError("failed to get instance of underlying transport").Base(err).AtError()
	}
	memcfg := &internet.MemoryStreamConfig{ProtocolName: config.UnderlyingTransportName, ProtocolSettings: instance}
	return newUDPAssemblerClient(ctx, config, memcfg), nil
}

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		serverConfig, ok := config.(*ServerConfig)
		if !ok {
			return nil, newError("not a ServerConfig")
		}
		return newUDPAssemblerServerFromConfig(ctx, serverConfig)
	}))
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		clientConfig, ok := config.(*ClientConfig)
		if !ok {
			return nil, newError("not a ClientConfig")
		}
		return newUDPAssemblerClientFromConfig(ctx, clientConfig)
	}))
}
