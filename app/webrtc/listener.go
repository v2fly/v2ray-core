package webrtc

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	pionice "github.com/pion/ice/v4"
	"github.com/pion/logging"
	pionstun "github.com/pion/stun/v3"
	pionwebrtc "github.com/pion/webrtc/v4"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	featuredns "github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type listenerRuntime interface {
	Tag() string
	NewPeerAPI() (*pionwebrtc.API, pionwebrtc.Configuration, error)
	RequestPortBlossom() bool
	AcceptPortBlossom() bool
	PortBlossomDuration() time.Duration
	RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround() bool
	BlastPorts(ip net.IP) error
	Close() error
}

const defaultPortBlossomDuration = 6 * time.Second

type activeListenerRuntime struct {
	ctx        context.Context
	dnsClient  featuredns.Client
	dispatcher routing.Dispatcher
	config     *LocalWebRTCListener

	mu          sync.RWMutex
	portBlaster interface{ BlastPorts(net.IP) error }
}

func (l *activeListenerRuntime) Tag() string {
	return l.config.Tag
}

func (l *activeListenerRuntime) RequestPortBlossom() bool {
	return l.config.GetRequestPortBlossom()
}

func (l *activeListenerRuntime) AcceptPortBlossom() bool {
	return l.config.GetAcceptPortBlossom()
}

func (l *activeListenerRuntime) PortBlossomDuration() time.Duration {
	if l.config == nil {
		return defaultPortBlossomDuration
	}
	return configuredPortBlossomDuration(l.config.GetPortBlossomDurationSec())
}

func (l *activeListenerRuntime) RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround() bool {
	if l.config == nil {
		return false
	}
	return l.config.GetRequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround()
}

func (l *activeListenerRuntime) NewPeerAPI() (*pionwebrtc.API, pionwebrtc.Configuration, error) {
	settingEngine := pionwebrtc.SettingEngine{}
	settingEngine.SetICETimeouts(defaultDisconnectedTimeout, defaultFailedTimeout, defaultKeepAliveInterval)
	settingEngine.SetNetworkTypes(activeListenerNetworks(l.config))
	settingEngine.SetICEMulticastDNSMode(pionice.MulticastDNSModeDisabled)

	cfg, err := activeListenerICEConfiguration(l.config)
	if err != nil {
		return nil, pionwebrtc.Configuration{}, err
	}

	var blaster interface{ BlastPorts(net.IP) error }
	if l.config.ConnectVia != "" {
		customNet, err := newConnectViaNet(l.ctx, l.dispatcher, l.dnsClient, l.config.ConnectVia, l.config.PacketEncoding)
		if err != nil {
			return nil, pionwebrtc.Configuration{}, err
		}
		settingEngine.SetNet(customNet)
		blaster = customNet
	} else {
		trackingNet, err := newTrackingNet()
		if err != nil {
			return nil, pionwebrtc.Configuration{}, err
		}
		settingEngine.SetNet(trackingNet)
		blaster = trackingNet
	}

	l.mu.Lock()
	l.portBlaster = blaster
	l.mu.Unlock()

	api := pionwebrtc.NewAPI(pionwebrtc.WithSettingEngine(settingEngine))

	return api, cfg, nil
}

func activeListenerICEConfiguration(config *LocalWebRTCListener) (pionwebrtc.Configuration, error) {
	iceServers := make([]pionwebrtc.ICEServer, 0, 1+len(config.TurnServers))
	if len(config.StunServers) > 0 {
		iceServers = append(iceServers, pionwebrtc.ICEServer{
			URLs: append([]string(nil), config.StunServers...),
		})
	}

	turnServers, err := activeListenerTURNICEServers(config)
	if err != nil {
		return pionwebrtc.Configuration{}, err
	}
	iceServers = append(iceServers, turnServers...)

	if len(iceServers) == 0 {
		return pionwebrtc.Configuration{}, newError("active listener ", config.Tag, " requires stun_servers or turn_servers")
	}

	return pionwebrtc.Configuration{
		ICEServers:         iceServers,
		ICETransportPolicy: pionwebrtc.ICETransportPolicyNoHost,
	}, nil
}

func activeListenerTURNICEServers(config *LocalWebRTCListener) ([]pionwebrtc.ICEServer, error) {
	if len(config.TurnServers) == 0 {
		return nil, nil
	}

	iceServers := make([]pionwebrtc.ICEServer, 0, len(config.TurnServers))
	for i, server := range config.TurnServers {
		if server == nil {
			return nil, newError("active listener ", config.Tag, " turn_servers[", i, "] is nil")
		}
		if server.GetUrl() == "" {
			return nil, newError("active listener ", config.Tag, " turn_servers[", i, "] requires url")
		}
		if server.GetUsername() == "" || server.GetPassword() == "" {
			return nil, newError("active listener ", config.Tag, " turn_servers[", i, "] require username and password")
		}

		uri, err := pionstun.ParseURI(server.GetUrl())
		if err != nil {
			return nil, newError("invalid TURN URL for active listener ", config.Tag, ": ", server.GetUrl()).Base(err)
		}
		if uri.Scheme != pionstun.SchemeTypeTURN || uri.Proto != pionstun.ProtoTypeUDP {
			return nil, newError("active listener ", config.Tag, " only supports udp-based TURN URLs, got ", server.GetUrl())
		}

		iceServers = append(iceServers, pionwebrtc.ICEServer{
			URLs:           []string{server.GetUrl()},
			Username:       server.GetUsername(),
			Credential:     server.GetPassword(),
			CredentialType: pionwebrtc.ICECredentialTypePassword,
		})
	}

	return iceServers, nil
}

