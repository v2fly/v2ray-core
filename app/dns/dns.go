//go:build !confonly
// +build !confonly

// Package dns is an implementation of core.DNS feature.
package dns

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"fmt"
	"strings"
	"sync"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
)

// DNS is a DNS rely server.
type DNS struct {
	sync.Mutex
	hosts         *StaticHosts
	clients       []*Client
	ctx           context.Context
	clientTags    map[string]bool
	fakeDNSEngine *FakeDNSEngine
	domainMatcher strmatcher.IndexMatcher
	matcherInfos  []DomainMatcherInfo
}

// DomainMatcherInfo contains information attached to index returned by Server.domainMatcher
type DomainMatcherInfo struct {
	clientIdx     uint16
	domainRuleIdx uint16
}

// New creates a new DNS server with given configuration.
func New(ctx context.Context, config *Config) (*DNS, error) {
	// Create static hosts
	hosts, err := NewStaticHosts(config.StaticHosts, config.Hosts)
	if err != nil {
		return nil, newError("failed to create hosts").Base(err)
	}

	// Create name servers from legacy configs
	clients := []*Client{}
	for _, endpoint := range config.NameServers {
		features.PrintDeprecatedFeatureWarning("simple DNS server")
		client, err := NewClient(ctx, &NameServer{Address: endpoint}, config)
		if err != nil {
			return nil, newError("failed to create client").Base(err)
		}
		clients = append(clients, client)
	}

	// Create name servers
	nsClientMap := map[int]int{}
	for nsIdx, ns := range config.NameServer {
		client, err := NewClient(ctx, ns, config)
		if err != nil {
			return nil, newError("failed to create client").Base(err)
		}
		nsClientMap[nsIdx] = len(clients)
		clients = append(clients, client)
	}

	// If there is no DNS client in config, add a `localhost` DNS client
	if len(clients) == 0 {
		clients = append(clients, NewLocalDNSClient())
	}

	s := &DNS{
		hosts:   hosts,
		clients: clients,
		ctx:     ctx,
	}

	// Establish members related to global DNS state
	s.clientTags = make(map[string]bool)
	for _, client := range clients {
		s.clientTags[client.tag] = true
	}
	if err := establishDomainRules(s, config, nsClientMap); err != nil {
		return nil, err
	}
	if err := establishExpectedIPs(s, config, nsClientMap); err != nil {
		return nil, err
	}
	if err := establishFakeDNS(s, config, nsClientMap); err != nil {
		return nil, err
	}

	return s, nil
}

func establishDomainRules(s *DNS, config *Config, nsClientMap map[int]int) error {
	domainRuleCount := 0
	for _, ns := range config.NameServer {
		domainRuleCount += len(ns.PrioritizedDomain)
	}
	// MatcherInfos is ensured to cover the maximum index domainMatcher could return, where matcher's index starts from 1
	matcherInfos := make([]DomainMatcherInfo, domainRuleCount+1)
	var domainMatcher strmatcher.IndexMatcher
	switch config.DomainMatcher {
	case "mph", "hybrid":
		newError("using mph domain matcher").AtDebug().WriteToLog()
		domainMatcher = strmatcher.NewMphIndexMatcher()
	case "linear":
		fallthrough
	default:
		newError("using default domain matcher").AtDebug().WriteToLog()
		domainMatcher = strmatcher.NewLinearIndexMatcher()
	}
	for nsIdx, ns := range config.NameServer {
		clientIdx := nsClientMap[nsIdx]
		var rules []string
		ruleCurr := 0
		ruleIter := 0
		for _, domain := range ns.PrioritizedDomain {
			domainRule, err := toStrMatcher(domain.Type, domain.Domain)
			if err != nil {
				return newError("failed to create prioritized domain").Base(err).AtWarning()
			}
			originalRuleIdx := ruleCurr
			if ruleCurr < len(ns.OriginalRules) {
				rule := ns.OriginalRules[ruleCurr]
				if ruleCurr >= len(rules) {
					rules = append(rules, rule.Rule)
				}
				ruleIter++
				if ruleIter >= int(rule.Size) {
					ruleIter = 0
					ruleCurr++
				}
			} else { // No original rule, generate one according to current domain matcher (majorly for compatibility with tests)
				rules = append(rules, domainRule.String())
				ruleCurr++
			}
			midx := domainMatcher.Add(domainRule)
			matcherInfos[midx] = DomainMatcherInfo{
				clientIdx:     uint16(clientIdx),
				domainRuleIdx: uint16(originalRuleIdx),
			}
			if err != nil {
				return newError("failed to create prioritized domain").Base(err).AtWarning()
			}
		}
		s.clients[clientIdx].domains = rules
	}
	if err := domainMatcher.Build(); err != nil {
		return err
	}
	s.domainMatcher = domainMatcher
	s.matcherInfos = matcherInfos
	return nil
}

