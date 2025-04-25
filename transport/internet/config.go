package internet

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/protoext"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
	"github.com/ghxhy/v2ray-core/v5/features"
)

type ConfigCreator func() interface{}

var (
	globalTransportConfigCreatorCache = make(map[string]ConfigCreator)
	globalTransportSettings           []*TransportConfig
)

const unknownProtocol = "unknown"

func transportProtocolToString(protocol TransportProtocol) string {
	switch protocol {
	case TransportProtocol_TCP:
		return "tcp"
	case TransportProtocol_UDP:
		return "udp"
	case TransportProtocol_HTTP:
		return "http"
	case TransportProtocol_MKCP:
		return "mkcp"
	case TransportProtocol_WebSocket:
		return "websocket"
	case TransportProtocol_DomainSocket:
		return "domainsocket"
	default:
		return unknownProtocol
	}
}

func RegisterProtocolConfigCreator(name string, creator ConfigCreator) error {
	if _, found := globalTransportConfigCreatorCache[name]; found {
		return newError("protocol ", name, " is already registered").AtError()
	}
	globalTransportConfigCreatorCache[name] = creator

	common.RegisterConfig(creator(), func(ctx context.Context, config interface{}) (interface{}, error) {
		return nil, newError("transport config should use CreateTransportConfig instead")
	})
	return nil
}

func CreateTransportConfig(name string) (interface{}, error) {
	creator, ok := globalTransportConfigCreatorCache[name]
	if !ok {
		return nil, newError("unknown transport protocol: ", name)
	}
	return creator(), nil
}

func (c *TransportConfig) GetTypedSettings() (interface{}, error) {
	return serial.GetInstanceOf(c.Settings)
}

func (c *TransportConfig) GetUnifiedProtocolName() string {
	if len(c.ProtocolName) > 0 {
		return c.ProtocolName
	}

	return transportProtocolToString(c.Protocol)
}

func (c *StreamConfig) GetEffectiveProtocol() string {
	if c == nil {
		return "tcp"
	}

	if len(c.ProtocolName) > 0 {
		return c.ProtocolName
	}

	return transportProtocolToString(c.Protocol)
}

func (c *StreamConfig) GetEffectiveTransportSettings() (interface{}, error) {
	protocol := c.GetEffectiveProtocol()
	return c.GetTransportSettingsFor(protocol)
}

func (c *StreamConfig) GetTransportSettingsFor(protocol string) (interface{}, error) {
	if c != nil {
		for _, settings := range c.TransportSettings {
			if settings.GetUnifiedProtocolName() == protocol {
				return settings.GetTypedSettings()
			}
		}
	}

	for _, settings := range globalTransportSettings {
		if settings.GetUnifiedProtocolName() == protocol {
			return settings.GetTypedSettings()
		}
	}

	return CreateTransportConfig(protocol)
}

func (c *StreamConfig) GetEffectiveSecuritySettings() (interface{}, error) {
	for _, settings := range c.SecuritySettings {
		if serial.V2Type(settings) == c.SecurityType {
			return serial.GetInstanceOf(settings)
		}
	}
	return serial.GetInstance(c.SecurityType)
}

func (c *StreamConfig) HasSecuritySettings() bool {
	return len(c.SecurityType) > 0
}

func ApplyGlobalTransportSettings(settings []*TransportConfig) error {
	features.PrintDeprecatedFeatureWarning("global transport settings")
	globalTransportSettings = settings
	return nil
}

func (c *ProxyConfig) HasTag() bool {
	return c != nil && len(c.Tag) > 0
}

func (m SocketConfig_TProxyMode) IsEnabled() bool {
	return m != SocketConfig_Off
}

func getOriginalMessageName(streamSettings *MemoryStreamConfig) string {
	msgOpts, err := protoext.GetMessageOptions(proto.MessageV2(streamSettings.ProtocolSettings).ProtoReflect().Descriptor())
	if err == nil {
		if msgOpts.TransportOriginalName != "" {
			return msgOpts.TransportOriginalName
		}
	}
	return ""
}
