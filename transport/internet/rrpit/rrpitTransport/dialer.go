//go:build !confonly
// +build !confonly

package rrpitTransport

import (
	"context"
	"fmt"
	"io"
	gonet "net"
	"sync"
	"time"

	piondtls "github.com/pion/dtls/v3"
	piondtlsnet "github.com/pion/dtls/v3/pkg/net"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	v2session "github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSessionManager"
)

const (
	rrpitTransportConnectionStateKey = "rrpit-transport-connection-state"
	rrpitClientSessionIdleTimeout    = 5 * time.Second
	rrpitSessionClassAttributeKey    = "rrpitSessionClass"
)

type transportConnectionState struct {
	scopedSessionMap    map[string]*persistentClientSession
	scopedSessionAccess sync.Mutex
	closed              bool
}

func (*transportConnectionState) IsTransientStorageLifecycleReceiver() {}

func (s *transportConnectionState) Close() error {
	s.scopedSessionAccess.Lock()
	s.closed = true
	sessions := make([]*persistentClientSession, 0, len(s.scopedSessionMap))
	for _, session := range s.scopedSessionMap {
		sessions = append(sessions, session)
	}
	s.scopedSessionMap = nil
	s.scopedSessionAccess.Unlock()

	var firstErr error
	for _, session := range sessions {
		if err := session.Close(); err != nil && firstErr == nil && err != io.ErrClosedPipe {
			firstErr = err
		}
	}
	return firstErr
}

func (s *transportConnectionState) getOrCreateSession(
	ctx context.Context,
	dest v2net.Destination,
	streamSettings *internet.MemoryStreamConfig,
) (*persistentClientSession, error) {
	key := rrpitTransportSessionKey(dest)

	s.scopedSessionAccess.Lock()
	if s.closed {
		s.scopedSessionAccess.Unlock()
		return nil, io.ErrClosedPipe
	}
	if existing := s.scopedSessionMap[key]; existing != nil && !existing.IsClosed() {
		s.scopedSessionAccess.Unlock()
		return existing, nil
	}
	s.scopedSessionAccess.Unlock()

	var session *persistentClientSession
	created, err := newPersistentClientSession(ctx, dest, streamSettings, rrpitClientSessionIdleTimeout, func() {
		s.removeSession(key, session)
	})
	if err != nil {
		return nil, err
	}
	session = created

	s.scopedSessionAccess.Lock()
	defer s.scopedSessionAccess.Unlock()

	if s.closed {
		_ = session.Close()
		return nil, io.ErrClosedPipe
	}
	if s.scopedSessionMap == nil {
		s.scopedSessionMap = make(map[string]*persistentClientSession)
	}
	if existing := s.scopedSessionMap[key]; existing != nil && !existing.IsClosed() {
		_ = session.Close()
		return existing, nil
	}
	s.scopedSessionMap[key] = session
	return session, nil
}

func (s *transportConnectionState) removeSession(key string, session *persistentClientSession) {
	if s == nil {
		return
	}

	s.scopedSessionAccess.Lock()
	defer s.scopedSessionAccess.Unlock()

	if s.scopedSessionMap == nil {
		return
	}
	if current := s.scopedSessionMap[key]; current == session {
		delete(s.scopedSessionMap, key)
	}
}

type persistentClientSession struct {
	owner        *transportSession
	openStream   func(rrpitBidirectionalSessionManager.SessionName) (gonet.Conn, error)
	closeSession func() error
	idleTimeout  time.Duration

	mu            sync.Mutex
	activeStreams int
	idleTimer     *time.Timer
	closed        bool
}

func (s *persistentClientSession) IsClosed() bool {
	if s == nil {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func (s *persistentClientSession) OpenConnection(ctx context.Context) (internet.Connection, error) {
	if s == nil {
		return nil, io.ErrClosedPipe
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, io.ErrClosedPipe
	}
	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}
	s.activeStreams++
	s.mu.Unlock()

	stream, err := s.openStream(sessionClassFromContext(ctx))
	if err != nil {
		s.releaseStream()
		_ = s.Close()
		return nil, err
	}

	conn := &ownedConn{
		Conn:    stream,
		owner:   s.owner,
		onClose: s.releaseStream,
		ctx:     ctx,
		done:    make(chan struct{}),
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			_ = conn.Close()
			return nil, err
		}
		if done := ctx.Done(); done != nil {
			go func() {
				select {
				case <-done:
					_ = conn.Close()
				case <-conn.done:
				}
			}()
		}
	}

	return conn, nil
}