func establishExpectedIPs(s *DNS, config *Config, nsClientMap map[int]int) error {
	geoipContainer := router.GeoIPMatcherContainer{}
	for nsIdx, ns := range config.NameServer {
		clientIdx := nsClientMap[nsIdx]
		var matchers []*router.GeoIPMatcher
		for _, geoip := range ns.Geoip {
			matcher, err := geoipContainer.Add(geoip)
			if err != nil {
				return newError("failed to create ip matcher").Base(err).AtWarning()
			}
			matchers = append(matchers, matcher)
		}
		s.clients[clientIdx].expectIPs = matchers
	}
	return nil
}

func establishFakeDNS(s *DNS, config *Config, nsClientMap map[int]int) error {
	fakeHolders := &fakedns.HolderMulti{}
	fakeDefault := (*fakedns.HolderMulti)(nil)
	if config.FakeDns != nil {
		defaultEngine, err := fakeHolders.AddPoolMulti(config.FakeDns)
		if err != nil {
			return newError("fail to create fake dns").Base(err).AtWarning()
		}
		fakeDefault = defaultEngine
	}
	for nsIdx, ns := range config.NameServer {
		clientIdx := nsClientMap[nsIdx]
		if ns.FakeDns == nil {
			continue
		}
		engine, err := fakeHolders.AddPoolMulti(ns.FakeDns)
		if err != nil {
			return newError("fail to create fake dns").Base(err).AtWarning()
		}
		s.clients[clientIdx].fakeDNS = NewFakeDNSServer(engine)
		s.clients[clientIdx].queryStrategy.FakeEnable = true
	}
	// Do not create FakeDNSEngine feature if no FakeDNS server is configured
	if fakeHolders.IsEmpty() {
		return nil
	}
	// Add FakeDNSEngine feature when DNS feature is added for the first time
	s.fakeDNSEngine = &FakeDNSEngine{dns: s, fakeHolders: fakeHolders, fakeDefault: fakeDefault}
	return core.RequireFeatures(s.ctx, func(client dns.Client) error {
		v := core.MustFromContext(s.ctx)
		if v.GetFeature(dns.FakeDNSEngineType()) != nil {
			return nil
		}
		if client, ok := client.(dns.ClientWithFakeDNS); ok {
			return v.AddFeature(client.AsFakeDNSEngine())
		}
		return nil
	})
}

// Type implements common.HasType.
func (*DNS) Type() interface{} {
	return dns.ClientType()
}

// Start implements common.Runnable.
func (s *DNS) Start() error {
	return nil
}

// Close implements common.Closable.
func (s *DNS) Close() error {
	return nil
}

// IsOwnLink implements proxy.dns.ownLinkVerifier
func (s *DNS) IsOwnLink(ctx context.Context) bool {
	inbound := session.InboundFromContext(ctx)
	return inbound != nil && s.clientTags[inbound.Tag]
}

// AsFakeDNSClient implements dns.ClientWithFakeDNS.
func (s *DNS) AsFakeDNSClient() dns.Client {
	return &FakeDNSClient{DNS: s}
}

// AsFakeDNSEngine implements dns.ClientWithFakeDNS.
func (s *DNS) AsFakeDNSEngine() dns.FakeDNSEngine {
	return s.fakeDNSEngine
}

