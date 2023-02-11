package dns

import (
	"context"
	"net/url"
	"strings"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

// Server is the interface for Name Server.
type Server interface {
	// Name of the Client.
	Name() string
	// QueryIP sends IP queries to its configured server.
	QueryIP(ctx context.Context, domain string, clientIP net.IP, option dns.IPOption, disableCache bool) ([]net.IP, error)
}

// Client is the interface for DNS client.
type Client struct {
	server   Server
	clientIP net.IP
	tag      string

	queryStrategy    dns.IPOption
	cacheStrategy    CacheStrategy
	fallbackStrategy FallbackStrategy

	domains   []string
	expectIPs []*router.GeoIPMatcher
	fakeDNS   Server
}

var errExpectedIPNonMatch = errors.New("expectIPs not match")

// NewServer creates a name server object according to the network destination url.
func NewServer(ctx context.Context, dest net.Destination, onCreated func(Server) error) error {
	onCreatedWithError := func(server Server, err error) error {
		if err != nil {
			return err
		}
		return onCreated(server)
	}
	if address := dest.Address; address.Family().IsDomain() {
		u, err := url.Parse(address.Domain())
		if err != nil {
			return err
		}
		switch {
		case strings.EqualFold(u.String(), "localhost"):
			return onCreated(NewLocalNameServer())
		case strings.EqualFold(u.String(), "fakedns"):
			return core.RequireFeatures(ctx, func(fakedns dns.FakeDNSEngine) error { return onCreated(NewFakeDNSServer(fakedns)) })
		case strings.EqualFold(u.Scheme, "https"): // DOH Remote mode
			return core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error { return onCreatedWithError(NewDoHNameServer(u, dispatcher)) })
		case strings.EqualFold(u.Scheme, "https+local"): // DOH Local mode
			return onCreated(NewDoHLocalNameServer(u))
		case strings.EqualFold(u.Scheme, "tcp"): // DNS-over-TCP Remote mode
			return core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error { return onCreatedWithError(NewTCPNameServer(u, dispatcher)) })
		case strings.EqualFold(u.Scheme, "tcp+local"): // DNS-over-TCP Local mode
			return onCreatedWithError(NewTCPLocalNameServer(u))
		case strings.EqualFold(u.Scheme, "quic+local"): // DNS-over-QUIC Local mode
			return onCreatedWithError(NewQUICNameServer(u))
		}
	}
	if dest.Network == net.Network_Unknown {
		dest.Network = net.Network_UDP
	}
	if dest.Network == net.Network_UDP { // UDP classic DNS mode
		return core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error { return onCreated(NewClassicNameServer(dest, dispatcher)) })
	}
	return newError("No available name server could be created from ", dest).AtWarning()
}

