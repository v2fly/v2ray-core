package mirrorenrollment

import (
	"context"
	"net"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment/httpenrollmentconfirmation"
)

func NewEnrollmentConfirmationClient(
	ctx context.Context,
	config *Config,
	serverIdentity []byte,
) (*EnrollmentConfirmationClient, error) {
	if ctx == nil {
		return nil, newError("context cannot be nil")
	}

	if config == nil {
		return nil, newError("config cannot be nil")
	}

	ecc := &EnrollmentConfirmationClient{
		ctx:            ctx,
		config:         config,
		serverIdentity: serverIdentity,
	}

	if err := ecc.init(); err != nil {
		return nil, newError("failed to initialize enrollment confirmation client").Base(err).AtError()
	}

	return ecc, nil
}

type EnrollmentConfirmationClient struct {
	ctx context.Context

	config *Config

	serverIdentity []byte

	primaryEnrollmentConfirmationClient tlsmirror.ConnectionEnrollmentConfirmation
}

func (c *EnrollmentConfirmationClient) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	return c.primaryEnrollmentConfirmationClient.VerifyConnectionEnrollment(req)
}

func (c *EnrollmentConfirmationClient) init() error {
	rtt, err := httpenrollmentconfirmation.NewClientRoundTripperForEnrollmentConfirmation(
		func(network, addr string) (net.Conn, error) {
			transportEnvironment := envctx.EnvironmentFromContext(c.ctx).(environment.TransportEnvironment)
			dialer := transportEnvironment.OutboundDialer()
			if dialer == nil {
				return nil, newError("no outbound dialer available in transport environment")
			}
			dest, err := v2net.ParseDestination(addr)
			if err != nil {
				return nil, newError("failed to parse destination address").Base(err).AtError()
			}
			dest.Network = v2net.Network_TCP
			return dialer(c.ctx, dest, c.config.PrimaryEgressOutbound)
		}, c.serverIdentity)
	if err != nil {
		return newError("failed to create HTTP round tripper for enrollment confirmation").Base(err).AtError()
	}
	c.primaryEnrollmentConfirmationClient, err = httpenrollmentconfirmation.NewHTTPEnrollmentConfirmationClientFromHTTPRoundTripper(rtt)
	if err != nil {
		return newError("failed to create HTTP enrollment confirmation client").Base(err).AtError()
	}
	return nil
}