func (l *activeListenerRuntime) Close() error {
	l.mu.Lock()
	l.portBlaster = nil
	l.mu.Unlock()
	return nil
}

func (l *activeListenerRuntime) BlastPorts(ip net.IP) error {
	l.mu.RLock()
	blaster := l.portBlaster
	l.mu.RUnlock()
	if blaster == nil {
		return newError("active listener ", l.config.Tag, " has no active packet sockets for port blossom")
	}
	return blaster.BlastPorts(ip)
}

type systemListenerRuntime struct {
	config *LocalWebRTCSystemListener

	packetConns []v2net.PacketConn
	udpMux      pionice.UDPMux
}

func newSystemListenerRuntime(ctx context.Context, config *LocalWebRTCSystemListener) (*systemListenerRuntime, error) {
	runtime := &systemListenerRuntime{config: config}

	if config.LocalPort == 0 {
		return runtime, nil
	}

	udpMux, packetConns, err := newSystemUDPMux(ctx, config)
	if err != nil {
		return nil, err
	}
	runtime.packetConns = packetConns
	runtime.udpMux = udpMux

	return runtime, nil
}

func (l *systemListenerRuntime) Tag() string {
	return l.config.Tag
}

func (l *systemListenerRuntime) RequestPortBlossom() bool {
	return l.config.GetRequestPortBlossom()
}

func (l *systemListenerRuntime) AcceptPortBlossom() bool {
	return l.config.GetAcceptPortBlossom()
}

func (l *systemListenerRuntime) PortBlossomDuration() time.Duration {
	if l.config == nil {
		return defaultPortBlossomDuration
	}
	return configuredPortBlossomDuration(l.config.GetPortBlossomDurationSec())
}

func (l *systemListenerRuntime) RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround() bool {
	return false
}

func (l *systemListenerRuntime) NewPeerAPI() (*pionwebrtc.API, pionwebrtc.Configuration, error) {
	settingEngine := pionwebrtc.SettingEngine{}
	settingEngine.SetLite(true)
	settingEngine.SetICETimeouts(defaultDisconnectedTimeout, defaultFailedTimeout, defaultKeepAliveInterval)
	settingEngine.SetNetworkTypes(systemListenerNetworks(l.config))
	settingEngine.SetICEMulticastDNSMode(pionice.MulticastDNSModeDisabled)

	if l.udpMux != nil {
		settingEngine.SetICEUDPMux(l.udpMux)
	}

	rules := systemListenerRewriteRules(l.config)
	if len(rules) > 0 {
		if err := settingEngine.SetICEAddressRewriteRules(rules...); err != nil {
			return nil, pionwebrtc.Configuration{}, newError("failed to configure ICE address rewrite rules for system listener ", l.config.Tag).Base(err)
		}
	}

	api := pionwebrtc.NewAPI(pionwebrtc.WithSettingEngine(settingEngine))

	cfg := pionwebrtc.Configuration{}
	if len(l.config.Ip) > 0 && l.config.LocalPort != 0 {
		cfg.ICEServers = nil
	}

	return api, cfg, nil
}

func (l *systemListenerRuntime) AddressString() string {
	if l.config == nil {
		return ""
	}

	host := v2net.IP(l.config.Ip).String()
	if host == "" || host == "<nil>" {
		host = "0.0.0.0"
	}
	return host + ":" + strconv.Itoa(int(l.config.LocalPort))
}

func (l *systemListenerRuntime) BlastPorts(ip net.IP) error {
	return blossomUDPPorts(l.packetConns, ip)
}

