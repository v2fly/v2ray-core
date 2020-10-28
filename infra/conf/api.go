package conf

import (
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/commander"
	loggerservice "github.com/v2fly/v2ray-core/v5/app/log/command"
	handlerservice "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	statsservice "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"github.com/v2fly/v2ray-core/v5/common/serial"
)

type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

func (c *APIConfig) Build() (*commander.Config, error) {
	if c.Tag == "" {
		return nil, newError("API tag can't be empty.")
	}

	services := make([]*serial.TypedMessage, 0, 16)
	for _, s := range c.Services {
		switch strings.ToLower(s) {
		case "handlerservice":
			services = append(services, serial.ToTypedMessage(&handlerservice.Config{}))
		case "loggerservice":
			services = append(services, serial.ToTypedMessage(&loggerservice.Config{}))
		case "statsservice":
			services = append(services, serial.ToTypedMessage(&statsservice.Config{}))
		}
	}

	return &commander.Config{
		Tag:     c.Tag,
		Service: services,
	}, nil
}
