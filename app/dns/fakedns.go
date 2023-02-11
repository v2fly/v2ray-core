//go:build !confonly
// +build !confonly

package dns

import (
	fakedns "github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/dns"
)

// FakeDNSClient is an implementation of dns.Client with FakeDNS enabled.
type FakeDNSClient struct {
	*DNS
}

// LookupIP implements dns.Client.
func (s *FakeDNSClient) LookupIP(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv4Enable: true, IPv6Enable: true, FakeEnable: true})
}

// LookupIPv4 implements dns.IPv4Lookup.
func (s *FakeDNSClient) LookupIPv4(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv4Enable: true, FakeEnable: true})
}

// LookupIPv6 implements dns.IPv6Lookup.
func (s *FakeDNSClient) LookupIPv6(domain string) ([]net.IP, error) {
	return s.lookupIPInternal(domain, dns.IPOption{IPv6Enable: true, FakeEnable: true})
}

// FakeDNSEngine is an implementation of dns.FakeDNSEngine based on a fully functional DNS.
type FakeDNSEngine struct {
	dns         *DNS
	fakeHolders *fakedns.HolderMulti
	fakeDefault *fakedns.HolderMulti
}

// Type implements common.HasType.
func (*FakeDNSEngine) Type() interface{} {
	return dns.FakeDNSEngineType()
}

// Start implements common.Runnable.
func (f *FakeDNSEngine) Start() error {
	return f.fakeHolders.Start()
}

// Close implements common.Closable.
func (f *FakeDNSEngine) Close() error {
	return f.fakeHolders.Close()
}

// GetFakeIPForDomain implements dns.FakeDNSEngine.
func (f *FakeDNSEngine) GetFakeIPForDomain(domain string) []net.Address {
	return f.GetFakeIPForDomain3(domain, true, true)
}

// GetDomainFromFakeDNS implements dns.FakeDNSEngine.
func (f *FakeDNSEngine) GetDomainFromFakeDNS(ip net.Address) string {
	return f.fakeHolders.GetDomainFromFakeDNS(ip)
}

// IsIPInIPPool implements dns.FakeDNSEngineRev0.
func (f *FakeDNSEngine) IsIPInIPPool(ip net.Address) bool {
	return f.fakeHolders.IsIPInIPPool(ip)
}

// GetFakeIPForDomain3 implements dns.FakeDNSEngineRev0.
func (f *FakeDNSEngine) GetFakeIPForDomain3(domain string, IPv4 bool, IPv6 bool) []net.Address { // nolint: gocritic
	option := dns.IPOption{IPv4Enable: IPv4, IPv6Enable: IPv6, FakeEnable: true}
	for _, client := range f.dns.sortClients(domain, option) {
		fakeServer, ok := client.fakeDNS.(*FakeDNSServer)
		if !ok {
			continue
		}
		fakeEngine, ok := fakeServer.fakeDNSEngine.(dns.FakeDNSEngineRev0)
		if !ok {
			return filterIP(fakeServer.fakeDNSEngine.GetFakeIPForDomain(domain), option)
		}
		return fakeEngine.GetFakeIPForDomain3(domain, IPv4, IPv6)
	}
	if f.fakeDefault != nil {
		return f.fakeDefault.GetFakeIPForDomain3(domain, IPv4, IPv6)
	}
	return nil
}
