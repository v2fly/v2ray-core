package router

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/dice"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/extension"
	"google.golang.org/protobuf/runtime/protoiface"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct {
	ctx         context.Context
	settings    *StrategyRandomConfig
	observatory extension.Observatory
}

func (s *RandomStrategy) GetPrincipleTarget(strings []string) []string {
	return strings
}

// NewRandomStrategy creates a new RandomStrategy with settings
func NewRandomStrategy(settings *StrategyRandomConfig) *RandomStrategy {
	return &RandomStrategy{
		settings: settings,
	}
}

func (s *RandomStrategy) InjectContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *RandomStrategy) PickOutbound(candidates []string) string {
	if s.settings.AliveOnly {
		if s.observatory == nil {
			core.RequireFeatures(s.ctx, func(observatory extension.Observatory) error {
				s.observatory = observatory
				return nil
			})
		}
		if s.observatory != nil {
			var observeReport protoiface.MessageV1
			var err error
			if s.settings.ObserverTag == "" {
				observeReport, err = s.observatory.GetObservation(s.ctx)
			} else {
				observeReport, err = common.Must2(s.observatory.(features.TaggedFeatures).GetFeaturesByTag(s.settings.ObserverTag)).(extension.Observatory).GetObservation(s.ctx)
			}
			if err == nil {
				outboundsList := outboundList(candidates)
				aliveTags := make([]string, 0)
				if result, ok := observeReport.(*observatory.ObservationResult); ok {
					status := result.Status
					for _, v := range status {
						// outbound is alive unless proven not
						if !(outboundsList.contains(v.OutboundTag) && !v.Alive) {
							aliveTags = append(aliveTags, v.OutboundTag)
						}
					}
					candidates = aliveTags
				}
			}
		}
	}

	count := len(candidates)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return candidates[dice.Roll(count)]
}

func init() {
	common.Must(common.RegisterConfig((*StrategyRandomConfig)(nil), nil))
}
