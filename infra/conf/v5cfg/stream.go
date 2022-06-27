package v5cfg

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func (s StreamConfig) BuildV5(ctx context.Context) (proto.Message, error) {
	config := &internet.StreamConfig{}

	if s.Transport == "" {
		s.Transport = "tcp"
	}
	if s.Security == "" {
		s.Security = "none"
	}

	if s.TransportSettings == nil {
		s.TransportSettings = []byte("{}")
	}
	transportConfigPack, err := loadHeterogeneousConfigFromRawJSON("transport", s.Transport, s.TransportSettings)
	if err != nil {
		return nil, newError("unable to load transport config").Base(err)
	}

	config.ProtocolName = s.Transport
	config.TransportSettings = append(config.TransportSettings, &internet.TransportConfig{
		ProtocolName: s.Transport,
		Settings:     serial.ToTypedMessage(transportConfigPack),
	})

	if s.Security != "none" {
		if s.SecuritySettings == nil {
			s.SecuritySettings = []byte("{}")
		}
		securityConfigPack, err := loadHeterogeneousConfigFromRawJSON("security", s.Security, s.SecuritySettings)
		if err != nil {
			return nil, newError("unable to load security config").Base(err)
		}
		securityConfigPackTypedMessage := serial.ToTypedMessage(securityConfigPack)
		config.SecurityType = serial.V2Type(securityConfigPackTypedMessage)
		config.SecuritySettings = append(config.SecuritySettings, securityConfigPackTypedMessage)
	}

	config.SocketSettings, err = s.SocketSettings.Build()
	if err != nil {
		return nil, newError("unable to build socket config").Base(err)
	}

	return config, nil
}
