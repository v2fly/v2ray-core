package webrtc

// This component is primarily machine generated

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	cbuf "github.com/v2fly/v2ray-core/v5/common/buf"
	v2errors "github.com/v2fly/v2ray-core/v5/common/errors"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/session"
	featuredns "github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

const (
	connectAttemptTimeout                   = 45 * time.Second
	initialCandidateWait                    = 100 * time.Millisecond
	signalPollInterval                      = time.Second
	remoteCandidateGatheringWorkaroundDelay = 500 * time.Millisecond
	reconnectBackoffMin                     = time.Second
	reconnectBackoffMax                     = 15 * time.Second
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		w := new(WebRTC)
		if err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher, outboundManager outbound.Manager, dnsClient featuredns.Client) error {
			return w.Init(ctx, config.(*Config), dispatcher, outboundManager, dnsClient)
		}); err != nil {
			return nil, err
		}
		return w, nil
	}))
}

type WebRTC struct {
	ctx            context.Context
	config         *Config
	dispatcher     routing.Dispatcher
	outboundManger outbound.Manager
	dnsClient      featuredns.Client

	cancel context.CancelFunc

	activeListeners map[string]*activeListenerRuntime
	systemListeners map[string]*systemListenerRuntime
	acceptors       []*acceptorRuntime
	remotes         []*remoteConnectionRuntime
}

func (w *WebRTC) Init(ctx context.Context, config *Config, dispatcher routing.Dispatcher, outboundManager outbound.Manager, dnsClient featuredns.Client) error {
	w.ctx = ctx
	w.config = config
	w.dispatcher = dispatcher
	w.outboundManger = outboundManager
	w.dnsClient = dnsClient
	return nil
}

func (w *WebRTC) Type() interface{} {
	return (*WebRTC)(nil)
}

func (w *WebRTC) Start() error {
	if w.config == nil {
		return newError("nil webrtc config")
	}

	runCtx, cancel := context.WithCancel(w.ctx)
	w.cancel = cancel

	activeListeners := make(map[string]*activeListenerRuntime, len(w.config.Listener))
	for _, cfg := range w.config.Listener {
		if cfg == nil {
			continue
		}
		if cfg.Tag == "" {
			cancel()
			return newError("active listener tag cannot be empty")
		}
		if _, found := activeListeners[cfg.Tag]; found {
			cancel()
			return newError("duplicate active listener tag ", cfg.Tag)
		}
		activeListeners[cfg.Tag] = &activeListenerRuntime{
			ctx:        runCtx,
			dnsClient:  w.dnsClient,
			dispatcher: w.dispatcher,
			config:     cfg,
		}
	}

	systemListeners := make(map[string]*systemListenerRuntime, len(w.config.SystemListener))
	for _, cfg := range w.config.SystemListener {
		if cfg == nil {
			continue
		}
		if cfg.Tag == "" {
			cancel()
			return newError("system listener tag cannot be empty")
		}
		if _, found := systemListeners[cfg.Tag]; found {
			cancel()
			return newError("duplicate system listener tag ", cfg.Tag)
		}
		runtime, err := newSystemListenerRuntime(runCtx, cfg)
		if err != nil {
			cancel()
			return err
		}
		systemListeners[cfg.Tag] = runtime
	}

	remoteConfigs := make(map[string]*ClientConfig, len(w.config.Remotes))
	for _, cfg := range w.config.Remotes {
		if cfg == nil {
			continue
		}
		if cfg.Tag == "" {
			cancel()
			return newError("remote tag cannot be empty")
		}
		if _, found := remoteConfigs[cfg.Tag]; found {
			cancel()
			return newError("duplicate remote tag ", cfg.Tag)
		}
		clientConfig, err := decodeClientConfig(cfg.ClientConfig)
		if err != nil {
			cancel()
			return err
		}
		remoteConfigs[cfg.Tag] = clientConfig
	}

	w.activeListeners = activeListeners
	w.systemListeners = systemListeners

	acceptors := make([]*acceptorRuntime, 0, len(w.config.Acceptors))
	for _, cfg := range w.config.Acceptors {
		if cfg == nil {
			continue
		}
		runtime, err := w.newAcceptorRuntime(runCtx, cfg)
		if err != nil {
			cancel()
			w.closeListeners()
			return err
		}
		if err := runtime.Start(); err != nil {
			cancel()
			_ = runtime.Close()
			w.closeAcceptors(acceptors)
			w.closeListeners()
			return err
		}
		acceptors = append(acceptors, runtime)
	}

	remotes := make([]*remoteConnectionRuntime, 0, len(w.config.Connection))
	for _, cfg := range w.config.Connection {
		if cfg == nil {
			continue
		}
		remoteCfg, found := remoteConfigs[cfg.RemoteTag]
		if !found {
			cancel()
			w.closeAcceptors(acceptors)
			w.closeListeners()
			return newError("unknown remote tag ", cfg.RemoteTag)
		}

		listener, err := w.listenerByTag(cfg.LocalListenerTag)
		if err != nil {
			cancel()
			w.closeAcceptors(acceptors)
			w.closeListeners()
			return err
		}

		runtime, err := newRemoteConnectionRuntime(runCtx, w.outboundManger, cfg, remoteCfg, listener)
		if err != nil {
			cancel()
			w.closeAcceptors(acceptors)
			w.closeListeners()
			return err
		}
		if err := runtime.Start(); err != nil {
			cancel()
			_ = runtime.Close()
			w.closeRemotes(remotes)
			w.closeAcceptors(acceptors)
			w.closeListeners()
			return err
		}
		remotes = append(remotes, runtime)
	}

	w.acceptors = acceptors
	w.remotes = remotes

	return nil
}

func (w *WebRTC) Close() error {
	if w.cancel != nil {
		w.cancel()
	}

	var errs []error
	errs = append(errs, w.closeRemotes(w.remotes)...)
	errs = append(errs, w.closeAcceptors(w.acceptors)...)
	errs = append(errs, w.closeListeners()...)
	return v2errors.Combine(errs...)
}

func (w *WebRTC) listenerByTag(tag string) (listenerRuntime, error) {
	if listener, found := w.activeListeners[tag]; found {
		return listener, nil
	}
	if listener, found := w.systemListeners[tag]; found {
		return listener, nil
	}
	return nil, newError("unknown local listener tag ", tag)
}

