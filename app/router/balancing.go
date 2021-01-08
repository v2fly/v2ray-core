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
	tag, err := b.strategy.PickOutbound()
	if err != nil {
		if b.fallbackTag != "" {
			newError("fallback to [", b.fallbackTag, "], due to error: ", err).AtInfo().WriteToLog()
			return b.fallbackTag, nil
		}
		return "", err
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
