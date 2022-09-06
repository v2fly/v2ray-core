package request

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type TransportClientAssembly interface {
	Tripper() Tripper
	AutoImplDialer() Dialer
}

type TransportClientAssemblyReceiver interface {
	OnTransportClientAssemblyReady(TransportClientAssembly)
}

type TransportServerAssembly interface {
	TripperReceiver() TripperReceiver
	SessionReceiver() SessionReceiver
	AutoImplListener() Listener
}

type TransportServerAssemblyReceiver interface {
	OnTransportServerAssemblyReady(TransportServerAssembly)
}

type SessionCreator interface {
	NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
}

type SessionReceiver interface {
	OnNewSession(ctx context.Context, sess Session, opts ...SessionOption) error
}

type Dialer interface {
	Dial(ctx context.Context) (net.Conn, error)
}

type Listener interface {
	Listen(ctx context.Context) (net.Listener, error)
}
