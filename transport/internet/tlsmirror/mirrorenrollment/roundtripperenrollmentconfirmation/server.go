package roundtripperenrollmentconfirmation

import (
	"context"
	"net"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewServer(ctx context.Context, config *ServerConfig) *Server {
	s := &Server{
		ctx:    ctx,
		config: config,
	}
	return s
}

type Server struct {
	config              *ServerConfig
	ctx                 context.Context
	enrollmentProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor
	rttServer           request.RoundTripperServer
}

func (s *Server) Listen(ctx context.Context) (v2net.Listener, error) {
	transportEnvironment := envctx.EnvironmentFromContext(s.ctx).(environment.TransportEnvironment)
	listener := transportEnvironment.Listener()
	addr, err := v2net.ParseDestination(s.config.Listen)
	if err != nil {
		panic(newError("invalid listen address " + s.config.Listen).Base(err).AtError())
	}
	netaddr := &net.TCPAddr{IP: addr.Address.IP(), Port: int(addr.Port)}
	l, err := listener.Listen(s.ctx, netaddr, nil)
	if err != nil {
		panic(newError("failed to listen on " + s.config.Listen).Base(err).AtError())
	}
	return l, nil
}

func (s *Server) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	enrollmentReq := &tlsmirror.EnrollmentConfirmationReq{}
	err = proto.Unmarshal(req.Data, enrollmentReq)
	if err != nil {
		return request.Response{}, newError("failed to unmarshal enrollment confirmation request").Base(err).AtError()
	}
	enrollmentResp, err := s.enrollmentProcessor.VerifyConnectionEnrollment(enrollmentReq)
	if err != nil {
		return request.Response{}, newError("failed to process enrollment confirmation request").Base(err).AtError()
	}
	respData, err := proto.Marshal(enrollmentResp)
	if err != nil {
		return request.Response{}, newError("failed to marshal enrollment confirmation response").Base(err).AtError()
	}
	return request.Response{
		Data: respData,
	}, nil
}

func (s *Server) TripperReceiver() request.TripperReceiver {
	return s
}

func (s *Server) SessionReceiver() request.SessionReceiver {
	return nil
}

func (s *Server) AutoImplListener() request.Listener {
	return s
}

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig)), nil
	}))
}
