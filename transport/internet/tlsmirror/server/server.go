package server

import (
	cryptoRand "crypto/rand"
	"math/big"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorbase"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type Server struct {
	config *Config

	listener net.Listener
	handler  internet.ConnHandler

	ctx context.Context

	explicitNonceCiphersuiteLookup *ciphersuiteLookuper
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
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) accept(clientConn net.Conn, serverConn net.Conn) {
	ctx, cancel := context.WithCancel(s.ctx)

	firstWriteDelay := time.Duration(0)
	if s.config.DeferFirstPayloadWriteTime != nil {
		firstWriteDelay = time.Duration(s.config.DeferFirstPayloadWriteTime.BaseNanoseconds)
		if s.config.DeferFirstPayloadWriteTime.UniformRandomMultiplierNanoseconds > 0 {
			uniformRandomAdd := big.NewInt(int64(s.config.DeferFirstPayloadWriteTime.UniformRandomMultiplierNanoseconds))
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
		ctx:             ctx,
		done:            cancel,
		localAddr:       clientConn.LocalAddr(),
		remoteAddr:      clientConn.RemoteAddr(),
		primaryKey:      s.config.PrimaryKey,
		handler:         s.onIncomingReadyConnection,
		readPipe:        make(chan []byte, 1),
		firstWrite:      true,
		firstWriteDelay: firstWriteDelay,
	}

	conn.mirrorConn = mirrorbase.NewMirroredTLSConn(ctx, clientConn, serverConn, conn.onC2SMessage, nil, conn,
		s.explicitNonceCiphersuiteLookup.Lookup)
}

func (s *Server) onIncomingReadyConnection(conn internet.Connection) {
	go s.handler(conn)
}

func NewServer(ctx context.Context, listener net.Listener, config *Config, handler internet.ConnHandler) *Server {
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

	return &Server{
		ctx:                            ctx,
		listener:                       listener,
		config:                         config,
		handler:                        handler,
		explicitNonceCiphersuiteLookup: explicitNonceCiphersuiteLookup,
	}
}