func (w *WebRTC) acceptorListener(config *Acceptor) (listenerRuntime, error) {
	if config == nil {
		return nil, newError("nil acceptor config")
	}

	if config.AcceptOnTag != "" {
		listener, err := w.listenerByTag(config.AcceptOnTag)
		if err == nil {
			return listener, nil
		}
		return nil, newError("unknown accept_on_tag ", config.AcceptOnTag, " for acceptor ", config.Tag)
	}

	if listener, found := w.systemListeners[config.Tag]; found {
		return listener, nil
	}
	if len(w.systemListeners) == 1 {
		for _, listener := range w.systemListeners {
			return listener, nil
		}
	}
	return nil, newError("unable to resolve system listener for acceptor ", config.Tag, "; set accept_on_tag explicitly")
}

func (w *WebRTC) newAcceptorRuntime(ctx context.Context, config *Acceptor) (*acceptorRuntime, error) {
	if config.Tag == "" {
		return nil, newError("acceptor tag cannot be empty")
	}
	serverConfig, err := decodeServerConfig(config.ServerConfig)
	if err != nil {
		return nil, err
	}
	if len(serverConfig.ServerIdentity) == 0 {
		return nil, newError("acceptor ", config.Tag, " is missing server_identity")
	}

	listener, err := w.acceptorListener(config)
	if err != nil {
		return nil, err
	}

	signaler, err := newSignaler(ctx, serverConfig.RoundTripperClient, serverConfig.SecurityConfig, serverConfig.Dest, serverConfig.OutboundTag)
	if err != nil {
		return nil, err
	}

	forwards := make(map[string]*UDPPortForwarderAcceptor, len(config.PortForwarderAccepts))
	for _, forward := range config.PortForwarderAccepts {
		if forward == nil {
			continue
		}
		if forward.Tag == "" {
			return nil, newError("acceptor port forward tag cannot be empty")
		}
		if _, found := forwards[forward.Tag]; found {
			return nil, newError("duplicate acceptor forward tag ", forward.Tag)
		}
		forwards[forward.Tag] = forward
	}

	childCtx, cancel := context.WithCancel(ctx)
	return &acceptorRuntime{
		ctx:            childCtx,
		cancel:         cancel,
		tag:            config.Tag,
		signaler:       signaler,
		serverIdentity: append([]byte(nil), serverConfig.ServerIdentity...),
		listener:       listener,
		forwards:       forwards,
		sessions:       make(map[string]*acceptorSession),
		responses:      make(chan queuedResponse, 32),
	}, nil
}

func (w *WebRTC) closeListeners() []error {
	var errs []error
	for _, listener := range w.systemListeners {
		errs = append(errs, listener.Close())
	}
	for _, listener := range w.activeListeners {
		errs = append(errs, listener.Close())
	}
	return errs
}

func (w *WebRTC) closeAcceptors(acceptors []*acceptorRuntime) []error {
	errs := make([]error, 0, len(acceptors))
	for _, acceptor := range acceptors {
		errs = append(errs, acceptor.Close())
	}
	return errs
}

func (w *WebRTC) closeRemotes(remotes []*remoteConnectionRuntime) []error {
	errs := make([]error, 0, len(remotes))
	for _, remote := range remotes {
		errs = append(errs, remote.Close())
	}
	return errs
}

type acceptorRuntime struct {
	ctx    context.Context
	cancel context.CancelFunc

	tag            string
	signaler       *signaler
	serverIdentity []byte
	listener       listenerRuntime
	forwards       map[string]*UDPPortForwarderAcceptor

	responses chan queuedResponse

	mu       sync.Mutex
	sessions map[string]*acceptorSession
}

type queuedResponse struct {
	replyTag  []byte
	sessionID []byte
	payload   []byte
}

func (a *acceptorRuntime) Start() error {
	go a.pollLoop()
	go a.responseLoop()
	return nil
}

func (a *acceptorRuntime) Close() error {
	a.cancel()

	a.mu.Lock()
	sessions := make([]*acceptorSession, 0, len(a.sessions))
	for _, session := range a.sessions {
		sessions = append(sessions, session)
	}
	a.mu.Unlock()

	var errs []error
	for _, session := range sessions {
		errs = append(errs, session.Close())
	}
	return v2errors.Combine(errs...)
}

func (a *acceptorRuntime) pollLoop() {
	backoff := reconnectBackoffMin
	for a.ctx.Err() == nil {
		newError(
			"acceptor signal poll send acceptor=", a.tag,
			" server_identity=", compactSignalToken(a.serverIdentity),
		).AtDebug().WriteToLog()
		respData, err := a.signaler.RoundTrip(a.ctx, append([]byte(nil), a.serverIdentity...), nil)
		if err != nil {
			newError("acceptor poll failed for ", a.tag).Base(err).AtWarning().WriteToLog()
			select {
			case <-a.ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = nextBackoff(backoff)
			continue
		}
		backoff = reconnectBackoffMin
		if len(respData) == 0 {
			newError(
				"acceptor signal poll recv empty acceptor=", a.tag,
				" server_identity=", compactSignalToken(a.serverIdentity),
			).AtDebug().WriteToLog()
			continue
		}

		req := new(ConnectionRequest)
		if err := proto.Unmarshal(respData, req); err != nil {
			newError("failed to decode signaling request").Base(err).AtWarning().WriteToLog()
			continue
		}
		newError(
			"acceptor signal poll recv acceptor=", a.tag,
			" reply_tag=", compactSignalToken(req.ReplyAddressTag),
			" session_id=", compactSignalToken(req.ConnectionSessionId),
			" has_sdp=", len(req.SessionDescription) > 0,
			" candidates=", len(req.Candidates),
			" request_port_blossom=", req.GetRequestPortBlossom(),
		).AtDebug().WriteToLog()

		if err := a.handleRequest(req); err != nil {
			newError("failed to handle signaling request").Base(err).AtWarning().WriteToLog()
		}
	}
}

func (a *acceptorRuntime) responseLoop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case item := <-a.responses:
			key := sessionKey(item.replyTag, item.sessionID)
			if !a.isCurrentSession(key, item.sessionID) {
				continue
			}

			routingTag := make([]byte, 0, len(a.serverIdentity)+len(item.replyTag))
			routingTag = append(routingTag, a.serverIdentity...)
			routingTag = append(routingTag, item.replyTag...)

			for a.ctx.Err() == nil {
				newError(
					"acceptor signal response send acceptor=", a.tag,
					" reply_tag=", compactSignalToken(item.replyTag),
					" session_id=", compactSignalToken(item.sessionID),
					" routing_tag=", compactSignalToken(routingTag),
					" payload_bytes=", len(item.payload),
				).AtDebug().WriteToLog()
				if _, err := a.signaler.RoundTrip(a.ctx, routingTag, item.payload); err != nil {
					newError("failed to send signaling response").Base(err).AtWarning().WriteToLog()
					select {
					case <-a.ctx.Done():
						return
					case <-time.After(reconnectBackoffMin):
					}
					continue
				}
				break
			}
		}
	}
}

