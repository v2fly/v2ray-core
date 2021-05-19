package router_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/protocol/http"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/features/routing"
	routing_session "github.com/v2fly/v2ray-core/v4/features/routing/session"
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

func withBackground() routing.Context {
	return &routing_session.Context{}
}

func withOutbound(outbound *session.Outbound) routing.Context {
	return &routing_session.Context{Outbound: outbound}
}

func withInbound(inbound *session.Inbound) routing.Context {
	return &routing_session.Context{Inbound: inbound}
}

func withContent(content *session.Content) routing.Context {
	return &routing_session.Context{Content: content}
}

func TestRoutingRule(t *testing.T) {
	type ruleTest struct {
		input  routing.Context
		output bool
	}

	cases := []struct {
		rule *router.RoutingRule
		test []ruleTest
	}{
		{
			rule: &router.RoutingRule{
				Domain: []*router.Domain{
					{
						Value: "v2fly.org",
						Type:  router.Domain_Plain,
					},
					{
						Value: "google.com",
						Type:  router.Domain_Domain,
					},
					{
						Value: "^facebook\\.com$",
						Type:  router.Domain_Regex,
					},
				},
			},
			test: []ruleTest{
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("v2fly.org"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("www.v2fly.org.www"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("v2ray.co"), 80)}),
					output: false,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("www.google.com"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("facebook.com"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.DomainAddress("www.facebook.com"), 80)}),
					output: false,
				},
				{
					input:  withBackground(),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				Cidr: []*router.CIDR{
					{
						Ip:     []byte{8, 8, 8, 8},
						Prefix: 32,
					},
					{
						Ip:     []byte{8, 8, 8, 8},
						Prefix: 32,
					},
					{
						Ip:     net.ParseAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334").IP(),
						Prefix: 128,
					},
				},
			},
			test: []ruleTest{
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("8.8.8.8"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("8.8.4.4"), 80)}),
					output: false,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), 80)}),
					output: true,
				},
				{
					input:  withBackground(),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				Geoip: []*router.GeoIP{
					{
						Cidr: []*router.CIDR{
							{
								Ip:     []byte{8, 8, 8, 8},
								Prefix: 32,
							},
							{
								Ip:     []byte{8, 8, 8, 8},
								Prefix: 32,
							},
							{
								Ip:     net.ParseAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334").IP(),
								Prefix: 128,
							},
						},
					},
				},
			},
			test: []ruleTest{
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("8.8.8.8"), 80)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("8.8.4.4"), 80)}),
					output: false,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), 80)}),
					output: true,
				},
				{
					input:  withBackground(),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				SourceCidr: []*router.CIDR{
					{
						Ip:     []byte{192, 168, 0, 0},
						Prefix: 16,
					},
				},
			},
			test: []ruleTest{
				{
					input:  withInbound(&session.Inbound{Source: net.TCPDestination(net.ParseAddress("192.168.0.1"), 80)}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{Source: net.TCPDestination(net.ParseAddress("10.0.0.1"), 80)}),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				UserEmail: []string{
					"admin@v2fly.org",
				},
			},
			test: []ruleTest{
				{
					input:  withInbound(&session.Inbound{User: &protocol.MemoryUser{Email: "admin@v2fly.org"}}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{User: &protocol.MemoryUser{Email: "love@v2fly.org"}}),
					output: false,
				},
				{
					input:  withBackground(),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				Protocol: []string{"http"},
			},
			test: []ruleTest{
				{
					input:  withContent(&session.Content{Protocol: (&http.SniffHeader{}).Protocol()}),
					output: true,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				InboundTag: []string{"test", "test1"},
			},
			test: []ruleTest{
				{
					input:  withInbound(&session.Inbound{Tag: "test"}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{Tag: "test2"}),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				PortList: &net.PortList{
					Range: []*net.PortRange{
						{From: 443, To: 443},
						{From: 1000, To: 1100},
					},
				},
			},
			test: []ruleTest{
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.LocalHostIP, 443)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.LocalHostIP, 1100)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.LocalHostIP, 1005)}),
					output: true,
				},
				{
					input:  withOutbound(&session.Outbound{Target: net.TCPDestination(net.LocalHostIP, 53)}),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				SourcePortList: &net.PortList{
					Range: []*net.PortRange{
						{From: 123, To: 123},
						{From: 9993, To: 9999},
					},
				},
			},
			test: []ruleTest{
				{
					input:  withInbound(&session.Inbound{Source: net.UDPDestination(net.LocalHostIP, 123)}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{Source: net.UDPDestination(net.LocalHostIP, 9999)}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{Source: net.UDPDestination(net.LocalHostIP, 9994)}),
					output: true,
				},
				{
					input:  withInbound(&session.Inbound{Source: net.UDPDestination(net.LocalHostIP, 53)}),
					output: false,
				},
			},
		},
		{
			rule: &router.RoutingRule{
				Protocol:   []string{"http"},
				Attributes: "attrs[':path'].startswith('/test')",
			},
			test: []ruleTest{
				{
					input:  withContent(&session.Content{Protocol: "http/1.1", Attributes: map[string]string{":path": "/test/1"}}),
					output: true,
				},
			},
		},
	}

	for _, test := range cases {
		cond, err := test.rule.BuildCondition()
		common.Must(err)

		for _, subtest := range test.test {
			actual := cond.Apply(subtest.input)
			if actual != subtest.output {
				t.Error("test case failed: ", subtest.input, " expected ", subtest.output, " but got ", actual)
			}
		}
	}
}

