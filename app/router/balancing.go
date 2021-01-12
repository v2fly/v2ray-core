package router

import (
	"v2ray.com/core/features/outbound"
	"v2ray.com/core/features/routing"
)

// Balancer represents a balancer
type Balancer struct {
	selectors     []string
	strategy      routing.BalancingStrategy
	healthChecker *HealthChecker
	ohm           outbound.Manager
	fallbackTag   string
}

// PickOutbound picks the tag of a outbound
func (b *Balancer) PickOutbound() (string, error) {
	candidates, err := b.SelectOutbounds()
	if err != nil {
		if b.fallbackTag != "" {
			newError("fallback to [", b.fallbackTag, "], due to error: ", err).AtInfo().WriteToLog()
			return b.fallbackTag, nil
		}
		return "", err
	}
	var tag string
	if !b.healthChecker.Settings.Enabled {
		// if not enabled, should ignore the results, since they could be outdated
		tag = b.strategy.PickOutbound(candidates, nil)
	} else {
		b.healthChecker.access.Lock()
		tag = b.strategy.PickOutbound(candidates, b.healthChecker.Results)
		b.healthChecker.access.Unlock()
	}
	if tag == "" {
		if b.fallbackTag != "" {
			newError("fallback to [", b.fallbackTag, "], due to empty tag returned").AtInfo().WriteToLog()
			return b.fallbackTag, nil
		}
		// will use default handler
		return "", newError("balancing strategy returns empty tag")
	}
	return tag, nil
}

// SelectOutbounds select outbounds with selectors of the Balancer
func (b *Balancer) SelectOutbounds() ([]string, error) {
	hs, ok := b.ohm.(outbound.HandlerSelector)
	if !ok {
		return nil, newError("outbound.Manager is not a HandlerSelector")
	}
	tags := hs.Select(b.selectors)
	return tags, nil
}