func (a *acceptorRuntime) handleRequest(req *ConnectionRequest) error {
	if len(req.ReplyAddressTag) == 0 {
		return newError("signaling request missing reply_address_tag")
	}
	if len(req.ConnectionSessionId) == 0 {
		return newError("signaling request missing connection_session_id")
	}

	key := sessionKey(req.ReplyAddressTag, req.ConnectionSessionId)
	session, created, err := a.getOrCreateSession(key, req)
	if err != nil {
		return err
	}
	session.setRequestPortBlossom(req.GetRequestPortBlossom())

	if created && len(req.SessionDescription) == 0 {
		return newError("new acceptor session missing offer SDP")
	}

	if created {
		answer, err := session.acceptOffer(req.SessionDescription)
		if err != nil {
			_ = session.Close()
			return err
		}
		session.setLocalDescription(answer)
		if err := session.addCandidates(req.Candidates); err != nil {
			_ = session.Close()
			return err
		}
		return session.queueCurrentResponse()
	}

	if len(req.Candidates) > 0 {
		if err := session.addCandidates(req.Candidates); err != nil {
			return err
		}
	}

	return session.queueCurrentResponse()
}

func (a *acceptorRuntime) getOrCreateSession(key string, req *ConnectionRequest) (*acceptorSession, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if session, found := a.sessions[key]; found {
		return session, false, nil
	}

	api, cfg, err := a.listener.NewPeerAPI()
	if err != nil {
		return nil, false, err
	}

	var (
		transportSession *acceptorTransportSession
		session          *acceptorSession
	)
	peer, err := newAnswererPeer(api, cfg, "acceptor:"+a.tag+"/"+compactSignalToken(req.ConnectionSessionId), func(candidate []byte) {
		if session == nil {
			return
		}
		session.addLocalCandidate(candidate)
	}, func() {
		if session == nil {
			return
		}
		session.setPeerConnected()
	}, func() {
		if session == nil {
			return
		}
		session.setLocalGatheringDone()
	}, func(channel *peerDataChannel) {
		if transportSession != nil {
			transportSession.handleChannel(channel)
		}
	})
	if err != nil {
		return nil, false, err
	}

	session = &acceptorSession{
		owner:              a,
		key:                key,
		replyTag:           append([]byte(nil), req.ReplyAddressTag...),
		sessionID:          append([]byte(nil), req.ConnectionSessionId...),
		peer:               peer,
		localCandidateSet:  make(map[string]struct{}),
		remoteCandidateSet: make(map[string]struct{}),
		portBlossomIPs:     make(map[string]v2net.IP),
		portBlossomStop:    make(chan struct{}),
	}
	transportSession = &acceptorTransportSession{
		ctx:     a.ctx,
		owner:   a,
		bridges: make(map[*peerDataChannel]*udpBridge),
	}
	session.transport = transportSession

	a.sessions[key] = session
	go func() {
		<-peer.Done()
		_ = session.Close()
	}()

	return session, true, nil
}

func (a *acceptorRuntime) removeSession(key string) {
	a.mu.Lock()
	delete(a.sessions, key)
	a.mu.Unlock()
}

func (a *acceptorRuntime) isCurrentSession(key string, sessionID []byte) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, found := a.sessions[key]
	if !found {
		return false
	}
	return string(session.sessionID) == string(sessionID)
}

func (a *acceptorRuntime) enqueueResponse(replyTag, sessionID []byte, response *ConnectionResponse) error {
	payload, err := proto.Marshal(response)
	if err != nil {
		return newError("failed to encode signaling response").Base(err)
	}

	select {
	case <-a.ctx.Done():
		return a.ctx.Err()
	case a.responses <- queuedResponse{
		replyTag:  append([]byte(nil), replyTag...),
		sessionID: append([]byte(nil), sessionID...),
		payload:   payload,
	}:
		return nil
	}
}

type acceptorSession struct {
	owner     *acceptorRuntime
	key       string
	replyTag  []byte
	sessionID []byte
	peer      *peerSession
	transport *acceptorTransportSession
	mu        sync.Mutex

	localDescription   []byte
	localCandidates    [][]byte
	localCandidateSet  map[string]struct{}
	remoteCandidateSet map[string]struct{}
	portBlossomIPs     map[string]v2net.IP
	localGatheringDone bool
	peerConnected      bool
	requestPortBlossom bool
	portBlossomRunning bool
	portBlossomStop    chan struct{}

	once                sync.Once
	portBlossomStopOnce sync.Once
}

func (s *acceptorSession) acceptOffer(offer []byte) ([]byte, error) {
	return s.peer.AcceptOffer(offer)
}

func (s *acceptorSession) setLocalDescription(answer []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.localDescription = append([]byte(nil), answer...)
}

func (s *acceptorSession) addLocalCandidate(candidate []byte) {
	if len(candidate) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(candidate)
	if _, found := s.localCandidateSet[key]; found {
		return
	}
	s.localCandidateSet[key] = struct{}{}
	s.localCandidates = append(s.localCandidates, append([]byte(nil), candidate...))
}

func (s *acceptorSession) setLocalGatheringDone() {
	s.mu.Lock()
	s.localGatheringDone = true
	s.mu.Unlock()
	newError(
		"acceptor signaling gathering complete acceptor=", s.owner.tag,
		" session_id=", compactSignalToken(s.sessionID),
	).AtDebug().WriteToLog()
}

func (s *acceptorSession) setPeerConnected() {
	s.mu.Lock()
	s.peerConnected = true
	s.mu.Unlock()
	s.stopPortBlossom()
	newError(
		"acceptor signaling peer connected acceptor=", s.owner.tag,
		" session_id=", compactSignalToken(s.sessionID),
	).AtDebug().WriteToLog()
}

func (s *acceptorSession) setRequestPortBlossom(enabled bool) {
	if !enabled {
		return
	}
	s.mu.Lock()
	s.requestPortBlossom = true
	s.mu.Unlock()
}

