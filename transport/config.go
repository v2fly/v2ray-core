package transport

import (
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

// Apply applies this Config.
func (c *Config) Apply() error {
	if c == nil {
		return nil
	}
	return internet.ApplyGlobalTransportSettings(c.TransportSettings)
}
