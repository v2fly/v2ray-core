package fakedns

import (
	gonet "net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/uuid"
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

func TestGetFakeIPForDomainConcurrently(t *testing.T) {
	fkdns, err := NewFakeDNSHolder()
	common.Must(err)

	total := 200
	addr := make([][]net.Address, total+1)
	var errg errgroup.Group
	for i := 0; i < total; i++ {
		errg.Go(testGetFakeIP(i, addr, fkdns))
	}
	errg.Wait()
	for i := 0; i < total; i++ {
		for j := i + 1; j < total; j++ {
			assert.NotEqual(t, addr[i][0].IP().String(), addr[j][0].IP().String())
		}
	}
}

func testGetFakeIP(index int, addr [][]net.Address, fkdns *Holder) func() error {
	return func() error {
		addr[index] = fkdns.GetFakeIPForDomain("fakednstest" + strconv.Itoa(index) + ".example.com")
		return nil
	}
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

func TestFakeDNSMultiAddPool(t *testing.T) {
	runTest := func(runTestBeforeStart bool) {
		fakeMulti, err := NewFakeDNSHolderMulti(&FakeDnsPoolMulti{
			Pools: []*FakeDnsPool{{
				IpPool:  "240.0.0.0/12",
				LruSize: 256,
			}, {
				IpPool:  "fddd:c5b4:ff5f:f4f0::/64",
				LruSize: 256,
			}},
		})
		common.Must(err)
		if !runTestBeforeStart {
			err = fakeMulti.Start()
			common.Must(err)
		}
		t.Run("ipv4_return_existing", func(t *testing.T) {
			pool, err := fakeMulti.AddPool(&FakeDnsPool{
				IpPool:  "240.0.0.1/12",
				LruSize: 256,
			})
			common.Must(err)
			if pool != fakeMulti.holders[0] {
				t.Error("HolderMulti.AddPool not returning same holder for existing IPv4 pool")
			}
		})
		t.Run("ipv6_return_existing", func(t *testing.T) {
			pool, err := fakeMulti.AddPool(&FakeDnsPool{
				IpPool:  "fddd:c5b4:ff5f:f4f0::1/64",
				LruSize: 256,
			})
			common.Must(err)
			if pool != fakeMulti.holders[1] {
				t.Error("HolderMulti.AddPool not returning same holder for existing IPv6 pool")
			}
		})
		t.Run("ipv4_reject_overlap", func(t *testing.T) {
			_, err := fakeMulti.AddPool(&FakeDnsPool{
				IpPool:  "240.8.0.0/13",
				LruSize: 256,
			})
			if err == nil {
				t.Error("HolderMulti.AddPool not rejecting IPv4 pool that is subnet of existing ones")
			}
			_, err = fakeMulti.AddPool(&FakeDnsPool{
				IpPool:  "240.0.0.0/11",
				LruSize: 256,
			})
			if err == nil {
				t.Error("HolderMulti.AddPool not rejecting IPv4 pool that contains existing ones")
			}
		})
		t.Run("new_pool", func(t *testing.T) {
			pool, err := fakeMulti.AddPool(&FakeDnsPool{
				IpPool:  "192.168.168.0/16",
				LruSize: 256,
			})
			common.Must(err)
			if pool != fakeMulti.holders[2] {
				t.Error("HolderMulti.AddPool not creating new holder for new IPv4 pool")
			}
		})
		t.Run("add_pool_multi", func(t *testing.T) {
			pools, err := fakeMulti.AddPoolMulti(&FakeDnsPoolMulti{
				Pools: []*FakeDnsPool{{
					IpPool:  "192.168.168.0/16",
					LruSize: 256,
				}, {
					IpPool:  "2001:1111::/64",
					LruSize: 256,
				}},
			})
			common.Must(err)
			if len(pools.holders) != 2 {
				t.Error("HolderMulti.AddPoolMutli not returning holderMulti that has the same length as passed PoolMulti config")
			}
			if pools.holders[0] != fakeMulti.holders[2] {
				t.Error("HolderMulti.AddPoolMulti not returning same holder for existing IPv4 pool 192.168.168.0/16")
			}
			if pools.holders[1] != fakeMulti.holders[3] {
				t.Error("HolderMulti.AddPoolMulti not creating new holder for new IPv6 pool 2001:1111::/64")
			}
		})
		if runTestBeforeStart {
			err = fakeMulti.Start()
			common.Must(err)
		}
	}
	t.Run("addPoolBeforeStart", func(t *testing.T) {
		runTest(true)
	})
	t.Run("addPoolAfterStart", func(t *testing.T) {
		runTest(false)
	})
}