func (s *acceptorSession) addCandidates(candidates [][]byte) error {
	for _, candidate := range candidates {
		startPortBlossom := false
		key := string(candidate)
		s.mu.Lock()
		if _, found := s.remoteCandidateSet[key]; found {
			s.mu.Unlock()
			continue
		}
		s.remoteCandidateSet[key] = struct{}{}
		if s.requestPortBlossom && s.owner.listener.AcceptPortBlossom() && !s.peerConnected {
			ip, err := candidateBlossomIP(candidate)
			if err != nil {
				newError("failed to derive candidate IP for port blossom").Base(err).AtDebug().WriteToLog()
			} else if _, found := s.portBlossomIPs[ip.String()]; !found {
				s.portBlossomIPs[ip.String()] = append(v2net.IP(nil), ip...)
			}
			if !s.portBlossomRunning && len(s.portBlossomIPs) > 0 {
				s.portBlossomRunning = true
				startPortBlossom = true
			}
		}
		s.mu.Unlock()
		if startPortBlossom {
			go s.portBlossomLoop()
		}
		if err := s.peer.AddICECandidate(candidate); err != nil {
			return err
		}
	}
	return nil
}

func (s *acceptorSession) portBlossomLoop() {
	blastTimeout := s.owner.listener.PortBlossomDuration()
	timeout := time.NewTimer(blastTimeout)
	defer timeout.Stop()

	s.blastKnownCandidateIPs()

	ticker := time.NewTicker(portBlossomRepeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.portBlossomStop:
			return
		case <-timeout.C:
			newError(
				"acceptor port blossom timeout session=", compactSignalToken(s.sessionID),
				" duration=", blastTimeout,
			).AtDebug().WriteToLog()
			s.stopPortBlossom()
			return
		case <-ticker.C:
			s.blastKnownCandidateIPs()
		}
	}
}

func (s *acceptorSession) blastKnownCandidateIPs() {
	targets := s.snapshotPortBlossomIPs()
	for _, ip := range targets {
		newError(
			"acceptor port blossom session=", compactSignalToken(s.sessionID),
			" ip=", ip.String(),
		).AtDebug().WriteToLog()
		if err := s.owner.listener.BlastPorts(ip); err != nil {
			newError("acceptor port blossom retry failed for candidate IP ", ip.String()).Base(err).AtDebug().WriteToLog()
		}
	}
}

func (s *acceptorSession) snapshotPortBlossomIPs() []v2net.IP {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.peerConnected || len(s.portBlossomIPs) == 0 {
		return nil
	}

	ips := make([]v2net.IP, 0, len(s.portBlossomIPs))
	for _, ip := range s.portBlossomIPs {
		ips = append(ips, append(v2net.IP(nil), ip...))
	}
	return ips
}

func (s *acceptorSession) stopPortBlossom() {
	s.portBlossomStopOnce.Do(func() {
		close(s.portBlossomStop)
	})
}

func (s *acceptorSession) queueResponse(resp *ConnectionResponse) error {
	return s.owner.enqueueResponse(s.replyTag, s.sessionID, resp)
}

func (s *acceptorSession) queueCurrentResponse() error {
	s.mu.Lock()
	resp := &ConnectionResponse{
		ReplyAddressTag:    append([]byte(nil), s.replyTag...),
		SessionDescription: append([]byte(nil), s.localDescription...),
		Candidates:         cloneByteSlices(s.localCandidates),
		StopPolling:        s.localGatheringDone || s.peerConnected,
	}
	s.mu.Unlock()
	return s.queueResponse(resp)
}

func (s *acceptorSession) Close() error {
	var err error
	s.once.Do(func() {
		s.stopPortBlossom()
		s.owner.removeSession(s.key)
		if s.transport != nil {
			err = s.transport.Close()
		}
		if s.peer != nil {
			_ = s.peer.Close()
		}
	})
	return err
}

type acceptorTransportSession struct {
	ctx   context.Context
	owner *acceptorRuntime

	mu      sync.Mutex
	bridges map[*peerDataChannel]*udpBridge
}

func (s *acceptorTransportSession) handleChannel(channel *peerDataChannel) {
	if channel == nil {
		return
	}

	forwardCfg, found := s.owner.forwards[channel.Label()]
	if !found {
		newError("unknown acceptor forward tag ", channel.Label()).AtWarning().WriteToLog()
		_ = channel.Close()
		return
	}

	bridge, err := newUDPBridge(s.ctx, channel, forwardCfg)
	if err != nil {
		newError("failed to open acceptor data channel bridge").Base(err).AtWarning().WriteToLog()
		_ = channel.Close()
		return
	}

	channel.SetMessageHandler(func(payload []byte) {
		if err := bridge.Write(payload); err != nil {
			newError("failed to write acceptor data channel packet").Base(err).AtWarning().WriteToLog()
			s.closeChannel(channel, true)
		}
	})

	s.mu.Lock()
	s.bridges[channel] = bridge
	s.mu.Unlock()

	go func() {
		<-channel.Done()
		s.closeChannel(channel, false)
	}()
}

func (s *acceptorTransportSession) closeChannel(channel *peerDataChannel, notifyPeer bool) {
	s.mu.Lock()
	bridge := s.bridges[channel]
	delete(s.bridges, channel)
	s.mu.Unlock()

	if bridge != nil {
		_ = bridge.Close(notifyPeer)
		return
	}

	if notifyPeer && channel != nil {
		_ = channel.Close()
	}
}

func (s *acceptorTransportSession) Close() error {
	s.mu.Lock()
	bridges := make([]*udpBridge, 0, len(s.bridges))
	for _, bridge := range s.bridges {
		bridges = append(bridges, bridge)
	}
	s.bridges = map[*peerDataChannel]*udpBridge{}
	s.mu.Unlock()

	var errs []error
	for _, bridge := range bridges {
		errs = append(errs, bridge.Close(false))
	}
	return v2errors.Combine(errs...)
}

type udpBridge struct {
	channel *peerDataChannel
	config  *UDPPortForwarderAcceptor
	conn    v2net.Conn
	writer  cbuf.Writer

	once sync.Once
}

func newUDPBridge(ctx context.Context, channel *peerDataChannel, config *UDPPortForwarderAcceptor) (*udpBridge, error) {
	dest := v2net.Destination{
		Address: v2net.IPAddress(config.Ip),
		Port:    v2net.Port(config.Port),
		Network: v2net.Network_UDP,
	}

	var (
		conn v2net.Conn
		err  error
	)
	if config.ConnectVia != "" {
		conn, err = internet.DialTaggedOutbound(ctx, dest, config.ConnectVia)
	} else {
		conn, err = internet.DialSystem(ctx, dest, nil)
	}
	if err != nil {
		return nil, newError("failed to dial UDP forward target").Base(err)
	}

	bridge := &udpBridge{
		channel: channel,
		config:  config,
		conn:    conn,
		writer:  &cbuf.SequentialWriter{Writer: conn},
	}
	go bridge.readLoop()
	return bridge, nil
}

