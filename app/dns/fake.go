package dns

import (
	"strconv"

	"v2ray.com/core/common/net"
)

var domainToIP = NewLru(65535)
var nextIP = 0

// GetFakeIPForDomain check and generate a fake IP for a domain name
func GetFakeIPForDomain(domain string) []net.Address {
	if v, ok := domainToIP.Get(domain); ok {
		return []net.Address{v.(net.Address)}
	}
	as := "240."
	as += strconv.Itoa((0xff0000&nextIP)>>16) + "."
	as += strconv.Itoa((0xff00&nextIP)>>8) + "."
	as += strconv.Itoa(0xff & nextIP)
	ip := net.ParseAddress(as)
	domainToIP.Put(domain, ip)
	nextIP = 0xffffff & (nextIP + 1)
	return []net.Address{ip}
}

// GetDomainFromFakeDNS check if an IP is a fake IP and have corresponding domain name
func GetDomainFromFakeDNS(ip net.Address) string {
	if !ip.Family().IsIP() || ip.String()[:3] != "240" {
		return ""
	}
	k, ok := domainToIP.GetKeyFromValue(ip)
	if ok {
		return k.(string)
	}
	return ""
}
