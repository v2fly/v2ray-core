package mirrorenrollment

import (
	"context"
	"net"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
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

	primaryEnrollmentConfirmationClient    tlsmirror.ConnectionEnrollmentConfirmation
	bootstrapEnrollmentConfirmationClients []tlsmirror.ConnectionEnrollmentConfirmation
}

func (c *EnrollmentConfirmationClient) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	resp, err := c.primaryEnrollmentConfirmationClient.VerifyConnectionEnrollment(req)
	if err == nil {
		if resp.Enrolled {
			newError("enrollment confirmation verification with primary enrollment successful").Base(err).WriteToLog()
		} else {
			newError("enrollment confirmation verification with primary enrollment over, not enrolled").Base(err).WriteToLog()
		}
		return resp, nil
	}
	newError("enrollment confirmation verification with primary enrollment failed").Base(err).WriteToLog()
	for _, bootstrapClient := range c.bootstrapEnrollmentConfirmationClients {
		resp, err := bootstrapClient.VerifyConnectionEnrollment(req)
		if err == nil {
			if resp.Enrolled {
				newError("enrollment confirmation verification with bootstrap enrollment successful").Base(err).WriteToLog()
			} else {
				newError("enrollment confirmation verification with bootstrap enrollment over, not enrolled").Base(err).WriteToLog()
			}

			return resp, nil
		}
		newError("enrollment confirmation verification with bootstrap enrollment failed").Base(err).WriteToLog()
	}
	return nil, newError("all enrollment confirmation clients failed").Base(err).AtError()
}

func (c *EnrollmentConfirmationClient) init() error {
	rtt, _, err := httpenrollmentconfirmation.NewClientRoundTripperForEnrollmentConfirmation(
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

			loopbackProtectedCtx := mirrorcommon.SetLoopbackProtectionFlagForContext(c.ctx, c.serverIdentity)
			primaryConfirmationOutbound := c.config.PrimaryEgressOutbound
			if primaryConfirmationOutbound == "" {
				primaryConfirmationOutbound = transportEnvironment.SelfProxyTag()
				loopbackProtectedCtx = mirrorcommon.SetSecondaryLoopbackProtectionFlagForContext(c.ctx, c.serverIdentity)
			}
			return dialer(loopbackProtectedCtx, dest, primaryConfirmationOutbound)
		}, c.serverIdentity)
	if err != nil {
		return newError("failed to create HTTP round tripper for enrollment confirmation").Base(err).AtError()
	}
	c.primaryEnrollmentConfirmationClient, err = httpenrollmentconfirmation.NewHTTPEnrollmentConfirmationClientFromHTTPRoundTripper(rtt)
	if err != nil {
		return newError("failed to create HTTP enrollment confirmation client").Base(err).AtError()
	}

	for _, bootstrapEnrollmentConfirmationConfig := range c.config.BootstrapEgressConfig {
		enrollment, err := serial.GetInstanceOf(bootstrapEnrollmentConfirmationConfig)
		if err != nil {
			return newError("failed to get instance of bootstrap enrollment confirmation config").Base(err).AtError()
		}

		loopbackProtectedCtx := mirrorcommon.SetLoopbackProtectionFlagForContext(c.ctx, c.serverIdentity)
		enrollmentInst, err := common.CreateObject(loopbackProtectedCtx, enrollment)
		if err != nil {
			return newError("failed to create bootstrap enrollment confirmation config").Base(err).AtError()
		}

		enrollmentConfirmation, ok := enrollmentInst.(tlsmirror.ConnectionEnrollmentConfirmation)
		if !ok {
			return newError("bootstrap enrollment confirmation config is not a valid ConnectionEnrollmentConfirmation")
		}

		if configReceiver, ok := enrollmentConfirmation.(tlsmirror.ConnectionEnrollmentConfirmationClientInstanceConfigReceiver); ok {
			configReceiver.OnConnectionEnrollmentConfirmationClientInstanceConfigReady(tlsmirror.ConnectionEnrollmentConfirmationClientInstanceConfig{
				DefaultOutboundTag: c.config.BootstrapEgressOutbound,
			})
		}
		c.bootstrapEnrollmentConfirmationClients = append(c.bootstrapEnrollmentConfirmationClients, enrollmentConfirmation)
	}
	return nil
}
