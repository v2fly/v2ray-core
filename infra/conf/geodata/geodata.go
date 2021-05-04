package geodata

import (
	"strings"

	"github.com/v2fly/v2ray-core/v4/app/router"
)

type loader struct {
	LoaderImplementation
}

func (l *loader) LoadGeoSite(list string) ([]*router.Domain, error) {
	return l.LoadGeoSiteWithAttr("geosite.dat", list)
}

func (l *loader) LoadGeoSiteWithAttr(file string, siteWithAttr string) ([]*router.Domain, error) {
	parts := strings.Split(siteWithAttr, "@")
	if len(parts) == 0 {
		return nil, newError("empty rule")
	}
	list := strings.TrimSpace(parts[0])
	attrVal := parts[1:]

	if len(list) == 0 {
		return nil, newError("empty listname in rule: ", siteWithAttr)
	}

	domains, err := l.LoadSite(file, list)
	if err != nil {
		return nil, err
	}

	attrs := parseAttrs(attrVal)
	if attrs.IsEmpty() {
		if strings.Contains(siteWithAttr, "@") {
			newError("empty attribute list: ", siteWithAttr)
		}
		return domains, nil
	}

	filteredDomains := make([]*router.Domain, 0, len(domains))
	hasAttrMatched := false
	for _, domain := range domains {
		if attrs.Match(domain) {
			hasAttrMatched = true
			filteredDomains = append(filteredDomains, domain)
		}
	}
	if !hasAttrMatched {
		newError("attribute match no rule: geosite:", siteWithAttr)
	}

	return filteredDomains, nil
}

func (l *loader) LoadGeoIP(country string) ([]*router.CIDR, error) {
	return l.LoadIP("geoip.dat", country)
}

var loaders map[string]func() LoaderImplementation

func RegisterGeoDataLoaderImplementationCreator(name string, loader func() LoaderImplementation) {
	if loaders == nil {
		loaders = map[string]func() LoaderImplementation{}
	}
	loaders[name] = loader
}

func getGeoDataLoaderImplementation(name string) (LoaderImplementation, error) {
	if geoLoader, ok := loaders[name]; ok {
		return geoLoader(), nil
	}
	return nil, newError("unable to locate GeoData loader ", name)
}

func GetGeoDataLoader(name string) (Loader, error) {
	loadImpl, err := getGeoDataLoaderImplementation(name)
	if err == nil {
		return &loader{loadImpl}, nil
	}
	return nil, err
}
