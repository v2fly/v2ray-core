package fakedns

import (
	gonet "net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

func TestNewFakeDnsHolder(_ *testing.T) {
	_, err := NewFakeDNSHolder()
	common.Must(err)
}

func TestFakeDnsHolderCreateMapping(t *testing.T) {
	fkdns, err := NewFakeDNSHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "198.18.0.0", addr[0].IP().String())
}

func TestFakeDnsHolderCreateMappingMany(t *testing.T) {
	fkdns, err := NewFakeDNSHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "198.18.0.0", addr[0].IP().String())

	addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
	assert.Equal(t, "198.18.0.1", addr2[0].IP().String())
}

func TestFakeDnsHolderCreateMappingManyAndResolve(t *testing.T) {
	fkdns, err := NewFakeDNSHolder()
	common.Must(err)

	{
		addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
		assert.Equal(t, "198.18.0.0", addr[0].IP().String())
	}

	{
		addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
		assert.Equal(t, "198.18.0.1", addr2[0].IP().String())
	}

	{
		result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("198.18.0.0"))
		assert.Equal(t, "fakednstest.v2fly.org", result)
	}

	{
		result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("198.18.0.1"))
		assert.Equal(t, "fakednstest2.v2fly.org", result)
	}
}

func TestFakeDnsHolderCreateMappingManySingleDomain(t *testing.T) {
	fkdns, err := NewFakeDNSHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "198.18.0.0", addr[0].IP().String())

	addr2 := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "198.18.0.0", addr2[0].IP().String())
}

func TestFakeDnsHolderCreateMappingAndRollOver(t *testing.T) {
	fkdns, err := NewFakeDNSHolderConfigOnly(&FakeDnsPool{
		IpPool:  "240.0.0.0/12",
		LruSize: 256,
	})
	common.Must(err)

	err = fkdns.Start()

	common.Must(err)

	{
		addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
		assert.Equal(t, "240.0.0.0", addr[0].IP().String())
	}

	{
		addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
		assert.Equal(t, "240.0.0.1", addr2[0].IP().String())
	}

	for i := 0; i <= 8192; i++ {
		{
			result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("240.0.0.0"))
			assert.Equal(t, "fakednstest.v2fly.org", result)
		}

		{
			result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("240.0.0.1"))
			assert.Equal(t, "fakednstest2.v2fly.org", result)
		}

		{
			uuid := uuid.New()
			domain := uuid.String() + ".fakednstest.v2fly.org"
			addr := fkdns.GetFakeIPForDomain(domain)
			rsaddr := addr[0].IP().String()

			result := fkdns.GetDomainFromFakeDNS(net.ParseAddress(rsaddr))
			assert.Equal(t, domain, result)
		}
	}
}

func TestFakeDNSMulti(t *testing.T) {
	fakeMulti, err := NewFakeDNSHolderMulti(&FakeDnsPoolMulti{
		Pools: []*FakeDnsPool{{
			IpPool:  "240.0.0.0/12",
			LruSize: 256,
		}, {
			IpPool:  "fddd:c5b4:ff5f:f4f0::/64",
			LruSize: 256,
		}},
	},
	)
	common.Must(err)

	err = fakeMulti.Start()

	common.Must(err)

	assert.Nil(t, err, "Should not throw error")
	_ = fakeMulti

	t.Run("checkInRange", func(t *testing.T) {
		t.Run("ipv4", func(t *testing.T) {
			inPool := fakeMulti.IsIPInIPPool(net.IPAddress([]byte{240, 0, 0, 5}))
			assert.True(t, inPool)
		})
		t.Run("ipv6", func(t *testing.T) {
			ip, err := gonet.ResolveIPAddr("ip", "fddd:c5b4:ff5f:f4f0::5")
			assert.Nil(t, err)
			inPool := fakeMulti.IsIPInIPPool(net.IPAddress(ip.IP))
			assert.True(t, inPool)
		})
		t.Run("ipv4_inverse", func(t *testing.T) {
			inPool := fakeMulti.IsIPInIPPool(net.IPAddress([]byte{241, 0, 0, 5}))
			assert.False(t, inPool)
		})
		t.Run("ipv6_inverse", func(t *testing.T) {
			ip, err := gonet.ResolveIPAddr("ip", "fcdd:c5b4:ff5f:f4f0::5")
			assert.Nil(t, err)
			inPool := fakeMulti.IsIPInIPPool(net.IPAddress(ip.IP))
			assert.False(t, inPool)
		})
	})

	t.Run("allocateTwoAddressForTwoPool", func(t *testing.T) {
		address := fakeMulti.GetFakeIPForDomain("fakednstest.v2fly.org")
		assert.Len(t, address, 2, "should be 2 address one for each pool")
		t.Run("eachOfThemShouldResolve:0", func(t *testing.T) {
			domain := fakeMulti.GetDomainFromFakeDNS(address[0])
			assert.Equal(t, "fakednstest.v2fly.org", domain)
		})
		t.Run("eachOfThemShouldResolve:1", func(t *testing.T) {
			domain := fakeMulti.GetDomainFromFakeDNS(address[1])
			assert.Equal(t, "fakednstest.v2fly.org", domain)
		})
	})

	t.Run("understandIPTypeSelector", func(t *testing.T) {
		t.Run("ipv4", func(t *testing.T) {
			address := fakeMulti.GetFakeIPForDomain3("fakednstestipv4.v2fly.org", true, false)
			assert.Len(t, address, 1, "should be 1 address")
			assert.True(t, address[0].Family().IsIPv4())
		})
		t.Run("ipv6", func(t *testing.T) {
			address := fakeMulti.GetFakeIPForDomain3("fakednstestipv6.v2fly.org", false, true)
			assert.Len(t, address, 1, "should be 1 address")
			assert.True(t, address[0].Family().IsIPv6())
		})
		t.Run("ipv46", func(t *testing.T) {
			address := fakeMulti.GetFakeIPForDomain3("fakednstestipv46.v2fly.org", true, true)
			assert.Len(t, address, 2, "should be 2 address")
			assert.True(t, address[0].Family().IsIPv4())
			assert.True(t, address[1].Family().IsIPv6())
		})
	})
}
