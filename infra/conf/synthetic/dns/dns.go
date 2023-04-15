package dns

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/dns"
	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
	rule2 "github.com/v2fly/v2ray-core/v5/infra/conf/rule"
)

type NameServerConfig struct {
	Address          *cfgcommon.Address
	ClientIP         *cfgcommon.Address
	Port             uint16
	Tag              string
	QueryStrategy    string
	CacheStrategy    string
	FallbackStrategy string
	SkipFallback     bool
	Domains          []string
	ExpectIPs        cfgcommon.StringList
	FakeDNS          FakeDNSConfigExtend

	cfgctx context.Context
}

func (c *NameServerConfig) UnmarshalJSON(data []byte) error {
	var address cfgcommon.Address
	if err := json.Unmarshal(data, &address); err == nil {
		c.Address = &address
		return nil
	}

	var advanced struct {
		Address          *cfgcommon.Address   `json:"address"`
		ClientIP         *cfgcommon.Address   `json:"clientIp"`
		Port             uint16               `json:"port"`
		Tag              string               `json:"tag"`
		QueryStrategy    string               `json:"queryStrategy"`
		CacheStrategy    string               `json:"cacheStrategy"`
		FallbackStrategy string               `json:"fallbackStrategy"`
		SkipFallback     bool                 `json:"skipFallback"`
		Domains          []string             `json:"domains"`
		ExpectIPs        cfgcommon.StringList `json:"expectIps"`
		FakeDNS          FakeDNSConfigExtend  `json:"fakedns"`
	}
	if err := json.Unmarshal(data, &advanced); err == nil {
		c.Address = advanced.Address
		c.ClientIP = advanced.ClientIP
		c.Port = advanced.Port
		c.Tag = advanced.Tag
		c.QueryStrategy = advanced.QueryStrategy
		c.CacheStrategy = advanced.CacheStrategy
		c.FallbackStrategy = advanced.FallbackStrategy
		c.SkipFallback = advanced.SkipFallback
		c.Domains = advanced.Domains
		c.ExpectIPs = advanced.ExpectIPs
		c.FakeDNS = advanced.FakeDNS
		return nil
	}

	return newError("failed to parse name server: ", string(data))
}

func toDomainMatchingType(t routercommon.Domain_Type) dns.DomainMatchingType {
	switch t {
	case routercommon.Domain_RootDomain:
		return dns.DomainMatchingType_Subdomain
	case routercommon.Domain_Full:
		return dns.DomainMatchingType_Full
	case routercommon.Domain_Plain:
		return dns.DomainMatchingType_Keyword
	case routercommon.Domain_Regex:
		return dns.DomainMatchingType_Regex
	default:
		panic("unknown domain type")
	}
}

func (c *NameServerConfig) BuildV5(ctx context.Context) (*dns.NameServer, error) {
	c.cfgctx = ctx
	return c.Build()
}

