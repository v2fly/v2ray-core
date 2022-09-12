package session

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

// OriginDestContext is an implementation of routing.Context,
// With target IPs derived from destination before overridden by sniffed domains, if available.
type OriginDestContext struct {
	*Context
}

// GetTargetIPs overrides original routing.Context's implementation.
func (ctx OriginDestContext) GetTargetIPs() []net.IP {
	if ctx.Content != nil && ctx.Content.OverriddenDestination != nil && ctx.Content.OverriddenDestination.Family().IsIP() {
		return []net.IP{ctx.Content.OverriddenDestination.IP()}
	}
	return ctx.Context.GetTargetIPs()
}

// ContextWithOriginDestination creates a new routing context with ability to retrieve original destination.
// Original IP destination can be retrieved by GetTargetIPs(), along with overridden domain by GetTargetDomain().
func ContextWithOriginDestination(ctx routing.Context) routing.Context {
	if ctx, ok := ctx.(*Context); ok {
		return OriginDestContext{ctx}
	}
	return ctx
}
