package v4

import (
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/proxy/wireguard"
)

type WireGuardClientConfig struct {
	Address        *cfgcommon.Address   `json:"address"`
	Port           uint16               `json:"port"`
	Network        cfgcommon.Network    `json:"network"`
	LocalAddresses cfgcommon.StringList `json:"localAddresses"`
	PrivateKey     string               `json:"privateKey"`
	PeerPublicKey  string               `json:"peerPublicKey"`
	PreSharedKey   string               `json:"preSharedKey"`
	MTU            uint32               `json:"mtu"`
	UserLevel      uint32               `json:"userLevel"`
}

func (v *WireGuardClientConfig) Build() (proto.Message, error) {
	config := &wireguard.Config{
		Server: &protocol.ServerEndpoint{
			Address: v.Address.Build(),
			Port:    uint32(v.Port),
		},
		Network:       v.Network.Build(),
		LocalAddress:  v.LocalAddresses,
		PrivateKey:    v.PrivateKey,
		PeerPublicKey: v.PeerPublicKey,
		PreSharedKey:  v.PreSharedKey,
		Mtu:           v.MTU,
		UserLevel:     v.UserLevel,
	}
	return config, nil
}
