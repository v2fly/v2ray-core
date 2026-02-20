package hysteria2

import (
	"context"

	"github.com/apernet/quic-go"
	"github.com/apernet/quic-go/http3"
	"github.com/v2fly/hysteria/core/v2/international/utils"
	hyServer "github.com/v2fly/hysteria/core/v2/server"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

// Listener is an internet.Listener that listens for TCP connections.
type Listener struct {
	hyServer hyServer.Server
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

func (l *Listener) StreamHijacker(ft http3.FrameType, conn *quic.Conn, stream *utils.QStream, err error) (bool, error) {
	// err always == nil

	tcpConn := &HyConn{
		stream: stream,
		local:  conn.LocalAddr(),
		remote: conn.RemoteAddr(),
	}
	l.addConn(tcpConn)
	return true, nil
}

func (l *Listener) UDPHijacker(entry *hyServer.UdpSessionEntry, originalAddr string) {
	addr, err := net.ResolveUDPAddr("udp", originalAddr)
	if err != nil {
		return
	}
	udpConn := &HyConn{
		IsUDPExtension:   true,
		IsServer:         true,
		ServerUDPSession: entry,
		remote:           addr,
		local:            l.rawConn.LocalAddr(),
	}
	l.addConn(udpConn)
}

// Listen creates a new Listener based on configurations.
func Listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	tlsConfig, err := GetServerTLSConfig(streamSettings)
	if err != nil {
		return nil, err
	}

	if address.Family().IsDomain() {
		return nil, nil
	}

	config := streamSettings.ProtocolSettings.(*Config)
	rawConn, err := internet.ListenSystemPacket(context.Background(),
		&net.UDPAddr{
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

	hyServer, err := hyServer.NewServer(&hyServer.Config{
		Conn:                  rawConn,
		TLSConfig:             *tlsConfig,
		DisableUDP:            !config.GetUseUdpExtension(),
		Authenticator:         &Authenticator{Password: config.GetPassword()},
		StreamHijacker:        listener.StreamHijacker, // acceptStreams
		BandwidthConfig:       hyServer.BandwidthConfig{MaxTx: config.Congestion.GetUpMbps() * MBps, MaxRx: config.GetCongestion().GetDownMbps() * MBps},
		UdpSessionHijacker:    listener.UDPHijacker, // acceptUDPSession
		IgnoreClientBandwidth: config.GetIgnoreClientBandwidth(),
	})
	if err != nil {
		return nil, err
	}

	listener.hyServer = hyServer
	go hyServer.Serve()
	return listener, nil
}

func GetServerTLSConfig(streamSettings *internet.MemoryStreamConfig) (*hyServer.TLSConfig, error) {
	config := tls.ConfigFromStreamSettings(streamSettings)
	if config == nil {
		return nil, newError(Hy2MustNeedTLS)
	}
	tlsConfig := config.GetTLSConfig()

	return &hyServer.TLSConfig{Certificates: tlsConfig.Certificates, GetCertificate: tlsConfig.GetCertificate}, nil
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
