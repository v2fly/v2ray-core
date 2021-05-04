package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/proxy/dns"
)

type DNSOutboundConfig struct {
	Network cfgcommon.Network  `json:"network"`
	Address *cfgcommon.Address `json:"address"`
	Port    uint16             `json:"port"`
}

func (c *DNSOutboundConfig) Build() (proto.Message, error) {
	config := &dns.Config{
		Server: &net.Endpoint{
			Network: c.Network.Build(),
			Port:    uint32(c.Port),
		},
	}
	if c.Address != nil {
		config.Server.Address = c.Address.Build()
	}
	return config, nil
}
