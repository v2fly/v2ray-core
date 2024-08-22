package dtls

import (
	"context"

	"github.com/pion/dtls/v2"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func dialDTLS(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)

	newError("dialing DTLS to ", dest).WriteToLog()

	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.Dialer()

	rawConn, err := dialer.Dial(ctx, nil, dest, streamSettings.SocketSettings)
	if err != nil {
		return nil, newError("failed to dial to dest: ", err).AtWarning().Base(err)
	}
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
		return nil, newError("unknow dtls mode")
	}
	return dtls.Client(rawConn, config)
}

func dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialDTLS(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial request to ", dest).Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, dial))
}
