package mirrorenrollment

import (
	"context"
	"net"

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

	return &EnrollmentConfirmationServer{
		ctx:                             ctx,
		config:                          config,
		enrollmentProcessor:             enrollmentProcessor,
		primaryIngressConnectionHandler: primaryIngressConnectionHandler,
	}, nil
}

type EnrollmentConfirmationServer struct {
	ctx context.Context

	config *Config

	enrollmentProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor

	primaryIngressConnectionHandler *httpenrollmentconfirmation.HTTPConnectionHub
}

func (s *EnrollmentConfirmationServer) HandlePrimaryIngressConnection(ctx context.Context, conn net.Conn) error {
	err := s.primaryIngressConnectionHandler.ServeConnection(ctx, conn)
	if err != nil {
		return newError("failed to handle primary ingress connection").Base(err).AtError()
	}
	return nil
}
