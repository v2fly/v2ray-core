package router

import (
	"time"

	"v2ray.com/core/features/outbound"
	"v2ray.com/core/features/routing"
)

// Balancer represents a balancer
type Balancer struct {
	selectors   []string
	strategy    routing.BalancingStrategy
	ohm         outbound.Manager
	fallbackTag string

	override overridden
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
	if o := b.override.Get(); o != nil {
		tag = b.strategy.Pick(o.selects)
	} else {
		tag = b.strategy.SelectAndPick(candidates)
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

func (b *Balancer) overrideSelecting(selects []string, validity time.Duration) error {
	if validity <= 0 {
		b.override.Clear()
		return nil
	}
	hs, ok := b.ohm.(outbound.HandlerSelector)
	if !ok {
		return newError("outbound.Manager is not a HandlerSelector")
	}
	tags := hs.Select(selects)
	if len(tags) == 0 {
		return newError("no outbound selected")
	}
	b.override.Put(tags, time.Now().Add(validity))
	return nil
}

// OverrideSelecting implements routing.BalancingOverrider
func (r *Router) OverrideSelecting(balancer string, selects []string, validity time.Duration) error {
	var b *Balancer
	for tag, bl := range r.balancers {
		if tag == balancer {
			b = bl
			break
		}
	}
	if b == nil {
		return newError("balancer '", balancer, "' not found")
	}
	err := b.overrideSelecting(selects, validity)
	if err != nil {
		return err
	}
	return nil
}
