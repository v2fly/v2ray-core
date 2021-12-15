package dispatcher

import (
	"context"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/features/routing"
	"github.com/v2fly/v2ray-core/v4/transport"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

var (
	_              routing.Dispatcher = (*SystemDispatcher)(nil)
	SystemInstance                    = &SystemDispatcher{}
)

type SystemDispatcher struct{}

func (s *SystemDispatcher) Type() interface{} {
	return routing.DispatcherType()
}

func (s *SystemDispatcher) Start() error {
	return nil
}

func (s *SystemDispatcher) Close() error {
	return nil
}

func (s *SystemDispatcher) Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error) {
	conn, err := internet.DialSystem(ctx, dest, nil)
	if err != nil {
		return nil, err
	}
	return &transport.Link{Reader: buf.NewReader(conn), Writer: buf.NewWriter(conn)}, nil
}
