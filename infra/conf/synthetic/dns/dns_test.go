package dns_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/v2fly/v2ray-core/v5/app/dns"
	"github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	_ "github.com/v2fly/v2ray-core/v5/infra/conf/geodata/standard"
	dns2 "github.com/v2fly/v2ray-core/v5/infra/conf/synthetic/dns"
)

func init() {
	const (
		geoipURL   = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"
		geositeURL = "https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat"
	)

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")
	geositePath := filepath.Join(tempPath, "geosite.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
	}
	if _, err := os.Stat(geositePath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geositeBytes, err := common.FetchHTTPContent(geositeURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geositePath, geositeBytes))
	}
}

func TestDNSConfigParsing(t *testing.T) {
	parserCreator := func() func(string) (protoiface.MessageV1, error) {
		return func(s string) (protoiface.MessageV1, error) {
			config := new(dns2.DNSConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"servers": [{
					"address": "8.8.8.8",
					"clientIp": "10.0.0.1",
					"port": 5353,
					"skipFallback": true,
					"domains": ["domain:v2fly.org"]
				}],
				"hosts": {
					"v2fly.org": "127.0.0.1",
					"www.v2fly.org": ["1.2.3.4", "5.6.7.8"],
					"domain:example.com": "google.com",
					"geosite:test": ["127.0.0.1", "127.0.0.2"],
					"keyword:google": ["8.8.8.8", "8.8.4.4"],
					"regexp:.*\\.com": "8.8.4.4"
				},
				"clientIp": "10.0.0.1",
				"queryStrategy": "UseIPv4",
				"disableCache": true,
				"disableFallback": true
			}`,
			Parser: parserCreator(),
			Output: &dns.Config{
				NameServer: []*dns.NameServer{
					{
						Address: &net.Endpoint{
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Ip{
									Ip: []byte{8, 8, 8, 8},
								},
							},
							Network: net.Network_UDP,
							Port:    5353,
						},
						ClientIp:     []byte{10, 0, 0, 1},
						SkipFallback: true,
						PrioritizedDomain: []*dns.NameServer_PriorityDomain{
							{
								Type:   dns.DomainMatchingType_Subdomain,
								Domain: "v2fly.org",
							},
						},
						OriginalRules: []*dns.NameServer_OriginalRule{
							{
								Rule: "domain:v2fly.org",
								Size: 1,
							},
						},
					},
				},
				StaticHosts: []*dns.HostMapping{
					{
						Type:          dns.DomainMatchingType_Subdomain,
						Domain:        "example.com",
						ProxiedDomain: "google.com",
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "test.example.com",
						Ip:     [][]byte{{127, 0, 0, 1}, {127, 0, 0, 2}},
					},
					{
						Type:   dns.DomainMatchingType_Keyword,
						Domain: "google",
						Ip:     [][]byte{{8, 8, 8, 8}, {8, 8, 4, 4}},
					},
					{
						Type:   dns.DomainMatchingType_Regex,
						Domain: ".*\\.com",
						Ip:     [][]byte{{8, 8, 4, 4}},
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "v2fly.org",
						Ip:     [][]byte{{127, 0, 0, 1}},
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "www.v2fly.org",
						Ip:     [][]byte{{1, 2, 3, 4}, {5, 6, 7, 8}},
					},
				},
				ClientIp:        []byte{10, 0, 0, 1},
				QueryStrategy:   dns.QueryStrategy_USE_IP4,
				DisableCache:    true,
				DisableFallback: true,
			},
		},
		{
			Input: `{
				"servers": [{
					"address": "fakedns",
					"tag": "fake",
					"queryStrategy": "UseIPv6",
					"fallbackStrategy": "disabledIfAnyMatch",
					"fakedns": true
				}, {
					"address": "8.8.8.8",
					"port": 5353,
					"tag": "local",
					"clientIp": "10.0.0.1",
					"queryStrategy": "UseIP",
					"cacheStrategy": "enabled",
					"fallbackStrategy": "disabled",
					"domains": ["domain:v2fly.org"],
					"fakedns": ["198.19.0.0/16", "fc01::/18"]
				}],
				"hosts": {
					"v2fly.org": "127.0.0.1",
					"www.v2fly.org": ["1.2.3.4", "5.6.7.8"],
					"domain:example.com": "google.com",
					"geosite:test": ["127.0.0.1", "127.0.0.2"],
					"keyword:google": ["8.8.8.8", "8.8.4.4"],
					"regexp:.*\\.com": "8.8.4.4"
				},
				"fakedns": [
					{ "ipPool": "198.18.0.0/16", "poolSize": 32768 },
					{ "ipPool": "fc00::/18", "poolSize": 32768 }
				],
				"tag": "global",
				"clientIp": "10.0.0.1",
				"queryStrategy": "UseIPv4",
				"cacheStrategy": "disabled",
				"fallbackStrategy": "enabled"
			}`,
			Parser: parserCreator(),
			Output: &dns.Config{
				NameServer: []*dns.NameServer{
					{
						Address: &net.Endpoint{
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Domain{
									Domain: "fakedns",
								},
							},
							Network: net.Network_UDP,
						},
						Tag:              "fake",
						QueryStrategy:    dns.QueryStrategy_USE_IP6.Enum(),
						FallbackStrategy: dns.FallbackStrategy_DisabledIfAnyMatch.Enum(),
						FakeDns: &fakedns.FakeDnsPoolMulti{
							Pools: []*fakedns.FakeDnsPool{},
						},
					},
					{
						Address: &net.Endpoint{
							Address: &net.IPOrDomain{
								Address: &net.IPOrDomain_Ip{
									Ip: []byte{8, 8, 8, 8},
								},
							},
							Network: net.Network_UDP,
							Port:    5353,
						},
						Tag:              "local",
						ClientIp:         []byte{10, 0, 0, 1},
						QueryStrategy:    dns.QueryStrategy_USE_IP.Enum(),
						CacheStrategy:    dns.CacheStrategy_CacheEnabled.Enum(),
						FallbackStrategy: dns.FallbackStrategy_Disabled.Enum(),
						PrioritizedDomain: []*dns.NameServer_PriorityDomain{
							{
								Type:   dns.DomainMatchingType_Subdomain,
								Domain: "v2fly.org",
							},
						},
						OriginalRules: []*dns.NameServer_OriginalRule{
							{
								Rule: "domain:v2fly.org",
								Size: 1,
							},
						},
						FakeDns: &fakedns.FakeDnsPoolMulti{
							Pools: []*fakedns.FakeDnsPool{
								{IpPool: "198.19.0.0/16", LruSize: 65535},
								{IpPool: "fc01::/18", LruSize: 65535},
							},
						},
					},
				},
				StaticHosts: []*dns.HostMapping{
					{
						Type:          dns.DomainMatchingType_Subdomain,
						Domain:        "example.com",
						ProxiedDomain: "google.com",
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "test.example.com",
						Ip:     [][]byte{{127, 0, 0, 1}, {127, 0, 0, 2}},
					},
					{
						Type:   dns.DomainMatchingType_Keyword,
						Domain: "google",
						Ip:     [][]byte{{8, 8, 8, 8}, {8, 8, 4, 4}},
					},
					{
						Type:   dns.DomainMatchingType_Regex,
						Domain: ".*\\.com",
						Ip:     [][]byte{{8, 8, 4, 4}},
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "v2fly.org",
						Ip:     [][]byte{{127, 0, 0, 1}},
					},
					{
						Type:   dns.DomainMatchingType_Full,
						Domain: "www.v2fly.org",
						Ip:     [][]byte{{1, 2, 3, 4}, {5, 6, 7, 8}},
					},
				},
				FakeDns: &fakedns.FakeDnsPoolMulti{
					Pools: []*fakedns.FakeDnsPool{
						{IpPool: "198.18.0.0/16", LruSize: 32768},
						{IpPool: "fc00::/18", LruSize: 32768},
					},
				},
				Tag:              "global",
				ClientIp:         []byte{10, 0, 0, 1},
				QueryStrategy:    dns.QueryStrategy_USE_IP4,
				CacheStrategy:    dns.CacheStrategy_CacheDisabled,
				FallbackStrategy: dns.FallbackStrategy_Enabled,
			},
		},
	})
}
