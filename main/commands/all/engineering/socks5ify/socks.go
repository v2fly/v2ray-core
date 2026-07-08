//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	stdnet "net"
	"net/url"
	"strconv"
	"strings"
)

func parseSocksServer(raw string, userOverride string, passOverride string) (socksServer, error) {
	if !strings.Contains(raw, "://") {
		raw = "socks5://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return socksServer{}, err
	}
	if parsed.Scheme != "socks5" && parsed.Scheme != "socks" {
		return socksServer{}, fmt.Errorf("unsupported SOCKS scheme %q", parsed.Scheme)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return socksServer{}, fmt.Errorf("SOCKS URL must not include a path")
	}
	host, portText, err := stdnet.SplitHostPort(parsed.Host)
	if err != nil {
		return socksServer{}, fmt.Errorf("invalid SOCKS address %q: %w", parsed.Host, err)
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return socksServer{}, fmt.Errorf("invalid SOCKS port %q", portText)
	}

	username := ""
	password := ""
	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}
	if userOverride != "" {
		username = userOverride
	}
	if passOverride != "" {
		password = passOverride
	}

	return socksServer{
		Host:     host,
		Port:     uint32(port),
		Username: username,
		Password: password,
	}, nil
}
