package v4

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/proxy/socks"
)

type SocksAccount struct {
	Username string `json:"user"`
	Password string `json:"pass"`
}

func (v *SocksAccount) Build() *socks.Account {
	return &socks.Account{
		Username: v.Username,
		Password: v.Password,
	}
}

const (
	AuthMethodNoAuth   = "noauth"
	AuthMethodUserPass = "password"
)

type SocksServerConfig struct {
	AuthMethod     string             `json:"auth"`
	Accounts       []*SocksAccount    `json:"accounts"`
	UDP            bool               `json:"udp"`
	Host           *cfgcommon.Address `json:"ip"`
	Timeout        uint32             `json:"timeout"`
	UserLevel      uint32             `json:"userLevel"`
	PacketEncoding string             `json:"packetEncoding"`
	DeferLastReply bool               `json:"deferLastReply,omitempty"`
}

func (v *SocksServerConfig) Build() (proto.Message, error) {
	config := new(socks.ServerConfig)
	switch v.AuthMethod {
	case AuthMethodNoAuth:
		config.AuthType = socks.AuthType_NO_AUTH
	case AuthMethodUserPass:
		config.AuthType = socks.AuthType_PASSWORD
	default:
		// newError("unknown socks auth method: ", v.AuthMethod, ". Default to noauth.").AtWarning().WriteToLog()
		config.AuthType = socks.AuthType_NO_AUTH
	}

	if len(v.Accounts) > 0 {
		config.Accounts = make(map[string]string, len(v.Accounts))
		for _, account := range v.Accounts {
			config.Accounts[account.Username] = account.Password
		}
	}

	config.UdpEnabled = v.UDP
	if v.Host != nil {
		config.Address = v.Host.Build()
	}

	config.Timeout = v.Timeout
	config.UserLevel = v.UserLevel

	switch v.PacketEncoding {
	case "Packet":
		config.PacketEncoding = packetaddr.PacketAddrType_Packet
	case "", "None":
		config.PacketEncoding = packetaddr.PacketAddrType_None
	}

	config.DeferLastReply = v.DeferLastReply

	return config, nil
}

type SocksRemoteConfig struct {
	Address *cfgcommon.Address `json:"address"`
	Port    uint16             `json:"port"`
	Users   []json.RawMessage  `json:"users"`
}

type SocksClientConfig struct {
	Servers []*SocksRemoteConfig `json:"servers"`
	Version string               `json:"version"`
}

func (v *SocksClientConfig) Build() (proto.Message, error) {
	config := new(socks.ClientConfig)
	config.Server = make([]*protocol.ServerEndpoint, len(v.Servers))
	switch strings.ToLower(v.Version) {
	case "4":
		config.Version = socks.Version_SOCKS4
	case "4a":
		config.Version = socks.Version_SOCKS4A
	case "", "5":
		config.Version = socks.Version_SOCKS5
	default:
		return nil, newError("failed to parse socks server version: ", v.Version).AtError()
	}
	for idx, serverConfig := range v.Servers {
		server := &protocol.ServerEndpoint{
			Address: serverConfig.Address.Build(),
			Port:    uint32(serverConfig.Port),
		}
		for _, rawUser := range serverConfig.Users {
			user := new(protocol.User)
			if err := json.Unmarshal(rawUser, user); err != nil {
				return nil, newError("failed to parse Socks user").Base(err).AtError()
			}
			account := new(SocksAccount)
			if err := json.Unmarshal(rawUser, account); err != nil {
				return nil, newError("failed to parse socks account").Base(err).AtError()
			}
			if config.Version != socks.Version_SOCKS5 && account.Password != "" {
				return nil, newError("password is only supported in socks5").AtError()
			}
			user.Account = serial.ToTypedMessage(account.Build())
			server.User = append(server.User, user)
		}
		config.Server[idx] = server
	}
	return config, nil
}