// LookupIP implements dns.Client.
func (s *DNS) LookupIP(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv4Enable: true, IPv6Enable: true, FakeEnable: false})
}

// LookupIPv4 implements dns.IPv4Lookup.
func (s *DNS) LookupIPv4(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv4Enable: true, FakeEnable: false})
}

// LookupIPv6 implements dns.IPv6Lookup.
func (s *DNS) LookupIPv6(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv6Enable: true, FakeEnable: false})
}

func (s *DNS) lookupIPInternal(domain string, option dns.IPOption) ([]net.IP, error) {
	if domain == "" {
		return nil, newError("empty domain name")
	}

	// Normalize the FQDN form query
	domain = strings.TrimSuffix(domain, ".")

	// Static host lookup
	switch addrs := s.hosts.Lookup(domain, option); {
	case addrs == nil: // Domain not recorded in static host
		break
	case len(addrs) == 0: // Domain recorded, but no valid IP returned (e.g. IPv4 address with only IPv6 enabled)
		return nil, dns.ErrEmptyResponse
	case len(addrs) == 1 && addrs[0].Family().IsDomain(): // Domain replacement
		newError("domain replaced: ", domain, " -> ", addrs[0].Domain()).WriteToLog()
		domain = addrs[0].Domain()
	default: // Successfully found ip records in static host
		newError("returning ", len(addrs), " IP(s) for domain ", domain, " -> ", addrs).WriteToLog()
		return toNetIP(addrs)
	}

	// Name servers lookup
	errs := []error{}
	for _, client := range s.sortClients(domain, option) {
		ips, err := client.QueryIP(s.ctx, domain, option)
		if len(ips) > 0 {
			return ips, nil
		}
		if err != nil {
			errs = append(errs, err)
		}
		if err != dns.ErrEmptyResponse { // ErrEmptyResponse is not seen as failure, so no failed log
			newError("failed to lookup ip for domain ", domain, " at server ", client.Name()).Base(err).WriteToLog()
		}
		if err != context.Canceled && err != context.DeadlineExceeded && err != errExpectedIPNonMatch {
			return nil, err // Only continue lookup for certain errors
		}
	}

	if len(errs) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return nil, newError("returning nil for domain ", domain).Base(errors.Combine(errs...))
}

func (s *DNS) sortClients(domain string, option dns.IPOption) []*Client {
	clients := make([]*Client, 0, len(s.clients))
	clientUsed := make([]bool, len(s.clients))
	clientIdxs := make([]int, 0, len(s.clients))
	domainRules := []string{}

	// Priority domain matching
	for _, match := range s.domainMatcher.Match(domain) {
		info := s.matcherInfos[match]
		client := s.clients[info.clientIdx]
		domainRule := client.domains[info.domainRuleIdx]
		domainRules = append(domainRules, fmt.Sprintf("%s(DNS idx:%d)", domainRule, info.clientIdx))
		switch {
		case clientUsed[info.clientIdx]:
			continue
		case !option.FakeEnable && isFakeDNS(client.server):
			continue
		}
		clientUsed[info.clientIdx] = true
		clients = append(clients, client)
		clientIdxs = append(clientIdxs, int(info.clientIdx))
	}

	// Default round-robin query
	hasDomainMatch := len(clients) > 0
	for idx, client := range s.clients {
		switch {
		case clientUsed[idx]:
			continue
		case !option.FakeEnable && isFakeDNS(client.server):
			continue
		case client.fallbackStrategy == FallbackStrategy_Disabled:
			continue
		case client.fallbackStrategy == FallbackStrategy_DisabledIfAnyMatch && hasDomainMatch:
			continue
		}
		clientUsed[idx] = true
		clients = append(clients, client)
		clientIdxs = append(clientIdxs, idx)
	}

	if len(domainRules) > 0 {
		newError("domain ", domain, " matches following rules: ", domainRules).AtDebug().WriteToLog()
	}
	if len(clientIdxs) > 0 {
		newError("domain ", domain, " will use DNS in order: ", s.formatClientNames(clientIdxs, option), " ", toReqTypes(option)).AtDebug().WriteToLog()
	}

	return clients
}

