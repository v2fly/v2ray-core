package meek

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/simple"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembly"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/httprt"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const protocolName = "meek"

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return nil, newError("meek is a transport")
	}))

	common.Must(internet.RegisterTransportDialer(protocolName, meekDial))
	common.Must(internet.RegisterTransportListener(protocolName, meekListen))
}

func meekDial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	meekSetting := streamSettings.ProtocolSettings.(*Config)
	simpleAssembler := &simple.ClientConfig{
		MaxWriteSize:             65536,
		WaitSubsequentWriteMs:    10,
		InitialPollingIntervalMs: 100,
		MaxPollingIntervalMs:     1000,
		MinPollingIntervalMs:     10,
		BackoffFactor:            1.5,
		FailedRetryIntervalMs:    1000,
	}
	httprtSetting := &httprt.ClientConfig{
		Http: &httprt.HTTPConfig{
			UrlPrefix: meekSetting.Url,
		},
	}
	request := &assembly.Config{
		Assembler:    serial.ToTypedMessage(simpleAssembler),
		Roundtripper: serial.ToTypedMessage(httprtSetting),
	}
	constructedSetting := &internet.MemoryStreamConfig{
		ProtocolName:     "request",
		ProtocolSettings: request,
		SecurityType:     streamSettings.SecurityType,
		SecuritySettings: streamSettings.SecuritySettings,
		SocketSettings:   streamSettings.SocketSettings,
	}
	return internet.Dial(ctx, dest, constructedSetting)
}

func meekListen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, callback internet.ConnHandler) (internet.Listener, error) {
	meekSetting := streamSettings.ProtocolSettings.(*Config)
	simpleAssembler := &simple.ServerConfig{MaxWriteSize: 65536}
	httprtSetting := &httprt.ServerConfig{NoDecodingSessionTag: true, Http: &httprt.HTTPConfig{UrlPrefix: meekSetting.Url}}
	request := &assembly.Config{
		Assembler:    serial.ToTypedMessage(simpleAssembler),
		Roundtripper: serial.ToTypedMessage(httprtSetting),
	}
	constructedSetting := &internet.MemoryStreamConfig{
		ProtocolName:     "request",
		ProtocolSettings: request,
		SecurityType:     streamSettings.SecurityType,
		SecuritySettings: streamSettings.SecuritySettings,
		SocketSettings:   streamSettings.SocketSettings,
	}
	return internet.ListenTCP(ctx, address, port, constructedSetting, callback)
}
