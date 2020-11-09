package dns

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"v2ray.com/core/common"
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
