package v4

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/proxy/shadowsocksr"
)

type ShadowsocksRServerConfig struct {
	Cipher         string                 `json:"method"`
	Password       string                 `json:"password"`
	Protocol       string                 `json:"protocol"`
	ProtocolParam  string                 `json:"protocol_param"`
	Obfs          string                 `json:"obfs"`
	ObfsParam     string                 `json:"obfs_param"`
	UDP           bool                   `json:"udp"`
	Level         byte                   `json:"level"`
	Email         string                 `json:"email"`
	NetworkList   *cfgcommon.NetworkList `json:"network"`
	PacketEncoding string                `json:"packetEncoding"`
}

func (v *ShadowsocksRServerConfig) Build() (proto.Message, error) {
	config := new(shadowsocksr.ServerConfig)
	config.UdpEnabled = v.UDP
	config.Network = v.NetworkList.Build()

	if v.Password == "" {
		return nil, newError("ShadowsocksR password is not specified.")
	}
	account := &shadowsocksr.Account{
		Password:      v.Password,
		Protocol:     v.Protocol,
		ProtocolParam: v.ProtocolParam,
		Obfs:         v.Obfs,
		ObfsParam:    v.ObfsParam,
	}
	account.CipherType = shadowsocksr.CipherTypeFromString(v.Cipher)
	if account.CipherType == shadowsocksr.CipherType_UNKNOWN {
		return nil, newError("unknown cipher method: ", v.Cipher)
	}

	config.User = &protocol.User{
		Email:   v.Email,
		Level:   uint32(v.Level),
		Account: serial.ToTypedMessage(account),
	}

	switch v.PacketEncoding {
	case "Packet":
		config.PacketEncoding = packetaddr.PacketAddrType_Packet
	case "", "None":
		config.PacketEncoding = packetaddr.PacketAddrType_None
	}

	return config, nil
}

type ShadowsocksRServerTarget struct {
	Address       *cfgcommon.Address `json:"address"`
	Port         uint16             `json:"port"`
	Cipher       string             `json:"method"`
	Password     string             `json:"password"`
	Protocol     string             `json:"protocol"`
	ProtocolParam string             `json:"protocol_param"`
	Obfs         string             `json:"obfs"`
	ObfsParam    string             `json:"obfs_param"`
	Email        string             `json:"email"`
	Level        byte               `json:"level"`
}

type ShadowsocksRClientConfig struct {
	Servers []*ShadowsocksRServerTarget `json:"servers"`
}

func (v *ShadowsocksRClientConfig) Build() (proto.Message, error) {
	config := new(shadowsocksr.ClientConfig)

	if len(v.Servers) == 0 {
		return nil, newError("0 ShadowsocksR server configured.")
	}

	serverSpecs := make([]*protocol.ServerEndpoint, len(v.Servers))
	for idx, server := range v.Servers {
		if server.Address == nil {
			return nil, newError("ShadowsocksR server address is not set.")
		}
		if server.Port == 0 {
			return nil, newError("Invalid ShadowsocksR port.")
		}
		if server.Password == "" {
			return nil, newError("ShadowsocksR password is not specified.")
		}
		account := &shadowsocksr.Account{
			Password:      server.Password,
			Protocol:     server.Protocol,
			ProtocolParam: server.ProtocolParam,
			Obfs:         server.Obfs,
			ObfsParam:    server.ObfsParam,
		}
		account.CipherType = shadowsocksr.CipherTypeFromString(server.Cipher)
		if account.CipherType == shadowsocksr.CipherType_UNKNOWN {
			return nil, newError("unknown cipher method: ", server.Cipher)
		}

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

	return config, nil
}
