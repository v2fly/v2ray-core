package router

import (
	"v2ray.com/core/features/outbound"
)

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	PickOutbound() (string, error)
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
	tag, err := b.strategy.PickOutbound()
	if err != nil {
		return "", err
	}
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
