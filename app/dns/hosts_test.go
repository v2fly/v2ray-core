package dns_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	. "github.com/ghxhy/v2ray-core/v5/app/dns"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/features/dns"
)

func TestStaticHosts(t *testing.T) {
	pb := []*HostMapping{
		{
			Type:   DomainMatchingType_Full,
			Domain: "v2fly.org",
			Ip: [][]byte{
				{1, 1, 1, 1},
			},
		},
		{
			Type:   DomainMatchingType_Full,
			Domain: "proxy.v2fly.org",
			Ip: [][]byte{
				{1, 2, 3, 4},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
			ProxiedDomain: "another-proxy.v2fly.org",
		},
		{
			Type:          DomainMatchingType_Full,
			Domain:        "proxy2.v2fly.org",
			ProxiedDomain: "proxy.v2fly.org",
		},
		{
			Type:   DomainMatchingType_Subdomain,
			Domain: "v2ray.cn",
			Ip: [][]byte{
				{2, 2, 2, 2},
			},
		},
		{
			Type:   DomainMatchingType_Subdomain,
			Domain: "baidu.com",
			Ip: [][]byte{
				{127, 0, 0, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
		},
	}

	hosts, err := NewStaticHosts(pb, nil)
	common.Must(err)

	{
		ips := hosts.Lookup("v2fly.org", dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: true,
		})
		if len(ips) != 1 {
			t.Error("expect 1 IP, but got ", len(ips))
		}
		if diff := cmp.Diff([]byte(ips[0].IP()), []byte{1, 1, 1, 1}); diff != "" {
			t.Error(diff)
		}
	}

	{
		domain := hosts.Lookup("proxy.v2fly.org", dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: false,
		})
		if len(domain) != 1 {
			t.Error("expect 1 domain, but got ", len(domain))
		}
		if diff := cmp.Diff(domain[0].Domain(), "another-proxy.v2fly.org"); diff != "" {
			t.Error(diff)
		}
	}

	{
		domain := hosts.Lookup("proxy2.v2fly.org", dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: false,
		})
		if len(domain) != 1 {
			t.Error("expect 1 domain, but got ", len(domain))
		}
		if diff := cmp.Diff(domain[0].Domain(), "another-proxy.v2fly.org"); diff != "" {
			t.Error(diff)
		}
	}

	{
		ips := hosts.Lookup("www.v2ray.cn", dns.IPOption{
			IPv4Enable: true,
			IPv6Enable: true,
		})
		if len(ips) != 1 {
			t.Error("expect 1 IP, but got ", len(ips))
		}
		if diff := cmp.Diff([]byte(ips[0].IP()), []byte{2, 2, 2, 2}); diff != "" {
			t.Error(diff)
		}
	}

	{
		ips := hosts.Lookup("baidu.com", dns.IPOption{
			IPv4Enable: false,
			IPv6Enable: true,
		})
		if len(ips) != 1 {
			t.Error("expect 1 IP, but got ", len(ips))
		}
		if diff := cmp.Diff([]byte(ips[0].IP()), []byte(net.LocalHostIPv6.IP())); diff != "" {
			t.Error(diff)
		}
	}
}
