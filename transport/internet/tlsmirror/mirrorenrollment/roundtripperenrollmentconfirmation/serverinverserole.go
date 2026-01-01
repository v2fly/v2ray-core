package roundtripperenrollmentconfirmation

import (
	"context"
	csrand "crypto/rand"
	"net"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewServerInverseRole(ctx context.Context, config *ServerInverseRoleConfig) (*ServerInverseRole, error) {
	if ctx == nil {
		return nil, newError("context cannot be nil")
	}

	if config == nil {
		return nil, newError("config cannot be nil")
	}

	rttClientConfig, err := serial.GetInstanceOf(config.RoundTripperClient)
	if err != nil {
		return nil, newError("failed to get instance of RoundTripperClient").Base(err)
	}

	rttClientI, err := common.CreateObject(ctx, rttClientConfig)
	if err != nil {
		return nil, newError("failed to create RoundTripperClient").Base(err)
	}

	rttClient, ok := rttClientI.(request.RoundTripperClient)
	if !ok {
		return nil, newError("RoundTripperClient is not a valid request.RoundTripperClient")
	}

	clientTemporaryIdentifier := make([]byte, 16)
	if _, err := csrand.Read(clientTemporaryIdentifier); err != nil {
		return nil, newError("failed to generate client temporary identifier").Base(err)
	}

	c := &ServerInverseRole{
		ctx:                       ctx,
		config:                    config,
		rttClient:                 rttClient,
		clientTemporaryIdentifier: clientTemporaryIdentifier,
	}

	rttClient.OnTransportClientAssemblyReady(c)

	go c.worker(ctx)

	return c, nil
}

type ServerInverseRole struct {
	config    *ServerInverseRoleConfig
	rttClient request.RoundTripperClient

	clientTemporaryIdentifier []byte

	ctx context.Context

	defaultOutboundTag string

	enrollmentProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor
}

func (c *ServerInverseRole) worker(ctx context.Context) {
	for ctx.Err() == nil {
		err := c.pollRemoteForEnrollment(ctx)
		if err != nil {
			newError("error polling remote for enrollment").Base(err).AtWarning().WriteToLog()
		}
	}
	newError("inverse role server quitted").AtWarning().WriteToLog()
}

func (c *ServerInverseRole) Dial(ctx context.Context) (net.Conn, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()
	if dialer == nil {
		return nil, newError("no outbound dialer available in transport environment")
	}
	dest, err := v2net.ParseDestination(c.config.Dest)
	if err != nil {
		return nil, newError("failed to parse destination address").Base(err).AtError()
	}
	dest.Network = v2net.Network_TCP
	conn, err := dialer(ctx, dest, c.config.OutboundTag)
	if err != nil {
		return nil, newError("failed to dial to destination").Base(err).AtError()
	}
	if c.config.SecurityConfig != nil {
		securityConfigSetting, err := serial.GetInstanceOf(c.config.SecurityConfig)
		if err != nil {
			return nil, newError("unable to get security config instance").Base(err)
		}
		securityEngine, err := common.CreateObject(c.ctx, securityConfigSetting)
		if err != nil {
			return nil, newError("unable to create security engine from security settings").Base(err)
		}
		securityEngineTyped, ok := securityEngine.(security.Engine)
		if !ok {
			return nil, newError("type assertion error when create security engine from security settings")
		}
		conn, err = securityEngineTyped.Client(conn, security.OptionWithDestination{Dest: dest})
		if err != nil {
			return nil, newError("unable to create security protocol client from security engine").Base(err)
		}
	}
	return conn, nil
}

func (c *ServerInverseRole) Tripper() request.Tripper {
	return c.rttClient
}

func (c *ServerInverseRole) AutoImplDialer() request.Dialer {
	return c
}

func (s *ServerInverseRole) OnConnectionEnrollmentConfirmationServerInstanceConfigReady(config tlsmirror.ConnectionEnrollmentConfirmationServerInstanceConfig) {
	s.enrollmentProcessor = config.EnrollmentProcessor
}

func (s *ServerInverseRole) pollRemoteForEnrollment(ctx context.Context) error {
	pollAs := s.config.ServerIdentity
	req := request.Request{
		ConnectionTag: pollAs,
	}
	resp, err := s.rttClient.RoundTrip(ctx, req)
	if err != nil {
		return newError("failed to poll remote for enrollment").Base(err)
	}
	if resp.Data == nil {
		return newError("no enrollment confirmation response received from remote")
	}
	enrollmentReq := &tlsmirror.EnrollmentConfirmationReq{}
	err = proto.Unmarshal(resp.Data, enrollmentReq)
	if err != nil {
		return newError("failed to unmarshal enrollment confirmation request").Base(err).AtError()
	}
	enrollmentResp, err := s.enrollmentProcessor.VerifyConnectionEnrollment(enrollmentReq)
	if err != nil {
		return newError("failed to process enrollment confirmation request").Base(err).AtError()
	}
	respData, err := proto.Marshal(enrollmentResp)
	if err != nil {
		return newError("failed to marshal enrollment confirmation response").Base(err).AtError()
	}
	respAs := append(s.config.ServerIdentity, enrollmentReq.ReplyAddressTag...) //nolint:gocritic
	_, err = s.rttClient.RoundTrip(ctx, request.Request{
		Data:          respData,
		ConnectionTag: respAs,
	})
	if err != nil {
		return newError("failed to send enrollment confirmation response back to remote").Base(err)
	}
	newError("successfully processed enrollment confirmation request").AtDebug().WriteToLog()
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ServerInverseRoleConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServerInverseRole(ctx, config.(*ServerInverseRoleConfig))
	}))
}
