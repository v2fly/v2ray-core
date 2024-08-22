package packetconn

import (
	"crypto/rand"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	net2 "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

type packet struct {
	addr string
	data []byte
}

type wrappedPacketConn struct {
	connLock *sync.Mutex
	conn     map[string]*serverSession

	readChan chan packet

	ctx    context.Context
	finish func()
}

func (w *wrappedPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	select {
	case pack := <-w.readChan:
		n := copy(p, pack.data)
		if n < len(pack.data) {
			return n, nil, io.ErrShortBuffer
		}
		return n, &net.UDPAddr{IP: net2.IP(pack.addr)}, nil
	case <-w.ctx.Done():
		return 0, nil, w.ctx.Err()
	}
}

func (w *wrappedPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	w.connLock.Lock()
	conn := w.conn[string(addr.(*net.UDPAddr).IP)]
	w.connLock.Unlock()
	return conn.Write(p)
}

func (w *wrappedPacketConn) Close() error {
	w.finish()
	return nil
}

func (w *wrappedPacketConn) LocalAddr() net.Addr {
	return nil
}

func (w *wrappedPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (w *wrappedPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (w *wrappedPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (w wrappedPacketConn) OnNewSession(ctx context.Context, sess request.Session, opts ...request.SessionOption) error {
	imaginaryAddr := net2.UDPAddr{
		IP:   net2.AnyIPv6.IP(),
		Port: 0,
	}
	rand.Read([]byte(imaginaryAddr.IP))
	session := newServerSession(ctx, sess, string(imaginaryAddr.IP), &w)
	w.connLock.Lock()
	w.conn[string(imaginaryAddr.IP)] = session
	w.connLock.Unlock()
	session.start()
	return nil
}

func newServerSession(ctx context.Context, sess request.Session, name string, listener *wrappedPacketConn) *serverSession {
	_ = ctx
	return &serverSession{session: sess, name: name, listener: listener}
}

type serverSession struct {
	name     string
	session  request.Session
	listener *wrappedPacketConn
}

func (s *serverSession) start() {
	go func() {
		for {
			select {
			case <-s.listener.ctx.Done():
				return
			default:
				buf := make([]byte, 2000)
				n, err := s.session.Read(buf)
				if err != nil || n > 2000 {
					return
				}
				s.listener.readChan <- packet{s.name, buf[:n]}
			}
		}
	}()
}

func (s *serverSession) Write(p []byte) (int, error) {
	return s.session.Write(p)
}

type udpAssemblerServer struct {
	ctx            context.Context
	streamSettings *internet.MemoryStreamConfig
	assembly       request.TransportServerAssembly
	req2packs      *requestToPacketConnServer
	listener       internet.Listener
}

func (u *udpAssemblerServer) Start() error {
	listener, err := u.listen(net2.LocalHostIP, 0)
	if err != nil {
		return newError("failed to listen").Base(err).AtError()
	}
	u.listener = listener
	return nil
}

func (u *udpAssemblerServer) Close() error {
	return u.listener.Close()
}

func (u *udpAssemblerServer) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	return u.req2packs.OnRoundTrip(ctx, req, opts...)
}

func (u *udpAssemblerServer) OnTransportServerAssemblyReady(assembly request.TransportServerAssembly) {
	u.assembly = assembly
}

func newUDPAssemblerServer(ctx context.Context, config *ServerConfig, streamSettings *internet.MemoryStreamConfig) *udpAssemblerServer {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	transportEnvironmentWrapped := &wrappedTransportEnvironment{TransportEnvironment: transportEnvironment}
	transportEnvironmentWrapped.server = newRequestToPacketConnServer(ctx, config)
	wrappedContext := envctx.ContextWithEnvironment(ctx, transportEnvironmentWrapped)
	return &udpAssemblerServer{ctx: wrappedContext, streamSettings: streamSettings, req2packs: transportEnvironmentWrapped.server}
}

func (u *udpAssemblerServer) listen(address net2.Address, port net2.Port) (internet.Listener, error) {
	return internet.ListenTCP(u.ctx, address, port, u.streamSettings, func(connection internet.Connection) {
		err := u.assembly.SessionReceiver().OnNewSession(u.ctx, connection)
		if err != nil {
			newError("failed to handle new session").Base(err).WriteToLog()
		}
	})
}
