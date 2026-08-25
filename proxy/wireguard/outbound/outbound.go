package outbound

import (
	"context"
	gonet "net"
	"sync"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/dualStack/happyEyeball"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	cnet "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/packetswitch/gvisorstack"
	"github.com/v2fly/v2ray-core/v5/common/packetswitch/interconnect"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/proxy/wireguard/wgcommon"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewWireguardOutbound(ctx context.Context, config *Config) (*WireguardOutbound, error) {
	if err := validateWireguardConfig(config); err != nil {
		return nil, err
	}
	w := &WireguardOutbound{
		ctx:    ctx,
		config: config,
	}
	// Acquire the DNS client feature.
	if err := core.RequireFeatures(ctx, func(d dns.Client) error {
		w.dnsClient = d
		return nil
	}); err != nil {
		return nil, newError("failed to require dns client feature").Base(err)
	}
	storage := envctx.EnvironmentFromContext(ctx).(environment.ProxyEnvironment).TransientStorage()

	udpState, err := NewClientConnState()
	if err != nil {
		return nil, newError("failed to create UDP connection state").Base(err)
	}
	if err := storage.Put(ctx, ConnectionState, udpState); err != nil {
		return nil, newError("failed to put connection state").Base(err)
	}
	return w, nil
}

func validateWireguardConfig(config *Config) error {
	if config == nil {
		return newError("nil config")
	}
	if config.GetRestricted() && config.GetListenOnSystemNetwork() {
		return newError("restricted WireGuard outbound cannot listen on the system network")
	}
	return nil
}

type WireguardOutbound struct {
	ctx    context.Context
	config *Config

	dnsClient dns.Client
}

type WireguardOutboundSession struct {
	ctx       context.Context
	config    *Config
	closeOnce sync.Once

	stack           *gvisorstack.WrappedStack
	wireguardDevice *wgcommon.WrappedWireguardDevice
	interconnect    *interconnect.NetworkLayerCable

	// system packet conn used when ListenOnSystemNetwork is true
	systemPacketConn internet.PacketConn

	dnsClient dns.Client
}

func (s *WireguardOutboundSession) initFromConfig(ctx context.Context, config *Config) error {
	if config == nil {
		return newError("nil config")
	}
	// create interconnect cable
	cable, err := interconnect.NewNetworkLayerCable(ctx)
	if err != nil {
		return newError("failed to create interconnect cable").Base(err)
	}
	s.interconnect = cable

	// create wireguard device wrapper
	wd, err := wgcommon.NewWrappedWireguardDevice(ctx, config.GetWgDevice())
	if err != nil {
		return newError("failed to create wireguard device").Base(err)
	}
	s.wireguardDevice = wd
	// attach device tunnel to left side of cable
	s.wireguardDevice.SetTunnel(cable.GetLSideDevice())

	// create gvisor stack wrapper if stack config is provided
	if config.GetStack() != nil {
		st, err := gvisorstack.NewStack(ctx, config.GetStack())
		if err != nil {
			return newError("failed to create gvisor stack").Base(err)
		}
		s.stack = st
		if err := s.stack.CreateStackFromNetworkLayerDevice(cable.GetRSideDevice()); err != nil {
			return newError("failed to create stack from network layer device").Base(err)
		}
	}

	return nil
}

const ConnectionState = "ConnectionState"

type ClientConnState struct {
	mu            sync.Mutex
	ready         *sync.Cond
	creating      bool
	active        int
	closed        bool
	session       *WireguardOutboundSession
	sessionCancel context.CancelFunc
}

func (c *ClientConnState) GetOrCreateSession(create func() (*WireguardOutboundSession, error)) (*WireguardOutboundSession, error) {
	return c.GetOrCreateSessionWithContext(context.Background(), func(context.Context) (*WireguardOutboundSession, error) {
		return create()
	})
}

func (c *ClientConnState) GetOrCreateSessionWithContext(
	ctx context.Context,
	create func(context.Context) (*WireguardOutboundSession, error),
) (*WireguardOutboundSession, error) {
	sess, _, err := c.getOrCreateSessionWithContext(ctx, create, false)
	return sess, err
}

func (c *ClientConnState) AcquireOrCreateSessionWithContext(
	ctx context.Context,
	create func(context.Context) (*WireguardOutboundSession, error),
) (*WireguardOutboundSession, func(), error) {
	return c.getOrCreateSessionWithContext(ctx, create, true)
}

