package geodata

import (
	"runtime"

	"github.com/v2fly/v2ray-core/v4/app/router"
)

var geoipcache GeoIPCache = make(map[string]*router.GeoIP)
var geositecache GeoSiteCache = make(map[string]*router.GeoSite)

func LoadIP(filename, country string) ([]*router.CIDR, error) {
	geoip, err := geoipcache.Unmarshal(filename, country)
	if err != nil {
		return nil, newError("failed to decode geodata file: ", filename).Base(err)
	}
	runtime.GC()
	return geoip.Cidr, nil
}

func LoadSite(filename, list string) ([]*router.Domain, error) {
	geosite, err := geositecache.Unmarshal(filename, list)
	if err != nil {
		return nil, newError("failed to decode geodata file: ", filename).Base(err)
	}
	runtime.GC()
	return geosite.Domain, nil
}
