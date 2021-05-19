package rule

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func parseDomainRule(ctx context.Context, domain string) ([]*router.Domain, error) {
	cfgEnv := cfgcommon.GetConfigureLoadingEnvironment(ctx)
	geoLoader := cfgEnv.GetGeoLoader()

	if strings.HasPrefix(domain, "geosite:") {
		list := domain[8:]
		if len(list) == 0 {
			return nil, newError("empty listname in rule: ", domain)
		}
		domains, err := geoLoader.LoadGeoSite(list)
		if err != nil {
			return nil, newError("failed to load geosite: ", list).Base(err)
		}

		return domains, nil
	}

	isExtDatFile := 0
	{
		const prefix = "ext:"
		if strings.HasPrefix(domain, prefix) {
			isExtDatFile = len(prefix)
		}
		const prefixQualified = "ext-domain:"
		if strings.HasPrefix(domain, prefixQualified) {
			isExtDatFile = len(prefixQualified)
		}
	}

	if isExtDatFile != 0 {
		kv := strings.Split(domain[isExtDatFile:], ":")
		if len(kv) != 2 {
			return nil, newError("invalid external resource: ", domain)
		}
		filename := kv[0]
		list := kv[1]
		domains, err := geoLoader.LoadGeoSiteWithAttr(filename, list)
		if err != nil {
			return nil, newError("failed to load external geosite: ", list, " from ", filename).Base(err)
		}

		return domains, nil
	}

	domainRule := new(router.Domain)
	switch {
	case strings.HasPrefix(domain, "regexp:"):
		regexpVal := domain[7:]
		if len(regexpVal) == 0 {
			return nil, newError("empty regexp type of rule: ", domain)
		}
		domainRule.Type = router.Domain_Regex
		domainRule.Value = regexpVal

	case strings.HasPrefix(domain, "domain:"):
		domainName := domain[7:]
		if len(domainName) == 0 {
			return nil, newError("empty domain type of rule: ", domain)
		}
		domainRule.Type = router.Domain_Domain
		domainRule.Value = domainName

	case strings.HasPrefix(domain, "full:"):
		fullVal := domain[5:]
		if len(fullVal) == 0 {
			return nil, newError("empty full domain type of rule: ", domain)
		}
		domainRule.Type = router.Domain_Full
		domainRule.Value = fullVal

	case strings.HasPrefix(domain, "keyword:"):
		keywordVal := domain[8:]
		if len(keywordVal) == 0 {
			return nil, newError("empty keyword type of rule: ", domain)
		}
		domainRule.Type = router.Domain_Plain
		domainRule.Value = keywordVal

	case strings.HasPrefix(domain, "dotless:"):
		domainRule.Type = router.Domain_Regex
		switch substr := domain[8:]; {
		case substr == "":
			domainRule.Value = "^[^.]*$"
		case !strings.Contains(substr, "."):
			domainRule.Value = "^[^.]*" + substr + "[^.]*$"
		default:
			return nil, newError("substr in dotless rule should not contain a dot: ", substr)
		}

	default:
		domainRule.Type = router.Domain_Plain
		domainRule.Value = domain
	}
	return []*router.Domain{domainRule}, nil
}

