package hysteria2

import (
	"context"
	"sync"

	hyClient "github.com/apernet/hysteria/core/v2/client"
	hyProtocol "github.com/apernet/hysteria/core/v2/international/protocol"
	"github.com/apernet/quic-go/quicvarint"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

var RunningClient map[net.Addr](hyClient.Client)
var ClientMutex sync.Mutex
var MBps uint64 = 1000000 / 8 // MByte

func GetClientTLSConfig(streamSettings *internet.MemoryStreamConfig) (*hyClient.TLSConfig, error) {
	config := tls.ConfigFromStreamSettings(streamSettings)
	if config == nil {
		return nil, newError(Hy2MustNeedTLS)
	}
	tlsConfig := config.GetTLSConfig()

	return &hyClient.TLSConfig{
		RootCAs:               tlsConfig.RootCAs,
		ServerName:            tlsConfig.ServerName,
		InsecureSkipVerify:    tlsConfig.InsecureSkipVerify,
		VerifyPeerCertificate: tlsConfig.VerifyPeerCertificate,
	}, nil
}

func ResolveAddress(dest net.Destination) (net.Addr, error) {
	var destAddr *net.UDPAddr
	if dest.Address.Family().IsIP() {
		destAddr = &net.UDPAddr{
			IP:   dest.Address.IP(),
			Port: int(dest.Port),
		}
	} else {
		addr, err := net.ResolveUDPAddr("udp", dest.NetAddr())
		if err != nil {
			return nil, err
		}
		destAddr = addr
	}
	return destAddr, nil
}

type connFactory struct {
	hyClient.ConnFactory

	NewFunc func(addr net.Addr) (net.PacketConn, error)
}

func (f *connFactory) New(addr net.Addr) (net.PacketConn, error) {
	return f.NewFunc(addr)
}

func NewHyClient(serverAddr net.Addr, streamSettings *internet.MemoryStreamConfig) (hyClient.Client, error) {
	tlsConfig, err := GetClientTLSConfig(streamSettings)
	if err != nil {
		return nil, err
	}

	config := streamSettings.ProtocolSettings.(*Config)
	client, _, err := hyClient.NewClient(&hyClient.Config{
		Auth:       config.GetPassword(),
		TLSConfig:  *tlsConfig,
		ServerAddr: serverAddr,
		ConnFactory: &connFactory{
			NewFunc: func(addr net.Addr) (net.PacketConn, error) {
				rawConn, err := internet.ListenSystemPacket(context.Background(), &net.UDPAddr{
					IP:   []byte{0, 0, 0, 0},
					Port: 0,
				}, streamSettings.SocketSettings)
				if err != nil {
					return nil, err
				}
				return rawConn.(*net.UDPConn), nil
			},
		},
		BandwidthConfig: hyClient.BandwidthConfig{MaxTx: config.Congestion.GetUpMbps() * MBps, MaxRx: config.GetCongestion().GetDownMbps() * MBps},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CloseHyClient(serverAddr net.Addr) error {
	ClientMutex.Lock()
	defer ClientMutex.Unlock()

	client, found := RunningClient[serverAddr]
	if found {
		delete(RunningClient, serverAddr)
		return client.Close()
	}
	return nil
}

func GetHyClient(serverAddr net.Addr, streamSettings *internet.MemoryStreamConfig) (hyClient.Client, error) {
	var err error
	var client hyClient.Client

	ClientMutex.Lock()
	client, found := RunningClient[serverAddr]
	ClientMutex.Unlock()
	if !found || !CheckHyClientHealthy(client) {
		if found {
			// retry
			CloseHyClient(serverAddr)
		}
		client, err = NewHyClient(serverAddr, streamSettings)
		if err != nil {
			return nil, err
		}
		ClientMutex.Lock()
		RunningClient[serverAddr] = client
		ClientMutex.Unlock()
	}
	return client, nil
}

func CheckHyClientHealthy(client hyClient.Client) bool {
	quicConn := client.GetQuicConn()
	if quicConn == nil {
		return false
	}
	select {
	case <-quicConn.Context().Done():
		return false
	default:
	}
	return true
}

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	config := streamSettings.ProtocolSettings.(*Config)

	serverAddr, err := ResolveAddress(dest)
	if err != nil {
		return nil, err
	}

	client, err := GetHyClient(serverAddr, streamSettings)
	if err != nil {
		CloseHyClient(serverAddr)
		return nil, err
	}

	quicConn := client.GetQuicConn()
	conn := &HyConn{
		local:  quicConn.LocalAddr(),
		remote: quicConn.RemoteAddr(),
	}

	outbound := session.OutboundFromContext(ctx)
	network := net.Network_TCP
	if outbound != nil {
		network = outbound.Target.Network
		conn.Target = outbound.Target
	}

	if network == net.Network_UDP && config.GetUseUdpExtension() { // only hysteria2 can use udpExtension
		conn.IsUDPExtension = true
		conn.IsServer = false
		conn.ClientUDPSession, err = client.UDP()
		if err != nil {
			CloseHyClient(serverAddr)
			return nil, err
		}
		return conn, nil
	}

	conn.stream, err = client.OpenStream()
	if err != nil {
		CloseHyClient(serverAddr)
		return nil, err
	}

	// write TCP frame type
	frameSize := int(quicvarint.Len(hyProtocol.FrameTypeTCPRequest))
	buf := make([]byte, frameSize)
	hyProtocol.VarintPut(buf, hyProtocol.FrameTypeTCPRequest)
	conn.stream.Write(buf)
	return conn, nil
}

func init() {
	RunningClient = make(map[net.Addr]hyClient.Client)
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