func (s *DNS) formatClientNames(clientIdxs []int, option dns.IPOption) []string {
	clientNames := make([]string, 0, len(clientIdxs))
	counter := make(map[string]uint, len(clientIdxs))
	for _, clientIdx := range clientIdxs {
		client := s.clients[clientIdx]
		var name string
		if option.With(client.queryStrategy).FakeEnable {
			name = fmt.Sprintf("%s(DNS idx:%d)", client.fakeDNS.Name(), clientIdx)
		} else {
			name = client.Name()
		}
		counter[name]++
		clientNames = append(clientNames, name)
	}
	for idx, clientIdx := range clientIdxs {
		name := clientNames[idx]
		if counter[name] > 1 {
			clientNames[idx] = fmt.Sprintf("%s(DNS idx:%d)", name, clientIdx)
		}
	}
	return clientNames
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))

	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		ctx = cfgcommon.NewConfigureLoadingContext(ctx)

		geoloadername := platform.NewEnvFlag("v2ray.conf.geoloader").GetValue(func() string {
			return "standard"
		})

		if loader, err := geodata.GetGeoDataLoader(geoloadername); err == nil {
			cfgcommon.SetGeoDataLoader(ctx, loader)
		} else {
			return nil, newError("unable to create geo data loader ").Base(err)
		}

		cfgEnv := cfgcommon.GetConfigureLoadingEnvironment(ctx)
		geoLoader := cfgEnv.GetGeoLoader()

		simplifiedConfig := config.(*SimplifiedConfig)
		for _, v := range simplifiedConfig.NameServer {
			for _, geo := range v.Geoip {
				if geo.Code != "" {
					filepath := "geoip.dat"
					if geo.FilePath != "" {
						filepath = geo.FilePath
					} else {
						geo.CountryCode = geo.Code
					}
					var err error
					geo.Cidr, err = geoLoader.LoadIP(filepath, geo.Code)
					if err != nil {
						return nil, newError("unable to load geoip").Base(err)
					}
				}
			}
		}

		var nameservers []*NameServer

		for _, v := range simplifiedConfig.NameServer {
			nameserver := &NameServer{
				Address:          v.Address,
				ClientIp:         net.ParseIP(v.ClientIp),
				Tag:              v.Tag,
				QueryStrategy:    v.QueryStrategy,
				CacheStrategy:    v.CacheStrategy,
				FallbackStrategy: v.FallbackStrategy,
				SkipFallback:     v.SkipFallback,
				Geoip:            v.Geoip,
			}
			for _, prioritizedDomain := range v.PrioritizedDomain {
				nameserver.PrioritizedDomain = append(nameserver.PrioritizedDomain, &NameServer_PriorityDomain{
					Type:   prioritizedDomain.Type,
					Domain: prioritizedDomain.Domain,
				})
			}
			nameservers = append(nameservers, nameserver)
		}

		var statichosts []*HostMapping

		for _, v := range simplifiedConfig.StaticHosts {
			statichost := &HostMapping{
				Type:          v.Type,
				Domain:        v.Domain,
				ProxiedDomain: v.ProxiedDomain,
			}
			for _, ip := range v.Ip {
				statichost.Ip = append(statichost.Ip, net.ParseIP(ip))
			}
			statichosts = append(statichosts, statichost)
		}

		fullConfig := &Config{
			StaticHosts:      statichosts,
			NameServer:       nameservers,
			ClientIp:         net.ParseIP(simplifiedConfig.ClientIp),
			Tag:              simplifiedConfig.Tag,
			DomainMatcher:    simplifiedConfig.DomainMatcher,
			QueryStrategy:    simplifiedConfig.QueryStrategy,
			CacheStrategy:    simplifiedConfig.CacheStrategy,
			FallbackStrategy: simplifiedConfig.FallbackStrategy,
			// Deprecated flags
			DisableCache:           simplifiedConfig.DisableCache,
			DisableFallback:        simplifiedConfig.DisableFallback,
			DisableFallbackIfMatch: simplifiedConfig.DisableFallbackIfMatch,
		}
		return common.CreateObject(ctx, fullConfig)
	}))
}
