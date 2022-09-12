package dns

import (
	"context"
	"net/url"
	"strings"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
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
}

var errExpectedIPNonMatch = errors.New("expectIPs not match")

// NewServer creates a name server object according to the network destination url.
func NewServer(dest net.Destination, dispatcher routing.Dispatcher) (Server, error) {
	if address := dest.Address; address.Family().IsDomain() {
		u, err := url.Parse(address.Domain())
		if err != nil {
			return nil, err
		}
		switch {
		case strings.EqualFold(u.String(), "localhost"):
			return NewLocalNameServer(), nil
		case strings.EqualFold(u.Scheme, "https"): // DOH Remote mode
			return NewDoHNameServer(u, dispatcher)
		case strings.EqualFold(u.Scheme, "https+local"): // DOH Local mode
			return NewDoHLocalNameServer(u), nil
		case strings.EqualFold(u.Scheme, "quic+local"): // DNS-over-QUIC Local mode
			return NewQUICNameServer(u)
		case strings.EqualFold(u.Scheme, "tcp"): // DNS-over-TCP Remote mode
			return NewTCPNameServer(u, dispatcher)
		case strings.EqualFold(u.Scheme, "tcp+local"): // DNS-over-TCP Local mode
			return NewTCPLocalNameServer(u)
		case strings.EqualFold(u.String(), "fakedns"):
			return NewFakeDNSServer(), nil
		}
	}
	if dest.Network == net.Network_Unknown {
		dest.Network = net.Network_UDP
	}
	if dest.Network == net.Network_UDP { // UDP classic DNS mode
		return NewClassicNameServer(dest, dispatcher), nil
	}
	return nil, newError("No available name server could be created from ", dest).AtWarning()
}

// NewClient creates a DNS client managing a name server with client IP, domain rules and expected IPs.
func NewClient(ctx context.Context, ns *NameServer, dns *Config) (*Client, error) {
	client := &Client{}

	// Create DNS server instance
	err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error {
		// Create a new server for each client for now
		server, err := NewServer(ns.Address.AsDestination(), dispatcher)
		if err != nil {
			return newError("failed to create nameserver").Base(err).AtWarning()
		}
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
			*ns.CacheStrategy = CacheStrategy_CacheDisabled
		}
	}
	if ns.FallbackStrategy == nil {
		ns.FallbackStrategy = new(FallbackStrategy)
		switch {
		case ns.SkipFallback:
			*ns.FallbackStrategy = FallbackStrategy_Disabled
		case dns.FallbackStrategy != FallbackStrategy_Enabled:
			*ns.FallbackStrategy = dns.FallbackStrategy
		case dns.DisableFallback:
			*ns.FallbackStrategy = FallbackStrategy_Disabled
		case dns.DisableFallbackIfMatch:
			*ns.FallbackStrategy = FallbackStrategy_DisabledIfAnyMatch
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

// QueryIP send DNS query to the name server with the client's IP.
func (c *Client) QueryIP(ctx context.Context, domain string, option dns.IPOption) ([]net.IP, error) {
	queryOption := option.With(c.queryStrategy)
	if !queryOption.IsValid() {
		newError(c.server.Name(), " returns empty answer: ", domain, ". ", toReqTypes(option)).AtInfo().WriteToLog()
		return nil, dns.ErrEmptyResponse
	}
	disableCache := c.cacheStrategy == CacheStrategy_CacheDisabled

	ctx = session.ContextWithInbound(ctx, &session.Inbound{Tag: c.tag})
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	ips, err := c.server.QueryIP(ctx, domain, c.clientIP, queryOption, disableCache)
	cancel()

	if err != nil {
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
