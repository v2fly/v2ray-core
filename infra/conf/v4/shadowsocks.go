package v4

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/proxy/shadowsocks"
)

type ShadowsocksServerConfig struct {
	Cipher      string                 `json:"method"`
	Password    string                 `json:"password"`
	UDP         bool                   `json:"udp"`
	Level       byte                   `json:"level"`
	Email       string                 `json:"email"`
	NetworkList *cfgcommon.NetworkList `json:"network"`
	IVCheck     bool                   `json:"ivCheck"`
	Plugin      string                 `json:"plugin"`
	PluginOpts  string                 `json:"pluginOpts"`
	PluginArgs  *cfgcommon.StringList  `json:"pluginArgs"`
}

func (v *ShadowsocksServerConfig) Build() (proto.Message, error) {
	config := new(shadowsocks.ServerConfig)
	config.UdpEnabled = v.UDP
	config.Network = v.NetworkList.Build()

	if v.Password == "" {
		return nil, newError("Shadowsocks password is not specified.")
	}
	account := &shadowsocks.Account{
		Password: v.Password,
		IvCheck:  v.IVCheck,
	}
	account.CipherType = shadowsocks.CipherFromString(v.Cipher)
	if account.CipherType == shadowsocks.CipherType_UNKNOWN {
		return nil, newError("unknown cipher method: ", v.Cipher)
	}

	config.User = &protocol.User{
		Email:   v.Email,
		Level:   uint32(v.Level),
		Account: serial.ToTypedMessage(account),
	}

	config.Plugin = v.Plugin
	config.PluginOpts = v.PluginOpts
	if v.PluginArgs != nil && len(*v.PluginArgs) > 0 {
		config.PluginArgs = *v.PluginArgs
	}

	return config, nil
}

type ShadowsocksServerTarget struct {
	Address  *cfgcommon.Address `json:"address"`
	Port     uint16             `json:"port"`
	Cipher   string             `json:"method"`
	Password string             `json:"password"`
	Email    string             `json:"email"`
	Ota      bool               `json:"ota"`
	Level    byte               `json:"level"`
	IVCheck  bool               `json:"ivCheck"`
}

type ShadowsocksClientConfig struct {
	Servers    []*ShadowsocksServerTarget `json:"servers"`
	Plugin     string                     `json:"plugin"`
	PluginOpts string                     `json:"pluginOpts"`
	PluginArgs *cfgcommon.StringList      `json:"pluginArgs"`
}

func (v *ShadowsocksClientConfig) Build() (proto.Message, error) {
	config := new(shadowsocks.ClientConfig)

	if len(v.Servers) == 0 {
		return nil, newError("0 Shadowsocks server configured.")
	}

	serverSpecs := make([]*protocol.ServerEndpoint, len(v.Servers))
	for idx, server := range v.Servers {
		if server.Address == nil {
			return nil, newError("Shadowsocks server address is not set.")
		}
		if server.Port == 0 {
			return nil, newError("Invalid Shadowsocks port.")
		}
		if server.Password == "" {
			return nil, newError("Shadowsocks password is not specified.")
		}
		account := &shadowsocks.Account{
			Password: server.Password,
		}
		account.CipherType = shadowsocks.CipherFromString(server.Cipher)
		if account.CipherType == shadowsocks.CipherType_UNKNOWN {
			return nil, newError("unknown cipher method: ", server.Cipher)
		}

		account.IvCheck = server.IVCheck

		ss := &protocol.ServerEndpoint{
			Address: server.Address.Build(),
			Port:    uint32(server.Port),
			User: []*protocol.User{
				{
					Level:   uint32(server.Level),
					Email:   server.Email,
					Account: serial.ToTypedMessage(account),
				},
			},
		}

		serverSpecs[idx] = ss
	}

	config.Server = serverSpecs
	config.Plugin = v.Plugin
	config.PluginOpts = v.PluginOpts
	if v.PluginArgs != nil && len(*v.PluginArgs) > 0 {
		config.PluginArgs = *v.PluginArgs
	}

	return config, nil
}
