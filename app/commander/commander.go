package commander

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"net"
	"sync"

	"google.golang.org/grpc"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/infra/conf/v5cfg"
)

type CommanderIfce interface {
	features.Feature

	ExtractGrpcServer() *grpc.Server
}

// Commander is a V2Ray feature that provides gRPC methods to external clients.
type Commander struct {
	sync.Mutex
	server   *grpc.Server
	services []Service
	ohm      outbound.Manager
	tag      string
}

// NewCommander creates a new Commander based on the given config.
func NewCommander(ctx context.Context, config *Config) (*Commander, error) {
	c := &Commander{
		tag: config.Tag,
	}

	common.Must(core.RequireFeatures(ctx, func(om outbound.Manager) {
		c.ohm = om
	}))

	for _, rawConfig := range config.Service {
		config, err := serial.GetInstanceOf(rawConfig)
		if err != nil {
			return nil, err
		}
		rawService, err := common.CreateObject(ctx, config)
		if err != nil {
			return nil, err
		}
		service, ok := rawService.(Service)
		if !ok {
			return nil, newError("not a Service.")
		}
		c.services = append(c.services, service)
	}

	return c, nil
}

// Type implements common.HasType.
func (c *Commander) Type() interface{} {
	return (*CommanderIfce)(nil)
}

// Start implements common.Runnable.
func (c *Commander) Start() error {
	c.Lock()
	c.server = grpc.NewServer()
	for _, service := range c.services {
		service.Register(c.server)
	}
	c.Unlock()

	listener := &OutboundListener{
		buffer: make(chan net.Conn, 4),
		done:   done.New(),
	}

	go func() {
		if err := c.server.Serve(listener); err != nil {
			newError("failed to start grpc server").Base(err).AtError().WriteToLog()
		}
	}()

	if err := c.ohm.RemoveHandler(context.Background(), c.tag); err != nil {
		newError("failed to remove existing handler").WriteToLog()
	}

	return c.ohm.AddHandler(context.Background(), &Outbound{
		tag:      c.tag,
		listener: listener,
	})
}

// Close implements common.Closable.
func (c *Commander) Close() error {
	c.Lock()
	defer c.Unlock()

	if c.server != nil {
		c.server.Stop()
		c.server = nil
	}

	return nil
}

// ExtractGrpcServer extracts the gRPC server from Commander.
// Private function for core code base.
func (c *Commander) ExtractGrpcServer() *grpc.Server {
	c.Lock()
	defer c.Unlock()
	return c.server
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		return NewCommander(ctx, cfg.(*Config))
	}))
}

func init() {
	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		simplifiedConfig := cfg.(*SimplifiedConfig)
		fullConfig := &Config{
			Tag:     simplifiedConfig.Tag,
			Service: nil,
		}
		for _, v := range simplifiedConfig.Name {
			pack, err := v5cfg.LoadHeterogeneousConfigFromRawJSON(ctx, "grpcservice", v, []byte("{}"))
			if err != nil {
				return nil, err
			}
			fullConfig.Service = append(fullConfig.Service, serial.ToTypedMessage(pack))
		}
		return common.CreateObject(ctx, fullConfig)
	}))
}