func (b *udpBridge) Write(payload []byte) error {
	packet := append([]byte(nil), payload...)
	return b.writer.WriteMultiBuffer(cbuf.MultiBuffer{cbuf.FromBytes(packet)})
}

func (b *udpBridge) readLoop() {
	reader := cbuf.NewPacketReader(b.conn)
	for {
		mb, err := reader.ReadMultiBuffer()
		if !mb.IsEmpty() {
			for _, buffer := range mb {
				payload := append([]byte(nil), buffer.Bytes()...)
				buffer.Release()
				if sendErr := b.channel.Send(payload); sendErr != nil {
					_ = b.Close(true)
					return
				}
			}
		}
		if err != nil {
			_ = b.Close(true)
			return
		}
	}
}

func (b *udpBridge) Close(notifyPeer bool) error {
	var err error
	b.once.Do(func() {
		if notifyPeer && b.channel != nil {
			_ = b.channel.Close()
		}
		err = b.conn.Close()
	})
	return err
}

type remoteConnectionRuntime struct {
	ctx    context.Context
	cancel context.CancelFunc

	name           string
	outboundManger outbound.Manager
	signaler       *signaler
	serverIdentity []byte
	listener       listenerRuntime
	forwards       []*UDPPortForwarderRemote

	outbounds []*remoteForwardOutbound

	mu         sync.Mutex
	current    *remoteTransportSession
	generation uint64
	notify     chan struct{}
}

func newRemoteConnectionRuntime(ctx context.Context, outboundManager outbound.Manager, config *RemoteConnections, clientConfig *ClientConfig, listener listenerRuntime) (*remoteConnectionRuntime, error) {
	signaler, err := newSignaler(ctx, clientConfig.RoundTripperClient, clientConfig.SecurityConfig, clientConfig.Dest, clientConfig.OutboundTag)
	if err != nil {
		return nil, err
	}
	if len(clientConfig.ServerIdentity) == 0 {
		return nil, newError("remote ", config.RemoteTag, " is missing server_identity")
	}

	childCtx, cancel := context.WithCancel(ctx)
	return &remoteConnectionRuntime{
		ctx:            childCtx,
		cancel:         cancel,
		name:           config.RemoteTag + "@" + config.LocalListenerTag,
		outboundManger: outboundManager,
		signaler:       signaler,
		serverIdentity: append([]byte(nil), clientConfig.ServerIdentity...),
		listener:       listener,
		forwards:       append([]*UDPPortForwarderRemote(nil), config.PortForward...),
		notify:         make(chan struct{}),
	}, nil
}

func (r *remoteConnectionRuntime) Start() error {
	registeredTags := make(map[string]struct{}, len(r.forwards))
	for _, forward := range r.forwards {
		if forward == nil {
			continue
		}
		if forward.AcceptConnectOn == "" {
			return newError("remote forward accept_connect_on cannot be empty")
		}
		if _, found := registeredTags[forward.AcceptConnectOn]; found {
			return newError("duplicate remote forward accept_connect_on ", forward.AcceptConnectOn)
		}
		registeredTags[forward.AcceptConnectOn] = struct{}{}
		handler := &remoteForwardOutbound{
			tag:        forward.AcceptConnectOn,
			runtime:    r,
			forwardTag: forward.Tag,
		}
		if err := handler.Start(); err != nil {
			return err
		}
		_ = r.outboundManger.RemoveHandler(context.Background(), forward.AcceptConnectOn)
		if err := r.outboundManger.AddHandler(context.Background(), handler); err != nil {
			return newError("failed to add outbound handler ", forward.AcceptConnectOn).Base(err)
		}
		r.outbounds = append(r.outbounds, handler)
	}

	go r.supervisor()
	return nil
}

func (r *remoteConnectionRuntime) Close() error {
	r.cancel()

	r.mu.Lock()
	current := r.current
	r.current = nil
	close(r.notify)
	r.notify = make(chan struct{})
	r.mu.Unlock()

	var errs []error
	if current != nil {
		errs = append(errs, current.Close())
	}

	for _, handler := range r.outbounds {
		if handler == nil {
			continue
		}
		errs = append(errs, r.outboundManger.RemoveHandler(context.Background(), handler.tag))
		errs = append(errs, handler.Close())
	}
	return v2errors.Combine(errs...)
}

