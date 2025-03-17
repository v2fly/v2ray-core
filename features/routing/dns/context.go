package dns

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

// ResolvableContext is an implementation of routing.Context, with domain resolving capability.
type ResolvableContext struct {
	routing.Context
	dnsClient   dns.Client
	resolvedIPs []net.IP
}

// fuck by b1gcat for inbound tag
func JoinInboundDomainTag(ctx any, domain string) (string, string) {
	newDomain := domain
	inBoundTag := ""

	switch ctx := ctx.(type) {
	case context.Context:
		if session.InboundFromContext(ctx) != nil {
			inBoundTag = session.InboundFromContext(ctx).Tag
		}
	case *ResolvableContext:
		inBoundTag = ctx.GetInboundTag()
	}

	if inBoundTag != "" {
		newDomain = fmt.Sprintf("dns-%s:%s", inBoundTag, domain)
	}

	return inBoundTag, newDomain
}

func SplitInboundDomainTag(domain string) (string, string) {
	info := strings.SplitN(domain, ":", 2)
	if len(info) == 2 {
		return info[0], info[1] //tag, domain
	}
	return "", domain
}

func MatchInboundDomainTag(inBoundtag, clientTag string) (matched bool) {
	if inBoundtag == "" || clientTag == "" {
		panic("BUG: inBoundtag or clientTag is empty:" + inBoundtag + "/" + clientTag)
	}

	tags := strings.SplitN(clientTag, ":", 2)
	if len(tags) == 2 && tags[1] == inBoundtag {
		matched = true
		return
	}

	return
}

// fuck by b1gcat for inbound tag end

// GetTargetIPs overrides original routing.Context's implementation.
func (ctx *ResolvableContext) GetTargetIPs() []net.IP {
	if ips := ctx.Context.GetTargetIPs(); len(ips) != 0 {
		return ips
	}

	if len(ctx.resolvedIPs) > 0 {
		return ctx.resolvedIPs
	}

	if domain := ctx.GetTargetDomain(); len(domain) != 0 {
		//add by b1gcat start
		_, domain = JoinInboundDomainTag(ctx, domain)
		// add by b1gcat end

		ips, err := ctx.dnsClient.LookupIP(domain)
		if err == nil {
			ctx.resolvedIPs = ips
			return ips
		}
		newError("resolve ip for ", domain).Base(err).WriteToLog()
	}

	return nil
}

// ContextWithDNSClient creates a new routing context with domain resolving capability.
// Resolved domain IPs can be retrieved by GetTargetIPs().
func ContextWithDNSClient(ctx routing.Context, client dns.Client) routing.Context {
	return &ResolvableContext{Context: ctx, dnsClient: client}
}
