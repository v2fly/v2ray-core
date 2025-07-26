package server

import (
	cryptoRand "crypto/rand"
	"math/big"
	"strings"
	"time"

	"golang.org/x/net/context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorbase"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type Server struct {
	config *Config

	listener net.Listener
	handler  internet.ConnHandler

	ctx context.Context

	explicitNonceCiphersuiteLookup *ciphersuiteLookuper

	enrollmentConfirmationListener *OutboundListener
	enrollmentConfirmationOutbound *Outbound

	obm outbound.Manager

	enrollmentConfirmationServer    *mirrorenrollment.EnrollmentConfirmationServer
	enrollmentConfirmationProcessor tlsmirror.ConnectionEnrollmentConfirmationProcessor
}

func (s *Server) process(conn net.Conn) {
	transportEnvironment := envctx.EnvironmentFromContext(s.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	port, err := net.PortFromInt(s.config.ForwardPort)
	if err != nil {
		newError("failed to parse port").Base(err).AtWarning().WriteToLog()
		return
	}

	address := net.ParseAddress(s.config.ForwardAddress)

	dest := net.TCPDestination(address, port)

	forwardConn, err := dialer(s.ctx, dest, s.config.ForwardTag)
	if err != nil {
		newError("failed to dial to destination").Base(err).AtWarning().WriteToLog()
		return
	}

	s.accept(conn, forwardConn)
}

func (s *Server) accepts() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "closed") {
				break
			}
			newError("failed to accepted raw connections").Base(err).AtWarning().WriteToLog()
			if strings.Contains(errStr, "too many") {
				time.Sleep(time.Millisecond * 500)
			}
			continue
		}
		go s.process(conn)
	}
}

func (s *Server) Close() error {
	if s.enrollmentConfirmationListener != nil {
		if err := s.enrollmentConfirmationListener.Close(); err != nil {
			newError("failed to close enrollment confirmation listener").Base(err).AtWarning().WriteToLog()
		}
	}
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) accept(clientConn net.Conn, serverConn net.Conn) {
	ctx, cancel := context.WithCancel(s.ctx)

	firstWriteDelay := time.Duration(0)
	if s.config.DeferInstanceDerivedWriteTime != nil {
		firstWriteDelay = time.Duration(s.config.DeferInstanceDerivedWriteTime.BaseNanoseconds)
		if s.config.DeferInstanceDerivedWriteTime.UniformRandomMultiplierNanoseconds > 0 {
			uniformRandomAdd := big.NewInt(int64(s.config.DeferInstanceDerivedWriteTime.UniformRandomMultiplierNanoseconds))
			uniformRandomAddBigInt, err := cryptoRand.Int(cryptoRand.Reader, uniformRandomAdd)
			if err != nil {
				newError("failed to generate random delay").Base(err).AtWarning().WriteToLog()
				return
			}
			uniformRandomAddU64 := uint64(uniformRandomAddBigInt.Int64())
			firstWriteDelay += time.Duration(uniformRandomAddU64)
		}
	}

	conn := &connState{
		ctx:                      ctx,
		done:                     cancel,
		localAddr:                clientConn.LocalAddr(),
		remoteAddr:               clientConn.RemoteAddr(),
		primaryKey:               s.config.PrimaryKey,
		handler:                  s.onIncomingReadyConnection,
		readPipe:                 make(chan []byte, 1),
		firstWrite:               true,
		firstWriteDelay:          firstWriteDelay,
		transportLayerPadding:    s.config.TransportLayerPadding,
		sequenceWatermarkEnabled: s.config.SequenceWatermarkingEnabled,
	}

	if s.config.ConnectionEnrolment != nil {
		conn.connectionEnrollmentEnabled = true
		conn.connectionEnrollmentProcessor = s.enrollmentConfirmationProcessor
	}

	conn.mirrorConn = mirrorbase.NewMirroredTLSConn(ctx, clientConn, serverConn, conn.onC2SMessage, conn.onS2CMessage, conn,
		s.explicitNonceCiphersuiteLookup.Lookup, conn.onC2SMessageTx, conn.onS2CMessageTx)
}

func (s *Server) onIncomingReadyConnection(conn internet.Connection) {
	go s.handler(conn)
}

func (s *Server) init() error {
	if err := core.RequireFeatures(s.ctx, func(om outbound.Manager) {
		s.obm = om
	}); err != nil {
		return err
	}

	if s.config.ConnectionEnrolment != nil {
		s.enrollmentConfirmationListener = NewOutboundListener()
		s.enrollmentConfirmationOutbound = NewOutbound(s.config.ConnectionEnrolment.PrimaryIngressOutbound,
			s.enrollmentConfirmationListener)

		if err := s.enrollmentConfirmationOutbound.Start(); err != nil {
			return newError("failed to start enrollment confirmation outbound").Base(err).AtWarning()
		}

		if err := s.obm.RemoveHandler(context.Background(), s.config.ConnectionEnrolment.PrimaryIngressOutbound); err != nil {
			newError("failed to remove existing handler").Base(err).AtDebug().WriteToLog()
		}

		err := s.obm.AddHandler(context.Background(), s.enrollmentConfirmationOutbound)
		if err != nil {
			return newError("failed to add outbound handler").Base(err)
		}

		s.enrollmentConfirmationProcessor, err = mirrorenrollment.NewServerEnrollmentProcessor(s.config.PrimaryKey)
		if err != nil {
			return newError("failed to create enrollment confirmation processor").Base(err).AtError()
		}

		s.enrollmentConfirmationServer, err = mirrorenrollment.NewEnrollmentConfirmationServer(s.ctx, s.config.ConnectionEnrolment,
			s.enrollmentConfirmationProcessor)
		if err != nil {
			return newError("failed to create enrollment confirmation server").Base(err).AtError()
		}

		go func() {
			for {
				conn, err := s.enrollmentConfirmationListener.Accept()
				if err != nil {
					newError("failed to accept enrollment confirmation connection").Base(err).AtWarning().WriteToLog()
					continue
				}
				go func() {
					if err := s.enrollmentConfirmationServer.HandlePrimaryIngressConnection(s.ctx, conn); err != nil {
						newError("failed to handle primary ingress connection for enrollment confirmation").Base(err).AtWarning().WriteToLog()
					}
				}()
			}
		}()
	}
	return nil
}

func NewServer(ctx context.Context, listener net.Listener, config *Config, handler internet.ConnHandler) (*Server, error) {
	var explicitNonceCiphersuiteLookup *ciphersuiteLookuper
	if len(config.ExplicitNonceCiphersuites) > 0 {
		var err error
		explicitNonceCiphersuiteLookup, err = newCipherSuiteLookuperFromUint32Array(config.ExplicitNonceCiphersuites)
		if err != nil {
			newError("failed to create explicit nonce ciphersuite lookuper").Base(err).AtWarning().WriteToLog()
		}
	} else {
		explicitNonceCiphersuiteLookup = newEmptyCipherSuiteLookuper()
		newError("no explicit nonce ciphersuites configured, all ciphersuites will be treated as non-explicit nonce").AtWarning().WriteToLog()
	}

	s := &Server{
		ctx:                            ctx,
		listener:                       listener,
		config:                         config,
		handler:                        handler,
		explicitNonceCiphersuiteLookup: explicitNonceCiphersuiteLookup,
	}

	if err := s.init(); err != nil {
		return nil, newError("failed to initialize TLS mirror server").Base(err).AtError()
	}

	return s, nil
}
