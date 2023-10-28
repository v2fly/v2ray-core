package packetconn

import (
	"io"
	gonet "net"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

type udpAssemblerClient struct {
	ctx            context.Context
	streamSettings *internet.MemoryStreamConfig
	assembly       request.TransportClientAssembly
	req2connc      *requestToPacketConnClient
}

func (u *udpAssemblerClient) NewSession(ctx context.Context, opts ...request.SessionOption) (request.Session, error) {
	return u.dial(net.Destination{})
}

func (u *udpAssemblerClient) OnTransportClientAssemblyReady(assembly request.TransportClientAssembly) {
	u.assembly = assembly
	u.req2connc.OnTransportClientAssemblyReady(assembly)
}

func newWrappedConn(in io.ReadWriteCloser) net.Conn {
	return wrappedConn{in}
}

type wrappedConn struct {
	io.ReadWriteCloser
}

func (w wrappedConn) LocalAddr() gonet.Addr {
	return nil
}

func (w wrappedConn) RemoteAddr() gonet.Addr {
	return nil
}

func (w wrappedConn) SetDeadline(t time.Time) error {
	return nil
}

func (w wrappedConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (w wrappedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func newWrappedPacketConn(ctx context.Context) *wrappedPacketConn {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	return &wrappedPacketConn{
		conn:     make(map[string]*serverSession),
		readChan: make(chan packet, 16), ctx: ctxWithCancel, finish: cancel, connLock: &sync.Mutex{},
	}
}

func newUDPAssemblerClient(ctx context.Context, config *ClientConfig, streamSettings *internet.MemoryStreamConfig) *udpAssemblerClient {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	transportEnvironmentWrapped := &wrappedTransportEnvironment{TransportEnvironment: transportEnvironment}
	transportEnvironmentWrapped.client, _ = newRequestToPacketConnClient(ctx, config)
	wrappedContext := envctx.ContextWithEnvironment(ctx, transportEnvironmentWrapped)
	return &udpAssemblerClient{ctx: wrappedContext, streamSettings: streamSettings, req2connc: transportEnvironmentWrapped.client}
}

func (u *udpAssemblerClient) dial(dest net.Destination) (internet.Connection, error) {
	_ = dest
	return internet.Dial(u.ctx, net.TCPDestination(net.LocalHostIP, 0), u.streamSettings)
}
