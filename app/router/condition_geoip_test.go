package router_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem"
)

func init() {
	const geoipURL = "https://github.com/v2fly/geoip/releases/download/202507050144/geoip.dat"

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip-202507050144.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
	}
}

func TestGeoIPMatcherContainer(t *testing.T) {
	container := &router.GeoIPMatcherContainer{}

	m1, err := container.Add(&routercommon.GeoIP{
		CountryCode: "CN",
	})
	common.Must(err)

	m2, err := container.Add(&routercommon.GeoIP{
		CountryCode: "US",
	})
	common.Must(err)

	m3, err := container.Add(&routercommon.GeoIP{
		CountryCode: "CN",
	})
	common.Must(err)

	if m1 != m3 {
		t.Error("expect same matcher for same geoip, but not")
	}

	if m1 == m2 {
		t.Error("expect different matcher for different geoip, but actually same")
	}
}

func TestGeoIPMatcher(t *testing.T) {
	cidrList := []*routercommon.CIDR{
		{Ip: []byte{0, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{10, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{100, 64, 0, 0}, Prefix: 10},
		{Ip: []byte{127, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{169, 254, 0, 0}, Prefix: 16},
		{Ip: []byte{172, 16, 0, 0}, Prefix: 12},
		{Ip: []byte{192, 0, 0, 0}, Prefix: 24},
		{Ip: []byte{192, 0, 2, 0}, Prefix: 24},
		{Ip: []byte{192, 168, 0, 0}, Prefix: 16},
		{Ip: []byte{192, 18, 0, 0}, Prefix: 15},
		{Ip: []byte{198, 51, 100, 0}, Prefix: 24},
		{Ip: []byte{203, 0, 113, 0}, Prefix: 24},
		{Ip: []byte{8, 8, 8, 8}, Prefix: 32},
		{Ip: []byte{91, 108, 4, 0}, Prefix: 16},
	}

	matcher := &router.GeoIPMatcher{}
	common.Must(matcher.Init(cidrList))

	testCases := []struct {
		Input  string
		Output bool
	}{
		{
			Input:  "192.168.1.1",
			Output: true,
		},
		{
			Input:  "192.0.0.0",
			Output: true,
		},
		{
			Input:  "192.0.1.0",
			Output: false,
		},
		{
			Input:  "0.1.0.0",
			Output: true,
		},
		{
			Input:  "1.0.0.1",
			Output: false,
		},
		{
			Input:  "8.8.8.7",
			Output: false,
		},
		{
			Input:  "8.8.8.8",
			Output: true,
		},
		{
			Input:  "2001:cdba::3257:9652",
			Output: false,
		},
		{
			Input:  "91.108.255.254",
			Output: true,
		},
	}

	for _, testCase := range testCases {
		ip := net.ParseAddress(testCase.Input).IP()
		actual := matcher.Match(ip)
		if actual != testCase.Output {
			t.Error("expect input", testCase.Input, "to be", testCase.Output, ", but actually", actual)
		}
	}
}

func TestGeoIPReverseMatcher(t *testing.T) {
	cidrList := []*routercommon.CIDR{
		{Ip: []byte{8, 8, 8, 8}, Prefix: 32},
		{Ip: []byte{91, 108, 4, 0}, Prefix: 16},
	}
	matcher := &router.GeoIPMatcher{}
	matcher.SetReverseMatch(true) // Reverse match
	common.Must(matcher.Init(cidrList))

	testCases := []struct {
		Input  string
		Output bool
	}{
		{
			Input:  "8.8.8.8",
			Output: false,
		},
		{
			Input:  "2001:cdba::3257:9652",
			Output: true,
		},
		{
			Input:  "91.108.255.254",
			Output: false,
		},
	}

	for _, testCase := range testCases {
		ip := net.ParseAddress(testCase.Input).IP()
		actual := matcher.Match(ip)
		if actual != testCase.Output {
			t.Error("expect input", testCase.Input, "to be", testCase.Output, ", but actually", actual)
		}
	}
}

func TestGeoIPMatcher4CN(t *testing.T) {
	ips, err := loadGeoIP("CN")
	common.Must(err)

	matcher := &router.GeoIPMatcher{}
	common.Must(matcher.Init(ips))

	if matcher.Match([]byte{8, 8, 8, 8}) {
		t.Error("expect CN geoip doesn't contain 8.8.8.8, but actually does")
	}
}

func TestGeoIPMatcher6US(t *testing.T) {
	ips, err := loadGeoIP("US")
	common.Must(err)

	matcher := &router.GeoIPMatcher{}
	common.Must(matcher.Init(ips))

	if !matcher.Match(net.ParseAddress("2001:4860:4860::8888").IP()) {
		t.Error("expect US geoip contain 2001:4860:4860::8888, but actually not")
	}
}

func loadGeoIP(country string) ([]*routercommon.CIDR, error) {
	geoipBytes, err := filesystem.ReadAsset("geoip-202507050144.dat")
	if err != nil {
		return nil, err
	}
	var geoipList routercommon.GeoIPList
	if err := proto.Unmarshal(geoipBytes, &geoipList); err != nil {
		return nil, err
	}

	for _, geoip := range geoipList.Entry {
		if strings.EqualFold(geoip.CountryCode, country) {
			return geoip.Cidr, nil
		}
	}

	panic("country not found: " + country)
}

func BenchmarkGeoIPMatcher4CN(b *testing.B) {
	ips, err := loadGeoIP("CN")
	common.Must(err)

	matcher := &router.GeoIPMatcher{}
	common.Must(matcher.Init(ips))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = matcher.Match([]byte{8, 8, 8, 8})
	}
}

func BenchmarkGeoIPMatcher6US(b *testing.B) {
	ips, err := loadGeoIP("US")
	common.Must(err)

	matcher := &router.GeoIPMatcher{}
	common.Must(matcher.Init(ips))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = matcher.Match(net.ParseAddress("2001:4860:4860::8888").IP())
	}
}
