package router

import (
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/dice"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct{}

func (s *RandomStrategy) GetPrincipleTarget(strings []string) []string {
	return strings
}

func (s *RandomStrategy) PickOutbound(candidates []string) string {
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
