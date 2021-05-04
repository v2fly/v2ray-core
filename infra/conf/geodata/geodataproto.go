package geodata

import "github.com/v2fly/v2ray-core/v4/app/router"

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type LoaderImplementation interface {
	LoadSite(filename, list string) ([]*router.Domain, error)
	LoadIP(filename, country string) ([]*router.CIDR, error)
}

type Loader interface {
	LoaderImplementation
	LoadGeoSite(list string) ([]*router.Domain, error)
	LoadGeoSiteWithAttr(file string, siteWithAttr string) ([]*router.Domain, error)
	LoadGeoIP(country string) ([]*router.CIDR, error)
}
