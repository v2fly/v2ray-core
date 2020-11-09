package dns

import (
	"strconv"

	"v2ray.com/core/common/cache"
	"v2ray.com/core/common/net"
)

var domainToIP = cache.NewLru(65535)
var nextIP = 0

// GetFakeIPForDomain check and generate a fake IP for a domain name
func GetFakeIPForDomain(domain string) []net.Address {
	if v, ok := domainToIP.Get(domain); ok {
		return []net.Address{v.(net.Address)}
	}
	var ip net.Address
	for {
		as := "240."
		as += strconv.Itoa((0xff0000&nextIP)>>16) + "."
		as += strconv.Itoa((0xff00&nextIP)>>8) + "."
		as += strconv.Itoa(0xff & nextIP)
		ip = net.ParseAddress(as)
		nextIP = 0xffffff & (nextIP + 1)
		// if we run for a long time, we may go back to beginning and start seeing the IP in use
		if _, ok := domainToIP.GetKeyFromValue(ip); !ok {
			break
		}
	}
	domainToIP.Put(domain, ip)
	return []net.Address{ip}
}

// GetDomainFromFakeDNS check if an IP is a fake IP and have corresponding domain name
func GetDomainFromFakeDNS(ip net.Address) string {
	if !ip.Family().IsIP() || ip.String()[:3] != "240" {
		return ""
	}
	if k, ok := domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}