func (c *ClientConnState) getOrCreateSessionWithContext(
	ctx context.Context,
	create func(context.Context) (*WireguardOutboundSession, error),
	acquire bool,
) (*WireguardOutboundSession, func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.mu.Lock()
	for c.creating {
		c.ready.Wait()
	}
	if c.closed {
		c.mu.Unlock()
		return nil, nil, newError("UDP connection state is closed")
	}
	if c.session != nil {
		sess := c.session
		release := c.acquireLocked(acquire)
		c.mu.Unlock()
		return sess, release, nil
	}
	sessionCtx, sessionCancel := context.WithCancel(ctx)
	c.creating = true
	c.sessionCancel = sessionCancel
	c.mu.Unlock()

	sess, err := create(sessionCtx)
	if err != nil {
		sessionCancel()
		c.mu.Lock()
		c.sessionCancel = nil
		c.creating = false
		c.ready.Broadcast()
		c.mu.Unlock()
		return nil, nil, newError("failed to initialize UDP State").Base(err)
	}
	if sess == nil {
		sessionCancel()
		c.mu.Lock()
		c.sessionCancel = nil
		c.creating = false
		c.ready.Broadcast()
		c.mu.Unlock()
		return nil, nil, newError("failed to initialize UDP State: nil session")
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		sessionCancel()
		_ = sess.Close()
		c.mu.Lock()
		c.sessionCancel = nil
		c.creating = false
		c.ready.Broadcast()
		c.mu.Unlock()
		return nil, nil, newError("UDP connection state is closed")
	}
	c.session = sess
	release := c.acquireLocked(acquire)
	c.creating = false
	c.ready.Broadcast()
	c.mu.Unlock()
	return sess, release, nil
}

func (c *ClientConnState) acquireLocked(acquire bool) func() {
	if !acquire {
		return nil
	}
	c.active++
	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			c.active--
			c.ready.Broadcast()
			c.mu.Unlock()
		})
	}
}

func (c *ClientConnState) IsTransientStorageLifecycleReceiver() {}

func (c *ClientConnState) Close() error {
	c.mu.Lock()
	c.closed = true
	cancel := c.sessionCancel
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}

	c.mu.Lock()
	for c.creating || c.active != 0 {
		c.ready.Wait()
	}
	sess := c.session
	c.session = nil
	c.sessionCancel = nil
	c.mu.Unlock()
	if sess == nil {
		return nil
	}
	return sess.Close()
}

func (s *WireguardOutboundSession) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		// Close interconnect devices first to stop any further packet injections.
		if s.interconnect != nil {
			_ = s.interconnect.GetLSideDevice().Close()
			_ = s.interconnect.GetRSideDevice().Close()
		}

		if s.systemPacketConn != nil {
			_ = s.systemPacketConn.Close()
		}

		if s.wireguardDevice != nil {
			_ = s.wireguardDevice.Close()
		}

		// Close stack last to quiesce any gVisor internal goroutines that may
		// hold references to PacketBuffers (prevents dec-ref races). Keep the
		// session pointers stable so concurrent shutdown paths cannot race on
		// nil assignments.
		if s.stack != nil {
			_ = s.stack.Close()
		}
	})

	return nil
}

const defaultOutboundNoReceiveTimeout = time.Minute

func outboundNoReceiveTimeout(config *Config) time.Duration {
	if config == nil || config.OutboundNoReceiveTimeoutSec == nil {
		return defaultOutboundNoReceiveTimeout
	}
	return time.Duration(config.GetOutboundNoReceiveTimeoutSec()) * time.Second
}

func createLogicalPacketConn(
	ctx context.Context,
	dispatcher routing.Dispatcher,
	encoding packetaddr.PacketAddrType,
) (cnet.PacketConn, error) {
	switch encoding {
	case packetaddr.PacketAddrType_None:
		return newWireguardPlainPacketConn(ctx, dispatcher)
	case packetaddr.PacketAddrType_Packet:
		return packetaddr.CreatePacketAddrConn(ctx, dispatcher, false)
	case packetaddr.PacketAddrType_Stream:
		return packetaddr.CreatePacketAddrConn(ctx, dispatcher, true)
	default:
		return nil, newError("unsupported outbound packet encoding: ", encoding)
	}
}

