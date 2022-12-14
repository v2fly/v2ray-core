package dns_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	. "github.com/v2fly/v2ray-core/v5/app/dns"
	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/app/policy"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	feature_dns "github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/proxy/freedom"
	"github.com/v2fly/v2ray-core/v5/testing/servers/udp"
)

func TestFakeDNS(t *testing.T) {
	port := udp.PickPort()

	dnsServer := dns.Server{
		Addr:    "127.0.0.1:" + port.String(),
		Net:     "udp",
		Handler: &staticHandler{},
		UDPSize: 1200,
	}

	go dnsServer.ListenAndServe()
	time.Sleep(time.Second)

	config := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&Config{
				NameServer: []*NameServer{
					{ // "fakedns"
						Address: &net.Endpoint{
							Network: net.Network_UDP,
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Domain{
									Domain: "fakedns",
								},
							},
							Port: uint32(53),
						},
					},
					{ // { "address": "127.0.0.1", "port": "<port>", "domains": ["domain:google.com"], "fakedns": "198.19.0.0/16", "fallbackStrategy": "disabled" }
						Address: &net.Endpoint{
							Network: net.Network_UDP,
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Ip{
									Ip: []byte{127, 0, 0, 1},
								},
							},
							Port: uint32(port),
						},
						PrioritizedDomain: []*NameServer_PriorityDomain{
							{Type: DomainMatchingType_Subdomain, Domain: "google.com"},
						},
						FakeDns: &fakedns.FakeDnsPoolMulti{
							Pools: []*fakedns.FakeDnsPool{
								{IpPool: "198.19.0.0/16", LruSize: 256},
							},
						},
						FallbackStrategy: FallbackStrategy_Disabled.Enum(),
					},
				},
				FakeDns: &fakedns.FakeDnsPoolMulti{ // "fakedns": "198.18.0.0/16"
					Pools: []*fakedns.FakeDnsPool{
						{IpPool: "198.18.0.0/16", LruSize: 256},
					},
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(&policy.Config{}),
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	v, err := core.New(config)
	common.Must(err)
	common.Must(v.Start())

	dnsClient := v.GetFeature(feature_dns.ClientType()).(feature_dns.Client)
	fakeClient := dnsClient.(feature_dns.ClientWithFakeDNS).AsFakeDNSClient()

	var fakeIPForFacebook net.IP
	var fakeIPForGoogle net.IP

	{ // Lookup facebook.com with Fake Client will return 198.18.0.0/16 (global fake pool)
		ips, err := fakeClient.LookupIP("facebook.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{198, 18, 0, 0}, Mask: net.CIDRMask(16, 8*net.IPv4len)}).Contains(ip) {
				t.Fatal("Lookup facebook.com with fake client not in global pool 198.18.0.0/16")
			}
		}
		fakeIPForFacebook = ips[0]
	}
	{ // Lookup facebook.com with Normal Client with return empty record (because UDP server matching "domain:google.com" are configured with fallback disabled)
		_, err := dnsClient.LookupIP("facebook.com")
		if err != feature_dns.ErrEmptyResponse {
			t.Fatal("Lookup facebook.com with normal client not returning empty response")
		}
	}
	{ // Lookup google.com with Fake Client will return 198.19.0.0/16 (local fake pool)
		ips, err := fakeClient.LookupIP("google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{198, 19, 0, 0}, Mask: net.CIDRMask(16, 8*net.IPv4len)}).Contains(ip) {
				t.Fatal("Lookup google.com with fake client not in global pool 198.19.0.0/16")
			}
		}
		fakeIPForGoogle = ips[0]
	}
	{ // Lookup google.com with Normal Client will return 8.8.8.8
		ips, err := dnsClient.LookupIP("google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		if r := cmp.Diff(ips, []net.IP{{8, 8, 8, 8}}); r != "" {
			t.Fatal("Lookup google.com with normal client not returning 8.8.8.8")
		}
	}

	fakeEngine := dnsClient.(feature_dns.ClientWithFakeDNS).AsFakeDNSEngine().(feature_dns.FakeDNSEngineRev0)
	{
		if !fakeEngine.IsIPInIPPool(net.IPAddress(fakeIPForFacebook)) {
			t.Fatal("Fake IP of domain facebook.com not in FakeDNSEngine's pool.")
		}
		if !fakeEngine.IsIPInIPPool(net.IPAddress(fakeIPForGoogle)) {
			t.Fatal("Fake IP of domain google.com not in FakeDNSEngine's pool.")
		}
	}
	{
		if domain := fakeEngine.GetDomainFromFakeDNS(net.IPAddress(fakeIPForFacebook)); domain != "facebook.com" {
			t.Fatal("Recover fake IP to get domain facebook.com failed.")
		}
		if domain := fakeEngine.GetDomainFromFakeDNS(net.IPAddress(fakeIPForGoogle)); domain != "google.com" {
			t.Fatal("Recover fake IP to get domain google.com failed.")
		}
	}
	{
		ips := fakeEngine.GetFakeIPForDomain("api.google.com")
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{198, 19, 0, 0}, Mask: net.CIDRMask(16, 8*net.IPv4len)}).Contains(ip.IP()) {
				t.Fatal("Fake IP for api.google.com not in local pool 198.19.0.0/16")
			}
		}
	}
	{
		ips := fakeEngine.GetFakeIPForDomain3("v2fly.org", true, false)
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{198, 18, 0, 0}, Mask: net.CIDRMask(16, 8*net.IPv4len)}).Contains(ip.IP()) {
				t.Fatal("Fake IP for v2fly.org not in global pool 198.18.0.0/16")
			}
		}
	}
}

