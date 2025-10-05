//go:build !confonly
// +build !confonly

package fakedns

import (
	"context"
	"math"
	"math/big"
	gonet "net"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/cache"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/dns"
)

type Holder struct {
	domainToIP cache.Lru
	nextIP     *big.Int
	mu         *sync.Mutex

	ipRange *gonet.IPNet

	config *FakeDnsPool
	closed bool // Flag, closed instance
}

func (fkdns *Holder) IsIPInIPPool(ip net.Address) bool {
	if fkdns == nil || fkdns.closed {
		return false
	}
	if ip.Family().IsDomain() {
		return false
	}
	if fkdns.ipRange == nil {
		return false
	}
	return fkdns.ipRange.Contains(ip.IP())
}

func (fkdns *Holder) GetFakeIPForDomain3(domain string, ipv4, ipv6 bool) []net.Address {
	if fkdns == nil || fkdns.closed || fkdns.ipRange == nil {
		return []net.Address{}
	}
	isIPv6 := fkdns.ipRange.IP.To4() == nil
	if (isIPv6 && ipv6) || (!isIPv6 && ipv4) {
		return fkdns.GetFakeIPForDomain(domain)
	}
	return []net.Address{}
}

func (*Holder) Type() interface{} {
	return dns.FakeDNSEngineType()
}

func (fkdns *Holder) Start() error {
	if fkdns == nil || fkdns.closed {
		return newError("Holder is closed")
	}
	if fkdns.config != nil && fkdns.config.IpPool != "" && fkdns.config.LruSize != 0 {
		return fkdns.initializeFromConfig()
	}
	return newError("invalid fakeDNS setting")
}

func (fkdns *Holder) Close() error {
	if fkdns == nil || fkdns.closed {
		return nil
	}
	fkdns.domainToIP = nil
	fkdns.nextIP = nil
	fkdns.ipRange = nil
	fkdns.mu = nil
	fkdns.closed = true
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
	return &Holder{nil, nil, nil, nil, conf}, nil
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
	fkdns.mu = new(sync.Mutex)
	return nil
}

// GetFakeIPForDomain checks and generate a fake IP for a domain name
func (fkdns *Holder) GetFakeIPForDomain(domain string) []net.Address {
	if fkdns == nil || fkdns.closed {
		return nil
	}
	fkdns.mu.Lock()
	defer fkdns.mu.Unlock()
	if v, ok := fkdns.domainToIP.Get(domain); ok {
		return []net.Address{v.(net.Address)}
	}
	var ip net.Address
	for {
		ip = net.IPAddress(fkdns.nextIP.Bytes())

		fkdns.nextIP = fkdns.nextIP.Add(fkdns.nextIP, big.NewInt(1))
		if fkdns.ipRange == nil || !fkdns.ipRange.Contains(fkdns.nextIP.Bytes()) {
			if fkdns.ipRange != nil {
				fkdns.nextIP = big.NewInt(0).SetBytes(fkdns.ipRange.IP)
			} else {
				break // ipRange nil, cancel
			}
		}

		// if we run for a long time, we may go back to beginning and start seeing the IP in use
		if _, ok := fkdns.domainToIP.GetKeyFromValue(ip); !ok {
			break
		}
	}
	fkdns.domainToIP.Put(domain, ip)
	return []net.Address{ip}
}

// GetDomainFromFakeDNS checks if an IP is a fake IP and have corresponding domain name
func (fkdns *Holder) GetDomainFromFakeDNS(ip net.Address) string {
	// nil-Checks, Panics preventing (Issue: Segfault in net.(*IPNet).Contains)
	if fkdns == nil || fkdns.closed || fkdns.ipRange == nil || fkdns.domainToIP == nil {
		return ""
	}
	if ip == nil || !ip.Family().IsIP() {
		return ""
	}
	ipBytes := ip.IP()
	if ipBytes == nil {
		return ""
	}
	// ipRange-Check
	if fkdns.ipRange == nil || !fkdns.ipRange.Contains(ipBytes) {
		return ""
	}
	if fkdns.domainToIP == nil {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		if str, ok := k.(string); ok {
			return str
		}
	}
	return ""
}

type HolderMulti struct {
	holders []*Holder
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

func (h *HolderMulti) IsEmpty() bool {
	return len(h.holders) == 0
}

func (h *HolderMulti) AddPool(poolConfig *FakeDnsPool) (*Holder, error) {
	_, newIPRange, err := gonet.ParseCIDR(poolConfig.IpPool)
	if err != nil {
		return nil, err
	}
	running := false
	for _, v := range h.holders {
		var ipRange *gonet.IPNet
		if v.ipRange != nil {
			ipRange = v.ipRange
			running = true
		} else {
			_, ipRange, err = gonet.ParseCIDR(v.config.IpPool)
			if err != nil {
				return nil, err
			}
		}
		if ipRange.String() == newIPRange.String() {
			return v, nil
		}
		if ipRange.Contains(newIPRange.IP) || newIPRange.Contains(ipRange.IP) {
			return nil, newError("Trying to add ip pool ", newIPRange, " that overlaps with existing ip pool ", ipRange)
		}
	}
	holder, err := NewFakeDNSHolderConfigOnly(poolConfig)
	if err != nil {
		return nil, err
	}
	if running {
		if err := holder.Start(); err != nil {
			return nil, err
		}
	}
	h.holders = append(h.holders, holder)
	return holder, nil
}

func (h *HolderMulti) AddPoolMulti(poolMultiConfig *FakeDnsPoolMulti) (*HolderMulti, error) {
	holderMulti := &HolderMulti{}
	for _, poolConfig := range poolMultiConfig.Pools {
		pool, err := h.AddPool(poolConfig)
		if err != nil {
			return nil, err
		}
		holderMulti.holders = append(holderMulti.holders, pool)
	}
	return holderMulti, nil // Returned holderMulti holds references to pools managed by `h`
}

func (h *HolderMulti) Type() interface{} {
	return dns.FakeDNSEngineType()
}

func (h *HolderMulti) Start() error {
	for _, v := range h.holders {
		if err := v.Start(); err != nil {
			return newError("Cannot start all fake dns pools").Base(err)
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

func (h *HolderMulti) createHolderGroups(conf *FakeDnsPoolMulti) error {
	for _, pool := range conf.Pools {
		_, err := h.AddPool(pool)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewFakeDNSHolderMulti(conf *FakeDnsPoolMulti) (*HolderMulti, error) {
	holderMulti := &HolderMulti{}
	if err := holderMulti.createHolderGroups(conf); err != nil {
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
