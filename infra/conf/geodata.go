package conf

import (
	"runtime"
	"strings"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common/geodata"
)

func loadGeoIP(country string) ([]*router.CIDR, error) {
	return geodata.LoadIP("geoip.dat", country)
}

func loadGeosite(list string) ([]*router.Domain, error) {
	return loadGeositeWithAttr("geosite.dat", list)
}

func loadGeositeWithAttr(filename string, siteWithAttr string) ([]*router.Domain, error) {
	parts := strings.Split(siteWithAttr, "@")
	if len(parts) == 0 {
		return nil, newError("empty rule")
	}
	list := strings.TrimSpace(parts[0])
	attrVal := parts[1:]

	if len(list) == 0 {
		return nil, newError("empty listname in rule: ", siteWithAttr)
	}

	domains, err := geodata.LoadSite(filename, list)
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

	runtime.GC()
	return filteredDomains, nil
}
