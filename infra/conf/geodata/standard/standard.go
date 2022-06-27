package standard

import (
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func loadIP(filename, country string) ([]*routercommon.CIDR, error) {
	geoipBytes, err := filesystem.ReadAsset(filename)
	if err != nil {
		return nil, newError("failed to open file: ", filename).Base(err)
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

	return nil, newError("country not found in ", filename, ": ", country)
}

func loadSite(filename, list string) ([]*routercommon.Domain, error) {
	geositeBytes, err := filesystem.ReadAsset(filename)
	if err != nil {
		return nil, newError("failed to open file: ", filename).Base(err)
	}
	var geositeList routercommon.GeoSiteList
	if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
		return nil, err
	}

	for _, site := range geositeList.Entry {
		if strings.EqualFold(site.CountryCode, list) {
			return site.Domain, nil
		}
	}

	return nil, newError("list not found in ", filename, ": ", list)
}

type standardLoader struct{}

func (d standardLoader) LoadSite(filename, list string) ([]*routercommon.Domain, error) {
	return loadSite(filename, list)
}

func (d standardLoader) LoadIP(filename, country string) ([]*routercommon.CIDR, error) {
	return loadIP(filename, country)
}

func init() {
	geodata.RegisterGeoDataLoaderImplementationCreator("standard", func() geodata.LoaderImplementation {
		return standardLoader{}
	})
}
