package memconservative

import (
	"os"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/app/router/routercommon"
	"github.com/ghxhy/v2ray-core/v5/common/platform"
)

type GeoIPCache map[string]*routercommon.GeoIP

func (g GeoIPCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoIPCache) Get(key string) *routercommon.GeoIP {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoIPCache) Set(key string, value *routercommon.GeoIP) {
	if g == nil {
		g = make(map[string]*routercommon.GeoIP)
	}
	g[key] = value
}

func (g GeoIPCache) Unmarshal(filename, code string) (*routercommon.GeoIP, error) {
	asset := platform.GetAssetLocation(filename)
	idx := strings.ToLower(asset + ":" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geoipBytes, err := Decode(asset, code)
	switch err {
	case nil:
		var geoip routercommon.GeoIP
		if err := proto.Unmarshal(geoipBytes, &geoip); err != nil {
			return nil, err
		}
		g.Set(idx, &geoip)
		return &geoip, nil

	case errCodeNotFound:
		return nil, newError("country code ", code, " not found in ", filename)

	case errFailedToReadBytes, errFailedToReadExpectedLenBytes,
		errInvalidGeodataFile, errInvalidGeodataVarintLength:
		newError("failed to decode geoip file: ", filename, ", fallback to the original ReadFile method")
		geoipBytes, err = os.ReadFile(asset)
		if err != nil {
			return nil, err
		}
		var geoipList routercommon.GeoIPList
		if err := proto.Unmarshal(geoipBytes, &geoipList); err != nil {
			return nil, err
		}
		for _, geoip := range geoipList.GetEntry() {
			if strings.EqualFold(code, geoip.GetCountryCode()) {
				g.Set(idx, geoip)
				return geoip, nil
			}
		}

	default:
		return nil, err
	}

	return nil, newError("country code ", code, " not found in ", filename)
}

type GeoSiteCache map[string]*routercommon.GeoSite

func (g GeoSiteCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoSiteCache) Get(key string) *routercommon.GeoSite {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoSiteCache) Set(key string, value *routercommon.GeoSite) {
	if g == nil {
		g = make(map[string]*routercommon.GeoSite)
	}
	g[key] = value
}

func (g GeoSiteCache) Unmarshal(filename, code string) (*routercommon.GeoSite, error) {
	asset := platform.GetAssetLocation(filename)
	idx := strings.ToLower(asset + ":" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geositeBytes, err := Decode(asset, code)
	switch err {
	case nil:
		var geosite routercommon.GeoSite
		if err := proto.Unmarshal(geositeBytes, &geosite); err != nil {
			return nil, err
		}
		g.Set(idx, &geosite)
		return &geosite, nil

	case errCodeNotFound:
		return nil, newError("list ", code, " not found in ", filename)

	case errFailedToReadBytes, errFailedToReadExpectedLenBytes,
		errInvalidGeodataFile, errInvalidGeodataVarintLength:
		newError("failed to decode geoip file: ", filename, ", fallback to the original ReadFile method")
		geositeBytes, err = os.ReadFile(asset)
		if err != nil {
			return nil, err
		}
		var geositeList routercommon.GeoSiteList
		if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
			return nil, err
		}
		for _, geosite := range geositeList.GetEntry() {
			if strings.EqualFold(code, geosite.GetCountryCode()) {
				g.Set(idx, geosite)
				return geosite, nil
			}
		}

	default:
		return nil, err
	}

	return nil, newError("list ", code, " not found in ", filename)
}
