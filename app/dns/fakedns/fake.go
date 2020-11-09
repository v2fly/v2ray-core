package fakedns

import (
	"context"
	"math/big"
	"v2ray.com/core/common/cache"
	"v2ray.com/core/common/net"
	dns2 "v2ray.com/core/features/dns"

	gonet "net"
)

type FakeDnsHolder struct {
	domainToIP cache.Lru
	nextIP     *big.Int

	ipRange *gonet.IPNet
}

var fakeDnsHolder, _ = NewFakeDnsHolder()

func NewFakeDnsHolder() (*FakeDnsHolder, error) {
	var ipRange *gonet.IPNet
	var currentIP *big.Int

	var lruSize = 65535

	if ipaddr, ipRangeResult, err := gonet.ParseCIDR("240.0.0.0/8"); err != nil {
		return nil, newError("Unable to parse CIDR for Fake DNS IP assignment").Base(err).AtError()
	} else {
		ipRange = ipRangeResult
		currentIP = big.NewInt(0).SetBytes(ipaddr)
		if ipaddr.To4() != nil {
			currentIP = big.NewInt(0).SetBytes(ipaddr.To4())
		}
	}

	ones, bits := ipRange.Mask.Size()
	rooms := bits - ones
	if lruSize >= 1<<rooms {
		return nil, newError("LRU size is bigger than subnet size").AtError()
	}

	return &FakeDnsHolder{cache.NewLru(lruSize), currentIP, ipRange}, nil
}

// GetFakeIPForDomain check and generate a fake IP for a domain name
func (fkdns *FakeDnsHolder) GetFakeIPForDomain(domain string) []net.Address {
	if v, ok := fkdns.domainToIP.Get(domain); ok {
		return []net.Address{v.(net.Address)}
	}
	var ip net.Address
	for {
		ip = net.IPAddress(fkdns.nextIP.Bytes())

		fkdns.nextIP = fkdns.nextIP.Add(fkdns.nextIP, big.NewInt(1))
		if !fkdns.ipRange.Contains(fkdns.nextIP.Bytes()) {
			fkdns.nextIP = big.NewInt(0).SetBytes(fkdns.ipRange.IP)
		}

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
	if !ip.Family().IsIP() || !fkdns.ipRange.Contains(ip.IP()) {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}

// GetDefaultFakeDnsFromContext will retrieve a FakeDnsHolder from context, local to that context
// TODO: Current a stub function, should not relay on global variable
func GetDefaultFakeDnsFromContext(ctx context.Context) dns2.FakeDnsEngine {
	return fakeDnsHolder
}
