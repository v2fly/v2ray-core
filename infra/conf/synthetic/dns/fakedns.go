package dns

import (
	"encoding/json"
	"net"

	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
)

type FakeDNSPoolElementConfig struct {
	IPPool  string `json:"ipPool"`
	LRUSize int64  `json:"poolSize"`
}

type FakeDNSConfig struct {
	pool  *FakeDNSPoolElementConfig
	pools []*FakeDNSPoolElementConfig
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON
func (f *FakeDNSConfig) UnmarshalJSON(data []byte) error {
	var pool FakeDNSPoolElementConfig
	var pools []*FakeDNSPoolElementConfig
	var ipPools []string
	switch {
	case json.Unmarshal(data, &pool) == nil:
		f.pool = &pool
	case json.Unmarshal(data, &pools) == nil:
		f.pools = pools
	case json.Unmarshal(data, &ipPools) == nil:
		f.pools = make([]*FakeDNSPoolElementConfig, 0, len(ipPools))
		for _, ipPool := range ipPools {
			_, ipNet, err := net.ParseCIDR(ipPool)
			if err != nil {
				return err
			}
			ones, bits := ipNet.Mask.Size()
			sizeInBits := bits - ones
			if sizeInBits > 16 { // At most 65536 ips for a IP pool
				sizeInBits = 16
			}
			f.pools = append(f.pools, &FakeDNSPoolElementConfig{
				IPPool:  ipPool,
				LRUSize: (1 << sizeInBits) - 1,
			})
		}
	default:
		return newError("invalid fakedns config")
	}
	return nil
}

func (f *FakeDNSConfig) Build() (*fakedns.FakeDnsPoolMulti, error) {
	fakeDNSPool := fakedns.FakeDnsPoolMulti{}

	if f.pool != nil {
		fakeDNSPool.Pools = append(fakeDNSPool.Pools, &fakedns.FakeDnsPool{
			IpPool:  f.pool.IPPool,
			LruSize: f.pool.LRUSize,
		})
		return &fakeDNSPool, nil
	}

	if f.pools != nil {
		for _, v := range f.pools {
			fakeDNSPool.Pools = append(fakeDNSPool.Pools, &fakedns.FakeDnsPool{IpPool: v.IPPool, LruSize: v.LRUSize})
		}
		return &fakeDNSPool, nil
	}

	return nil, newError("no valid FakeDNS config")
}

type FakeDNSConfigExtend struct { // Adds boolean value parsing for "fakedns" config
	*FakeDNSConfig
}

func (f *FakeDNSConfigExtend) UnmarshalJSON(data []byte) error {
	var enabled bool
	if json.Unmarshal(data, &enabled) == nil {
		if enabled {
			f.FakeDNSConfig = &FakeDNSConfig{pools: []*FakeDNSPoolElementConfig{}}
		}
		return nil
	}
	return json.Unmarshal(data, &f.FakeDNSConfig)
}
