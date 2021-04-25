package geodata

import (
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common/platform"
)

type GeoIPCache map[string]*router.GeoIP

func (g GeoIPCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoIPCache) Get(key string) *router.GeoIP {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoIPCache) Set(key string, value *router.GeoIP) {
	if g == nil {
		g = make(map[string]*router.GeoIP)
	}
	g[key] = value
}

func (g GeoIPCache) Unmarshal(filename, code string) (*router.GeoIP, error) {
	filename = platform.GetAssetLocation(filename)
	idx := strings.ToUpper(filename + "|" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geoipBytes, err := Decode(filename, code)
	if err != nil {
		return nil, err
	}
	var geoip router.GeoIP
	if err := proto.Unmarshal(geoipBytes, &geoip); err != nil {
		return nil, err
	}

	g.Set(idx, &geoip)

	return &geoip, nil
}

type GeoSiteCache map[string]*router.GeoSite

func (g GeoSiteCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoSiteCache) Get(key string) *router.GeoSite {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoSiteCache) Set(key string, value *router.GeoSite) {
	if g == nil {
		g = make(map[string]*router.GeoSite)
	}
	g[key] = value
}

func (g GeoSiteCache) Unmarshal(filename, code string) (*router.GeoSite, error) {
	filename = platform.GetAssetLocation(filename)
	idx := strings.ToUpper(filename + "|" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geositeBytes, err := Decode(filename, code)
	if err != nil {
		return nil, err
	}
	var geosite router.GeoSite
	if err := proto.Unmarshal(geositeBytes, &geosite); err != nil {
		return nil, err
	}

	g.Set(idx, &geosite)

	return &geosite, nil
}
