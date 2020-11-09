package dns

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"v2ray.com/core/common"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/uuid"
)

func TestNewFakeDnsHolder(t *testing.T) {
	_, err := NewFakeDnsHolder()
	common.Must(err)
}

func TestFakeDnsHolderCreateMapping(t *testing.T) {
	fkdns, err := NewFakeDnsHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "240.0.0.0", addr[0].IP().String())
}

func TestFakeDnsHolderCreateMappingMany(t *testing.T) {
	fkdns, err := NewFakeDnsHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "240.0.0.0", addr[0].IP().String())

	addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
	assert.Equal(t, "240.0.0.1", addr2[0].IP().String())
}

func TestFakeDnsHolderCreateMappingManyAndResolve(t *testing.T) {
	fkdns, err := NewFakeDnsHolder()
	common.Must(err)

	{
		addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
		assert.Equal(t, "240.0.0.0", addr[0].IP().String())
	}

	{
		addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
		assert.Equal(t, "240.0.0.1", addr2[0].IP().String())
	}

	{
		result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("240.0.0.0"))
		assert.Equal(t, "fakednstest.v2fly.org", result)
	}

	{
		result := fkdns.GetDomainFromFakeDNS(net.ParseAddress("240.0.0.1"))
		assert.Equal(t, "fakednstest2.v2fly.org", result)
	}
}

func TestFakeDnsHolderCreateMappingManySingleDomain(t *testing.T) {
	fkdns, err := NewFakeDnsHolder()
	common.Must(err)

	addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "240.0.0.0", addr[0].IP().String())

	addr2 := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
	assert.Equal(t, "240.0.0.0", addr2[0].IP().String())
}

func TestFakeDnsHolderCreateMappingAndRollOver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DNS Holder RollOver test in short mode. ~190s")
	}

	fkdns, err := NewFakeDnsHolder()
	common.Must(err)

	{
		addr := fkdns.GetFakeIPForDomain("fakednstest.v2fly.org")
		assert.Equal(t, "240.0.0.0", addr[0].IP().String())
	}

	{
		addr2 := fkdns.GetFakeIPForDomain("fakednstest2.v2fly.org")
		assert.Equal(t, "240.0.0.1", addr2[0].IP().String())
	}

	for i := 0; i <= 33554432; i++ {
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