func (l *systemListenerRuntime) Close() error {
	if l.udpMux != nil {
		_ = l.udpMux.Close()
	}
	var err error
	for _, packetConn := range l.packetConns {
		if packetConn == nil {
			continue
		}
		if closeErr := packetConn.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

func activeListenerNetworks(config *LocalWebRTCListener) []pionwebrtc.NetworkType {
	useIPv4 := config.GetUseIpv4()
	useIPv6 := config.GetUseIpv6()
	if !useIPv4 && !useIPv6 {
		return []pionwebrtc.NetworkType{
			pionwebrtc.NetworkTypeUDP4,
			pionwebrtc.NetworkTypeUDP6,
		}
	}

	networks := make([]pionwebrtc.NetworkType, 0, 2)
	if useIPv4 {
		networks = append(networks, pionwebrtc.NetworkTypeUDP4)
	}
	if useIPv6 {
		networks = append(networks, pionwebrtc.NetworkTypeUDP6)
	}
	return networks
}

func systemListenerNetworks(config *LocalWebRTCSystemListener) []pionwebrtc.NetworkType {
	networks := make([]pionwebrtc.NetworkType, 0, 2)
	if ip := net.IP(config.GetIp()); len(ip) > 0 {
		networks = append(networks, pionwebrtc.NetworkTypeUDP4)
	}
	if ip := net.IP(config.GetIpv6()); len(ip) > 0 {
		networks = append(networks, pionwebrtc.NetworkTypeUDP6)
	}
	if len(networks) == 0 {
		return []pionwebrtc.NetworkType{
			pionwebrtc.NetworkTypeUDP4,
			pionwebrtc.NetworkTypeUDP6,
		}
	}
	return networks
}

func systemListenerRewriteRules(config *LocalWebRTCSystemListener) []pionwebrtc.ICEAddressRewriteRule {
	rules := make([]pionwebrtc.ICEAddressRewriteRule, 0, 2)
	if ip := net.IP(config.GetIp()); len(ip) > 0 && !ip.IsUnspecified() {
		rules = append(rules, pionwebrtc.ICEAddressRewriteRule{
			External:        []string{ip.String()},
			AsCandidateType: pionwebrtc.ICECandidateTypeHost,
			Mode:            pionwebrtc.ICEAddressRewriteReplace,
			Networks:        []pionwebrtc.NetworkType{pionwebrtc.NetworkTypeUDP4},
		})
	}
	if ip := net.IP(config.GetIpv6()); len(ip) > 0 && !ip.IsUnspecified() {
		rules = append(rules, pionwebrtc.ICEAddressRewriteRule{
			External:        []string{ip.String()},
			AsCandidateType: pionwebrtc.ICECandidateTypeHost,
			Mode:            pionwebrtc.ICEAddressRewriteReplace,
			Networks:        []pionwebrtc.NetworkType{pionwebrtc.NetworkTypeUDP6},
		})
	}
	return rules
}

func configuredPortBlossomDuration(seconds uint32) time.Duration {
	if seconds == 0 {
		return defaultPortBlossomDuration
	}
	return time.Duration(seconds) * time.Second
}

func newSystemUDPMux(ctx context.Context, config *LocalWebRTCSystemListener) (pionice.UDPMux, []v2net.PacketConn, error) {
	logger := logging.NewDefaultLoggerFactory().NewLogger("app/webrtc")
	muxes := make([]pionice.UDPMux, 0, 2)
	packetConns := make([]v2net.PacketConn, 0, 2)

	if ip := net.IP(config.GetIp()); len(ip) > 0 {
		packetConn, err := internet.ListenSystemPacket(ctx, &v2net.UDPAddr{
			IP:   v2net.AnyIP.IP(),
			Port: int(config.LocalPort),
		}, nil)
		if err != nil {
			closePacketConns(packetConns)
			return nil, nil, newError("failed to bind IPv4 socket for system WebRTC listener ", config.Tag).Base(err)
		}
		packetConns = append(packetConns, packetConn)
		muxes = append(muxes, pionwebrtc.NewICEUDPMux(logger, packetConn))
	}

	if ip := net.IP(config.GetIpv6()); len(ip) > 0 {
		packetConn, err := internet.ListenSystemPacket(ctx, &v2net.UDPAddr{
			IP:   v2net.AnyIPv6.IP(),
			Port: int(config.LocalPort),
			Zone: "",
		}, nil)
		if err != nil {
			closePacketConns(packetConns)
			return nil, nil, newError("failed to bind IPv6 socket for system WebRTC listener ", config.Tag).Base(err)
		}
		packetConns = append(packetConns, packetConn)
		muxes = append(muxes, pionwebrtc.NewICEUDPMux(logger, packetConn))
	}

	if len(muxes) == 0 {
		listenIP := v2net.AnyIP.IP()
		packetConn, err := internet.ListenSystemPacket(ctx, &v2net.UDPAddr{
			IP:   listenIP,
			Port: int(config.LocalPort),
		}, nil)
		if err != nil {
			return nil, nil, newError("failed to listen for system WebRTC listener ", config.Tag).Base(err)
		}
		packetConns = append(packetConns, packetConn)
		return pionwebrtc.NewICEUDPMux(logger, packetConn), packetConns, nil
	}

	if len(muxes) == 1 {
		return muxes[0], packetConns, nil
	}

	return pionice.NewMultiUDPMuxDefault(muxes...), packetConns, nil
}

func closePacketConns(packetConns []v2net.PacketConn) {
	for _, packetConn := range packetConns {
		if packetConn != nil {
			_ = packetConn.Close()
		}
	}
}
