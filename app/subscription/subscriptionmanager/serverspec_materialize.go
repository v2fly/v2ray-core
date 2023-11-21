package subscriptionmanager

import (
	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func (s *SubscriptionManagerImpl) materialize(subscriptionName, tagName string, serverSpec *specs.SubscriptionServerConfig) (*core.OutboundHandlerConfig, error) {
	outboundConf, err := s.getOutboundTemplateForSubscriptionName(subscriptionName)
	if err != nil {
		return nil, newError("failed to get outbound template for subscription name: ", err)
	}

	senderSettingsIfcd, err := serial.GetInstanceOf(outboundConf.SenderSettings)
	if err != nil {
		return nil, newError("failed to get sender settings: ", err)
	}
	senderSettings := senderSettingsIfcd.(*proxyman.SenderConfig)

	if serverSpec.Configuration.Transport != "" {
		senderSettings.StreamSettings.ProtocolName = serverSpec.Configuration.Transport
		senderSettings.StreamSettings.TransportSettings = append(senderSettings.StreamSettings.TransportSettings,
			&internet.TransportConfig{ProtocolName: serverSpec.Configuration.Transport, Settings: serverSpec.Configuration.TransportSettings})
	}

	if serverSpec.Configuration.Security != "" {
		senderSettings.StreamSettings.SecurityType = serverSpec.Configuration.Security
		senderSettings.StreamSettings.SecuritySettings = append(senderSettings.StreamSettings.SecuritySettings,
			serverSpec.Configuration.SecuritySettings)
	}

	outboundConf.SenderSettings = serial.ToTypedMessage(senderSettings)

	outboundConf.ProxySettings = serverSpec.Configuration.ProtocolSettings

	outboundConf.Tag = tagName

	return outboundConf, nil
}

func (s *SubscriptionManagerImpl) getOutboundTemplateForSubscriptionName(subscriptionName string) (*core.OutboundHandlerConfig, error) { //nolint: unparam
	senderSetting := &proxyman.SenderConfig{
		DomainStrategy: proxyman.SenderConfig_AS_IS, StreamSettings: &internet.StreamConfig{},
	}

	return &core.OutboundHandlerConfig{SenderSettings: serial.ToTypedMessage(senderSetting)}, nil
}
