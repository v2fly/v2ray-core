package mekya

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/packetconn"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembly"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/httprt"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const protocolName = "mekya"

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return nil, newError("meek is a transport")
	}))

	common.Must(internet.RegisterTransportDialer(protocolName, mekyaDial))
	common.Must(internet.RegisterTransportListener(protocolName, mekyaListen))
}

func mekyaDial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	mekyaSetting := streamSettings.ProtocolSettings.(*Config)
	packetConnAssembler := &packetconn.ClientConfig{}
	packetConnAssembler.PollingIntervalInitial = mekyaSetting.PollingIntervalInitial
	packetConnAssembler.MaxRequestSize = mekyaSetting.MaxRequestSize
	packetConnAssembler.MaxWriteDelay = mekyaSetting.MaxWriteDelay
	packetConnAssembler.UnderlyingTransportName = "kcp"
	packetConnAssembler.UnderlyingTransportSetting = serial.ToTypedMessage(mekyaSetting.Kcp)

	httprtSetting := &httprt.ClientConfig{
		Http: &httprt.HTTPConfig{
			UrlPrefix: mekyaSetting.Url,
		},
		H2PoolSize: mekyaSetting.H2PoolSize,
	}

	request := &assembly.Config{
		Assembler:    serial.ToTypedMessage(packetConnAssembler),
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

func mekyaListen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, callback internet.ConnHandler) (internet.Listener, error) {
	mekyaSetting := streamSettings.ProtocolSettings.(*Config)
	packetConnAssembler := &packetconn.ServerConfig{}
	packetConnAssembler.MaxWriteSize = mekyaSetting.MaxWriteSize
	packetConnAssembler.MaxSimultaneousWriteConnection = mekyaSetting.MaxSimultaneousWriteConnection
	packetConnAssembler.MaxWriteDurationMs = mekyaSetting.MaxWriteDurationMs
	packetConnAssembler.PacketWritingBuffer = mekyaSetting.PacketWritingBuffer
	packetConnAssembler.UnderlyingTransportName = "kcp"
	packetConnAssembler.UnderlyingTransportSetting = serial.ToTypedMessage(mekyaSetting.Kcp)

	request := &assembly.Config{
		Assembler:    serial.ToTypedMessage(packetConnAssembler),
		Roundtripper: serial.ToTypedMessage(&httprt.ServerConfig{}),
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
