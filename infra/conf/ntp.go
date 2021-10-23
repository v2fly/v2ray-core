package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/ntp"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	ntpproxy "github.com/v2fly/v2ray-core/v4/proxy/ntp"
)

type NTPConfig struct {
	Address      *cfgcommon.Address
	Port         uint16 `json:"port"`
	SyncInterval uint32 `json:"syncInterval"`
	InboundTag   string `json:"inboundTag"`
}

func (n *NTPConfig) Build() (proto.Message, error) {
	c := &ntp.Config{
		Address: &net.Endpoint{
			Network: net.Network_UDP,
			Address: n.Address.Build(),
			Port:    uint32(n.Port),
		},
		SyncInterval: n.SyncInterval,
		InboundTag:   n.InboundTag,
	}
	if c.SyncInterval == 0 {
		c.SyncInterval = 1440
	}
	if c.InboundTag == "" {
		c.InboundTag = "ntp"
	}
	return c, nil
}

type NTPOutboundConfig struct {
	UserLevel uint32 `json:"userLevel"`
}

func (n *NTPOutboundConfig) Build() (proto.Message, error) {
	return &ntpproxy.Config{
		UserLevel: n.UserLevel,
	}, nil
}
