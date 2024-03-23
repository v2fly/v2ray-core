package hysteria2

import (
	"context"

	hy "github.com/apernet/hysteria/core/client"
	hyProtocol "github.com/apernet/hysteria/core/international/protocol"
	"github.com/apernet/quic-go/quicvarint"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

const (
	FrameTypeTCPRequest = 0x401
)

var RunningClient map[net.Destination](hy.Client)

func InitTLSConifg(streamSettings *internet.MemoryStreamConfig) (*hy.TLSConfig, error) {
	tlsSetting := CheckTLSConfig(streamSettings, true)
	if tlsSetting == nil {
		tlsSetting = &tls.Config{
			ServerName:    internalDomain,
			AllowInsecure: true,
		}
	}
	res := &hy.TLSConfig{
		ServerName:         tlsSetting.ServerName,
		InsecureSkipVerify: tlsSetting.AllowInsecure,
	}
	return res, nil
}

func InitAddress(dest net.Destination) (net.Addr, error) {
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

func NewHyClient(dest net.Destination, streamSettings *internet.MemoryStreamConfig) (hy.Client, error) {
	tlsConfig, err := InitTLSConifg(streamSettings)
	if err != nil {
		return nil, err
	}

	serverAddr, err := InitAddress(dest)
	if err != nil {
		return nil, err
	}

	config := streamSettings.ProtocolSettings.(*Config)
	client, _, err := hy.NewClient(&hy.Config{
		TLSConfig:  *tlsConfig,
		Auth:       config.GetPassword(),
		ServerAddr: serverAddr,
	})
	if err != nil {
		return nil, err
	}

	RunningClient[dest] = client
	return client, nil
}

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	config := streamSettings.ProtocolSettings.(*Config)

	var client hy.Client
	client, found := RunningClient[dest]
	if !found {
		var err error
		client, err = NewHyClient(dest, streamSettings)
		if err != nil {
			return nil, err
		}
	}
	quicConn := client.GetQuicConn()

	outbound := session.OutboundFromContext(ctx)
	network := net.Network_TCP
	if outbound != nil {
		network = outbound.Target.Network
	}

	conn := &HyConn{
		local:    quicConn.LocalAddr(),
		remote:   quicConn.RemoteAddr(),
		quicConn: quicConn,
	}

	if network == net.Network_UDP {
		if !config.GetUdp() {
			return nil, newError("UDP extension is not enabled.")
		}
		conn.UseUDPExtension = true
		return conn, nil
	}

	stream, err := client.OpenStream()
	if err != nil {
		delete(RunningClient, dest)
		client.Close()
		return nil, err
	}

	conn.stream = stream

	// write TCP frame type
	frameSize := int(quicvarint.Len(FrameTypeTCPRequest))
	buf := make([]byte, frameSize)
	hyProtocol.VarintPut(buf, FrameTypeTCPRequest)
	stream.Write(buf)
	return conn, nil
}

func init() {
	RunningClient = make(map[net.Destination]hy.Client)
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
