package mirrorenrollment

import (
	"context"
	"net"

	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment/httpenrollmentconfirmation"
)

func NewEnrollmentConfirmationServer(ctx context.Context, config *Config, enrollmentProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor) (*EnrollmentConfirmationServer, error) {
	if ctx == nil {
		return nil, newError("context cannot be nil")
	}

	if config == nil {
		return nil, newError("config cannot be nil")
	}

	if enrollmentProcessor == nil {
		return nil, newError("enrollment processor cannot be nil")
	}

	enrollmentHandler, err := httpenrollmentconfirmation.NewHTTPEnrollmentConfirmationServerFromConnectionEnrollmentConfirmation(enrollmentProcessor)
	if err != nil {
		return nil, newError("failed to create HTTP enrollment confirmation server").Base(err).AtError()
	}

	primaryIngressConnectionHandler := httpenrollmentconfirmation.NewHTTPConnectionHub(enrollmentHandler)

	s := &EnrollmentConfirmationServer{
		ctx:                             ctx,
		config:                          config,
		enrollmentProcessor:             enrollmentProcessor,
		primaryIngressConnectionHandler: primaryIngressConnectionHandler,
	}

	return s, nil
}

type EnrollmentConfirmationServer struct {
	ctx context.Context

	config *Config

	enrollmentProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor

	primaryIngressConnectionHandler    *httpenrollmentconfirmation.HTTPConnectionHub
	bootstrapIngressConnectionHandlers []tlsmirror.ConnectionEnrollmentConfirmationServerInstanceConfigReceiver
}

func (s *EnrollmentConfirmationServer) HandlePrimaryIngressConnection(ctx context.Context, conn net.Conn) error {
	err := s.primaryIngressConnectionHandler.ServeConnection(ctx, conn)
	if err != nil {
		return newError("failed to handle primary ingress connection").Base(err).AtError()
	}
	return nil
}

func (s *EnrollmentConfirmationServer) init() error {
	for _, handler := range s.config.BootstrapEgressConfig {
		bootstrapEnrollmentHandler, err := serial.GetInstanceOf(handler)
		if err != nil {
			return newError("failed to get instance of bootstrap enrollment handler").Base(err).AtError()
		}
		bootstrapEnrollmentHandlerTyped, ok := bootstrapEnrollmentHandler.(tlsmirror.ConnectionEnrollmentConfirmationServerInstanceConfigReceiver)
		if !ok {
			return newError("bootstrap enrollment handler is not a valid ConnectionEnrollmentConfirmationServerInstanceConfigReceiver")
		}

		bootstrapEnrollmentHandlerTyped.OnConnectionEnrollmentConfirmationServerInstanceConfigReady(
			tlsmirror.ConnectionEnrollmentConfirmationServerInstanceConfig{
				EnrollmentProcessor: s.enrollmentProcessor,
			})
		s.bootstrapIngressConnectionHandlers = append(s.bootstrapIngressConnectionHandlers, bootstrapEnrollmentHandlerTyped)
	}
	return nil
}
