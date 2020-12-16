package link

import (
	"errors"
	"fmt"
	"strings"

	"v2ray.com/core/infra/conf"
)

// Link is the interface for v2ray links, like VMessLink
type Link interface {
	//  returns the tag of the link
	Tag() string
	// Detail returns human readable string of VmessLink
	Detail() string
	// ToOutbound converts the vmess link to *OutboundDetourConfig
	ToOutbound() *conf.OutboundDetourConfig
	// ToString unmarshals Link to string
	ToString() string
}

// Parse parse link string to Link
func Parse(arg string) (Link, error) {
	ps, err := getParsers(arg)
	if err != nil {
		return nil, err
	}
	errs := new(strings.Builder)
	errs.WriteString("collected errors:")
	for _, p := range ps {
		lk, err := p.Parse(arg)
		if err == nil {
			return lk, nil
		}
		errs.WriteString(fmt.Sprintf("\n  not a valid %s link: %s", p.Name, err))
	}
	return nil, errors.New(errs.String())
}
