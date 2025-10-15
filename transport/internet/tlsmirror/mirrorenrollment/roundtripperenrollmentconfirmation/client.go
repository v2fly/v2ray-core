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

func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
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

	c := &Client{
		ctx:                       ctx,
		config:                    config,
		rttClient:                 rttClient,
		clientTemporaryIdentifier: clientTemporaryIdentifier,
	}

	rttClient.OnTransportClientAssemblyReady(c)

	return c, nil
}

type Client struct {
	config    *ClientConfig
	rttClient request.RoundTripperClient

	clientTemporaryIdentifier []byte

	ctx context.Context

	defaultOutboundTag string
}

func (c *Client) OnConnectionEnrollmentConfirmationClientInstanceConfigReady(config tlsmirror.ConnectionEnrollmentConfirmationClientInstanceConfig) {
	c.defaultOutboundTag = config.DefaultOutboundTag
}

func (c *Client) Dial(ctx context.Context) (net.Conn, error) {
	transportEnvironment := envctx.EnvironmentFromContext(c.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()
	if dialer == nil {
		return nil, newError("no outbound dialer available in transport environment")
	}
	dest, err := v2net.ParseDestination(c.config.Dest)
	if err != nil {
		return nil, newError("failed to parse destination address").Base(err).AtError()
	}
	dest.Network = v2net.Network_TCP
	conn, err := dialer(c.ctx, dest, c.config.OutboundTag)
	if err != nil {
		return nil, newError("failed to dial to destination").Base(err).AtError()
	}
	if c.config.SecurityConfig != nil {
		securityEngine, err := common.CreateObject(c.ctx, c.config.SecurityConfig)
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

func (c *Client) Tripper() request.Tripper {
	return c.rttClient
}

func (c *Client) AutoImplDialer() request.Dialer {
	return c
}

func (c *Client) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	connectionTagServerID := req.ServerIdentifier
	if c.config.ServerIdentity != nil {
		connectionTagServerID = c.config.ServerIdentity
	}
	connectionTag := append(connectionTagServerID, c.clientTemporaryIdentifier...) //nolint:gocritic
	wrappedData, err := proto.Marshal(req)
	if err != nil {
		return nil, newError("failed to marshal enrollment confirmation request").Base(err)
	}
	wreq := request.Request{
		Data:          wrappedData,
		ConnectionTag: connectionTag,
	}
	resp, err := c.rttClient.RoundTrip(c.ctx, wreq)
	if err != nil {
		return nil, newError("failed to perform round trip").Base(err)
	}
	confirmationResp := &tlsmirror.EnrollmentConfirmationResp{}
	if err := proto.Unmarshal(resp.Data, confirmationResp); err != nil {
		return nil, newError("failed to unmarshal enrollment confirmation response").Base(err)
	}
	return confirmationResp, nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
