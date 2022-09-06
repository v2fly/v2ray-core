package transportcommon

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/transport/internet/security"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func DialWithSecuritySettings(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.Dialer()
	conn, err := dialer.Dial(ctx, nil, dest, streamSettings.SocketSettings)
	if err != nil {
		return nil, newError("failed to dial to ", dest).Base(err)
	}
	securityEngine, err := security.CreateSecurityEngineFromSettings(ctx, streamSettings)
	if err != nil {
		return nil, newError("unable to create security engine").Base(err)
	}

	if securityEngine != nil {
		conn, err = securityEngine.Client(conn, security.OptionWithDestination{Dest: dest})
		if err != nil {
			return nil, newError("unable to create security protocol client from security engine").Base(err)
		}
	}
	return internet.Connection(conn), nil
}
