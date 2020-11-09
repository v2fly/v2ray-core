package dns

import "v2ray.com/core/common/net"

type FakeDns interface {
	GetFakeIPForDomain(domain string) []net.Address
	GetDomainFromFakeDNS(ip net.Address) string
}
