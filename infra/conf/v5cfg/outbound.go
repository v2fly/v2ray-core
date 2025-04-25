package v5cfg

import (
	"context"

	"github.com/golang/protobuf/proto"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
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
	switch c.DomainStrategy {
	case "UseIP":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP
	case "UseIP4":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP4
	case "UseIP6":
		senderSettings.DomainStrategy = proxyman.SenderConfig_USE_IP6
	case "AsIs", "":
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
