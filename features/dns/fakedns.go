package dns

import (
	"v2ray.com/core/common/net"
	"v2ray.com/core/features"
)

type FakeDnsEngine interface {
	features.Feature
	GetFakeIPForDomain(domain string) []net.Address
	GetDomainFromFakeDNS(ip net.Address) string
}
