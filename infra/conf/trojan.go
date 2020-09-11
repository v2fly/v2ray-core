package conf

import (
	"strconv"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/trojan"
)

type TrojanServerTarget struct {
	Address  *Address `json:"address"`
	Port     uint16   `json:"port"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Level    byte     `json:"level"`
}

type TrojanClientConfig struct {
	Servers []*TrojanServerTarget `json:"servers"`
}

// Build implements Buildable
func (c *TrojanClientConfig) Build() (proto.Message, error) {

	config := new(trojan.ClientConfig)

	if len(c.Servers) == 0 {
		return nil, newError("0 Trojan server configured.")
	}

	serverSpecs := make([]*protocol.ServerEndpoint, len(c.Servers))
	for idx, rec := range c.Servers {
		if rec.Address == nil {
			return nil, newError("Trojan server address is not set.")
		}
		if rec.Port == 0 {
			return nil, newError("Invalid Trojan port.")
		}
		if rec.Password == "" {
			return nil, newError("Trojan password is not specified.")
		}
		account := &trojan.Account{
			Password: rec.Password,
		}
		trojan := &protocol.ServerEndpoint{
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

		serverSpecs[idx] = trojan
	}

	config.Server = serverSpecs

	return config, nil
}

type TrojanInboundFallback struct {
	Type string `json:"type"`
	Dest string `json:"dest"`
}

type TrojanUserConfig struct {
	Password string `json:"password"`
	Level    byte   `json:"level"`
	Email    string `json:"email"`
}

type TrojanServerConfig struct {
	Users    []*TrojanUserConfig    `json:"users"`
	Fallback *TrojanInboundFallback `json:"fallback"`
}

// Build implements Buildable
func (c *TrojanServerConfig) Build() (proto.Message, error) {

	config := new(trojan.ServerConfig)

	if len(c.Users) == 0 {
		return nil, newError("No trojan user settings.")
	}

	config.Users = make([]*protocol.User, len(c.Users))
	for idx, rawUser := range c.Users {
		user := new(protocol.User)
		account := &trojan.Account{
			Password: rawUser.Password,
		}

		user.Email = rawUser.Email
		user.Level = uint32(rawUser.Level)
		user.Account = serial.ToTypedMessage(account)
		config.Users[idx] = user
	}

	if c.Fallback != nil {
		fb := &trojan.Fallback{
			Dest: c.Fallback.Dest,
		}

		if fb.Type == "" && fb.Dest != "" {
			switch fb.Dest[0] {
			case '@', '/':
				fb.Type = "unix"
			default:
				if _, err := strconv.Atoi(fb.Dest); err == nil {
					fb.Dest = "127.0.0.1:" + fb.Dest
				}
				if _, _, err := net.SplitHostPort(fb.Dest); err == nil {
					fb.Type = "tcp"
				}
			}
		}
		if fb.Type == "" {
			return nil, newError("please fill in a valid value for trojan fallback type")
		}

		config.Fallback = fb
	}

	return config, nil
}