func (c *NameServerConfig) Build() (*dns.NameServer, error) {
	cfgctx := c.cfgctx

	if c.Address == nil {
		return nil, newError("NameServer address is not specified.")
	}

	var domains []*dns.NameServer_PriorityDomain
	var originalRules []*dns.NameServer_OriginalRule

	for _, rule := range c.Domains {
		parsedDomain, err := rule2.ParseDomainRule(cfgctx, rule)
		if err != nil {
			return nil, newError("invalid domain rule: ", rule).Base(err)
		}

		for _, pd := range parsedDomain {
			domains = append(domains, &dns.NameServer_PriorityDomain{
				Type:   toDomainMatchingType(pd.Type),
				Domain: pd.Value,
			})
		}
		originalRules = append(originalRules, &dns.NameServer_OriginalRule{
			Rule: rule,
			Size: uint32(len(parsedDomain)),
		})
	}

	geoipList, err := rule2.ToCidrList(cfgctx, c.ExpectIPs)
	if err != nil {
		return nil, newError("invalid IP rule: ", c.ExpectIPs).Base(err)
	}

	var fakeDNS *fakedns.FakeDnsPoolMulti
	if c.FakeDNS.FakeDNSConfig != nil {
		fake, err := c.FakeDNS.FakeDNSConfig.Build()
		if err != nil {
			return nil, newError("failed to build fakedns").Base(err)
		}
		fakeDNS = fake
	}

	var myClientIP []byte
	if c.ClientIP != nil {
		if !c.ClientIP.Family().IsIP() {
			return nil, newError("not an IP address:", c.ClientIP.String())
		}
		myClientIP = []byte(c.ClientIP.IP())
	}

	queryStrategy := new(dns.QueryStrategy)
	switch strings.ToLower(c.QueryStrategy) {
	case "useip", "use_ip", "use-ip":
		*queryStrategy = dns.QueryStrategy_USE_IP
	case "useip4", "useipv4", "use_ip4", "use_ipv4", "use_ip_v4", "use-ip4", "use-ipv4", "use-ip-v4":
		*queryStrategy = dns.QueryStrategy_USE_IP4
	case "useip6", "useipv6", "use_ip6", "use_ipv6", "use_ip_v6", "use-ip6", "use-ipv6", "use-ip-v6":
		*queryStrategy = dns.QueryStrategy_USE_IP6
	default:
		queryStrategy = nil
	}

	cacheStrategy := new(dns.CacheStrategy)
	switch strings.ToLower(c.CacheStrategy) {
	case "enabled":
		*cacheStrategy = dns.CacheStrategy_CacheEnabled
	case "disabled":
		*cacheStrategy = dns.CacheStrategy_CacheDisabled
	default:
		cacheStrategy = nil
	}

	fallbackStrategy := new(dns.FallbackStrategy)
	switch strings.ToLower(c.FallbackStrategy) {
	case "enabled":
		*fallbackStrategy = dns.FallbackStrategy_Enabled
	case "disabled":
		*fallbackStrategy = dns.FallbackStrategy_Disabled
	case "disabledifanymatch", "disabled_if_any_match", "disabled-if-any-match":
		*fallbackStrategy = dns.FallbackStrategy_DisabledIfAnyMatch
	default:
		fallbackStrategy = nil
	}

	return &dns.NameServer{
		Address: &net.Endpoint{
			Network: net.Network_UDP,
			Address: c.Address.Build(),
			Port:    uint32(c.Port),
		},
		ClientIp:          myClientIP,
		Tag:               c.Tag,
		SkipFallback:      c.SkipFallback,
		QueryStrategy:     queryStrategy,
		CacheStrategy:     cacheStrategy,
		FallbackStrategy:  fallbackStrategy,
		PrioritizedDomain: domains,
		Geoip:             geoipList,
		OriginalRules:     originalRules,
		FakeDns:           fakeDNS,
	}, nil
}

var typeMap = map[routercommon.Domain_Type]dns.DomainMatchingType{
	routercommon.Domain_Full:       dns.DomainMatchingType_Full,
	routercommon.Domain_RootDomain: dns.DomainMatchingType_Subdomain,
	routercommon.Domain_Plain:      dns.DomainMatchingType_Keyword,
	routercommon.Domain_Regex:      dns.DomainMatchingType_Regex,
}

// DNSConfig is a JSON serializable object for dns.Config.
type DNSConfig struct { // nolint: revive
	Servers                []*NameServerConfig     `json:"servers"`
	Hosts                  map[string]*HostAddress `json:"hosts"`
	FakeDNS                *FakeDNSConfig          `json:"fakedns"`
	DomainMatcher          string                  `json:"domainMatcher"`
	ClientIP               *cfgcommon.Address      `json:"clientIp"`
	Tag                    string                  `json:"tag"`
	QueryStrategy          string                  `json:"queryStrategy"`
	CacheStrategy          string                  `json:"cacheStrategy"`
	FallbackStrategy       string                  `json:"fallbackStrategy"`
	DisableCache           bool                    `json:"disableCache"`
	DisableFallback        bool                    `json:"disableFallback"`
	DisableFallbackIfMatch bool                    `json:"disableFallbackIfMatch"`
	cfgctx                 context.Context
}

