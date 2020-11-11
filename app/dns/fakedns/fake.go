package fakedns

import (
	"context"
	"math/big"
	gonet "net"

	"v2ray.com/core/common"
	"v2ray.com/core/common/cache"
	"v2ray.com/core/common/net"
	"v2ray.com/core/features/dns"
)

type FakeDNSHolder struct {
	domainToIP cache.Lru
	nextIP     *big.Int

	ipRange *gonet.IPNet

	config *FakeDnsPool
}

func (fkdns *FakeDNSHolder) Type() interface{} {
	return (*dns.FakeDNSEngine)(nil)
}

func (fkdns *FakeDNSHolder) Start() error {
	return fkdns.initializeFromConfig()
}

func (fkdns *FakeDNSHolder) Close() error {
	fkdns.domainToIP = nil
	fkdns.nextIP = nil
	fkdns.ipRange = nil
	return nil
}

func NewFakeDNSHolder() (*FakeDNSHolder, error) {
	var fkdns *FakeDNSHolder
	if fkdnsw, err := NewFakeDNSHolderConfigOnly(nil); err != nil {
		return nil, newError("Unable to create Fake Dns Engine").Base(err).AtError()
	} else {
		fkdnsw.initialize("240.0.0.0/8", 65535)
		fkdns = fkdnsw
	}
	return fkdns, nil
}

func NewFakeDNSHolderConfigOnly(conf *FakeDnsPool) (*FakeDNSHolder, error) {
	return &FakeDNSHolder{nil, nil, nil, conf}, nil
}

func (fkdns *FakeDNSHolder) initializeFromConfig() error {
	return fkdns.initialize(fkdns.config.IpPool, int(fkdns.config.LruSize))
}

func (fkdns *FakeDNSHolder) initialize(ipPoolCidr string, lruSize int) error {
	var ipRange *gonet.IPNet
	var currentIP *big.Int

	if ipaddr, ipRangeResult, err := gonet.ParseCIDR(ipPoolCidr); err != nil {
		return newError("Unable to parse CIDR for Fake DNS IP assignment").Base(err).AtError()
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
		return newError("LRU size is bigger than subnet size").AtError()
	}
	fkdns.domainToIP = cache.NewLru(lruSize)
	fkdns.ipRange = ipRange
	fkdns.nextIP = currentIP
	return nil
}

// GetFakeIPForDomain check and generate a fake IP for a domain name
func (fkdns *FakeDNSHolder) GetFakeIPForDomain(domain string) []net.Address {
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
func (fkdns *FakeDNSHolder) GetDomainFromFakeDNS(ip net.Address) string {
	if !ip.Family().IsIP() || !fkdns.ipRange.Contains(ip.IP()) {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}

func init() {
	common.Must(common.RegisterConfig((*FakeDnsPool)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		if f, err := NewFakeDNSHolderConfigOnly(config.(*FakeDnsPool)); err != nil {
			return nil, err
		} else {
			return f, nil
		}

	}))
}
