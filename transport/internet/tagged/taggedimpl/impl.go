//go:build !confonly
// +build !confonly

package taggedimpl

import (
	"context"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/features/routing"
	"github.com/ghxhy/v2ray-core/v5/transport/internet/tagged"
)

func DialTaggedOutbound(ctx context.Context, dest net.Destination, tag string) (net.Conn, error) {
	var dispatcher routing.Dispatcher
	if core.FromContext(ctx) == nil {
		return nil, newError("Instance context variable is not in context, dial denied. ")
	}
	if err := core.RequireFeatures(ctx, func(dispatcherInstance routing.Dispatcher) {
		dispatcher = dispatcherInstance
	}); err != nil {
		return nil, newError("Required Feature dispatcher not resolved").Base(err)
	}

	content := new(session.Content)
	content.SkipDNSResolve = true

	ctx = session.ContextWithContent(ctx, content)
	ctx = session.SetForcedOutboundTagToContext(ctx, tag)

	r, err := dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return nil, err
	}
	var readerOpt net.ConnectionOption
	if dest.Network == net.Network_TCP {
		readerOpt = net.ConnectionOutputMulti(r.Reader)
	} else {
		readerOpt = net.ConnectionOutputMultiUDP(r.Reader)
	}
	return net.NewConnection(net.ConnectionInputMulti(r.Writer), readerOpt), nil
}

func init() {
	tagged.Dialer = DialTaggedOutbound
}
