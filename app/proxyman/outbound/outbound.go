package outbound

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"strings"
	"sync"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/errors"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/features/outbound"
)

// Manager is to manage all outbound handlers.
type Manager struct {
	access           sync.RWMutex
	defaultHandler   outbound.Handler
	taggedHandler    map[string]outbound.Handler
	untaggedHandlers []outbound.Handler
	running          bool
}

// New creates a new Manager.
func New(ctx context.Context, config *proxyman.OutboundConfig) (*Manager, error) {
	m := &Manager{
		taggedHandler: make(map[string]outbound.Handler),
	}
	return m, nil
}

// Type implements common.HasType.
func (m *Manager) Type() interface{} {
	return outbound.ManagerType()
}

// Start implements core.Feature
func (m *Manager) Start() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = true

	for _, h := range m.taggedHandler {
		if err := h.Start(); err != nil {
			return err
		}
	}

	for _, h := range m.untaggedHandlers {
		if err := h.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Close implements core.Feature
func (m *Manager) Close() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = false

	var errs []error
	for _, h := range m.taggedHandler {
		errs = append(errs, h.Close())
	}

	for _, h := range m.untaggedHandlers {
		errs = append(errs, h.Close())
	}

	return errors.Combine(errs...)
}

// GetDefaultHandler implements outbound.Manager.
func (m *Manager) GetDefaultHandler() outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()

	return m.defaultHandler
}

// GetHandler implements outbound.Manager.
func (m *Manager) GetHandler(tag string) outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()
	if handler, found := m.taggedHandler[tag]; found {
		return handler
	}
	return nil
}

// AddHandler implements outbound.Manager.
func (m *Manager) AddHandler(ctx context.Context, handler outbound.Handler) error {
	m.access.Lock()
	defer m.access.Unlock()
	tag := handler.Tag()

	if m.defaultHandler == nil ||
		(len(tag) > 0 && tag == m.defaultHandler.Tag()) {
		m.defaultHandler = handler
	}

	if len(tag) > 0 {
		if oldHandler, found := m.taggedHandler[tag]; found {
			errors.New("will replace the existed outbound with the tag: " + tag).AtWarning().WriteToLog()
			_ = oldHandler.Close()
		}
		m.taggedHandler[tag] = handler
	} else {
		m.untaggedHandlers = append(m.untaggedHandlers, handler)
	}

	if m.running {
		return handler.Start()
	}

	return nil
}

// RemoveHandler implements outbound.Manager.
func (m *Manager) RemoveHandler(ctx context.Context, tag string) error {
	if tag == "" {
		return common.ErrNoClue
	}
	m.access.Lock()
	defer m.access.Unlock()

	if handler, found := m.taggedHandler[tag]; found {
		if err := handler.Close(); err != nil {
			newError("failed to close handler ", tag).Base(err).AtWarning().WriteToLog(session.ExportIDToError(ctx))
		}
		delete(m.taggedHandler, tag)
		if m.defaultHandler != nil && m.defaultHandler.Tag() == tag {
			m.defaultHandler = nil
		}
		return nil
	}

	return common.ErrNoClue
}

// Select implements outbound.HandlerSelector.
func (m *Manager) Select(selectors []string) []string {
	m.access.RLock()
	defer m.access.RUnlock()

	tags := make([]string, 0, len(selectors))

	for tag := range m.taggedHandler {
		match := false
		for _, selector := range selectors {
			if strings.HasPrefix(tag, selector) {
				match = true
				break
			}
		}
		if match {
			tags = append(tags, tag)
		}
	}

	return tags
}

func init() {
	common.Must(common.RegisterConfig((*proxyman.OutboundConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*proxyman.OutboundConfig))
	}))
	common.Must(common.RegisterConfig((*core.OutboundHandlerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewHandler(ctx, config.(*core.OutboundHandlerConfig))
	}))
}
