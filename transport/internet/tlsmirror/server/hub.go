package server

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"
)

func ListenTLSMirror(ctx context.Context, address net.Address, port net.Port,
	streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler,
) (internet.Listener, error) {
	tlsMirrorSettings := streamSettings.ProtocolSettings.(*Config)
	listener, err := transportcommon.ListenWithSecuritySettings(ctx, address, port, streamSettings)
	if err != nil {
		return nil, newError("failed to listen TLS mirror").Base(err)
	}

	tlsMirrorServer, err := NewServer(ctx, listener, tlsMirrorSettings, handler)
	if err != nil {
		return nil, newError("failed to create TLS mirror server").Base(err)
	}

	go tlsMirrorServer.accepts()

	return tlsMirrorServer, nil
}

const protocolName = "tlsmirror"

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, ListenTLSMirror))
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