// NewClient creates a DNS client managing a name server with client IP, domain rules and expected IPs.
func NewClient(ctx context.Context, ns *NameServer, dns *Config) (*Client, error) {
	client := &Client{}

	// Create DNS server instance
	err := NewServer(ctx, ns.Address.AsDestination(), func(server Server) error {
		client.server = server
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Initialize fields with default values
	if len(ns.Tag) == 0 {
		ns.Tag = dns.Tag
		if len(ns.Tag) == 0 {
			ns.Tag = generateRandomTag()
		}
	}
	if len(ns.ClientIp) == 0 {
		ns.ClientIp = dns.ClientIp
	}
	if ns.QueryStrategy == nil {
		ns.QueryStrategy = &dns.QueryStrategy
	}
	if ns.CacheStrategy == nil {
		ns.CacheStrategy = new(CacheStrategy)
		switch {
		case dns.CacheStrategy != CacheStrategy_CacheEnabled:
			*ns.CacheStrategy = dns.CacheStrategy
		case dns.DisableCache:
			features.PrintDeprecatedFeatureWarning("DNS disableCache settings")
			*ns.CacheStrategy = CacheStrategy_CacheDisabled
		}
	}
	if ns.FallbackStrategy == nil {
		ns.FallbackStrategy = new(FallbackStrategy)
		switch {
		case ns.SkipFallback:
			features.PrintDeprecatedFeatureWarning("DNS server skipFallback settings")
			*ns.FallbackStrategy = FallbackStrategy_Disabled
		case dns.FallbackStrategy != FallbackStrategy_Enabled:
			*ns.FallbackStrategy = dns.FallbackStrategy
		case dns.DisableFallback:
			features.PrintDeprecatedFeatureWarning("DNS disableFallback settings")
			*ns.FallbackStrategy = FallbackStrategy_Disabled
		case dns.DisableFallbackIfMatch:
			features.PrintDeprecatedFeatureWarning("DNS disableFallbackIfMatch settings")
			*ns.FallbackStrategy = FallbackStrategy_DisabledIfAnyMatch
		}
	}
	if (ns.FakeDns != nil && len(ns.FakeDns.Pools) == 0) || // Use globally configured fake ip pool if: 1. `fakedns` config is set, but empty(represents { "fakedns": true } in JSON settings);
		ns.FakeDns == nil && strings.EqualFold(ns.Address.Address.GetDomain(), "fakedns") { // 2. `fakedns` config not set, but server address is `fakedns`(represents { "address": "fakedns" } in JSON settings).
		if dns.FakeDns != nil {
			ns.FakeDns = dns.FakeDns
		} else {
			ns.FakeDns = &fakedns.FakeDnsPoolMulti{}
			queryStrategy := toIPOption(*ns.QueryStrategy)
			if queryStrategy.IPv4Enable {
				ns.FakeDns.Pools = append(ns.FakeDns.Pools, &fakedns.FakeDnsPool{
					IpPool:  "198.18.0.0/15",
					LruSize: 65535,
				})
			}
			if queryStrategy.IPv6Enable {
				ns.FakeDns.Pools = append(ns.FakeDns.Pools, &fakedns.FakeDnsPool{
					IpPool:  "fc00::/18",
					LruSize: 65535,
				})
			}
		}
	}

	// Priotize local domains with specific TLDs or without any dot to local DNS
	if strings.EqualFold(ns.Address.Address.GetDomain(), "localhost") {
		ns.PrioritizedDomain = append(ns.PrioritizedDomain, localTLDsAndDotlessDomains...)
		ns.OriginalRules = append(ns.OriginalRules, localTLDsAndDotlessDomainsRule)
	}

	if len(ns.ClientIp) > 0 {
		newError("DNS: client ", ns.Address.Address.AsAddress(), " uses clientIP ", net.IP(ns.ClientIp).String()).AtInfo().WriteToLog()
	}

	client.clientIP = ns.ClientIp
	client.tag = ns.Tag
	client.queryStrategy = toIPOption(*ns.QueryStrategy)
	client.cacheStrategy = *ns.CacheStrategy
	client.fallbackStrategy = *ns.FallbackStrategy
	return client, nil
}

// Name returns the server name the client manages.
func (c *Client) Name() string {
	return c.server.Name()
}

// QueryIP send DNS query to the name server with the client's IP and IP options.
func (c *Client) QueryIP(ctx context.Context, domain string, option dns.IPOption) ([]net.IP, error) {
	queryOption := option.With(c.queryStrategy)
	if !queryOption.IsValid() {
		newError(c.server.Name(), " returns empty answer: ", domain, ". ", toReqTypes(option)).AtInfo().WriteToLog()
		return nil, dns.ErrEmptyResponse
	}
	server := c.server
	if queryOption.FakeEnable && c.fakeDNS != nil {
		server = c.fakeDNS
	}
	disableCache := c.cacheStrategy == CacheStrategy_CacheDisabled

	ctx = session.ContextWithInbound(ctx, &session.Inbound{Tag: c.tag})
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	ips, err := server.QueryIP(ctx, domain, c.clientIP, queryOption, disableCache)
	cancel()

	if err != nil || queryOption.FakeEnable {
		return ips, err
	}
	return c.MatchExpectedIPs(domain, ips)
}

// MatchExpectedIPs matches queried domain IPs with expected IPs and returns matched ones.
func (c *Client) MatchExpectedIPs(domain string, ips []net.IP) ([]net.IP, error) {
	if len(c.expectIPs) == 0 {
		return ips, nil
	}
	newIps := []net.IP{}
	for _, ip := range ips {
		for _, matcher := range c.expectIPs {
			if matcher.Match(ip) {
				newIps = append(newIps, ip)
				break
			}
		}
	}
	if len(newIps) == 0 {
		return nil, errExpectedIPNonMatch
	}
	newError("domain ", domain, " expectIPs ", newIps, " matched at server ", c.Name()).AtDebug().WriteToLog()
	return newIps, nil
}
