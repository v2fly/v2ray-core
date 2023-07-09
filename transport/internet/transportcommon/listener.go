package transportcommon

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func ListenWithSecuritySettings(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig) (
	net.Listener, error,
) {
	var l net.Listener

	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	transportListener := transportEnvironment.Listener()

	if port == net.Port(0) { // unix
		listener, err := transportListener.Listen(ctx, &net.UnixAddr{
			Name: address.Domain(),
			Net:  "unix",
		}, streamSettings.SocketSettings)
		if err != nil {
			return nil, newError("failed to listen unix domain socket on ", address).Base(err)
		}
		newError("listening unix domain socket on ", address).WriteToLog(session.ExportIDToError(ctx))
		l = listener
	} else { // tcp
		listener, err := transportListener.Listen(ctx, &net.TCPAddr{
			IP:   address.IP(),
			Port: int(port),
		}, streamSettings.SocketSettings)
		if err != nil {
			return nil, newError("failed to listen TCP on ", address, ":", port).Base(err)
		}
		newError("listening TCP on ", address, ":", port).WriteToLog(session.ExportIDToError(ctx))
		l = listener
	}

	if streamSettings.SocketSettings != nil && streamSettings.SocketSettings.AcceptProxyProtocol {
		newError("accepting PROXY protocol").AtWarning().WriteToLog(session.ExportIDToError(ctx))
	}
	return l, nil
}