func TestFakeDNSEmptyGlobalConfig(t *testing.T) {
	config := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&Config{
				NameServer: []*NameServer{
					{ // "fakedns"
						Address: &net.Endpoint{
							Network: net.Network_UDP,
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Domain{
									Domain: "fakedns",
								},
							},
						},
						QueryStrategy: QueryStrategy_USE_IP4.Enum(),
					},
					{ // "localhost"
						Address: &net.Endpoint{
							Network: net.Network_UDP,
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Domain{
									Domain: "localhost",
								},
							},
						},
						QueryStrategy: QueryStrategy_USE_IP6.Enum(),
						PrioritizedDomain: []*NameServer_PriorityDomain{
							{Type: DomainMatchingType_Subdomain, Domain: "google.com"},
						},
						FakeDns: &fakedns.FakeDnsPoolMulti{Pools: []*fakedns.FakeDnsPool{}}, // "fakedns": true
					},
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(&policy.Config{}),
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	v, err := core.New(config)
	common.Must(err)
	common.Must(v.Start())

	dnsClient := v.GetFeature(feature_dns.ClientType()).(feature_dns.Client)
	fakeClient := dnsClient.(feature_dns.ClientWithFakeDNS).AsFakeDNSClient()

	{ // Lookup facebook.com will return 198.18.0.0/15 (default IPv4 pool)
		ips, err := fakeClient.LookupIP("facebook.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{198, 18, 0, 0}, Mask: net.CIDRMask(15, 8*net.IPv4len)}).Contains(ip) {
				t.Fatal("Lookup facebook.com with fake client not in default IPv4 pool 198.18.0.0/15")
			}
		}
	}
	{ // Lookup google.com will return fc00::/18 (default IPv6 pool)
		ips, err := fakeClient.LookupIP("google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
		for _, ip := range ips {
			if !(&net.IPNet{IP: net.IP{0xfc, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Mask: net.CIDRMask(18, 8*net.IPv6len)}).Contains(ip) {
				t.Fatal("Lookup google.com with fake client not in default IPv6 pool fc00::/18")
			}
		}
	}
}
