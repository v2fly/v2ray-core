//go:build !confonly
// +build !confonly

package grpc

import (
	"context"
	gonet "net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/grpc/encoding"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
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

type transportConnectionState struct {
	scopedDialerMap    map[net.Destination]*grpc.ClientConn
	scopedDialerAccess sync.Mutex
}

func (t *transportConnectionState) IsTransientStorageLifecycleReceiver() {
}

func (t *transportConnectionState) Close() error {
	t.scopedDialerAccess.Lock()
	defer t.scopedDialerAccess.Unlock()
	for _, conn := range t.scopedDialerMap {
		_ = conn.Close()
	}
	t.scopedDialerMap = nil
	return nil
}

type dialerCanceller func()

func dialgRPC(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	grpcSettings := streamSettings.ProtocolSettings.(*Config)

	config := tls.ConfigFromStreamSettings(streamSettings)

	transportCredentials := insecure.NewCredentials()
	if config != nil {
		transportCredentials = credentials.NewTLS(config.GetTLSConfig(tls.WithDestination(dest)))
	}
	dialOption := grpc.WithTransportCredentials(transportCredentials)

	conn, canceller, err := getGrpcClient(ctx, dest, dialOption, streamSettings)
	if err != nil {
		return nil, newError("Cannot dial grpc").Base(err)
	}
	client := encoding.NewGunServiceClient(conn)
	gunService, err := client.(encoding.GunServiceClientX).TunCustomName(ctx, grpcSettings.ServiceName)
	if err != nil {
		canceller()
		return nil, newError("Cannot dial grpc").Base(err)
	}
	return encoding.NewGunConn(gunService, nil), nil
}

func getGrpcClient(ctx context.Context, dest net.Destination, dialOption grpc.DialOption, streamSettings *internet.MemoryStreamConfig) (*grpc.ClientConn, dialerCanceller, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	state, err := transportEnvironment.TransientStorage().Get(ctx, "grpc-transport-connection-state")
	if err != nil {
		state = &transportConnectionState{}
		transportEnvironment.TransientStorage().Put(ctx, "grpc-transport-connection-state", state)
		state, err = transportEnvironment.TransientStorage().Get(ctx, "grpc-transport-connection-state")
		if err != nil {
			return nil, nil, newError("failed to get grpc transport connection state").Base(err)
		}
	}
	stateTyped := state.(*transportConnectionState)

	stateTyped.scopedDialerAccess.Lock()
	defer stateTyped.scopedDialerAccess.Unlock()

	if stateTyped.scopedDialerMap == nil {
		stateTyped.scopedDialerMap = make(map[net.Destination]*grpc.ClientConn)
	}

	canceller := func() {
		stateTyped.scopedDialerAccess.Lock()
		defer stateTyped.scopedDialerAccess.Unlock()
		delete(stateTyped.scopedDialerMap, dest)
	}

	if client, found := stateTyped.scopedDialerMap[dest]; found && client.GetState() != connectivity.Shutdown {
		return client, canceller, nil
	}

	conn, err := grpc.NewClient(
		dest.Address.String()+":"+dest.Port.String(),
		dialOption,
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
			return internet.DialSystem(detachedContext, net.TCPDestination(address, port), streamSettings.SocketSettings)
		}),
		grpc.WithDisableServiceConfig(),
	)
	canceller = func() {
		stateTyped.scopedDialerAccess.Lock()
		defer stateTyped.scopedDialerAccess.Unlock()
		delete(stateTyped.scopedDialerMap, dest)
		if err != nil {
			conn.Close()
		}
	}
	stateTyped.scopedDialerMap[dest] = conn
	return conn, canceller, err
}
