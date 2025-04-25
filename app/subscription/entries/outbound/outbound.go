package outbound

import (
	"github.com/ghxhy/v2ray-core/v5/app/subscription/entries"
	"github.com/ghxhy/v2ray-core/v5/app/subscription/specs"
	"github.com/ghxhy/v2ray-core/v5/common"
)

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

// NewOutboundEntriesParser internal api
func NewOutboundEntriesParser() entries.Converter {
	return newOutboundEntriesParser()
}

func newOutboundEntriesParser() entries.Converter {
	return &outboundEntriesParser{}
}

type outboundEntriesParser struct{}

func (o *outboundEntriesParser) ConvertToAbstractServerConfig(rawConfig []byte, kindHint string) (*specs.SubscriptionServerConfig, error) {
	parser := specs.NewOutboundParser()
	outbound, err := parser.ParseOutboundConfig(rawConfig)
	if err != nil {
		return nil, newError("failed to parse outbound config").Base(err).AtWarning()
	}
	return parser.ToSubscriptionServerConfig(outbound)
}

func init() {
	common.Must(entries.RegisterConverter("outbound", newOutboundEntriesParser()))
}