func toCidrList(ctx context.Context, ips cfgcommon.StringList) ([]*router.GeoIP, error) {
	cfgEnv := cfgcommon.GetConfigureLoadingEnvironment(ctx)
	geoLoader := cfgEnv.GetGeoLoader()

	var geoipList []*router.GeoIP
	var customCidrs []*router.CIDR

	for _, ip := range ips {
		if strings.HasPrefix(ip, "geoip:") {
			country := ip[6:]
			isReverseMatch := false
			if strings.HasPrefix(ip, "geoip:!") {
				country = ip[7:]
				isReverseMatch = true
			}
			if len(country) == 0 {
				return nil, newError("empty country name in rule")
			}
			geoip, err := geoLoader.LoadGeoIP(country)
			if err != nil {
				return nil, newError("failed to load geoip: ", country).Base(err)
			}

			geoipList = append(geoipList, &router.GeoIP{
				CountryCode:  strings.ToUpper(country),
				Cidr:         geoip,
				ReverseMatch: isReverseMatch,
			})

			continue
		}

		isExtDatFile := 0
		{
			const prefix = "ext:"
			if strings.HasPrefix(ip, prefix) {
				isExtDatFile = len(prefix)
			}
			const prefixQualified = "ext-ip:"
			if strings.HasPrefix(ip, prefixQualified) {
				isExtDatFile = len(prefixQualified)
			}
		}

		if isExtDatFile != 0 {
			kv := strings.Split(ip[isExtDatFile:], ":")
			if len(kv) != 2 {
				return nil, newError("invalid external resource: ", ip)
			}

			filename := kv[0]
			country := kv[1]
			if len(filename) == 0 || len(country) == 0 {
				return nil, newError("empty filename or empty country in rule")
			}

			isReverseMatch := false
			if strings.HasPrefix(country, "!") {
				country = country[1:]
				isReverseMatch = true
			}
			geoip, err := geoLoader.LoadIP(filename, country)
			if err != nil {
				return nil, newError("failed to load geoip: ", country, " from ", filename).Base(err)
			}

			geoipList = append(geoipList, &router.GeoIP{
				CountryCode:  strings.ToUpper(filename + "_" + country),
				Cidr:         geoip,
				ReverseMatch: isReverseMatch,
			})

			continue
		}

		ipRule, err := ParseIP(ip)
		if err != nil {
			return nil, newError("invalid IP: ", ip).Base(err)
		}
		customCidrs = append(customCidrs, ipRule)
	}

	if len(customCidrs) > 0 {
		geoipList = append(geoipList, &router.GeoIP{
			Cidr: customCidrs,
		})
	}

	return geoipList, nil
}

func parseFieldRule(ctx context.Context, msg json.RawMessage) (*router.RoutingRule, error) {
	type RawFieldRule struct {
		RouterRule
		Domain     *cfgcommon.StringList  `json:"domain"`
		Domains    *cfgcommon.StringList  `json:"domains"`
		IP         *cfgcommon.StringList  `json:"ip"`
		Port       *cfgcommon.PortList    `json:"port"`
		Network    *cfgcommon.NetworkList `json:"network"`
		SourceIP   *cfgcommon.StringList  `json:"source"`
		SourcePort *cfgcommon.PortList    `json:"sourcePort"`
		User       *cfgcommon.StringList  `json:"user"`
		InboundTag *cfgcommon.StringList  `json:"inboundTag"`
		Protocols  *cfgcommon.StringList  `json:"protocol"`
		Attributes string                 `json:"attrs"`
	}
	rawFieldRule := new(RawFieldRule)
	err := json.Unmarshal(msg, rawFieldRule)
	if err != nil {
		return nil, err
	}

	rule := new(router.RoutingRule)
	switch {
	case len(rawFieldRule.OutboundTag) > 0:
		rule.TargetTag = &router.RoutingRule_Tag{
			Tag: rawFieldRule.OutboundTag,
		}
	case len(rawFieldRule.BalancerTag) > 0:
		rule.TargetTag = &router.RoutingRule_BalancingTag{
			BalancingTag: rawFieldRule.BalancerTag,
		}
	default:
		return nil, newError("neither outboundTag nor balancerTag is specified in routing rule")
	}

	if rawFieldRule.DomainMatcher != "" {
		rule.DomainMatcher = rawFieldRule.DomainMatcher
	}

	if rawFieldRule.Domain != nil {
		for _, domain := range *rawFieldRule.Domain {
			rules, err := parseDomainRule(ctx, domain)
			if err != nil {
				return nil, newError("failed to parse domain rule: ", domain).Base(err)
			}
			rule.Domain = append(rule.Domain, rules...)
		}
	}

	if rawFieldRule.Domains != nil {
		for _, domain := range *rawFieldRule.Domains {
			rules, err := parseDomainRule(ctx, domain)
			if err != nil {
				return nil, newError("failed to parse domain rule: ", domain).Base(err)
			}
			rule.Domain = append(rule.Domain, rules...)
		}
	}

	if rawFieldRule.IP != nil {
		geoipList, err := toCidrList(ctx, *rawFieldRule.IP)
		if err != nil {
			return nil, err
		}
		rule.Geoip = geoipList
	}

	if rawFieldRule.Port != nil {
		rule.PortList = rawFieldRule.Port.Build()
	}

	if rawFieldRule.Network != nil {
		rule.Networks = rawFieldRule.Network.Build()
	}

	if rawFieldRule.SourceIP != nil {
		geoipList, err := toCidrList(ctx, *rawFieldRule.SourceIP)
		if err != nil {
			return nil, err
		}
		rule.SourceGeoip = geoipList
	}

	if rawFieldRule.SourcePort != nil {
		rule.SourcePortList = rawFieldRule.SourcePort.Build()
	}

	if rawFieldRule.User != nil {
		for _, s := range *rawFieldRule.User {
			rule.UserEmail = append(rule.UserEmail, s)
		}
	}

	if rawFieldRule.InboundTag != nil {
		for _, s := range *rawFieldRule.InboundTag {
			rule.InboundTag = append(rule.InboundTag, s)
		}
	}

	if rawFieldRule.Protocols != nil {
		for _, s := range *rawFieldRule.Protocols {
			rule.Protocol = append(rule.Protocol, s)
		}
	}

	if len(rawFieldRule.Attributes) > 0 {
		rule.Attributes = rawFieldRule.Attributes
	}

	return rule, nil
}

