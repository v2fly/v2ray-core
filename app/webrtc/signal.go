package webrtc

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
)

type signaler struct {
	ctx         context.Context
	rttClient   request.RoundTripperClient
	dest        string
	outboundTag string
	securityCfg *anypb.Any
}

func newSignaler(ctx context.Context, roundTripperCfg, securityCfg *anypb.Any, dest, outboundTag string) (*signaler, error) {
	if roundTripperCfg == nil {
		return nil, newError("missing round_tripper_client config")
	}

	instance, err := serial.GetInstanceOf(roundTripperCfg)
	if err != nil {
		return nil, newError("failed to get RoundTripperClient config").Base(err)
	}

	object, err := common.CreateObject(ctx, instance)
	if err != nil {
		return nil, newError("failed to create RoundTripperClient").Base(err)
	}

	rttClient, ok := object.(request.RoundTripperClient)
	if !ok {
		return nil, newError("configured object is not a request.RoundTripperClient")
	}

	s := &signaler{
		ctx:         ctx,
		rttClient:   rttClient,
		dest:        dest,
		outboundTag: outboundTag,
		securityCfg: securityCfg,
	}
	rttClient.OnTransportClientAssemblyReady(s)
	return s, nil
}

func (s *signaler) Tripper() request.Tripper {
	return s.rttClient
}

func (s *signaler) AutoImplDialer() request.Dialer {
	return s
}

func (s *signaler) Dial(ctx context.Context) (v2net.Conn, error) {
	if ctx == nil {
		ctx = s.ctx
	}

	instanceNetwork := envctx.EnvironmentFromContext(ctx).(environment.InstanceNetworkCapabilitySet)
	dialer := instanceNetwork.OutboundDialer()
	if dialer == nil {
		return nil, newError("no outbound dialer available in environment")
	}

	dest, err := v2net.ParseDestination(s.dest)
	if err != nil {
		return nil, newError("failed to parse signaling destination").Base(err)
	}
	dest.Network = v2net.Network_TCP

	conn, err := dialer(ctx, dest, s.outboundTag)
	if err != nil {
		return nil, newError("failed to dial signaling destination").Base(err)
	}

	if s.securityCfg == nil {
		return conn, nil
	}

	securityConfig, err := serial.GetInstanceOf(s.securityCfg)
	if err != nil {
		_ = conn.Close()
		return nil, newError("failed to decode security config").Base(err)
	}

	securityObject, err := common.CreateObject(s.ctx, securityConfig)
	if err != nil {
		_ = conn.Close()
		return nil, newError("failed to create security engine").Base(err)
	}

	engine, ok := securityObject.(security.Engine)
	if !ok {
		_ = conn.Close()
		return nil, newError("configured security object is not a security.Engine")
	}

	securedConn, err := engine.Client(conn, security.OptionWithDestination{Dest: dest})
	if err != nil {
		_ = conn.Close()
		return nil, newError("failed to secure signaling connection").Base(err)
	}

	return securedConn, nil
}

func (s *signaler) RoundTrip(ctx context.Context, routingTag, payload []byte) ([]byte, error) {
	resp, err := s.rttClient.RoundTrip(ctx, request.Request{
		ConnectionTag: routingTag,
		Data:          payload,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