func (w *WireguardOutbound) createSession(ctx context.Context, dialer internet.Dialer) (*WireguardOutboundSession, error) {
	s := &WireguardOutboundSession{
		ctx:       ctx,
		config:    w.config,
		dnsClient: w.dnsClient,
	}
	fail := func(err error) (*WireguardOutboundSession, error) {
		_ = s.Close()
		return nil, err
	}

	if err := s.initFromConfig(ctx, w.config); err != nil {
		return fail(err)
	}

	if w.config.GetListenOnSystemNetwork() {
		packetConn, err := internet.ListenSystemPacket(ctx, &gonet.UDPAddr{IP: cnet.AnyIP.IP(), Port: 0}, nil)
		if err != nil {
			return fail(newError("failed to listen on system network").Base(err))
		}
		s.systemPacketConn = packetConn
		s.wireguardDevice.SetConn(packetConn)
	} else {
		encoding := w.config.GetOutboundPacketEncoding()
		switch encoding {
		case packetaddr.PacketAddrType_None, packetaddr.PacketAddrType_Packet, packetaddr.PacketAddrType_Stream:
		default:
			return fail(newError("unsupported outbound packet encoding: ", encoding))
		}
		dispatcher, err := newDialerDispatcher(w, dialer)
		if err != nil {
			return fail(err)
		}
		factory := func(ctx context.Context) (cnet.PacketConn, error) {
			return createLogicalPacketConn(ctx, dispatcher, encoding)
		}
		bind := wgcommon.NewReconnectingNetPacketConnToWg(ctx, factory, outboundNoReceiveTimeout(w.config))
		s.wireguardDevice.SetBind(bind)
	}

	if err := s.wireguardDevice.InitDevice(); err != nil {
		return fail(newError("failed to init wireguard device").Base(err))
	}
	if err := s.wireguardDevice.SetupDeviceWithoutPeers(); err != nil {
		return fail(newError("failed to setup wireguard device").Base(err))
	}
	if err := s.wireguardDevice.AddOrReplacePeers(s.config.GetWgDevice().GetPeers()); err != nil {
		return fail(newError("failed to add peers").Base(err))
	}
	if err := s.wireguardDevice.Up(); err != nil {
		return fail(newError("failed to bring up wireguard device").Base(err))
	}
	return s, nil
}

