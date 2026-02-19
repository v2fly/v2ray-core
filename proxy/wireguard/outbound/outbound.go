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
	"github.com/v2fly/v2ray-core/v5/proxy/wireguard/wgcommon"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewWireguardOutbound(ctx context.Context, config *Config) (*WireguardOutbound, error) {
	w := &WireguardOutbound{
		ctx:    ctx,
		config: config,
	}
	// Acquire dns client feature if available
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

type WireguardOutbound struct {
	ctx    context.Context
	config *Config

	dnsClient dns.Client
}

type WireguardOutboundSession struct {
	ctx    context.Context
	config *Config

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
	session  *WireguardOutboundSession
	initOnce *sync.Once
	mu       sync.Mutex
}

func (c *ClientConnState) GetOrCreateSession(create func() (*WireguardOutboundSession, error)) (*WireguardOutboundSession, error) {
	var errOuter error
	c.initOnce.Do(func() {
		sess, err := create()
		if err != nil {
			errOuter = err
			return
		}
		c.mu.Lock()
		c.session = sess
		c.mu.Unlock()
	})
	if errOuter != nil {
		return nil, newError("failed to initialize UDP State").Base(errOuter)
	}
	return c.session, nil
}

func (c *ClientConnState) IsTransientStorageLifecycleReceiver() {}

func (c *ClientConnState) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session == nil {
		return nil
	}
	sess := c.session
	c.session = nil

	// close interconnect devices first to stop any further packet injections
	if sess.interconnect != nil {
		_ = sess.interconnect.GetLSideDevice().Close()
		_ = sess.interconnect.GetRSideDevice().Close()
		sess.interconnect = nil
	}

	// close system packet conn
	if sess.systemPacketConn != nil {
		_ = sess.systemPacketConn.Close()
		sess.systemPacketConn = nil
	}

	// close wireguard device
	if sess.wireguardDevice != nil {
		_ = sess.wireguardDevice.Close()
		sess.wireguardDevice = nil
	}

	// Close stack last to quiesce any gVisor internal goroutines that may
	// hold references to PacketBuffers (prevents dec-ref races).
	if sess.stack != nil {
		_ = sess.stack.Close()
		sess.stack = nil
	}

	return nil
}

func (w *WireguardOutbound) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	// keep dialer for address family preference when resolving domain
	_ = dialer
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
	sess, err := clientState.GetOrCreateSession(func() (*WireguardOutboundSession, error) {
		s := &WireguardOutboundSession{ctx: ctx, config: w.config}
		s.dnsClient = w.dnsClient
		if err := s.initFromConfig(ctx, w.config); err != nil {
			return nil, err
		}

		if !w.config.ListenOnSystemNetwork {
			// SORRRRRY, I tried but it was v2ray's udp support was too difficult to work with
			return nil, newError("unimplemented: listenOnSystemNetwork=false is not implemented yet")
		}

		packetConn, err := internet.ListenSystemPacket(w.ctx, &gonet.UDPAddr{IP: cnet.AnyIP.IP(), Port: 0}, nil)
		if err != nil {
			return nil, newError("failed to listen on system network").Base(err)
		}

		s.systemPacketConn = packetConn
		s.wireguardDevice.SetConn(packetConn)

		// initialize wireguard device now that conn present
		if err := s.wireguardDevice.InitDevice(); err != nil {
			return nil, newError("failed to init wireguard device").Base(err)
		}
		if err := s.wireguardDevice.SetupDeviceWithoutPeers(); err != nil {
			return nil, newError("failed to setup wireguard device").Base(err)
		}
		if err := s.wireguardDevice.AddOrReplacePeers(s.config.WgDevice.GetPeers()); err != nil {
			return nil, newError("failed to add peers").Base(err)
		}
		if err := s.wireguardDevice.Up(); err != nil {
			return nil, newError("failed to bring up wireguard device").Base(err)
		}
		return s, nil
	})
	if err != nil {
		return newError("failed to create or fetch session").Base(err)
	}

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
		pc, err := sess.stack.ListenUDP(ctx, cnet.UDPDestination(cnet.AnyIP, 0))
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
	if !ok || clientState.session == nil {
		return nil
	}
	_ = clientState.Close()
	return nil
}

func NewClientConnState() (*ClientConnState, error) {
	return &ClientConnState{initOnce: &sync.Once{}}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewWireguardOutbound(ctx, config.(*Config))
	}))
}
