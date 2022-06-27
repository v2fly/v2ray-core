package log

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"reflect"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/log"
)

// Instance is a log.Handler that handles logs.
type Instance struct {
	sync.RWMutex
	config       *Config
	accessLogger log.Handler
	errorLogger  log.Handler
	followers    map[reflect.Value]func(msg log.Message)
	active       bool
}

// New creates a new log.Instance based on the given config.
func New(ctx context.Context, config *Config) (*Instance, error) {
	if config.Error == nil {
		config.Error = &LogSpecification{Type: LogType_Console, Level: log.Severity_Warning}
	}

	if config.Access == nil {
		config.Access = &LogSpecification{Type: LogType_None}
	}

	g := &Instance{
		config: config,
		active: false,
	}
	log.RegisterHandler(g)

	// start logger instantly on inited
	// other modules would log during init
	if err := g.startInternal(); err != nil {
		return nil, err
	}

	newError("Logger started").AtDebug().WriteToLog()
	return g, nil
}

func (g *Instance) initAccessLogger() error {
	handler, err := createHandler(g.config.Access.Type, HandlerCreatorOptions{
		Path: g.config.Access.Path,
	})
	if err != nil {
		return err
	}
	g.accessLogger = handler
	return nil
}

func (g *Instance) initErrorLogger() error {
	handler, err := createHandler(g.config.Error.Type, HandlerCreatorOptions{
		Path: g.config.Error.Path,
	})
	if err != nil {
		return err
	}
	g.errorLogger = handler
	return nil
}

// Type implements common.HasType.
func (*Instance) Type() interface{} {
	return (*Instance)(nil)
}

func (g *Instance) startInternal() error {
	g.Lock()
	defer g.Unlock()

	if g.active {
		return nil
	}

	g.active = true

	if err := g.initAccessLogger(); err != nil {
		return newError("failed to initialize access logger").Base(err).AtWarning()
	}
	if err := g.initErrorLogger(); err != nil {
		return newError("failed to initialize error logger").Base(err).AtWarning()
	}

	return nil
}

// Start implements common.Runnable.Start().
func (g *Instance) Start() error {
	return g.startInternal()
}

// AddFollower implements log.Follower.
func (g *Instance) AddFollower(f func(msg log.Message)) {
	g.Lock()
	defer g.Unlock()
	if g.followers == nil {
		g.followers = make(map[reflect.Value]func(msg log.Message))
	}
	g.followers[reflect.ValueOf(f)] = f
}

// RemoveFollower implements log.Follower.
func (g *Instance) RemoveFollower(f func(msg log.Message)) {
	g.Lock()
	defer g.Unlock()
	delete(g.followers, reflect.ValueOf(f))
}

// Handle implements log.Handler.
func (g *Instance) Handle(msg log.Message) {
	g.RLock()
	defer g.RUnlock()

	if !g.active {
		return
	}

	for _, f := range g.followers {
		f(msg)
	}

	switch msg := msg.(type) {
	case *log.AccessMessage:
		if g.accessLogger != nil {
			g.accessLogger.Handle(msg)
		}
	case *log.GeneralMessage:
		if g.errorLogger != nil && msg.Severity <= g.config.Error.Level {
			g.errorLogger.Handle(msg)
		}
	default:
		// Swallow
	}
}

// Close implements common.Closable.Close().
func (g *Instance) Close() error {
	newError("Logger closing").AtDebug().WriteToLog()

	g.Lock()
	defer g.Unlock()

	if !g.active {
		return nil
	}

	g.active = false

	common.Close(g.accessLogger)
	g.accessLogger = nil

	common.Close(g.errorLogger)
	g.errorLogger = nil

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
