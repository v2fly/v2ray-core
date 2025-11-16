package server

import (
	"context"
	cryptoRand "crypto/rand"
	"math/big"
	"time"

	"github.com/golang/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorbase"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/tlstrafficgen"
)

func newPersistentMirrorTLSDialer(ctx context.Context, config *Config, serverAddress net.Destination, overrideSecuritySetting proto.Message) (*persistentMirrorTLSDialer, error) {
	persistentDialer := &persistentMirrorTLSDialer{
		ctx:                        ctx,
		serverAddress:              serverAddress,
		overridingSecuritySettings: overrideSecuritySetting,
	}

	err := persistentDialer.init(ctx, config)
	if err != nil {
		return nil, newError("failed to initialize persistent mirror TLS dialer").Base(err)
	}

	return persistentDialer, nil
}

type persistentMirrorTLSDialer struct {
	ctx context.Context

	config *Config

	requestNewConnection func(ctx context.Context) error
	incomingConnections  chan net.Conn

	listener *OutboundListener
	outbound *Outbound

	serverAddress              net.Destination
	overridingSecuritySettings proto.Message

	trafficGenerator *tlstrafficgen.TrafficGenerator

	obm outbound.Manager

	explicitNonceCiphersuiteLookup *ciphersuiteLookuper

	enrollmentConfirmationClient *mirrorenrollment.EnrollmentConfirmationClient

	enrollmentServerIdentifier []byte
}

func (d *persistentMirrorTLSDialer) init(ctx context.Context, config *Config) error {
	if err := core.RequireFeatures(ctx, func(om outbound.Manager) {
		d.obm = om
	}); err != nil {
		return err
	}

	d.requestNewConnection = func(ctx context.Context) error {
		return nil
	}

	d.ctx = ctx
	d.config = config

	d.incomingConnections = make(chan net.Conn, 4)
	d.listener = NewOutboundListener()
	d.outbound = NewOutbound(d.config.CarrierConnectionTag, d.listener)

	if len(d.config.ExplicitNonceCiphersuites) > 0 {
		var err error
		d.explicitNonceCiphersuiteLookup, err = newCipherSuiteLookuperFromUint32Array(d.config.ExplicitNonceCiphersuites)
		if err != nil {
			return newError("failed to create explicit nonce ciphersuite lookuper").Base(err)
		}
	} else {
		d.explicitNonceCiphersuiteLookup = newEmptyCipherSuiteLookuper()
		newError("no explicit nonce ciphersuites configured, all ciphersuites will be treated as non-explicit nonce").AtWarning().WriteToLog()
	}

	go func() {
		err := d.outbound.Start()
		if err != nil {
			newError("failed to start outbound listener").Base(err).AtWarning().WriteToLog()
			return
		}

		if err := d.obm.RemoveHandler(context.Background(), d.config.CarrierConnectionTag); err != nil {
			newError("failed to remove existing handler").WriteToLog()
		}

		err = d.obm.AddHandler(context.Background(), d.outbound)
		if err != nil {
			newError("failed to add outbound handler").Base(err).AtWarning().WriteToLog()
			return
		}

		for {
			var ctx context.Context
			conn, err := d.listener.Accept()
			if err != nil {
				break
			}
			if ctxGetter, ok := conn.(connectionContextGetter); ok {
				ctx = ctxGetter.GetConnectionContext()
			} else {
				ctx = d.ctx
				newError("connection does not implement connectionContextGetter, using default context").AtError().WriteToLog()
			}
			d.handleIncomingCarrierConnection(ctx, conn)
		}
	}()

	if d.config.EmbeddedTrafficGenerator != nil {
		if d.overridingSecuritySettings != nil && d.config.EmbeddedTrafficGenerator.SecuritySettings == nil {
			d.config.EmbeddedTrafficGenerator.SecuritySettings = serial.ToTypedMessage(d.overridingSecuritySettings)
		}
		d.trafficGenerator = tlstrafficgen.NewTrafficGenerator(d.ctx, d.config.EmbeddedTrafficGenerator,
			d.serverAddress, d.config.CarrierConnectionTag)

		d.requestNewConnection = func(ctx context.Context) error {
			go func() {
				err := d.trafficGenerator.GenerateNextTraffic(d.ctx)
				if err != nil {
					newError("failed to generate next traffic").Base(err).AtWarning().WriteToLog()
				} else {
					newError("traffic generation request sent").AtDebug().WriteToLog()
				}
			}()
			return nil
		}
	}

	if d.config.ConnectionEnrolment != nil {
		enrollmentServerIdentifier, err := mirrorenrollment.DeriveEnrollmentServerIdentifier(d.config.PrimaryKey)
		if err != nil {
			return newError("failed to derive enrollment server identifier").Base(err).AtError()
		}
		d.enrollmentServerIdentifier = enrollmentServerIdentifier
		d.enrollmentConfirmationClient, err = mirrorenrollment.NewEnrollmentConfirmationClient(d.ctx, d.config.ConnectionEnrolment, enrollmentServerIdentifier)
		if err != nil {
			return newError("failed to create enrollment confirmation client").Base(err).AtError()
		}
	}

	return nil
}