func loadGeoSite(country string) ([]*router.Domain, error) {
	geositeBytes, err := filesystem.ReadAsset("geosite.dat")
	if err != nil {
		return nil, err
	}
	var geositeList router.GeoSiteList
	if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
		return nil, err
	}

	for _, site := range geositeList.Entry {
		if strings.EqualFold(site.CountryCode, country) {
			return site.Domain, nil
		}
	}

	return nil, errors.New("country not found: " + country)
}

func TestChinaSites(t *testing.T) {
	domains, err := loadGeoSite("CN")
	common.Must(err)

	matcher, err := router.NewDomainMatcher(domains)
	common.Must(err)
	acMatcher, err := router.NewMphMatcherGroup(domains)
	common.Must(err)

	type TestCase struct {
		Domain string
		Output bool
	}
	testCases := []TestCase{
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "164.com",
			Output: false,
		},
		{
			Domain: "164.com",
			Output: false,
		},
	}

	for i := 0; i < 1024; i++ {
		testCases = append(testCases, TestCase{Domain: strconv.Itoa(i) + ".not-exists.com", Output: false})
	}

	for _, testCase := range testCases {
		r1 := matcher.ApplyDomain(testCase.Domain)
		r2 := acMatcher.ApplyDomain(testCase.Domain)
		if r1 != testCase.Output {
			t.Error("DomainMatcher expected output ", testCase.Output, " for domain ", testCase.Domain, " but got ", r1)
		} else if r2 != testCase.Output {
			t.Error("ACDomainMatcher expected output ", testCase.Output, " for domain ", testCase.Domain, " but got ", r2)
		}
	}
}

func BenchmarkMphDomainMatcher(b *testing.B) {
	domains, err := loadGeoSite("CN")
	common.Must(err)

	matcher, err := router.NewMphMatcherGroup(domains)
	common.Must(err)

	type TestCase struct {
		Domain string
		Output bool
	}
	testCases := []TestCase{
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "164.com",
			Output: false,
		},
		{
			Domain: "164.com",
			Output: false,
		},
	}

	for i := 0; i < 1024; i++ {
		testCases = append(testCases, TestCase{Domain: strconv.Itoa(i) + ".not-exists.com", Output: false})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range testCases {
			_ = matcher.ApplyDomain(testCase.Domain)
		}
	}
}

func BenchmarkDomainMatcher(b *testing.B) {
	domains, err := loadGeoSite("CN")
	common.Must(err)

	matcher, err := router.NewDomainMatcher(domains)
	common.Must(err)

	type TestCase struct {
		Domain string
		Output bool
	}
	testCases := []TestCase{
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "163.com",
			Output: true,
		},
		{
			Domain: "164.com",
			Output: false,
		},
		{
			Domain: "164.com",
			Output: false,
		},
	}

	for i := 0; i < 1024; i++ {
		testCases = append(testCases, TestCase{Domain: strconv.Itoa(i) + ".not-exists.com", Output: false})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range testCases {
			_ = matcher.ApplyDomain(testCase.Domain)
		}
	}
}

func BenchmarkMultiGeoIPMatcher(b *testing.B) {
	var geoips []*router.GeoIP

	{
		ips, err := loadGeoIP("CN")
		common.Must(err)
		geoips = append(geoips, &router.GeoIP{
			CountryCode: "CN",
			Cidr:        ips,
		})
	}

	{
		ips, err := loadGeoIP("JP")
		common.Must(err)
		geoips = append(geoips, &router.GeoIP{
			CountryCode: "JP",
			Cidr:        ips,
		})
	}

	{
		ips, err := loadGeoIP("CA")
		common.Must(err)
		geoips = append(geoips, &router.GeoIP{
			CountryCode: "CA",
			Cidr:        ips,
		})
	}

	{
		ips, err := loadGeoIP("US")
		common.Must(err)
		geoips = append(geoips, &router.GeoIP{
			CountryCode: "US",
			Cidr:        ips,
		})
	}

	matcher, err := router.NewMultiGeoIPMatcher(geoips, false)
	common.Must(err)

	ctx := withOutbound(&session.Outbound{Target: net.TCPDestination(net.ParseAddress("8.8.8.8"), 80)})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = matcher.Apply(ctx)
	}
}