type HostAddress struct {
	addr  *cfgcommon.Address
	addrs []*cfgcommon.Address
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON
func (h *HostAddress) UnmarshalJSON(data []byte) error {
	addr := new(cfgcommon.Address)
	var addrs []*cfgcommon.Address
	switch {
	case json.Unmarshal(data, &addr) == nil:
		h.addr = addr
	case json.Unmarshal(data, &addrs) == nil:
		h.addrs = addrs
	default:
		return newError("invalid address")
	}
	return nil
}

func getHostMapping(ha *HostAddress) *dns.HostMapping {
	if ha.addr != nil {
		if ha.addr.Family().IsDomain() {
			return &dns.HostMapping{
				ProxiedDomain: ha.addr.Domain(),
			}
		}
		return &dns.HostMapping{
			Ip: [][]byte{ha.addr.IP()},
		}
	}

	ips := make([][]byte, 0, len(ha.addrs))
	for _, addr := range ha.addrs {
		if addr.Family().IsDomain() {
			return &dns.HostMapping{
				ProxiedDomain: addr.Domain(),
			}
		}
		ips = append(ips, []byte(addr.IP()))
	}
	return &dns.HostMapping{
		Ip: ips,
	}
}

func (c *DNSConfig) BuildV5(ctx context.Context) (*dns.Config, error) {
	c.cfgctx = ctx
	return c.Build()
}

// Build implements Buildable
func (c *DNSConfig) Build() (*dns.Config, error) {
	if c.cfgctx == nil {
		c.cfgctx = cfgcommon.NewConfigureLoadingContext(context.Background())

		geoloadername := platform.NewEnvFlag("v2ray.conf.geoloader").GetValue(func() string {
			return "standard"
		})

		if loader, err := geodata.GetGeoDataLoader(geoloadername); err == nil {
			cfgcommon.SetGeoDataLoader(c.cfgctx, loader)
		} else {
			return nil, newError("unable to create geo data loader ").Base(err)
		}
	}

	cfgEnv := cfgcommon.GetConfigureLoadingEnvironment(c.cfgctx)
	geoLoader := cfgEnv.GetGeoLoader()

	config := &dns.Config{
		Tag:                    c.Tag,
		DisableCache:           c.DisableCache,
		DisableFallback:        c.DisableFallback,
		DisableFallbackIfMatch: c.DisableFallbackIfMatch,
		DomainMatcher:          c.DomainMatcher,
	}

	if c.ClientIP != nil {
		if !c.ClientIP.Family().IsIP() {
			return nil, newError("not an IP address:", c.ClientIP.String())
		}
		config.ClientIp = []byte(c.ClientIP.IP())
	}

	config.QueryStrategy = dns.QueryStrategy_USE_IP
	switch strings.ToLower(c.QueryStrategy) {
	case "useip", "use_ip", "use-ip":
		config.QueryStrategy = dns.QueryStrategy_USE_IP
	case "useip4", "useipv4", "use_ip4", "use_ipv4", "use_ip_v4", "use-ip4", "use-ipv4", "use-ip-v4":
		config.QueryStrategy = dns.QueryStrategy_USE_IP4
	case "useip6", "useipv6", "use_ip6", "use_ipv6", "use_ip_v6", "use-ip6", "use-ipv6", "use-ip-v6":
		config.QueryStrategy = dns.QueryStrategy_USE_IP6
	}

	config.CacheStrategy = dns.CacheStrategy_CacheEnabled
	switch strings.ToLower(c.CacheStrategy) {
	case "enabled":
		config.CacheStrategy = dns.CacheStrategy_CacheEnabled
	case "disabled":
		config.CacheStrategy = dns.CacheStrategy_CacheDisabled
	}

	config.FallbackStrategy = dns.FallbackStrategy_Enabled
	switch strings.ToLower(c.FallbackStrategy) {
	case "enabled":
		config.FallbackStrategy = dns.FallbackStrategy_Enabled
	case "disabled":
		config.FallbackStrategy = dns.FallbackStrategy_Disabled
	case "disabledifanymatch", "disabled_if_any_match", "disabled-if-any-match":
		config.FallbackStrategy = dns.FallbackStrategy_DisabledIfAnyMatch
	}

	for _, server := range c.Servers {
		server.cfgctx = c.cfgctx
		ns, err := server.Build()
		if err != nil {
			return nil, newError("failed to build nameserver").Base(err)
		}
		config.NameServer = append(config.NameServer, ns)
	}

	if c.Hosts != nil {
		mappings := make([]*dns.HostMapping, 0, 20)

		domains := make([]string, 0, len(c.Hosts))
		for domain := range c.Hosts {
			domains = append(domains, domain)
		}
		sort.Strings(domains)

		for _, domain := range domains {
			switch {
			case strings.HasPrefix(domain, "domain:"):
				domainName := domain[7:]
				if len(domainName) == 0 {
					return nil, newError("empty domain type of rule: ", domain)
				}
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Subdomain
				mapping.Domain = domainName
				mappings = append(mappings, mapping)

			case strings.HasPrefix(domain, "geosite:"):
				listName := domain[8:]
				if len(listName) == 0 {
					return nil, newError("empty geosite rule: ", domain)
				}
				geositeList, err := geoLoader.LoadGeoSite(listName)
				if err != nil {
					return nil, newError("failed to load geosite: ", listName).Base(err)
				}
				for _, d := range geositeList {
					mapping := getHostMapping(c.Hosts[domain])
					mapping.Type = typeMap[d.Type]
					mapping.Domain = d.Value
					mappings = append(mappings, mapping)
				}

			case strings.HasPrefix(domain, "regexp:"):
				regexpVal := domain[7:]
				if len(regexpVal) == 0 {
					return nil, newError("empty regexp type of rule: ", domain)
				}
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Regex
				mapping.Domain = regexpVal
				mappings = append(mappings, mapping)

			case strings.HasPrefix(domain, "keyword:"):
				keywordVal := domain[8:]
				if len(keywordVal) == 0 {
					return nil, newError("empty keyword type of rule: ", domain)
				}
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Keyword
				mapping.Domain = keywordVal
				mappings = append(mappings, mapping)

			case strings.HasPrefix(domain, "full:"):
				fullVal := domain[5:]
				if len(fullVal) == 0 {
					return nil, newError("empty full domain type of rule: ", domain)
				}
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Full
				mapping.Domain = fullVal
				mappings = append(mappings, mapping)

			case strings.HasPrefix(domain, "dotless:"):
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Regex
				switch substr := domain[8:]; {
				case substr == "":
					mapping.Domain = "^[^.]*$"
				case !strings.Contains(substr, "."):
					mapping.Domain = "^[^.]*" + substr + "[^.]*$"
				default:
					return nil, newError("substr in dotless rule should not contain a dot: ", substr)
				}
				mappings = append(mappings, mapping)

			case strings.HasPrefix(domain, "ext:"):
				kv := strings.Split(domain[4:], ":")
				if len(kv) != 2 {
					return nil, newError("invalid external resource: ", domain)
				}
				filename := kv[0]
				list := kv[1]
				geositeList, err := geoLoader.LoadGeoSiteWithAttr(filename, list)
				if err != nil {
					return nil, newError("failed to load domain list: ", list, " from ", filename).Base(err)
				}
				for _, d := range geositeList {
					mapping := getHostMapping(c.Hosts[domain])
					mapping.Type = typeMap[d.Type]
					mapping.Domain = d.Value
					mappings = append(mappings, mapping)
				}

			default:
				mapping := getHostMapping(c.Hosts[domain])
				mapping.Type = dns.DomainMatchingType_Full
				mapping.Domain = domain
				mappings = append(mappings, mapping)
			}
		}

		config.StaticHosts = append(config.StaticHosts, mappings...)
	}

	if c.FakeDNS != nil {
		fakeDNS, err := c.FakeDNS.Build()
		if err != nil {
			return nil, newError("failed to build fakedns").Base(err)
		}
		config.FakeDns = fakeDNS
	}

	return config, nil
}
