package grpc

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/grpc/encoding"
	"github.com/v2fly/v2ray-core/v4/transport/internet/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"time"
)

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialgRPC(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial WebSocket").Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}

func dialgRPC(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	grpcSettings := streamSettings.ProtocolSettings.(*Config)

	config := tls.ConfigFromStreamSettings(streamSettings)
	var dialOption = grpc.WithInsecure()

	if config != nil {
		dialOption = grpc.WithTransportCredentials(credentials.NewTLS(config.GetTLSConfig()))
	}

	conn, err := grpc.Dial(
		dest.Address.String()+":"+dest.Port.String(),
		dialOption,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  500 * time.Millisecond,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   19 * time.Millisecond,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
	)

	if err != nil {
		return nil, newError("Cannot dial grpc").Base(err)
	}
	client := encoding.NewGunServiceClient(conn)
	gunservice, err := client.(encoding.GunServiceClientX).TunCustomName(ctx, grpcSettings.ServiceName)
	if err != nil {
		return nil, newError("Cannot dial grpc").Base(err)
	}
	return encoding.NewClientConn(gunservice), nil
}
