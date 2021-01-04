package router

import (
	"v2ray.com/core/features/outbound"
)

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	PickOutbound(tags []string) string
}

// Balancer represents a balancer
type Balancer struct {
	selectors     []string
	strategy      BalancingStrategy
	healthChecker *HealthChecker
	ohm           outbound.Manager
}

// PickOutbound picks the tag of a outbound
func (b *Balancer) PickOutbound() (string, error) {
	tags, err := b.SelectOutbounds()
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", newError("no available outbounds selected")
	}
	tag := b.strategy.PickOutbound(tags)
	if tag == "" {
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
