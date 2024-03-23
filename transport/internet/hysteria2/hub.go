package hysteria2

import (
	"context"

	hy "github.com/apernet/hysteria/core/server"
	"github.com/apernet/quic-go"
	"github.com/apernet/quic-go/http3"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

// Listener is an internet.Listener that listens for TCP connections.
type Listener struct {
	hyServer hy.Server
	rawConn  net.PacketConn
	addConn  internet.ConnHandler
}

// Addr implements internet.Listener.Addr.
func (l *Listener) Addr() net.Addr {
	return l.rawConn.LocalAddr()
}

// Close implements internet.Listener.Close.
func (l *Listener) Close() error {
	return l.hyServer.Close()
}

func (l *Listener) ProxyStreamHijacker(ft http3.FrameType, conn quic.Connection, stream quic.Stream, err error) (bool, error) {
	// err always == nil

	tcpConn := &HyConn{
		stream: stream,
		local:  conn.LocalAddr(),
		remote: conn.RemoteAddr(),
	}
	l.addConn(tcpConn)
	return true, nil
}

// Listen creates a new Listener based on configurations.
func Listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	if address.Family().IsDomain() {
		return nil, nil
	}

	config := streamSettings.ProtocolSettings.(*Config)
	rawConn, err := internet.ListenSystemPacket(context.Background(), &net.UDPAddr{
		IP:   address.IP(),
		Port: int(port),
	}, streamSettings.SocketSettings)
	if err != nil {
		return nil, err
	}

	listener := &Listener{
		rawConn: rawConn,
		addConn: handler,
	}

	hyServer, err := hy.NewServer(&hy.Config{
		Conn:                  rawConn,
		TLSConfig:             *GetTLSConfig(streamSettings),
		Authenticator:         &Authenticator{Password: config.GetPassword()},
		IgnoreClientBandwidth: config.GetIgnoreClientBandwidth(),
		DisableUDP:            !config.GetUdp(),
		StreamHijacker:        listener.ProxyStreamHijacker, // acceptStreams
	})
	if err != nil {
		return nil, err
	}

	listener.hyServer = hyServer
	go hyServer.Serve()
	return listener, nil
}

func CheckTLSConfig(streamSettings *internet.MemoryStreamConfig, isClient bool) *tls.Config {
	if streamSettings == nil || streamSettings.SecuritySettings == nil {
		return nil
	}
	tlsSetting := streamSettings.SecuritySettings.(*tls.Config)
	if tlsSetting.ServerName == "" || (len(tlsSetting.Certificate) == 0 && !isClient) {
		return nil
	}
	return tlsSetting
}

func GetTLSConfig(streamSettings *internet.MemoryStreamConfig) *hy.TLSConfig {
	tlsSetting := CheckTLSConfig(streamSettings, false)
	if tlsSetting == nil {
		tlsSetting = &tls.Config{
			Certificate: []*tls.Certificate{
				tls.ParseCertificate(
					cert.MustGenerate(nil, cert.DNSNames(internalDomain), cert.CommonName(internalDomain)),
				),
			},
		}
	}
	return &hy.TLSConfig{Certificates: tlsSetting.GetTLSConfig().Certificates}
}

type Authenticator struct {
	Password string
}

func (a *Authenticator) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string) {
	if auth == a.Password || a.Password == "" {
		return true, "user"
	}
	return false, ""
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
