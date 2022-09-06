package systemnetworkimpl

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func NewSystemNetworkImpl(listener internet.SystemListener, dialer internet.SystemDialer) environment.SystemNetworkCapabilitySet {
	return &systemNetworkImpl{dialer: dialer, listener: listener}
}

type systemDefaultDialer struct{}

func (s systemDefaultDialer) Listen(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.Listener, error) {
	return internet.ListenSystem(ctx, addr, sockopt)
}

func (s systemDefaultDialer) ListenPacket(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.PacketConn, error) {
	return internet.ListenSystemPacket(ctx, addr, sockopt)
}

func (s systemDefaultDialer) Dial(ctx context.Context, source net.Address, destination net.Destination, sockopt *internet.SocketConfig) (net.Conn, error) {
	return internet.DialSystem(ctx, destination, sockopt)
}

func NewSystemNetworkDefault() environment.SystemNetworkCapabilitySet {
	systemDefault := systemDefaultDialer{}
	return &systemNetworkImpl{dialer: systemDefault, listener: systemDefault}
}

type systemNetworkImpl struct {
	listener internet.SystemListener
	dialer   internet.SystemDialer
}

func (s systemNetworkImpl) Dialer() internet.SystemDialer {
	return s.dialer
}

func (s systemNetworkImpl) Listener() internet.SystemListener {
	return s.listener
}

func NewSystemListenerWithDefaultOpt(listener internet.SystemListener, opt *internet.SocketConfig) internet.SystemListener {
	return systemListenerWithDefaultOpt{SystemListener: listener, opt: opt}
}

type systemListenerWithDefaultOpt struct {
	internet.SystemListener
	opt *internet.SocketConfig
}

func (s systemListenerWithDefaultOpt) Listen(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.Listener, error) {
	if sockopt == nil {
		return s.Listen(ctx, addr, s.opt)
	}
	return s.Listen(ctx, addr, sockopt)
}

func (s systemListenerWithDefaultOpt) ListenPacket(ctx context.Context, addr net.Addr, sockopt *internet.SocketConfig) (net.PacketConn, error) {
	if sockopt == nil {
		return s.ListenPacket(ctx, addr, s.opt)
	}
	return s.ListenPacket(ctx, addr, sockopt)
}

func NewSystemDialerWithDefaultOpt(listener internet.SystemDialer, opt *internet.SocketConfig) internet.SystemDialer {
	return systemDialerWithDefaultOpt{SystemDialer: listener, opt: opt}
}

type systemDialerWithDefaultOpt struct {
	internet.SystemDialer
	opt *internet.SocketConfig
}

func (s systemDialerWithDefaultOpt) Dial(ctx context.Context, source net.Address, destination net.Destination, sockopt *internet.SocketConfig) (net.Conn, error) {
	if sockopt == nil {
		return s.Dial(ctx, source, destination, s.opt)
	}
	return s.Dial(ctx, source, destination, sockopt)
}
