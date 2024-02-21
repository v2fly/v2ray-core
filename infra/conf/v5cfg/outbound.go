package v5cfg

import (
	"context"
	"strings"

	"github.com/golang/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func (c OutboundConfig) BuildV5(ctx context.Context) (proto.Message, error) {
	senderSettings := &proxyman.SenderConfig{}

	if c.SendThrough != nil {
		address := c.SendThrough
		if address.Family().IsDomain() {
			return nil, newError("unable to send through: " + address.String())
		}
		senderSettings.Via = address.Build()
	}

	if c.StreamSetting != nil {
		ss, err := c.StreamSetting.BuildV5(ctx)
		if err != nil {
			return nil, err
		}
		senderSettings.StreamSettings = ss.(*internet.StreamConfig)
	}

	if c.ProxySettings != nil {
		ps, err := c.ProxySettings.Build()
		if err != nil {
			return nil, newError("invalid outbound detour proxy settings.").Base(err)
		}
		senderSettings.ProxySettings = ps
	}

	if c.MuxSettings != nil {
		senderSettings.MultiplexSettings = c.MuxSettings.Build()
	}

	senderSettings.DomainStrategy = proxyman.SenderConfig_AS_IS
	switch strings.ToLower(c.DomainStrategy) {
	case "useip", "use_ip", "use-ip":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP
	case "useip4", "useipv4", "use_ip4", "use_ipv4", "use_ip_v4", "use-ip4", "use-ipv4", "use-ip-v4":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP4
	case "useip6", "useipv6", "use_ip6", "use_ipv6", "use_ip_v6", "use-ip6", "use-ipv6", "use-ip-v6":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP6
	case "asis", "":
	default:
		return nil, newError("unknown domain strategy: ", c.DomainStrategy)
	}

	if c.Settings == nil {
		c.Settings = []byte("{}")
	}

	outboundConfigPack, err := loadHeterogeneousConfigFromRawJSON("outbound", c.Protocol, c.Settings)
	if err != nil {
		return nil, newError("unable to load outbound protocol config").Base(err)
	}

	return &core.OutboundHandlerConfig{
		SenderSettings: serial.ToTypedMessage(senderSettings),
		Tag:            c.Tag,
		ProxySettings:  serial.ToTypedMessage(outboundConfigPack),
	}, nil
}
