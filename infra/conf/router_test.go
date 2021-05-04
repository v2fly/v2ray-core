package conf_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	_ "unsafe"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	. "github.com/v2fly/v2ray-core/v4/infra/conf"
)

func init() {
	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	common.Must(os.MkdirAll(tempPath, 0755))

	if _, err := os.Stat(platform.GetAssetLocation("geoip.dat")); err != nil && errors.Is(err, fs.ErrNotExist) {
		if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
			geoipBytes, err := common.FetchHTTPContent(geoipURL)
			common.Must(err)
			common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
		}
	}
}

//go:linkname toCidrList github.com/v2fly/v2ray-core/v4/infra/conf.toCidrList
func toCidrList(ips StringList) ([]*router.GeoIP, error)

func TestToCidrList(t *testing.T) {
	t.Log(os.Getenv("v2ray.location.asset"))

	common.Must(filesystem.CopyFile(platform.GetAssetLocation("geoiptestrouter.dat"), platform.GetAssetLocation("geoip.dat")))

	ips := StringList([]string{
		"geoip:us",
		"geoip:cn",
		"geoip:!cn",
		"ext:geoiptestrouter.dat:!cn",
		"ext:geoiptestrouter.dat:ca",
		"ext-ip:geoiptestrouter.dat:!cn",
		"ext-ip:geoiptestrouter.dat:!ca",
	})

	_, err := toCidrList(ips)
	if err != nil {
		t.Fatalf("Failed to parse geoip list, got %s", err)
	}
}

func TestRouterConfig(t *testing.T) {
	createParser := func() func(string) (proto.Message, error) {
		return func(s string) (proto.Message, error) {
			config := new(RouterConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"strategy": "rules",
				"settings": {
					"domainStrategy": "AsIs",
					"rules": [
						{
							"type": "field",
							"domain": [
								"baidu.com",
								"qq.com"
							],
							"outboundTag": "direct"
						},
						{
							"type": "field",
							"domains": [
								"v2fly.org",
								"github.com"
							],
							"outboundTag": "direct"
						},
						{
							"type": "field",
							"ip": [
								"10.0.0.0/8",
								"::1/128"
							],
							"outboundTag": "test"
						},{
							"type": "field",
							"port": "53, 443, 1000-2000",
							"outboundTag": "test"
						},{
							"type": "field",
							"port": 123,
							"outboundTag": "test"
						}
					]
				},
				"balancers": [
					{
						"tag": "b1",
						"selector": ["test"]
					}
				]
			}`,
			Parser: createParser(),
			Output: &router.Config{
				DomainStrategy: router.Config_AsIs,
				BalancingRule: []*router.BalancingRule{
					{
						Tag:              "b1",
						OutboundSelector: []string{"test"},
						Strategy:         "random",
					},
				},
				Rule: []*router.RoutingRule{
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "baidu.com",
							},
							{
								Type:  router.Domain_Plain,
								Value: "qq.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "v2fly.org",
							},
							{
								Type:  router.Domain_Plain,
								Value: "github.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Geoip: []*router.GeoIP{
							{
								Cidr: []*router.CIDR{
									{
										Ip:     []byte{10, 0, 0, 0},
										Prefix: 8,
									},
									{
										Ip:     []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
										Prefix: 128,
									},
								},
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "test",
						},
					},
					{
						PortList: &net.PortList{
							Range: []*net.PortRange{
								{From: 53, To: 53},
								{From: 443, To: 443},
								{From: 1000, To: 2000},
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "test",
						},
					},
					{
						PortList: &net.PortList{
							Range: []*net.PortRange{
								{From: 123, To: 123},
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "test",
						},
					},
				},
			},
		},
		{
			Input: `{
				"strategy": "rules",
				"settings": {
					"domainStrategy": "IPIfNonMatch",
					"rules": [
						{
							"type": "field",
							"domain": [
								"baidu.com",
								"qq.com"
							],
							"outboundTag": "direct"
						},
						{
							"type": "field",
							"domains": [
								"v2fly.org",
								"github.com"
							],
							"outboundTag": "direct"
						},
						{
							"type": "field",
							"ip": [
								"10.0.0.0/8",
								"::1/128"
							],
							"outboundTag": "test"
						}
					]
				}
			}`,
			Parser: createParser(),
			Output: &router.Config{
				DomainStrategy: router.Config_IpIfNonMatch,
				Rule: []*router.RoutingRule{
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "baidu.com",
							},
							{
								Type:  router.Domain_Plain,
								Value: "qq.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "v2fly.org",
							},
							{
								Type:  router.Domain_Plain,
								Value: "github.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Geoip: []*router.GeoIP{
							{
								Cidr: []*router.CIDR{
									{
										Ip:     []byte{10, 0, 0, 0},
										Prefix: 8,
									},
									{
										Ip:     []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
										Prefix: 128,
									},
								},
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "test",
						},
					},
				},
			},
		},
		{
			Input: `{
				"domainStrategy": "AsIs",
				"rules": [
					{
						"type": "field",
						"domain": [
							"baidu.com",
							"qq.com"
						],
						"outboundTag": "direct"
					},
					{
						"type": "field",
						"domains": [
							"v2fly.org",
							"github.com"
						],
						"outboundTag": "direct"
					},
					{
						"type": "field",
						"ip": [
							"10.0.0.0/8",
							"::1/128"
						],
						"outboundTag": "test"
					}
				]
			}`,
			Parser: createParser(),
			Output: &router.Config{
				DomainStrategy: router.Config_AsIs,
				Rule: []*router.RoutingRule{
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "baidu.com",
							},
							{
								Type:  router.Domain_Plain,
								Value: "qq.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Domain: []*router.Domain{
							{
								Type:  router.Domain_Plain,
								Value: "v2fly.org",
							},
							{
								Type:  router.Domain_Plain,
								Value: "github.com",
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
					},
					{
						Geoip: []*router.GeoIP{
							{
								Cidr: []*router.CIDR{
									{
										Ip:     []byte{10, 0, 0, 0},
										Prefix: 8,
									},
									{
										Ip:     []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
										Prefix: 128,
									},
								},
							},
						},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "test",
						},
					},
				},
			},
		},
	})
}
