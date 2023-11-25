package specs

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/registry"
	"github.com/v2fly/v2ray-core/v5/common/serial"
)

func NewOutboundParser() *OutboundParser {
	return &OutboundParser{}
}

type OutboundParser struct{}

func (p *OutboundParser) ParseOutboundConfig(rawConfig []byte) (*OutboundConfig, error) {
	skeleton := &OutboundConfig{}
	decoder := json.NewDecoder(bytes.NewReader(rawConfig))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(skeleton)
	if err != nil {
		return nil, newError("failed to parse outbound config skeleton").Base(err)
	}
	return skeleton, nil
}

func (p *OutboundParser) toAbstractServerSpec(config *OutboundConfig) (*ServerConfiguration, error) {
	serverConfig := &ServerConfiguration{}
	serverConfig.Protocol = config.Protocol
	{
		protocolSettings, err := loadHeterogeneousConfigFromRawJSONRestricted("outbound", config.Protocol, config.Settings)
		if err != nil {
			return nil, newError("failed to parse protocol settings").Base(err)
		}
		serverConfig.ProtocolSettings = serial.ToTypedMessage(protocolSettings)
	}

	if config.StreamSetting != nil {
		if config.StreamSetting.Transport == "" {
			config.StreamSetting.Transport = "tcp"
		}
		if config.StreamSetting.Security == "" {
			config.StreamSetting.Security = "none"
		}
		{
			serverConfig.Transport = config.StreamSetting.Transport
			transportSettings, err := loadHeterogeneousConfigFromRawJSONRestricted(
				"transport", config.StreamSetting.Transport, config.StreamSetting.TransportSettings)
			if err != nil {
				return nil, newError("failed to parse transport settings").Base(err)
			}
			serverConfig.TransportSettings = serial.ToTypedMessage(transportSettings)
		}
		if config.StreamSetting.Security != "none" {
			securitySettings, err := loadHeterogeneousConfigFromRawJSONRestricted(
				"security", config.StreamSetting.Security, config.StreamSetting.SecuritySettings)
			if err != nil {
				return nil, newError("failed to parse security settings").Base(err)
			}

			serverConfig.SecuritySettings = serial.ToTypedMessage(securitySettings)
			serverConfig.Security = serial.V2Type(serverConfig.SecuritySettings)
		}
	}
	return serverConfig, nil
}

func (p *OutboundParser) ToSubscriptionServerConfig(config *OutboundConfig) (*SubscriptionServerConfig, error) {
	serverSpec, err := p.toAbstractServerSpec(config)
	if err != nil {
		return nil, newError("unable to parse server specification").Base(err)
	}
	return &SubscriptionServerConfig{
		Configuration: serverSpec,
		Metadata:      config.Metadata,
	}, nil
}

func loadHeterogeneousConfigFromRawJSONRestricted(interfaceType, name string, rawJSON json.RawMessage) (proto.Message, error) {
	ctx := context.TODO()
	ctx = registry.CreateRestrictedModeContext(ctx)
	if len(rawJSON) == 0 {
		rawJSON = []byte("{}")
	}
	return registry.LoadImplementationByAlias(ctx, interfaceType, name, []byte(rawJSON))
}
