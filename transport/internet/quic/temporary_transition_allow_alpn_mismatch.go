// Machine Generated Code
package quic

import (
	"crypto/tls"
	"fmt"
	"slices"
	"strconv"
	"strings"

	core "github.com/v2fly/v2ray-core/v5"
)

func extractXY(s string) (x, y string, err error) {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid format: %q", s)
	}
	return parts[0], parts[1], nil
}

func setupTLSConfigForALPNMismatch(tlsConf *tls.Config) *tls.Config {
	// This workaround is only present for next 6 mid-version release
	if major, middle, err := extractXY(core.Version()); err != nil {
		return tlsConf
	} else {
		if major != "5" {
			return tlsConf
		}
		if middleInt, err := strconv.ParseInt(middle, 10, 64); err != nil || middleInt > 60 {
			return tlsConf
		}
	}

	// Initialize the session ticket keys before cloning, so connections continue
	// to share them. See https://github.com/golang/go/issues/60506.
	_, _ = tlsConf.DecryptTicket(nil, tls.ConnectionState{})

	conf := tlsConf.Clone()
	getConfigForClient := conf.GetConfigForClient
	conf.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		selected := tlsConf
		if getConfigForClient != nil {
			c, err := getConfigForClient(info)
			if err != nil {
				return nil, err
			}
			if c != nil {
				selected = c
			}
		}
		if len(info.SupportedProtos) == 0 {
			return selected, nil
		}
		for _, proto := range selected.NextProtos {
			if slices.Contains(info.SupportedProtos, proto) {
				return selected, nil
			}
		}
		// Only change this connection's configuration. The selected configuration
		// may be shared with other connections or returned by an application callback.
		selected = selected.Clone()
		selected.NextProtos = []string{info.SupportedProtos[0]}
		return selected, nil
	}
	return conf
}
