//go:build !confonly
// +build !confonly

package rrpitTransport

import (
	"context"
	"fmt"
	"io"
	gonet "net"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transportdtls "github.com/v2fly/v2ray-core/v5/transport/internet/dtls"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSessionManager"
)

type Listener struct {
	config   *Config
	handler  internet.ConnHandler
	channels []resolvedChannel

	mu        sync.Mutex
	listeners []internet.Listener
	sessions  map[transportSessionID]*transportSession
	closed    bool
}

func (l *Listener) Close() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	listeners := append([]internet.Listener(nil), l.listeners...)
	l.listeners = nil
	sessions := make([]*transportSession, 0, len(l.sessions))
	for _, session := range l.sessions {
		sessions = append(sessions, session)
	}
	l.sessions = nil
	l.mu.Unlock()

	var firstErr error
	for _, listener := range listeners {
		if err := listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for _, session := range sessions {
		if err := session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (l *Listener) Addr() gonet.Addr {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.listeners) == 0 {
		return nil
	}
	return l.listeners[0].Addr()
}

func (l *Listener) handleAcceptedChannel(conn gonet.Conn, channelIndex int) {
	go l.serveAcceptedChannel(conn, channelIndex)
}

func (l *Listener) serveAcceptedChannel(conn gonet.Conn, channelIndex int) {
	sessionID, err := readTransportSessionID(conn)
	if err != nil {
		_ = conn.Close()
		return
	}

	owner, err := l.getOrCreateSession(sessionID)
	if err != nil {
		_ = conn.Close()
		return
	}

	if err := owner.attachChannel(conn, channelIndex, rrpitChannelConfig(l.channels[channelIndex].transport), channelReadIdleTimeout(l.channels[channelIndex])); err != nil {
		_ = owner.Close()
	}
}

func (l *Listener) getOrCreateSession(sessionID transportSessionID) (*transportSession, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, io.ErrClosedPipe
	}
	if existing := l.sessions[sessionID]; existing != nil {
		return existing, nil
	}

	owner, err := newTransportSession("server", sessionID, l.config, false, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if l.sessions != nil {
			delete(l.sessions, sessionID)
		}
	}, nil)
	if err != nil {
		return nil, err
	}

	if l.sessions == nil {
		l.sessions = map[transportSessionID]*transportSession{}
	}
	l.sessions[sessionID] = owner

	go l.acceptStreams(owner, rrpitBidirectionalSessionManager.InteractiveStream)
	go l.acceptStreams(owner, rrpitBidirectionalSessionManager.BackgroundStream)
	return owner, nil
}

func (l *Listener) acceptStreams(owner *transportSession, sessionClass rrpitBidirectionalSessionManager.SessionName) {
	for {
		stream, err := owner.AcceptStream(sessionClass)
		if err != nil {
			_ = owner.Close()
			return
		}
		l.handler(&ownedConn{Conn: stream, owner: owner})
	}
}

func Listen(ctx context.Context, address v2net.Address, port v2net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	config, ok := streamSettings.ProtocolSettings.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid rrpit transport config")
	}

	channels, err := resolveListenChannels(address, port, config)
	if err != nil {
		return nil, err
	}

	listener := &Listener{
		config:   config,
		handler:  handler,
		channels: channels,
		sessions: map[transportSessionID]*transportSession{},
	}

	for index, channel := range channels {
		index := index
		dtlsListener, err := transportdtls.ListenDTLS(
			ctx,
			channel.address,
			channel.port,
			makeDTLSStreamSettings(channel, streamSettings.SocketSettings),
			func(conn internet.Connection) {
				listener.handleAcceptedChannel(conn, index)
			},
		)
		if err != nil {
			_ = listener.Close()
			return nil, err
		}
		listener.listeners = append(listener.listeners, dtlsListener)
	}

	return listener, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