func ParseRule(ctx context.Context, msg json.RawMessage) (*router.RoutingRule, error) {
	rawRule := new(RouterRule)
	err := json.Unmarshal(msg, rawRule)
	if err != nil {
		return nil, newError("invalid router rule").Base(err)
	}
	if strings.EqualFold(rawRule.Type, "field") {
		fieldrule, err := parseFieldRule(ctx, msg)
		if err != nil {
			return nil, newError("invalid field rule").Base(err)
		}
		return fieldrule, nil
	}

	return nil, newError("unknown router rule type: ", rawRule.Type)
}

func ParseIP(s string) (*router.CIDR, error) {
	var addr, mask string
	i := strings.Index(s, "/")
	if i < 0 {
		addr = s
	} else {
		addr = s[:i]
		mask = s[i+1:]
	}
	ip := net.ParseAddress(addr)
	switch ip.Family() {
	case net.AddressFamilyIPv4:
		bits := uint32(32)
		if len(mask) > 0 {
			bits64, err := strconv.ParseUint(mask, 10, 32)
			if err != nil {
				return nil, newError("invalid network mask for router: ", mask).Base(err)
			}
			bits = uint32(bits64)
		}
		if bits > 32 {
			return nil, newError("invalid network mask for router: ", bits)
		}
		return &router.CIDR{
			Ip:     []byte(ip.IP()),
			Prefix: bits,
		}, nil
	case net.AddressFamilyIPv6:
		bits := uint32(128)
		if len(mask) > 0 {
			bits64, err := strconv.ParseUint(mask, 10, 32)
			if err != nil {
				return nil, newError("invalid network mask for router: ", mask).Base(err)
			}
			bits = uint32(bits64)
		}
		if bits > 128 {
			return nil, newError("invalid network mask for router: ", bits)
		}
		return &router.CIDR{
			Ip:     []byte(ip.IP()),
			Prefix: bits,
		}, nil
	default:
		return nil, newError("unsupported address for router: ", s)
	}
}

func ParseDomainRule(ctx context.Context, domain string) ([]*router.Domain, error) {
	return parseDomainRule(ctx, domain)
}

func ToCidrList(ctx context.Context, ips cfgcommon.StringList) ([]*router.GeoIP, error) {
	return toCidrList(ctx, ips)
}

type RouterRule struct {
	Type        string `json:"type"`
	OutboundTag string `json:"outboundTag"`
	BalancerTag string `json:"balancerTag"`

	DomainMatcher string `json:"domainMatcher"`
}
