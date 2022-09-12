package session

import (
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

// OriginDestContext is an implementation of routing.Context,
// With target IPs derived from destination before overridden by sniffed domains, if available.
type OriginDestContext struct {
	routing.Context
	originIP []net.IP
}

// Unwrap implements routing.Context.
func (ctx *OriginDestContext) Unwrap() routing.Context {
	return ctx.Context
}

// GetTargetIPs overrides original routing.Context's implementation.
func (ctx *OriginDestContext) GetTargetIPs() []net.IP {
	if len(ctx.originIP) > 0 {
		return ctx.originIP
	}
	return ctx.Context.GetTargetIPs()
}

// ContextWithOriginDestination creates a new routing context with ability to retrieve original destination.
// Original IP destination can be retrieved by GetTargetIPs(), along with overridden domain by GetTargetDomain().
func ContextWithOriginDestination(context routing.Context) (routing.Context, bool) {
	for ctx := context; ctx != nil; ctx = ctx.Unwrap() {
		if ctx, ok := ctx.(*Context); ok && ctx.Content != nil && ctx.Content.OverriddenDestination != nil && ctx.Content.OverriddenDestination.Family().IsIP() {
			return &OriginDestContext{Context: context, originIP: []net.IP{ctx.Content.OverriddenDestination.IP()}}, true
		}
	}
	return context, false
}
