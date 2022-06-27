package sniffer

import (
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type SniffingConfig struct {
	Enabled      bool                  `json:"enabled"`
	DestOverride *cfgcommon.StringList `json:"destOverride"`
	MetadataOnly bool                  `json:"metadataOnly"`
}

// Build implements Buildable.
func (c *SniffingConfig) Build() (*proxyman.SniffingConfig, error) {
	var p []string
	if c.DestOverride != nil {
		for _, domainOverride := range *c.DestOverride {
			switch strings.ToLower(domainOverride) {
			case "http":
				p = append(p, "http")
			case "tls", "https", "ssl":
				p = append(p, "tls")
			case "quic":
				p = append(p, "quic")
			case "fakedns":
				p = append(p, "fakedns")
			case "fakedns+others":
				p = append(p, "fakedns+others")
			default:
				return nil, newError("unknown protocol: ", domainOverride)
			}
		}
	}

	return &proxyman.SniffingConfig{
		Enabled:             c.Enabled,
		DestinationOverride: p,
		MetadataOnly:        c.MetadataOnly,
	}, nil
}