func (r *remoteConnectionRuntime) supervisor() {
	backoff := reconnectBackoffMin
	for r.ctx.Err() == nil {
		attemptCtx, cancel := context.WithTimeout(r.ctx, connectAttemptTimeout)
		session, err := r.connectOnce(attemptCtx)
		cancel()
		if err != nil {
			newError("failed to establish WebRTC session for ", r.name).Base(err).AtWarning().WriteToLog()
			select {
			case <-r.ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = nextBackoff(backoff)
			continue
		}

		backoff = reconnectBackoffMin
		r.setCurrent(session)
		select {
		case <-r.ctx.Done():
			_ = session.Close()
			return
		case <-session.Done():
			newError("webrtc remote session ended for ", r.name, "; reconnecting").AtDebug().WriteToLog()
			r.clearCurrent(session.generation)
			_ = session.Close()
		}
	}
}

func (r *remoteConnectionRuntime) connectOnce(ctx context.Context) (*remoteTransportSession, error) {
	api, cfg, err := r.listener.NewPeerAPI()
	if err != nil {
		return nil, err
	}

	replyTag, err := randomBytes(16)
	if err != nil {
		return nil, err
	}
	sessionID, err := randomBytes(16)
	if err != nil {
		return nil, err
	}

	candidateCh := make(chan []byte, 16)
	gatheringDoneCh := make(chan struct{})
	sessionCtx, sessionCancel := context.WithCancel(r.ctx)
	var gatheringDoneOnce sync.Once

	var transportSession *remoteTransportSession
	peer, err := newOffererPeer(api, cfg, "remote:"+r.name+"/"+compactSignalToken(sessionID), func(candidate []byte) {
		select {
		case <-sessionCtx.Done():
			return
		case candidateCh <- append([]byte(nil), candidate...):
		default:
			newError("dropping ICE candidate due to full channel").AtWarning().WriteToLog()
		}
	}, nil, func() {
		gatheringDoneOnce.Do(func() {
			close(gatheringDoneCh)
		})
	}, nil)
	if err != nil {
		sessionCancel()
		return nil, err
	}

	offer, err := peer.CreateOffer()
	if err != nil {
		sessionCancel()
		_ = peer.Close()
		return nil, err
	}

	transportSession = newRemoteTransportSession(peer, sessionCancel)
	go r.signalLoop(sessionCtx, peer, replyTag, sessionID, offer, candidateCh, gatheringDoneCh)

	if err := peer.WaitReady(ctx); err != nil {
		sessionCancel()
		_ = peer.Close()
		return nil, err
	}

	return transportSession, nil
}

func (r *remoteConnectionRuntime) signalLoop(ctx context.Context, peer *peerSession, replyTag, sessionID, offer []byte, candidateCh <-chan []byte, gatheringDoneCh chan struct{}) {
	var (
		localCandidates         [][]byte
		localCandidateSet       = make(map[string]struct{})
		pendingRemoteCandidates [][]byte
		remoteCandidateSet      = make(map[string]struct{})
		localGatheringDone      bool
	)
	answerApplied := false
	candidateSendGate := newLocalCandidateSendGate(
		r.listener.RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround(),
		remoteCandidateGatheringWorkaroundDelay,
		nil,
	)
	localGatheringDoneSignal := gatheringDoneCh
	routingTag := make([]byte, 0, len(replyTag)+len(r.serverIdentity))
	routingTag = append(routingTag, replyTag...)
	routingTag = append(routingTag, r.serverIdentity...)

	localCandidates = collectInitialCandidates(ctx, candidateCh, &localGatheringDoneSignal, &localGatheringDone, localCandidateSet, localCandidates, initialCandidateWait)

	for ctx.Err() == nil {
		localCandidates = drainLocalCandidates(candidateCh, &localGatheringDoneSignal, &localGatheringDone, localCandidateSet, localCandidates)
		sendableCandidates := localCandidates
		if !candidateSendGate.AllowSend() {
			sendableCandidates = nil
		}
		pending := &ConnectionRequest{
			ReplyAddressTag:     append([]byte(nil), replyTag...),
			ConnectionSessionId: append([]byte(nil), sessionID...),
			SessionDescription:  append([]byte(nil), offer...),
			Candidates:          cloneByteSlices(sendableCandidates),
			RequestPortBlossom:  r.listener.RequestPortBlossom(),
		}
		newError(
			"client signal send remote=", r.name,
			" reply_tag=", compactSignalToken(replyTag),
			" session_id=", compactSignalToken(sessionID),
			" routing_tag=", compactSignalToken(routingTag),
			" has_sdp=", len(pending.SessionDescription) > 0,
			" candidates=", len(pending.Candidates),
			" request_port_blossom=", pending.RequestPortBlossom,
		).AtDebug().WriteToLog()
		payload, err := proto.Marshal(pending)
		if err != nil {
			newError("failed to encode signaling request").Base(err).AtWarning().WriteToLog()
			_ = peer.Close()
			return
		}

		respData, err := r.signaler.RoundTrip(ctx, routingTag, payload)
		if err != nil {
			newError("remote signaling round trip failed").Base(err).AtWarning().WriteToLog()
			if !waitSignalPoll(ctx, candidateSendGate.NextPollDelay(signalPollInterval)) {
				return
			}
			continue
		}

		if len(respData) == 0 {
			newError(
				"client signal recv empty remote=", r.name,
				" reply_tag=", compactSignalToken(replyTag),
				" session_id=", compactSignalToken(sessionID),
			).AtDebug().WriteToLog()
			if !waitSignalPoll(ctx, candidateSendGate.NextPollDelay(signalPollInterval)) {
				return
			}
			continue
		}

		resp := new(ConnectionResponse)
		if err := proto.Unmarshal(respData, resp); err != nil {
			newError("failed to decode signaling response").Base(err).AtWarning().WriteToLog()
			if !waitSignalPoll(ctx, candidateSendGate.NextPollDelay(signalPollInterval)) {
				return
			}
			continue
		}
		newError(
			"client signal recv remote=", r.name,
			" reply_tag=", compactSignalToken(resp.ReplyAddressTag),
			" session_id=", compactSignalToken(sessionID),
			" has_sdp=", len(resp.SessionDescription) > 0,
			" candidates=", len(resp.Candidates),
			" stop_polling=", resp.GetStopPolling(),
		).AtDebug().WriteToLog()

		if len(resp.SessionDescription) > 0 && !answerApplied {
			if err := peer.ApplyAnswer(resp.SessionDescription); err != nil {
				newError("failed to apply remote answer").Base(err).AtWarning().WriteToLog()
				_ = peer.Close()
				return
			}
			answerApplied = true
			for _, candidate := range pendingRemoteCandidates {
				if err := peer.AddICECandidate(candidate); err != nil {
					newError("failed to apply buffered remote ICE candidate").Base(err).AtWarning().WriteToLog()
					_ = peer.Close()
					return
				}
			}
			pendingRemoteCandidates = nil
		}

		for _, candidate := range resp.Candidates {
			key := string(candidate)
			if _, found := remoteCandidateSet[key]; found {
				continue
			}
			remoteCandidateSet[key] = struct{}{}
			if !answerApplied {
				pendingRemoteCandidates = append(pendingRemoteCandidates, append([]byte(nil), candidate...))
				continue
			}
			if err := peer.AddICECandidate(candidate); err != nil {
				newError("failed to apply remote ICE candidate").Base(err).AtWarning().WriteToLog()
				_ = peer.Close()
				return
			}
		}

		localCandidates = drainLocalCandidates(candidateCh, &localGatheringDoneSignal, &localGatheringDone, localCandidateSet, localCandidates)
		if resp.GetStopPolling() && answerApplied {
			if candidateSendGate.StartCountdown() {
				newError(
					"client signal candidate holdback start remote=", r.name,
					" reply_tag=", compactSignalToken(replyTag),
					" session_id=", compactSignalToken(sessionID),
					" release_after=", remoteCandidateGatheringWorkaroundDelay,
				).AtDebug().WriteToLog()
			}
			if !candidateSendGate.AllowSend() {
				if !waitSignalPoll(ctx, candidateSendGate.NextPollDelay(signalPollInterval)) {
					return
				}
				continue
			}
			if len(localCandidates) > len(pending.Candidates) {
				if len(pending.Candidates) == 0 {
					newError(
						"client signal candidate holdback release remote=", r.name,
						" reply_tag=", compactSignalToken(replyTag),
						" session_id=", compactSignalToken(sessionID),
						" candidates=", len(localCandidates),
					).AtDebug().WriteToLog()
				}
				continue
			}
			if !waitForSignalResume(ctx, r.name, replyTag, sessionID, candidateCh, &localGatheringDoneSignal, &localGatheringDone, localCandidateSet, &localCandidates, len(pending.Candidates)) {
				return
			}
			continue
		}

		if !waitSignalPoll(ctx, candidateSendGate.NextPollDelay(signalPollInterval)) {
			return
		}
	}
}

func collectInitialCandidates(
	ctx context.Context,
	candidates <-chan []byte,
	gatheringDoneCh *chan struct{},
	localGatheringDone *bool,
	seen map[string]struct{},
	acc [][]byte,
	wait time.Duration,
) [][]byte {
	if wait <= 0 {
		return drainLocalCandidates(candidates, gatheringDoneCh, localGatheringDone, seen, acc)
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return acc
		case <-*gatheringDoneCh:
			*localGatheringDone = true
			*gatheringDoneCh = nil
			return drainLocalCandidates(candidates, gatheringDoneCh, localGatheringDone, seen, acc)
		case <-timer.C:
			return drainLocalCandidates(candidates, gatheringDoneCh, localGatheringDone, seen, acc)
		case candidate, ok := <-candidates:
			if !ok {
				return acc
			}
			if len(candidate) == 0 {
				continue
			}
			key := string(candidate)
			if _, found := seen[key]; found {
				continue
			}
			seen[key] = struct{}{}
			acc = append(acc, append([]byte(nil), candidate...))
		}
	}
}

