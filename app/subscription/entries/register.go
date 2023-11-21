package entries

import "github.com/v2fly/v2ray-core/v5/app/subscription/specs"

type ConverterRegistry struct {
	knownConverters map[string]Converter
	parent          *ConverterRegistry
}

var globalConverterRegistry = &ConverterRegistry{knownConverters: map[string]Converter{}}

func RegisterConverter(kind string, converter Converter) error {
	return globalConverterRegistry.RegisterConverter(kind, converter)
}

func GetOverlayConverterRegistry() *ConverterRegistry {
	return globalConverterRegistry.GetOverlayConverterRegistry()
}

func (c *ConverterRegistry) RegisterConverter(kind string, converter Converter) error {
	if _, found := c.knownConverters[kind]; found {
		return newError("converter already registered for kind ", kind)
	}
	c.knownConverters[kind] = converter
	return nil
}

func (c *ConverterRegistry) TryAllConverters(rawConfig []byte, prioritizedConverter, kindHint string) (*specs.SubscriptionServerConfig, error) {
	if prioritizedConverter != "" {
		if converter, found := c.knownConverters[prioritizedConverter]; found {
			serverConfig, err := converter.ConvertToAbstractServerConfig(rawConfig, kindHint)
			if err == nil {
				return serverConfig, nil
			}
		}
	}

	for _, converter := range c.knownConverters {
		serverConfig, err := converter.ConvertToAbstractServerConfig(rawConfig, kindHint)
		if err == nil {
			return serverConfig, nil
		}
	}
	if c.parent != nil {
		if serverConfig, err := c.parent.TryAllConverters(rawConfig, prioritizedConverter, kindHint); err == nil {
			return serverConfig, nil
		}
	}
	return nil, newError("no converter found for config")
}

func (c *ConverterRegistry) GetOverlayConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{knownConverters: map[string]Converter{}, parent: c}
}
