//go:build !confonly
// +build !confonly

package grpc

import (
	"context"
	"io"
	gonet "net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/grpc/encoding"
	"github.com/v2fly/v2ray-core/v4/transport/internet/tls"
)

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialgRPC(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial Grpc").Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}

type dialerCanceller func()

var (
	globalDialerMap    map[net.Destination]*grpc.ClientConn
	globalDialerAccess sync.Mutex
)

func dialgRPC(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	grpcSettings := streamSettings.ProtocolSettings.(*Config)

	config := tls.ConfigFromStreamSettings(streamSettings)
	dialOption := grpc.WithInsecure()

	if config != nil {
		dialOption = grpc.WithTransportCredentials(credentials.NewTLS(config.GetTLSConfig()))
	}

	conn, canceller, err := getGrpcClient(ctx, dest, dialOption, grpcSettings)
	if err != nil {
		return nil, newError("Cannot dial grpc").Base(err)
	}
	client := encoding.NewGunServiceClient(conn)

	switch grpcSettings.Mode {
	case Mode_Gun:
		gunService, err := client.(encoding.GunServiceClientX).TunCustomName(ctx, grpcSettings.ServiceName)
		if err != nil {
			canceller()
			return nil, newError("Cannot dial grpc").Base(err)
		}
		return encoding.NewGunConn(gunService, nil), nil
	case Mode_Multi:
		gunService, err := client.(encoding.GunServiceClientX).TunMultiCustomName(ctx, grpcSettings.ServiceName)
		if err != nil {
			canceller()
			return nil, newError("Cannot dial grpc").Base(err)
		}
		conn, _ := encoding.NewMultiConn(gunService)
		return conn, nil
	case Mode_Raw:
		gunService, err := client.(encoding.GunServiceClientX).TunRawCustomName(ctx, grpcSettings.ServiceName, grpc.CallContentSubtype("raw"))
		if err != nil {
			canceller()
			return nil, newError("Cannot dial grpc").Base(err)
		}
		conn, _ := encoding.NewRawConn(gunService)
		return conn, nil
	}
	return nil, io.EOF
}

func getGrpcClient(ctx context.Context, dest net.Destination, dialOption grpc.DialOption, grpcSettings *Config) (*grpc.ClientConn, dialerCanceller, error) {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	if globalDialerMap == nil {
		globalDialerMap = make(map[net.Destination]*grpc.ClientConn)
	}

	canceller := func() {
		globalDialerAccess.Lock()
		defer globalDialerAccess.Unlock()
		delete(globalDialerMap, dest)
	}

	// TODO Should support chain proxy to the same destination
	if client, found := globalDialerMap[dest]; found && client.GetState() != connectivity.Shutdown {
		return client, canceller, nil
	}

	grpcOptions := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  500 * time.Millisecond,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   19 * time.Second,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
		grpc.WithContextDialer(func(ctxGrpc context.Context, s string) (gonet.Conn, error) {
			rawHost, rawPort, err := net.SplitHostPort(s)
			if err != nil {
				return nil, err
			}
			if len(rawPort) == 0 {
				rawPort = "443"
			}
			port, err := net.PortFromString(rawPort)
			if err != nil {
				return nil, err
			}
			address := net.ParseAddress(rawHost)
			detachedContext := core.ToBackgroundDetachedContext(ctx)
			return internet.DialSystem(detachedContext, net.TCPDestination(address, port), nil)
		}),
		dialOption,
	}
	if grpcSettings.IdleTimeout > 0 || grpcSettings.HealthCheckTimeout > 0 || grpcSettings.PermitWithoutStream {
		grpcOptions = append(grpcOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * time.Duration(grpcSettings.IdleTimeout),
			Timeout:             time.Second * time.Duration(grpcSettings.HealthCheckTimeout),
			PermitWithoutStream: grpcSettings.PermitWithoutStream,
		}))
	}

	conn, err := grpc.Dial(dest.Address.String()+":"+dest.Port.String(), grpcOptions...)
	globalDialerMap[dest] = conn
	return conn, canceller, err
}
