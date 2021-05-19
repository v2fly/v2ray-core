package conf_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/v2fly/v2ray-core/v4/app/dns"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/standard"
)

func init() {
	const (
		geoipURL   = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"
		geositeURL = "https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat"
	)

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
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
			config := new(conf.DNSConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}

	runMultiTestCase(t, []TestCase{
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
				StaticHosts: []*dns.Config_HostMapping{
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
	})
}
