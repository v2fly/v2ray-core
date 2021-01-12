package router

import (
	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/routing"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct{}

// GetInfo implements the BalancingStrategy.
func (s *RandomStrategy) GetInfo(tags []string, results map[string]*routing.HealthCheckResult) *routing.StrategyInfo {
	selects := s.SelectOutbounds(tags, results)
	selectsCount := len(selects)
	// append other outbounds to selected tags
	for _, t := range tags {
		if findSliceIndex(selects, t) < 0 {
			selects = append(selects, t)
		}
	}
	items := getHealthRTT(selects, results)
	return &routing.StrategyInfo{
		Name:        "Random",
		ValueTitles: []string{"RTT"},
		Selects:     items[:selectsCount],
		Others:      items[selectsCount:],
	}
}

// PickOutbound implements the BalancingStrategy.
// It picks an outbound from all tags (or alive tags if health check enabled) randomly
func (s *RandomStrategy) PickOutbound(candidates []string, results map[string]*routing.HealthCheckResult) string {
	tags := s.SelectOutbounds(candidates, results)
	count := len(tags)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return tags[dice.Roll(count)]
}

// SelectOutbounds implements BalancingStrategy
func (s *RandomStrategy) SelectOutbounds(candidates []string, results map[string]*routing.HealthCheckResult) []string {
	if results == nil {
		return candidates
	}
	aliveTags := make([]string, 0)
	for _, tag := range candidates {
		r, ok := results[tag]
		if ok && r.FailCount > 0 {
			continue
		}
		aliveTags = append(aliveTags, tag)
	}
	if len(aliveTags) == 0 {
		newError("random: no outbound alive").AtInfo().WriteToLog()
	}
	return aliveTags
}
