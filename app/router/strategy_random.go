package router

import (
	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/routing"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct{}

// GetInformation implements the routing.BalancingStrategy.
func (s *RandomStrategy) GetInformation(tags []string) *routing.StrategyInfo {
	items := make([]*routing.OutboundInfo, 0)
	for _, tag := range tags {
		items = append(items, &routing.OutboundInfo{Tag: tag})
	}
	return &routing.StrategyInfo{
		Settings:    []string{"random"},
		ValueTitles: nil,
		Selects:     items,
		Others:      nil,
	}
}

// PickOutbound implements the routing.BalancingStrategy.
func (s *RandomStrategy) PickOutbound(candidates []string) string {
	count := len(candidates)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return candidates[dice.Roll(count)]
}
