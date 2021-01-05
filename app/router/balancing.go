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
	tags, err := b.SelectOutbounds()
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", newError("no available outbounds selected")
	}
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

// SelectOutboundsAlive select alive outbounds according to the
// selectors and health chekerer of the Balancer. If health chekerer
// not enabled, it's equivalent to SelectOutbounds()
func (b *Balancer) SelectOutboundsAlive() ([]string, error) {
	tags, err := b.SelectOutbounds()
	if !b.healthChecker.Settings.Enabled {
		return tags, err
	}
	if err != nil || len(tags) == 0 {
		return nil, err
	}
	aliveTags := make([]string, 0)
	for _, tag := range tags {
		r, ok := b.healthChecker.Results[tag]
		if !ok {
			continue
		}
		if r.AverageRTT <= 0 {
			continue
		}
		aliveTags = append(aliveTags, tag)
	}
	return aliveTags, nil
}