func (r *remoteConnectionRuntime) setCurrent(session *remoteTransportSession) {
	r.mu.Lock()
	oldNotify := r.notify
	r.generation++
	session.generation = r.generation
	r.current = session
	r.notify = make(chan struct{})
	r.mu.Unlock()
	close(oldNotify)
}

func (r *remoteConnectionRuntime) clearCurrent(generation uint64) {
	r.mu.Lock()
	if r.current != nil && r.current.generation == generation {
		oldNotify := r.notify
		r.current = nil
		r.notify = make(chan struct{})
		r.mu.Unlock()
		close(oldNotify)
		return
	}
	r.mu.Unlock()
}

func (r *remoteConnectionRuntime) waitForSession(ctx context.Context) (*remoteTransportSession, error) {
	for {
		r.mu.Lock()
		current := r.current
		notify := r.notify
		r.mu.Unlock()

		if current != nil {
			return current, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-notify:
		}
	}
}

type remoteTransportSession struct {
	peer       *peerSession
	cancel     context.CancelFunc
	generation uint64

	mu     sync.Mutex
	closed bool
}

func newRemoteTransportSession(peer *peerSession, cancel context.CancelFunc) *remoteTransportSession {
	session := &remoteTransportSession{
		peer:   peer,
		cancel: cancel,
	}
	go func() {
		<-peer.Done()
		_ = session.Close()
	}()
	return session
}

func (s *remoteTransportSession) Done() <-chan struct{} {
	return s.peer.Done()
}

func (s *remoteTransportSession) OpenStream(tag string) (*remoteForwardStream, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errors.New("transport session closed")
	}
	s.mu.Unlock()

	channel, err := s.peer.OpenDataChannel(tag)
	if err != nil {
		return nil, err
	}

	respCh := make(chan []byte, 32)
	channel.SetMessageHandler(func(payload []byte) {
		select {
		case respCh <- append([]byte(nil), payload...):
		default:
			newError("dropping remote response packet due to full channel").AtWarning().WriteToLog()
		}
	})
	go func() {
		<-channel.Done()
		close(respCh)
	}()

	return &remoteForwardStream{
		channel: channel,
		respCh:  respCh,
	}, nil
}

func (s *remoteTransportSession) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	cancel := s.cancel
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	return s.peer.Close()
}

type remoteForwardStream struct {
	channel *peerDataChannel
	respCh  chan []byte
}

func (s *remoteForwardStream) Send(payload []byte) error {
	return s.channel.Send(payload)
}

func (s *remoteForwardStream) Responses() <-chan []byte {
	return s.respCh
}

func (s *remoteForwardStream) Close() error {
	return s.channel.Close()
}

type remoteForwardOutbound struct {
	tag        string
	runtime    *remoteConnectionRuntime
	forwardTag string

	mu     sync.RWMutex
	closed bool
}

func (o *remoteForwardOutbound) Tag() string {
	return o.tag
}

func (o *remoteForwardOutbound) Start() error {
	o.mu.Lock()
	o.closed = false
	o.mu.Unlock()
	return nil
}

func (o *remoteForwardOutbound) Close() error {
	o.mu.Lock()
	o.closed = true
	o.mu.Unlock()
	return nil
}

func (o *remoteForwardOutbound) Dispatch(ctx context.Context, link *transport.Link) {
	o.mu.RLock()
	closed := o.closed
	o.mu.RUnlock()
	if closed {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return
	}

	outboundMeta := session.OutboundFromContext(ctx)
	if outboundMeta == nil || !outboundMeta.Target.IsValid() || outboundMeta.Target.Network != v2net.Network_UDP {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		newError("remote forward outbound only supports UDP").AtWarning().WriteToLog()
		return
	}

	o.handleLink(ctx, link)
}

func (o *remoteForwardOutbound) handleLink(ctx context.Context, link *transport.Link) {
	defer common.Interrupt(link.Writer)

	var (
		current      *remoteForwardStream
		responseDone chan struct{}
	)

	bindStream := func(stream *remoteForwardStream) {
		if responseDone != nil {
			close(responseDone)
		}
		responseDone = make(chan struct{})
		go func(done <-chan struct{}, responses <-chan []byte) {
			for {
				select {
				case <-done:
					return
				case payload, ok := <-responses:
					if !ok {
						return
					}
					if err := writePacket(link.Writer, payload); err != nil {
						return
					}
				}
			}
		}(responseDone, stream.Responses())
	}

	sendPacket := func(payload []byte) error {
		for {
			if current == nil {
				session, err := o.runtime.waitForSession(ctx)
				if err != nil {
					return err
				}
				stream, err := session.OpenStream(o.forwardTag)
				if err != nil {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(reconnectBackoffMin):
						continue
					}
				}
				current = stream
				bindStream(stream)
			}

			if err := current.Send(payload); err != nil {
				_ = current.Close()
				current = nil
				continue
			}
			return nil
		}
	}

	err := readPackets(link.Reader, func(packet []byte) error {
		return sendPacket(packet)
	})
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
		newError("remote forward link closed").Base(err).AtDebug().WriteToLog()
	}

	if responseDone != nil {
		close(responseDone)
	}
	if current != nil {
		_ = current.Close()
	}
}

