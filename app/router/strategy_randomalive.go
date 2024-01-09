package router

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/dice"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

// RandomStrategy represents a random balancing strategy
type RandomAliveStrategy struct {
	ctx         context.Context
	observatory extension.Observatory

	config *StrategyRandomAliveConfig
}

func (s *RandomAliveStrategy) GetPrincipleTarget(strings []string) []string {
	return strings
}

func (s *RandomAliveStrategy) InjectContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *RandomAliveStrategy) PickOutbound(candidates []string) string {
	if s.observatory == nil {
		common.Must(core.RequireFeatures(s.ctx, func(observatory extension.Observatory) error {
			s.observatory = observatory
			return nil
		}))
	}
	observeReport, err := s.observatory.GetObservation(s.ctx)
	if err != nil {
		newError("cannot get observe report").Base(err).WriteToLog()
		return ""
	}
	outboundsList := outboundList(candidates)

	aliveTags := make([]string, 0)
	if result, ok := observeReport.(*observatory.ObservationResult); ok {
		status := result.Status
		for _, v := range status {
			if outboundsList.contains(v.OutboundTag) && v.Alive {
				aliveTags = append(aliveTags, v.OutboundTag)
			}
		}
		count := len(aliveTags)
		if count == 0 {
			// goes to fallbackTag
			return ""
		}
		return aliveTags[dice.Roll(count)]
	}

	return ""
}

func init() {
	common.Must(common.RegisterConfig((*StrategyRandomAliveConfig)(nil), nil))
}
