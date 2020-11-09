package dns

import "v2ray.com/core/common/net"

type FakeDnsEngine interface {
	GetFakeIPForDomain(domain string) []net.Address
	GetDomainFromFakeDNS(ip net.Address) string
}
