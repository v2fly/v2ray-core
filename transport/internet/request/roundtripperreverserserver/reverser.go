package roundtripperreverserserver

import (
	"context"
	"net"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewReverser(ctx context.Context, config *Config) (*Reverser, error) {
	reverser := &Reverser{
		ctx:    ctx,
		config: config,
	}
	if err := reverser.init(); err != nil {
		return nil, newError("failed to initialize Reverser").Base(err).AtError()
	}
	return reverser, nil
}

type Reverser struct {
	ctx           context.Context
	config        *Config
	rttServer     request.RoundTripperServer
	reverser      request.ReverserImpl
	accessChecker request.ReverserAccessChecker
}

func (s *Reverser) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	serverIntent := len(req.ConnectionTag) == 16
	if serverIntent {
		serverPublic, err := s.accessChecker.CheckReverserAccess(ctx, req.ConnectionTag)
		if err != nil {
			return request.Response{}, newError("reverser access check failed").Base(err).AtError()
		}
		reverserImpl, err := s.reverser.OnAuthenticatedServerIntentRoundTrip(ctx, serverPublic, req, opts...)
		if err != nil {
			return request.Response{}, newError("failed to handle authenticated server round trip").Base(err).AtError()
		}
		return reverserImpl, nil
	}
	if len(req.ConnectionTag) != 32 {
		return request.Response{}, newError("invalid ConnectionTag length")
	}
	reverserImpl, err := s.reverser.OnOtherRoundTrip(ctx, req, opts...)
	if err != nil {
		return request.Response{}, newError("failed to handle client round trip").Base(err).AtError()
	}
	return reverserImpl, nil
}

func (s *Reverser) Listen(ctx context.Context) (v2net.Listener, error) {
	systemNetworkCapabilitySet := envctx.EnvironmentFromContext(s.ctx).(environment.SystemNetworkCapabilitySet)
	listener := systemNetworkCapabilitySet.Listener()
	addr, err := v2net.ParseDestination(s.config.Listen)
	if err != nil {
		return nil, newError("invalid listen address " + s.config.Listen).Base(err).AtError()
	}
	netaddr := &net.TCPAddr{IP: addr.Address.IP(), Port: int(addr.Port)}
	l, err := listener.Listen(ctx, netaddr, nil)
	if err != nil {
		return nil, newError("failed to listen on " + s.config.Listen).Base(err).AtError()
	}
	return l, nil
}

func (s *Reverser) init() error {
	if s.config == nil {
		return newError("nil ServerConfig")
	}

	if s.config.AccessPassphrase == "" {
		return newError("empty AccessPassphrase in ServerConfig")
	}
	accessChecker, err := NewPasswordAccessChecker(s.config.AccessPassphrase)
	if err != nil {
		return newError("failed to create AccessChecker").Base(err).AtError()
	}
	s.accessChecker = accessChecker

	ReverserImplInst, err := NewReverserImpl()
	if err != nil {
		return newError("failed to create ReverserImpl").Base(err).AtError()
	}
	s.reverser = ReverserImplInst

	if s.config.RoundTripperServer == nil {
		return newError("nil RoundTripperServer in ServerConfig")
	}
	RoundTripperServerConfig, err := serial.GetInstanceOf(s.config.RoundTripperServer)
	if err != nil {
		return newError("failed to get instance of RoundTripperServer").Base(err).AtError()
	}
	RoundTripperServerObj, err := common.CreateObject(s.ctx, RoundTripperServerConfig)
	if err != nil {
		return newError("failed to create RoundTripperServer").Base(err).AtError()
	}
	RoundTripperServerTyped, ok := RoundTripperServerObj.(request.RoundTripperServer)
	if !ok {
		return newError("RoundTripperServer is not a valid request.RoundTripperServer")
	}
	s.rttServer = RoundTripperServerTyped
	s.rttServer.OnTransportServerAssemblyReady(s)
	if err := s.rttServer.Start(); err != nil {
		return newError("failed to start RoundTripperServer").Base(err).AtError()
	}
	return nil
}

func (s *Reverser) TripperReceiver() request.TripperReceiver {
	return s
}

func (s *Reverser) SessionReceiver() request.SessionReceiver {
	return nil
}

func (s *Reverser) AutoImplListener() request.Listener {
	return s
}
