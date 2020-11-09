package dns

import (
	"context"
	"strconv"

	"v2ray.com/core/common/cache"
	"v2ray.com/core/common/net"
)

type FakeDnsHolder struct {
	domainToIP cache.Lru
	nextIP     int
}

var fakednsHolder = NewFakeDnsHolder()

func NewFakeDnsHolder() *FakeDnsHolder {
	return &FakeDnsHolder{cache.NewLru(65535), 0}
}

// GetFakeIPForDomain check and generate a fake IP for a domain name
func (fkdns *FakeDnsHolder) GetFakeIPForDomain(domain string) []net.Address {
	if v, ok := fkdns.domainToIP.Get(domain); ok {
		return []net.Address{v.(net.Address)}
	}
	var ip net.Address
	for {
		as := "240."
		as += strconv.Itoa((0xff0000&fkdns.nextIP)>>16) + "."
		as += strconv.Itoa((0xff00&fkdns.nextIP)>>8) + "."
		as += strconv.Itoa(0xff & fkdns.nextIP)
		ip = net.ParseAddress(as)
		fkdns.nextIP = 0xffffff & (fkdns.nextIP + 1)
		// if we run for a long time, we may go back to beginning and start seeing the IP in use
		if _, ok := fkdns.domainToIP.GetKeyFromValue(ip); !ok {
			break
		}
	}
	fkdns.domainToIP.Put(domain, ip)
	return []net.Address{ip}
}

// GetDomainFromFakeDNS check if an IP is a fake IP and have corresponding domain name
func (fkdns *FakeDnsHolder) GetDomainFromFakeDNS(ip net.Address) string {
	if !ip.Family().IsIP() || ip.String()[:3] != "240" {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}

// GetDefaultFakeDnsFromContext will retrieve a FakeDnsHolder from context, local to that context
// TODO: Current a stub function, should not relay on global variable
func GetDefaultFakeDnsFromContext(ctx context.Context) FakeDns {
	return fakednsHolder
}
