// +build !confonly

package fakedns

import (
	"context"
	"math"
	"math/big"
	gonet "net"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/cache"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/features/dns"
)

type Holder struct {
	domainToIP cache.Lru
	nextIP     *big.Int

	ipRange *gonet.IPNet

	config *FakeDnsPool
}

func (fkdns *Holder) IsIPInIPPool(ip net.Address) bool {
	if ip.Family().IsDomain() {
		return false
	}
	return fkdns.ipRange.Contains(ip.IP())
}

func (fkdns *Holder) GetFakeIPForDomain3(domain string, ipv4, ipv6 bool) []net.Address {
	isIPv6 := fkdns.ipRange.IP.To4() == nil
	if (isIPv6 && ipv6) || (!isIPv6 && ipv4) {
		return fkdns.GetFakeIPForDomain(domain)
	}
	return []net.Address{}
}

func (*Holder) Type() interface{} {
	return (*dns.FakeDNSEngine)(nil)
}

func (fkdns *Holder) Start() error {
	if fkdns.config != nil && fkdns.config.IpPool != "" && fkdns.config.LruSize != 0 {
		return fkdns.initializeFromConfig()
	}
	return nil
}

func (fkdns *Holder) Close() error {
	fkdns.domainToIP = nil
	fkdns.nextIP = nil
	fkdns.ipRange = nil
	return nil
}

func NewFakeDNSHolder() (*Holder, error) {
	var fkdns *Holder
	var err error

	if fkdns, err = NewFakeDNSHolderConfigOnly(nil); err != nil {
		return nil, newError("Unable to create Fake Dns Engine").Base(err).AtError()
	}
	err = fkdns.initialize("198.18.0.0/15", 65535)
	if err != nil {
		return nil, err
	}
	return fkdns, nil
}

func NewFakeDNSHolderConfigOnly(conf *FakeDnsPool) (*Holder, error) {
	return &Holder{nil, nil, nil, conf}, nil
}

func (fkdns *Holder) initializeFromConfig() error {
	return fkdns.initialize(fkdns.config.IpPool, int(fkdns.config.LruSize))
}

func (fkdns *Holder) initialize(ipPoolCidr string, lruSize int) error {
	var ipRange *gonet.IPNet
	var ipaddr gonet.IP
	var currentIP *big.Int
	var err error

	if ipaddr, ipRange, err = gonet.ParseCIDR(ipPoolCidr); err != nil {
		return newError("Unable to parse CIDR for Fake DNS IP assignment").Base(err).AtError()
	}

	currentIP = big.NewInt(0).SetBytes(ipaddr)
	if ipaddr.To4() != nil {
		currentIP = big.NewInt(0).SetBytes(ipaddr.To4())
	}

	ones, bits := ipRange.Mask.Size()
	rooms := bits - ones
	if math.Log2(float64(lruSize)) >= float64(rooms) {
		return newError("LRU size is bigger than subnet size").AtError()
	}
	fkdns.domainToIP = cache.NewLru(lruSize)
	fkdns.ipRange = ipRange
	fkdns.nextIP = currentIP
	return nil
}

// GetFakeIPForDomain check and generate a fake IP for a domain name
func (fkdns *Holder) GetFakeIPForDomain(domain string) []net.Address {
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
func (fkdns *Holder) GetDomainFromFakeDNS(ip net.Address) string {
	if !ip.Family().IsIP() || !fkdns.ipRange.Contains(ip.IP()) {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}

type HolderMulti struct {
	holders []*Holder

	config *FakeDnsPoolMulti
}

func (h *HolderMulti) IsIPInIPPool(ip net.Address) bool {
	if ip.Family().IsDomain() {
		return false
	}
	for _, v := range h.holders {
		if v.IsIPInIPPool(ip) {
			return true
		}
	}
	return false
}

func (h *HolderMulti) GetFakeIPForDomain3(domain string, ipv4, ipv6 bool) []net.Address {
	var ret []net.Address
	for _, v := range h.holders {
		ret = append(ret, v.GetFakeIPForDomain3(domain, ipv4, ipv6)...)
	}
	return ret
}

func (h *HolderMulti) GetFakeIPForDomain(domain string) []net.Address {
	var ret []net.Address
	for _, v := range h.holders {
		ret = append(ret, v.GetFakeIPForDomain(domain)...)
	}
	return ret
}

func (h *HolderMulti) GetDomainFromFakeDNS(ip net.Address) string {
	for _, v := range h.holders {
		if domain := v.GetDomainFromFakeDNS(ip); domain != "" {
			return domain
		}
	}
	return ""
}

func (h *HolderMulti) Type() interface{} {
	return (*dns.FakeDNSEngine)(nil)
}

func (h *HolderMulti) Start() error {
	for _, v := range h.holders {
		if v.config != nil && v.config.IpPool != "" && v.config.LruSize != 0 {
			if err := v.Start(); err != nil {
				return newError("Cannot start all fake dns pools").Base(err)
			}
		}
	}
	return nil
}

func (h *HolderMulti) Close() error {
	for _, v := range h.holders {
		if err := v.Close(); err != nil {
			return newError("Cannot close all fake dns pools").Base(err)
		}
	}
	return nil
}

func (h *HolderMulti) createHolderGroups() error {
	for _, v := range h.config.Pools {
		holder, err := NewFakeDNSHolderConfigOnly(v)
		if err != nil {
			return err
		}
		h.holders = append(h.holders, holder)
	}
	return nil
}

func NewFakeDNSHolderMulti(conf *FakeDnsPoolMulti) (*HolderMulti, error) {
	holderMulti := &HolderMulti{nil, conf}
	if err := holderMulti.createHolderGroups(); err != nil {
		return nil, err
	}
	return holderMulti, nil
}

func init() {
	common.Must(common.RegisterConfig((*FakeDnsPool)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		var f *Holder
		var err error
		if f, err = NewFakeDNSHolderConfigOnly(config.(*FakeDnsPool)); err != nil {
			return nil, err
		}
		return f, nil
	}))

	common.Must(common.RegisterConfig((*FakeDnsPoolMulti)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		var f *HolderMulti
		var err error
		if f, err = NewFakeDNSHolderMulti(config.(*FakeDnsPoolMulti)); err != nil {
			return nil, err
		}
		return f, nil
	}))
}
