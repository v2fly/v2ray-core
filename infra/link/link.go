package link

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"v2ray.com/core/infra/conf"
)

// Link is the interface for v2ray links, like VMessLink
type Link interface {
	ToOutbound() *conf.OutboundDetourConfig
	Tag() string
	Detail() string
}

// Parse parse link string to Link
func Parse(arg string) (Link, error) {
	u, err := url.Parse(arg)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, errors.New("invalid link")
	}
	switch {
	case strings.EqualFold(u.Scheme, "vmess"):
		lk, err := NewVmessLink(arg)
		if err != nil {
			return nil, err
		}
		return lk, nil
	default:
		return nil, fmt.Errorf("unsupported link scheme: %s", u.Scheme)
	}
}