func (d *persistentMirrorTLSDialer) handleIncomingCarrierConnection(ctx context.Context, conn net.Conn) {
	transportEnvironment := envctx.EnvironmentFromContext(d.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	forwardConn, err := dialer(d.ctx, d.serverAddress, d.config.ForwardTag)
	if err != nil {
		newError("failed to dial to destination").Base(err).AtWarning().WriteToLog()
		return
	}

	var firstWriteDelay time.Duration
	if d.config.DeferInstanceDerivedWriteTime != nil {
		firstWriteDelay = time.Duration(d.config.DeferInstanceDerivedWriteTime.BaseNanoseconds)
		if d.config.DeferInstanceDerivedWriteTime.UniformRandomMultiplierNanoseconds > 0 {
			uniformRandomAdd := big.NewInt(int64(d.config.DeferInstanceDerivedWriteTime.UniformRandomMultiplierNanoseconds))
			uniformRandomAddBigInt, err := cryptoRand.Int(cryptoRand.Reader, uniformRandomAdd)
			if err != nil {
				newError("failed to generate random delay").Base(err).AtWarning().WriteToLog()
				return
			}
			uniformRandomAddU64 := uint64(uniformRandomAddBigInt.Int64())
			firstWriteDelay += time.Duration(uniformRandomAddU64)
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	cconnState := &clientConnState{
		ctx:                      ctx,
		done:                     cancel,
		localAddr:                conn.LocalAddr(),
		remoteAddr:               conn.RemoteAddr(),
		handler:                  d.handleIncomingReadyConnection,
		primaryKey:               d.config.PrimaryKey,
		readPipe:                 make(chan []byte, 1),
		firstWrite:               true,
		firstWriteDelay:          firstWriteDelay,
		transportLayerPadding:    d.config.TransportLayerPadding,
		sequenceWatermarkEnabled: d.config.SequenceWatermarkingEnabled,
	}

	cconnState.mirrorConn = mirrorbase.NewMirroredTLSConn(ctx, conn, forwardConn, cconnState.onC2SMessage, cconnState.onS2CMessage, conn,
		d.explicitNonceCiphersuiteLookup.Lookup, cconnState.onC2SMessageTx, cconnState.onS2CMessageTx)
}

type connectionContextGetter interface {
	GetConnectionContext() context.Context
}

type verifyConnectionEnrollment interface {
	VerifyConnectionEnrollmentWithProcessor(connectionEnrollmentConfirmationClient tlsmirror.ConnectionEnrollmentConfirmation) error
}

func (d *persistentMirrorTLSDialer) handleIncomingReadyConnection(conn internet.Connection) {
	go func() {
		if d.config.ConnectionEnrolment != nil {
			if enrollableConn, ok := conn.(verifyConnectionEnrollment); ok {
				if d.enrollmentConfirmationClient != nil {
					err := enrollableConn.VerifyConnectionEnrollmentWithProcessor(d.enrollmentConfirmationClient)
					if err != nil {
						newError("failed to verify connection enrollment").Base(err).AtWarning().WriteToLog()
						return
					}
				} else {
					newError("enrollment confirmation client is not set, connection rejected").AtWarning().WriteToLog()
					return
				}
			} else {
				newError("connection does not implement verifyConnectionEnrollment, connection rejected").AtWarning().WriteToLog()
				return
			}
		}
		var waitedForReady bool
		if getter, ok := conn.(connectionContextGetter); ok {
			ctx := getter.GetConnectionContext()

			if managedConnectionController := ctx.Value(tlsmirror.TrafficGeneratorManagedConnectionContextKey); managedConnectionController != nil {
				if controller, ok := managedConnectionController.(tlsmirror.TrafficGeneratorManagedConnection); ok {
					select { // nolint: staticcheck
					case <-controller.WaitConnectionReady().Done():
						waitedForReady = true
						// TODO: connection might become invalid and never ready, handle this case
						if controller.IsConnectionInvalidated() {
							newError("connection is invalidated, skipping").AtWarning().WriteToLog()
							return
						}
					case <-ctx.Done():
						return
					case <-d.ctx.Done():
						return
					}
				}
			}
		}
		if !waitedForReady {
			newError("unable to wait for connection ready, please verify your setup").AtWarning().WriteToLog()
		}
		d.incomingConnections <- conn
	}()
}

func (d *persistentMirrorTLSDialer) Dial(ctx context.Context,
	dest net.Destination, settings *internet.MemoryStreamConfig,
) (internet.Connection, error) {
	if len(d.enrollmentServerIdentifier) > 0 {
		if mirrorcommon.IsLoopbackProtectionEnabled(ctx, d.enrollmentServerIdentifier) {
			return nil, newError("loopback protection: refusing to dial to self")
		}
	}

	var recvConn net.Conn
	select {
	case conn := <-d.incomingConnections:
		recvConn = conn
	default:
		err := d.requestNewConnection(ctx)
		if err != nil {
			return nil, newError("failed to request new connection").Base(err)
		}
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		select { // nolint: staticcheck
		case conn := <-d.incomingConnections:
			recvConn = conn
		case <-timer.C:
			return nil, newError("timeout waiting for incoming connection")
		case <-ctx.Done():
			return nil, newError("context done while waiting for incoming connection")
		}
	}

	if recvConn == nil {
		return nil, newError("failed to receive connection")
	}

	return recvConn, nil
}

func Dial(ctx context.Context, dest net.Destination, settings *internet.MemoryStreamConfig) (internet.Connection, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer, err := transportEnvironment.TransientStorage().Get(ctx, "persistentDialer")
	if err != nil {
		var securitySetting proto.Message
		if settings.SecurityType != "" && settings.SecurityType != "none" {
			securitySetting = settings.SecuritySettings.(proto.Message)
		}
		config := settings.ProtocolSettings.(*Config)
		detachedContext := core.ToBackgroundDetachedContext(ctx)
		dialer, err = newPersistentMirrorTLSDialer(detachedContext, config, dest, securitySetting)
		if err != nil {
			return nil, newError("failed to create persistent mirror TLS dialer").Base(err)
		}
		err = transportEnvironment.TransientStorage().Put(ctx, "persistentDialer", dialer)
		if err != nil {
			return nil, newError("failed to put persistent dialer").Base(err)
		}
	}
	conn, err := dialer.(*persistentMirrorTLSDialer).Dial(ctx, dest, settings)
	if err != nil {
		return nil, newError("failed to dial").Base(err)
	}
	return conn, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
