//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	stdnet "net"
	"strings"
)

type bindFileFlags []bindFile

func (f *bindFileFlags) String() string {
	if f == nil {
		return ""
	}
	parts := make([]string, 0, len(*f))
	for _, item := range *f {
		parts = append(parts, item.Source+":"+item.Target)
	}
	return strings.Join(parts, ",")
}

func (f *bindFileFlags) Set(raw string) error {
	source, target, ok := strings.Cut(raw, ":")
	if !ok || source == "" || target == "" {
		return fmt.Errorf("bind-file must be source:target")
	}
	*f = append(*f, bindFile{Source: source, Target: target})
	return nil
}

func parseTunProtocolConfig(name string, host string, guest string, prefix int, ipv6 bool) (tunProtocolConfig, error) {
	hostIP, err := parseIPFlag(name+"-host", host, ipv6)
	if err != nil {
		return tunProtocolConfig{}, err
	}
	guestIP, err := parseIPFlag(name+"-guest", guest, ipv6)
	if err != nil {
		return tunProtocolConfig{}, err
	}
	maxPrefix := 32
	if ipv6 {
		maxPrefix = 128
	}
	if prefix < 0 || prefix > maxPrefix {
		return tunProtocolConfig{}, fmt.Errorf("-%s-prefix must be between 0 and %d, got %d", name, maxPrefix, prefix)
	}
	return tunProtocolConfig{
		Host:   hostIP,
		Guest:  guestIP,
		Prefix: prefix,
	}, nil
}

func parseIPFlag(name string, raw string, ipv6 bool) (string, error) {
	ip := stdnet.ParseIP(raw)
	if ip == nil {
		return "", fmt.Errorf("-%s must be a valid IP address, got %q", name, raw)
	}
	if ipv6 {
		if ip.To4() != nil || ip.To16() == nil {
			return "", fmt.Errorf("-%s must be an IPv6 address, got %q", name, raw)
		}
		return ip.String(), nil
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return "", fmt.Errorf("-%s must be an IPv4 address, got %q", name, raw)
	}
	return ip4.String(), nil
}
