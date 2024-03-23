package v4

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/proxy/hysteria2"
)

// Hysteria2ServerTarget is configuration of a single hysteria2 server
type Hysteria2ServerTarget struct {
	Address *cfgcommon.Address `json:"address"`
	Port    uint16             `json:"port"`
	Email   string             `json:"email"`
	Level   byte               `json:"level"`
}

// Hysteria2ClientConfig is configuration of hysteria2 servers
type Hysteria2ClientConfig struct {
	Servers []*Hysteria2ServerTarget `json:"servers"`
}

// Build implements Buildable
func (c *Hysteria2ClientConfig) Build() (proto.Message, error) {
	config := new(hysteria2.ClientConfig)

	if len(c.Servers) == 0 {
		return nil, newError("0 Hysteria2 server configured.")
	}

	serverSpecs := make([]*protocol.ServerEndpoint, len(c.Servers))
	for idx, rec := range c.Servers {
		if rec.Address == nil {
			return nil, newError("Hysteria2 server address is not set.")
		}
		if rec.Port == 0 {
			return nil, newError("Invalid Hysteria2 port.")
		}
		account := &hysteria2.Account{}
		hysteria2 := &protocol.ServerEndpoint{
			Address: rec.Address.Build(),
			Port:    uint32(rec.Port),
			User: []*protocol.User{
				{
					Level:   uint32(rec.Level),
					Email:   rec.Email,
					Account: serial.ToTypedMessage(account),
				},
			},
		}

		serverSpecs[idx] = hysteria2
	}

	config.Server = serverSpecs

	return config, nil
}

// Hysteria2UserConfig is user configuration
type Hysteria2UserConfig struct {
	Level byte   `json:"level"`
	Email string `json:"email"`
}

// Hysteria2ServerConfig is Inbound configuration
type Hysteria2ServerConfig struct {
	Clients []*Hysteria2UserConfig `json:"clients"`
}

// Build implements Buildable
func (c *Hysteria2ServerConfig) Build() (proto.Message, error) {
	config := new(hysteria2.ServerConfig)
	config.Users = make([]*protocol.User, len(c.Clients))
	for idx, rawUser := range c.Clients {
		user := new(protocol.User)
		account := &hysteria2.Account{}

		user.Email = rawUser.Email
		user.Level = uint32(rawUser.Level)
		user.Account = serial.ToTypedMessage(account)
		config.Users[idx] = user
	}
	return config, nil
}
