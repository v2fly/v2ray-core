// Package proxy contains all proxies used by V2Ray.
//
// To implement an inbound or outbound proxy, one needs to do the following:
// 1. Implement the interface(s) below.
// 2. Register a config creator through common.RegisterConfig.
package proxy

import (
	"context"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/features/routing"
	"github.com/ghxhy/v2ray-core/v5/transport"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

// A timeout for reading the first payload from the client, used in 0-RTT optimizations.
const FirstPayloadTimeout = 100 * time.Millisecond

// An Inbound processes inbound connections.
type Inbound interface {
	// Network returns a list of networks that this inbound supports. Connections with not-supported networks will not be passed into Process().
	Network() []net.Network

	// Process processes a connection of given network. If necessary, the Inbound can dispatch the connection to an Outbound.
	Process(context.Context, net.Network, internet.Connection, routing.Dispatcher) error
}

// An Outbound process outbound connections.
type Outbound interface {
	// Process processes the given connection. The given dialer may be used to dial a system outbound connection.
	Process(context.Context, *transport.Link, internet.Dialer) error
}

// UserManager is the interface for Inbounds and Outbounds that can manage their users.
type UserManager interface {
	// AddUser adds a new user.
	AddUser(context.Context, *protocol.MemoryUser) error

	// RemoveUser removes a user by email.
	RemoveUser(context.Context, string) error
}

type GetInbound interface {
	GetInbound() Inbound
}

type GetOutbound interface {
	GetOutbound() Outbound
}
