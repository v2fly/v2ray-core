//go:build !confonly
// +build !confonly

package dns

import (
	"golang.org/x/net/dns/dnsmessage"

	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/strmatcher"
	"github.com/ghxhy/v2ray-core/v5/common/uuid"
	"github.com/ghxhy/v2ray-core/v5/features/dns"
)

var typeMap = map[DomainMatchingType]strmatcher.Type{
	DomainMatchingType_Full:      strmatcher.Full,
	DomainMatchingType_Subdomain: strmatcher.Domain,
	DomainMatchingType_Keyword:   strmatcher.Substr,
	DomainMatchingType_Regex:     strmatcher.Regex,
}

// References:
// https://www.iana.org/assignments/special-use-domain-names/special-use-domain-names.xhtml
// https://unix.stackexchange.com/questions/92441/whats-the-difference-between-local-home-and-lan
var localTLDsAndDotlessDomains = []*NameServer_PriorityDomain{
	{Type: DomainMatchingType_Regex, Domain: "^[^.]+$"}, // This will only match domains without any dot
	{Type: DomainMatchingType_Subdomain, Domain: "local"},
	{Type: DomainMatchingType_Subdomain, Domain: "localdomain"},
	{Type: DomainMatchingType_Subdomain, Domain: "localhost"},
	{Type: DomainMatchingType_Subdomain, Domain: "lan"},
	{Type: DomainMatchingType_Subdomain, Domain: "home.arpa"},
	{Type: DomainMatchingType_Subdomain, Domain: "example"},
	{Type: DomainMatchingType_Subdomain, Domain: "invalid"},
	{Type: DomainMatchingType_Subdomain, Domain: "test"},
}

var localTLDsAndDotlessDomainsRule = &NameServer_OriginalRule{
	Rule: "geosite:private",
	Size: uint32(len(localTLDsAndDotlessDomains)),
}

func toStrMatcher(t DomainMatchingType, domain string) (strmatcher.Matcher, error) {
	strMType, f := typeMap[t]
	if !f {
		return nil, newError("unknown mapping type", t).AtWarning()
	}
	matcher, err := strMType.New(domain)
	if err != nil {
		return nil, newError("failed to create str matcher").Base(err)
	}
	return matcher, nil
}

func toNetIP(addrs []net.Address) ([]net.IP, error) {
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if addr.Family().IsIP() {
			ips = append(ips, addr.IP())
		} else {
			return nil, newError("Failed to convert address", addr, "to Net IP.").AtWarning()
		}
	}
	return ips, nil
}

func toIPOption(s QueryStrategy) dns.IPOption {
	return dns.IPOption{
		IPv4Enable: s == QueryStrategy_USE_IP || s == QueryStrategy_USE_IP4,
		IPv6Enable: s == QueryStrategy_USE_IP || s == QueryStrategy_USE_IP6,
		FakeEnable: false,
	}
}

func toReqTypes(option dns.IPOption) []dnsmessage.Type {
	var reqTypes []dnsmessage.Type
	if option.IPv4Enable {
		reqTypes = append(reqTypes, dnsmessage.TypeA)
	}
	if option.IPv6Enable {
		reqTypes = append(reqTypes, dnsmessage.TypeAAAA)
	}
	return reqTypes
}

func generateRandomTag() string {
	id := uuid.New()
	return "v2ray.system." + id.String()
}