func (w *WireguardOutbound) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	if isWireguardUnderlayFor(ctx, w) {
		return newError("wireguard outbound cannot use itself as its underlay")
	}
	storage := envctx.EnvironmentFromContext(w.ctx).(environment.ProxyEnvironment).TransientStorage()
	stateIfc, err := storage.Get(ctx, ConnectionState)
	if err != nil {
		return newError("failed to get connection state").Base(err)
	}
	clientState, ok := stateIfc.(*ClientConnState)
	if !ok {
		return newError("bad connection state")
	}

	// create session if needed
	createSession := func(ctx context.Context) (*WireguardOutboundSession, error) {
		return w.createSession(ctx, dialer)
	}
	sessionBaseCtx := preserveWireguardUnderlayMarkers(w.ctx, ctx)
	sess, releaseSession, err := clientState.AcquireOrCreateSessionWithContext(sessionBaseCtx, createSession)
	if err != nil {
		return newError("failed to create or fetch session").Base(err)
	}
	defer releaseSession()

	// Tie every user of the shared session to its lifetime. ClientConnState
	// cancels this context and waits for leases before closing the gVisor stack,
	// avoiding concurrent wrapper teardown and use.
	sessionCtx := sess.ctx
	if sessionCtx == nil {
		sessionCtx = context.Background()
	}
	processCtx, cancelProcess := context.WithCancel(ctx)
	stopSessionCancel := context.AfterFunc(sessionCtx, cancelProcess)
	if sessionCtx.Err() != nil {
		cancelProcess()
	}
	defer stopSessionCancel()
	defer cancelProcess()
	ctx = processCtx

	{
		debugData, err := sess.wireguardDevice.Debug()
		if err != nil {
			newError("failed to debug wireguard device").Base(err).WriteToLog(session.ExportIDToError(ctx))
		}
		newError("wireguard device debug: \n", debugData).AtDebug().WriteToLog(session.ExportIDToError(ctx))
	}

	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified")
	}
	destination := outbound.Target

	// require gVisor stack to process network-level connections
	if sess.stack == nil {
		return newError("gvisor stack is not configured for wireguard outbound")
	}

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, time.Second*300)
	defer cancel()

	if packetConn, err := packetaddr.ToPacketAddrConn(link, destination); err == nil {
		defer func() { _ = packetConn.Close() }()
		pc, err := sess.stack.ListenUDP(ctx, cnet.UDPDestination(nil, 0))
		if err != nil {
			return newError("failed to create udp session in stack").Base(err)
		}
		defer func() { _ = pc.Close() }()

		// Run copy loops and explicitly close resources afterwards to avoid leaks.
		err = nil
		func() {
			requestDone := func() error {
				protocolWriter := pc
				return udp.CopyPacketConn(protocolWriter, packetConn, udp.UpdateActivity(timer))
			}
			responseDone := func() error {
				protocolReader := pc
				return udp.CopyPacketConn(packetConn, protocolReader, udp.UpdateActivity(timer))
			}
			responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
			err = task.Run(ctx, requestDone, responseDoneAndCloseWriter)
		}()

		if err != nil {
			return newError("connection ends").Base(err)
		}
		return nil
	}

	switch destination.Network {
	case cnet.Network_TCP:
		// Dial TCP inside the virtual stack
		ips := w.resolveDNSName(ctx, destination, sess)

		var dialedConn gonet.Conn
		if len(ips) == 0 {
			conn, err := sess.stack.DialTCP(ctx, destination)
			if err != nil {
				return newError("failed to dial tcp in stack").Base(err)
			}
			dialedConn = conn
			newError("dialed ", destination, " with no DNS resolution").AtDebug().WriteToLog(session.ExportIDToError(ctx))
		} else {
			conn, err := happyEyeball.RacingDialer(ctx, destination, ips, func(ctx context.Context, domainDestination cnet.Destination, ips cnet.IP) (internet.Connection, error) {
				dest := cnet.Destination{Network: domainDestination.Network, Address: cnet.IPAddress(ips), Port: domainDestination.Port}
				return sess.stack.DialTCP(ctx, dest)
			}, true, time.Millisecond*300)
			if err != nil {
				return newError("failed to dial tcp in stack with racing dialer").Base(err)
			}
			dialedConn = conn
		}

		defer func() { _ = dialedConn.Close() }()

		requestDone := func() error {
			writer := buf.NewWriter(dialedConn)
			if err := buf.Copy(link.Reader, writer, buf.UpdateActivity(timer)); err != nil {
				return newError("failed to copy request").Base(err)
			}
			return nil
		}

		responseDone := func() error {
			reader := buf.NewReader(dialedConn)
			if err := buf.Copy(reader, link.Writer, buf.UpdateActivity(timer)); err != nil {
				return newError("failed to copy response").Base(err)
			}
			return nil
		}

		if err := task.Run(ctx, requestDone, task.OnSuccess(responseDone, task.Close(link.Writer))); err != nil {
			return newError("connection ends").Base(err)
		}
		return nil

	case cnet.Network_UDP:
		// Create a packet conn on the stack and use mono-dest adapter
		pc, err := sess.stack.ListenUDP(ctx, cnet.UDPDestination(nil, 0))
		if err != nil {
			return newError("failed to create udp session in stack").Base(err)
		}
		mono := udp.NewMonoDestUDPConn(pc, &gonet.UDPAddr{IP: destination.Address.IP(), Port: int(destination.Port)})

		requestDone := func() error {
			return buf.Copy(link.Reader, mono, buf.UpdateActivity(timer))
		}
		responseDone := func() error {
			return buf.Copy(mono, link.Writer, buf.UpdateActivity(timer))
		}

		if err := task.Run(ctx, requestDone, task.OnSuccess(responseDone, task.Close(link.Writer))); err != nil {
			_ = pc.Close()
			return newError("connection ends").Base(err)
		}
		return nil

	default:
		return newError("unsupported network: ", destination.Network)
	}
}

func (w *WireguardOutbound) resolveDNSName(ctx context.Context, destination cnet.Destination, sess *WireguardOutboundSession) []cnet.IP {
	// resolve domain names using dns client if necessary
	if destination.Address != nil && destination.Address.Family().IsDomain() && sess.dnsClient != nil {
		domain := destination.Address.Domain()
		opt := dns.IPOption{
			IPv4Enable: sess.config.DomainStrategy == Config_USE_IP || sess.config.DomainStrategy == Config_USE_IP4,
			IPv6Enable: sess.config.DomainStrategy == Config_USE_IP || sess.config.DomainStrategy == Config_USE_IP6,
			FakeEnable: false,
		}
		ips, err := dns.LookupIPWithOption(sess.dnsClient, domain, opt)
		if err != nil {
			newError("failed to get IP address for domain ", domain).Base(err).WriteToLog(session.ExportIDToError(ctx))
		}
		return ips
	}
	return nil
}

func (w *WireguardOutbound) Close() error {
	storage := envctx.EnvironmentFromContext(w.ctx).(environment.ProxyEnvironment).TransientStorage()
	stateIfc, err := storage.Get(context.Background(), ConnectionState)
	if err != nil || stateIfc == nil {
		return nil
	}
	clientState, ok := stateIfc.(*ClientConnState)
	if !ok {
		return nil
	}
	return clientState.Close()
}

func NewClientConnState() (*ClientConnState, error) {
	state := &ClientConnState{}
	state.ready = sync.NewCond(&state.mu)
	return state, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewWireguardOutbound(ctx, config.(*Config))
	}))
}
