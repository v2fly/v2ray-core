package entries

import "github.com/ghxhy/v2ray-core/v5/app/subscription/specs"

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

type Converter interface {
	ConvertToAbstractServerConfig(rawConfig []byte, kindHint string) (*specs.SubscriptionServerConfig, error)
}