func (s *persistentClientSession) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return io.ErrClosedPipe
	}
	s.closed = true
	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}
	s.mu.Unlock()

	if s.closeSession != nil {
		return s.closeSession()
	}
	return nil
}

func (s *persistentClientSession) releaseStream() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeStreams > 0 {
		s.activeStreams--
	}
	if s.closed || s.activeStreams != 0 || s.idleTimeout <= 0 || s.idleTimer != nil {
		return
	}

	s.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		s.mu.Lock()
		if s.closed || s.activeStreams != 0 {
			s.idleTimer = nil
			s.mu.Unlock()
			return
		}
		s.idleTimer = nil
		s.mu.Unlock()
		_ = s.Close()
	})
}

func rrpitTransportSessionKey(dest v2net.Destination) string {
	return dest.String()
}

func getRRPITTransportConnectionState(ctx context.Context) (*transportConnectionState, error) {
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	state, err := transportEnvironment.TransientStorage().Get(ctx, rrpitTransportConnectionStateKey)
	if err != nil {
		state = &transportConnectionState{}
		if putErr := transportEnvironment.TransientStorage().Put(ctx, rrpitTransportConnectionStateKey, state); putErr != nil {
			return nil, putErr
		}
		state, err = transportEnvironment.TransientStorage().Get(ctx, rrpitTransportConnectionStateKey)
		if err != nil {
			return nil, err
		}
	}
	typed, ok := state.(*transportConnectionState)
	if !ok {
		return nil, fmt.Errorf("invalid rrpit transport connection state %T", state)
	}
	return typed, nil
}

func newPersistentClientSession(
	ctx context.Context,
	dest v2net.Destination,
	streamSettings *internet.MemoryStreamConfig,
	idleTimeout time.Duration,
	onClose func(),
) (*persistentClientSession, error) {
	config, ok := streamSettings.ProtocolSettings.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid rrpit transport config")
	}

	channels, err := resolveDialChannels(dest, config)
	if err != nil {
		return nil, err
	}

	id, err := newSessionID()
	if err != nil {
		return nil, err
	}

	session := &persistentClientSession{idleTimeout: idleTimeout}
	owner, err := newTransportSession("client", id, config, true, func() {
		session.mu.Lock()
		session.closed = true
		if session.idleTimer != nil {
			session.idleTimer.Stop()
			session.idleTimer = nil
		}
		session.mu.Unlock()
		if onClose != nil {
			onClose()
		}
	})
	if err != nil {
		return nil, err
	}

	for index, channel := range channels {
		rawConn, err := internet.DialSystem(ctx, v2net.UDPDestination(channel.address, channel.port), streamSettings.SocketSettings)
		if err != nil {
			_ = owner.Close()
			return nil, err
		}

		dtlsConn, err := piondtls.Client(
			piondtlsnet.PacketConnFromConn(rawConn),
			rawConn.RemoteAddr(),
			makePionDTLSConfig(channel, id),
		)
		if err != nil {
			_ = rawConn.Close()
			_ = owner.Close()
			return nil, err
		}
		if err := dtlsConn.Handshake(); err != nil {
			_ = dtlsConn.Close()
			_ = owner.Close()
			return nil, err
		}

		if err := owner.addChannel(dtlsConn, index, rrpitChannelConfig(channel.transport)); err != nil {
			_ = owner.Close()
			return nil, err
		}
	}

	session.owner = owner
	session.openStream = owner.OpenStreamByClass
	session.closeSession = owner.Close
	return session, nil
}

func sessionClassFromContext(ctx context.Context) rrpitBidirectionalSessionManager.SessionName {
	if ctx == nil {
		return rrpitBidirectionalSessionManager.InteractiveStream
	}
	content := v2session.ContentFromContext(ctx)
	if content == nil {
		return rrpitBidirectionalSessionManager.InteractiveStream
	}
	switch content.Attribute(rrpitSessionClassAttributeKey) {
	case "background":
		return rrpitBidirectionalSessionManager.BackgroundStream
	default:
		return rrpitBidirectionalSessionManager.InteractiveStream
	}
}

func Dial(ctx context.Context, dest v2net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	detachedContext := core.ToBackgroundDetachedContext(ctx)

	state, err := getRRPITTransportConnectionState(detachedContext)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		session, err := state.getOrCreateSession(detachedContext, dest, streamSettings)
		if err != nil {
			return nil, err
		}

		conn, err := session.OpenConnection(ctx)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		_ = session.Close()
	}

	if lastErr == nil {
		lastErr = io.ErrClosedPipe
	}
	return nil, lastErr
}

var _ storage.TransientStorageLifecycleReceiver = (*transportConnectionState)(nil)

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