func decodeClientConfig(raw *anypb.Any) (*ClientConfig, error) {
	if raw == nil {
		return nil, newError("missing remote client_config")
	}
	instance, err := serial.GetInstanceOf(raw)
	if err != nil {
		return nil, newError("failed to decode remote client_config").Base(err)
	}
	cfg, ok := instance.(*ClientConfig)
	if !ok {
		return nil, newError("remote client_config is not a ClientConfig")
	}
	return cfg, nil
}

func decodeServerConfig(raw *anypb.Any) (*ServerInverseRoleConfig, error) {
	if raw == nil {
		return nil, newError("missing acceptor server_config")
	}
	instance, err := serial.GetInstanceOf(raw)
	if err != nil {
		return nil, newError("failed to decode acceptor server_config").Base(err)
	}
	cfg, ok := instance.(*ServerInverseRoleConfig)
	if !ok {
		return nil, newError("acceptor server_config is not a ServerInverseRoleConfig")
	}
	return cfg, nil
}

func readPackets(reader cbuf.Reader, handler func([]byte) error) error {
	for {
		mb, err := reader.ReadMultiBuffer()
		if !mb.IsEmpty() {
			for _, buffer := range mb {
				payload := append([]byte(nil), buffer.Bytes()...)
				buffer.Release()
				if len(payload) == 0 {
					continue
				}
				if handlerErr := handler(payload); handlerErr != nil {
					return handlerErr
				}
			}
		}
		if err != nil {
			return err
		}
	}
}

func writePacket(writer cbuf.Writer, payload []byte) error {
	packet := append([]byte(nil), payload...)
	return writer.WriteMultiBuffer(cbuf.MultiBuffer{cbuf.FromBytes(packet)})
}

func drainLocalCandidates(candidates <-chan []byte, gatheringDoneCh *chan struct{}, localGatheringDone *bool, seen map[string]struct{}, acc [][]byte) [][]byte {
	for {
		select {
		case <-*gatheringDoneCh:
			*localGatheringDone = true
			*gatheringDoneCh = nil
		case candidate, ok := <-candidates:
			if !ok {
				return acc
			}
			if len(candidate) == 0 {
				continue
			}
			key := string(candidate)
			if _, found := seen[key]; found {
				continue
			}
			seen[key] = struct{}{}
			acc = append(acc, append([]byte(nil), candidate...))
		default:
			return acc
		}
	}
}

func drainCandidateSet(candidates <-chan []byte, seen map[string]struct{}, acc [][]byte) [][]byte {
	var (
		gatheringDoneCh chan struct{}
		localDone       bool
	)
	return drainLocalCandidates(candidates, &gatheringDoneCh, &localDone, seen, acc)
}

func waitForSignalResume(
	ctx context.Context,
	remoteName string,
	replyTag []byte,
	sessionID []byte,
	candidates <-chan []byte,
	gatheringDoneCh *chan struct{},
	localGatheringDone *bool,
	seen map[string]struct{},
	acc *[][]byte,
	sentCandidateCount int,
) bool {
	for {
		*acc = drainLocalCandidates(candidates, gatheringDoneCh, localGatheringDone, seen, *acc)
		if len(*acc) > sentCandidateCount {
			newError(
				"client signal resume remote=", remoteName,
				" reply_tag=", compactSignalToken(replyTag),
				" session_id=", compactSignalToken(sessionID),
				" reason=new_local_candidate",
				" candidates=", len(*acc),
			).AtDebug().WriteToLog()
			return true
		}
		if *localGatheringDone {
			newError(
				"client signal finish remote=", remoteName,
				" reply_tag=", compactSignalToken(replyTag),
				" session_id=", compactSignalToken(sessionID),
				" reason=acceptor_stop_polling_and_local_gathering_complete",
			).AtDebug().WriteToLog()
			return false
		}

		newError(
			"client signal pause remote=", remoteName,
			" reply_tag=", compactSignalToken(replyTag),
			" session_id=", compactSignalToken(sessionID),
			" reason=acceptor_stop_polling",
		).AtDebug().WriteToLog()

		select {
		case <-ctx.Done():
			return false
		case <-*gatheringDoneCh:
			*localGatheringDone = true
			*gatheringDoneCh = nil
		case candidate, ok := <-candidates:
			if !ok {
				continue
			}
			if len(candidate) == 0 {
				continue
			}
			key := string(candidate)
			if _, found := seen[key]; found {
				continue
			}
			seen[key] = struct{}{}
			*acc = append(*acc, append([]byte(nil), candidate...))
		}
	}
}

func cloneByteSlices(in [][]byte) [][]byte {
	if len(in) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(in))
	for _, item := range in {
		out = append(out, append([]byte(nil), item...))
	}
	return out
}

func randomBytes(size int) ([]byte, error) {
	out := make([]byte, size)
	if _, err := io.ReadFull(crand.Reader, out); err != nil {
		return nil, newError("failed to read random bytes").Base(err)
	}
	return out, nil
}

func sessionKey(replyTag, sessionID []byte) string {
	return string(replyTag) + "|" + string(sessionID)
}

func nextBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		return reconnectBackoffMin
	}
	current *= 2
	if current > reconnectBackoffMax {
		return reconnectBackoffMax
	}
	return current
}

type localCandidateSendGate struct {
	enabled bool
	started bool

	releaseDelay time.Duration
	releaseAt    time.Time
	now          func() time.Time
}

func newLocalCandidateSendGate(enabled bool, releaseDelay time.Duration, now func() time.Time) localCandidateSendGate {
	if now == nil {
		now = time.Now
	}
	return localCandidateSendGate{
		enabled:      enabled,
		releaseDelay: releaseDelay,
		now:          now,
	}
}

func (g *localCandidateSendGate) AllowSend() bool {
	if !g.enabled {
		return true
	}
	if !g.started {
		return false
	}
	return !g.now().Before(g.releaseAt)
}

func (g *localCandidateSendGate) StartCountdown() bool {
	if !g.enabled || g.started {
		return false
	}
	g.started = true
	g.releaseAt = g.now().Add(g.releaseDelay)
	return true
}

func (g *localCandidateSendGate) NextPollDelay(defaultDelay time.Duration) time.Duration {
	if !g.enabled || !g.started {
		return defaultDelay
	}
	remaining := g.releaseAt.Sub(g.now())
	if remaining <= 0 || remaining >= defaultDelay {
		return defaultDelay
	}
	return remaining
}

func waitSignalPoll(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func compactSignalToken(data []byte) string {
	if len(data) == 0 {
		return "-"
	}
	text := base64.RawURLEncoding.EncodeToString(data)
	if len(text) > 16 {
		return text[:16]
	}
	return text
}
