package dtls

import (
	"context"
	"errors"
	"io"
	gonet "net"
	"sync"
	"time"

	"github.com/pion/dtls/v3"
	dtlsnet "github.com/pion/dtls/v3/pkg/net"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

type Listener struct {
	config *Config

	sync.Mutex
	addConn internet.ConnHandler
	hub     *udp.Hub

	sessions map[ConnectionID]*dTLSConnWrapped
}

func (l *Listener) Close() error {
	return l.hub.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.hub.Addr()
}

type ConnectionID struct {
	Remote net.Address
	Port   net.Port
}

func newDTLSServerConn(src net.Destination, parent *Listener) *dTLSConn {
	ctx := context.Background()
	ctx, finish := context.WithCancel(ctx)
	return &dTLSConn{
		src:      src,
		parent:   parent,
		readChan: make(chan *buf.Buffer, 256),
		ctx:      ctx,
		finish:   finish,
	}
}

type dTLSConnWrapped struct {
	unencryptedConn *dTLSConn
	dTLSConn        *dtls.Conn
	closeOnce       sync.Once
	closeErr        error
}

func (c *dTLSConnWrapped) Read(b []byte) (int, error) {
	if c == nil || c.dTLSConn == nil {
		return 0, io.ErrClosedPipe
	}
	return c.dTLSConn.Read(b)
}

func (c *dTLSConnWrapped) Write(b []byte) (int, error) {
	if c == nil || c.dTLSConn == nil {
		return 0, io.ErrClosedPipe
	}
	return c.dTLSConn.Write(b)
}

func (c *dTLSConnWrapped) Close() error {
	if c == nil {
		return nil
	}
	c.closeOnce.Do(func() {
		switch {
		case c.dTLSConn != nil:
			c.closeErr = c.dTLSConn.Close()
		case c.unencryptedConn != nil:
			c.closeErr = c.unencryptedConn.Close()
		}
		if errors.Is(c.closeErr, dtls.ErrConnClosed) {
			c.closeErr = nil
		}
	})
	return c.closeErr
}

func (c *dTLSConnWrapped) LocalAddr() gonet.Addr {
	if c == nil || c.dTLSConn == nil {
		return nil
	}
	return c.dTLSConn.LocalAddr()
}

func (c *dTLSConnWrapped) RemoteAddr() gonet.Addr {
	if c == nil || c.dTLSConn == nil {
		return nil
	}
	return c.dTLSConn.RemoteAddr()
}

func (c *dTLSConnWrapped) SetDeadline(t time.Time) error {
	if c == nil || c.dTLSConn == nil {
		return io.ErrClosedPipe
	}
	return c.dTLSConn.SetDeadline(t)
}

func (c *dTLSConnWrapped) SetReadDeadline(t time.Time) error {
	if c == nil || c.dTLSConn == nil {
		return io.ErrClosedPipe
	}
	return c.dTLSConn.SetReadDeadline(t)
}

func (c *dTLSConnWrapped) SetWriteDeadline(t time.Time) error {
	if c == nil || c.dTLSConn == nil {
		return io.ErrClosedPipe
	}
	return c.dTLSConn.SetWriteDeadline(t)
}

func (c *dTLSConnWrapped) ClientIdentity() []byte {
	if c == nil || c.dTLSConn == nil {
		return nil
	}
	state, ready := c.dTLSConn.ConnectionState()
	if !ready {
		return nil
	}
	return append([]byte(nil), state.IdentityHint...)
}

type dTLSConn struct {
	src    net.Destination
	parent *Listener

	readChan chan *buf.Buffer
	ctx      context.Context
	finish   func()
}

func (l *dTLSConn) Read(b []byte) (n int, err error) {
	select {
	case pack := <-l.readChan:
		n := copy(b, pack.Bytes())
		defer pack.Release()
		if n < int(pack.Len()) {
			return n, io.ErrShortBuffer
		}
		return n, nil
	case <-l.ctx.Done():
		return 0, l.ctx.Err()
	}
}

func (l *dTLSConn) Write(b []byte) (n int, err error) {
	return l.parent.hub.WriteTo(b, l.src)
}

func (l *dTLSConn) Close() error {
	l.finish()
	l.parent.Remove(l.src)
	return nil
}

func (l *dTLSConn) LocalAddr() gonet.Addr {
	return nil
}

func (l *dTLSConn) RemoteAddr() gonet.Addr {
	return &net.UDPAddr{
		IP:   l.src.Address.IP(),
		Port: int(l.src.Port.Value()),
	}
}

func (l *dTLSConn) SetDeadline(t time.Time) error {
	return nil
}

func (l *dTLSConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (l *dTLSConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (l *dTLSConn) OnReceive(payload *buf.Buffer) {
	select {
	case l.readChan <- payload:
	default:
	}
}

func NewListener(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (*Listener, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)
	hub, err := udp.ListenUDP(ctx, address, port, streamSettings, udp.HubCapacity(1024))
	if err != nil {
		return nil, err
	}
	l := &Listener{
		addConn:  addConn,
		config:   transportConfiguration,
		sessions: make(map[ConnectionID]*dTLSConnWrapped),
	}
	l.Lock()
	l.hub = hub
	l.Unlock()
	newError("listening on ", address, ":", port).WriteToLog()

	go l.handlePackets()
	return l, err
}

func (l *Listener) handlePackets() {
	receive := l.hub.Receive()
	for payload := range receive {
		l.OnReceive(payload.Payload, payload.Source)
	}
}

func newDTLSConnWrapped(unencryptedConnection *dTLSConn, transportConfiguration *Config) (*dtls.Conn, error) {
	config := &dtls.Config{}
	config.MTU = int(transportConfiguration.Mtu)
	config.ReplayProtectionWindow = int(transportConfiguration.ReplayProtectionWindow)

	switch transportConfiguration.Mode {
	case DTLSMode_PSK:
		config.PSK = func(bytes []byte) ([]byte, error) {
			return transportConfiguration.Psk, nil
		}
		config.PSKIdentityHint = []byte("")
		config.CipherSuites = []dtls.CipherSuiteID{dtls.TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256}
	default:
		newError("unknown dtls mode").WriteToLog()
	}
	dtlsConn, err := dtls.Server(
		dtlsnet.PacketConnFromConn(unencryptedConnection),
		unencryptedConnection.RemoteAddr(),
		config,
	)
	if err != nil {
		return nil, newError("unable to create dtls server conn").Base(err)
	}
	return dtlsConn, err
}

func (l *Listener) OnReceive(payload *buf.Buffer, src net.Destination) {
	id := ConnectionID{
		Remote: src.Address,
		Port:   src.Port,
	}
	l.Lock()
	defer l.Unlock()
	conn, found := l.sessions[id]
	if !found {
		var err error
		unEncryptedConn := newDTLSServerConn(src, l)
		conn = &dTLSConnWrapped{unencryptedConn: unEncryptedConn}
		l.sessions[id] = conn
		go func() {
			conn.dTLSConn, err = newDTLSConnWrapped(unEncryptedConn, l.config)
			if err != nil {
				newError("unable to accept new dtls connection").Base(err).WriteToLog()
				return
			}
			if err := conn.dTLSConn.Handshake(); err != nil {
				newError("unable to complete dtls handshake").Base(err).WriteToLog()
				_ = conn.Close()
				return
			}
			l.addConn(internet.Connection(conn))
		}()
	}
	conn.unencryptedConn.OnReceive(payload)
}

func (l *Listener) Remove(src net.Destination) {
	l.Lock()
	defer l.Unlock()
	id := ConnectionID{
		Remote: src.Address,
		Port:   src.Port,
	}
	delete(l.sessions, id)
}

func ListenDTLS(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	return NewListener(ctx, address, port, streamSettings, addConn)
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, ListenDTLS))
}
