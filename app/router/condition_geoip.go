package router

import (
	"net/netip"

	"go4.org/netipx"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

type GeoIPMatcher struct {
	countryCode  string
	reverseMatch bool
	ip4          *netipx.IPSet
	ip6          *netipx.IPSet
}

func (m *GeoIPMatcher) Init(cidrs []*routercommon.CIDR) error {
	var builder4, builder6 netipx.IPSetBuilder
	for _, cidr := range cidrs {
		netaddrIP, ok := netip.AddrFromSlice(cidr.GetIp())
		if !ok {
			return newError("invalid IP address ", cidr)
		}
		netaddrIP = netaddrIP.Unmap()
		ipPrefix := netip.PrefixFrom(netaddrIP, int(cidr.GetPrefix()))

		switch {
		case netaddrIP.Is4():
			builder4.AddPrefix(ipPrefix)
		case netaddrIP.Is6():
			builder6.AddPrefix(ipPrefix)
		}
	}

	var err error
	m.ip4, err = builder4.IPSet()
	if err != nil {
		return err
	}
	m.ip6, err = builder6.IPSet()
	if err != nil {
		return err
	}

	return nil
}

func (m *GeoIPMatcher) SetReverseMatch(isReverseMatch bool) {
	m.reverseMatch = isReverseMatch
}

func (m *GeoIPMatcher) match4(ip net.IP) bool {
	nip, ok := netipx.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ip4.Contains(nip)
}

func (m *GeoIPMatcher) match6(ip net.IP) bool {
	nip, ok := netipx.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ip6.Contains(nip)
}

// Match returns true if the given ip is included by the GeoIP.
func (m *GeoIPMatcher) Match(ip net.IP) bool {
	isMatched := false
	switch len(ip) {
	case net.IPv4len:
		isMatched = m.match4(ip)
	case net.IPv6len:
		isMatched = m.match6(ip)
	}
	if m.reverseMatch {
		return !isMatched
	}
	return isMatched
}

// GeoIPMatcherContainer is a container for GeoIPMatchers. It keeps unique copies of GeoIPMatcher by country code.
type GeoIPMatcherContainer struct {
	matchers []*GeoIPMatcher
}

// Add adds a new GeoIP set into the container.
// If the country code of GeoIP is not empty, GeoIPMatcherContainer will try to find an existing one, instead of adding a new one.
func (c *GeoIPMatcherContainer) Add(geoip *routercommon.GeoIP) (*GeoIPMatcher, error) {
	if geoip.CountryCode != "" {
		for _, m := range c.matchers {
			if m.countryCode == geoip.CountryCode && m.reverseMatch == geoip.InverseMatch {
				return m, nil
			}
		}
	}

	m := &GeoIPMatcher{
		countryCode:  geoip.CountryCode,
		reverseMatch: geoip.InverseMatch,
	}
	if err := m.Init(geoip.Cidr); err != nil {
		return nil, err
	}
	if geoip.CountryCode != "" {
		c.matchers = append(c.matchers, m)
	}
	return m, nil
}

var globalGeoIPContainer GeoIPMatcherContainer
