package router

import (
	"sync"
	"time"

	"v2ray.com/core/features/outbound"
	"v2ray.com/core/features/routing"
)

type overridden struct {
	access  sync.RWMutex
	selects []string
	until   time.Time
}

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
	if b.checkOverride() {
		tag = b.strategy.Pick(b.override.selects)
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
	b.override.access.Lock()
	defer b.override.access.Unlock()
	if validity <= 0 {
		// do not clear selects here, wait for checkOverride()
		// b.override.selects = nil
		b.override.until = time.Time{}
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
	b.override.selects = tags
	b.override.until = time.Now().Add(validity)
	return nil
}

func (b *Balancer) checkOverride() bool {
	if len(b.override.selects) == 0 {
		return false
	}
	if time.Now().Before(b.override.until) {
		return true
	}
	b.override.access.Lock()
	b.override.selects = nil
	b.override.until = time.Time{}
	b.override.access.Unlock()
	// restart scheduler
	checker, ok := b.strategy.(routing.HealthChecker)
	if ok {
		checker.StartScheduler(b.SelectOutbounds)
	}
	return false
}

// OverrideSelecting implements routing.BalancingOverrider
func (r *Router) OverrideSelecting(balancer string, selects []string, validity time.Duration, stop bool) error {
	var b *Balancer
	for tag, bl := range r.balancers {
		if tag == balancer {
			b = bl
			break
		}
	}
	if b == nil {
		return newError("balancer ", balancer, " not found")
	}
	err := b.overrideSelecting(selects, validity)
	if err != nil {
		return err
	}
	// check to restart scheduler if it's a remove action
	if !b.checkOverride() {
		return nil
	}
	if !stop {
		return nil
	}
	checker, ok := b.strategy.(routing.HealthChecker)
	if !ok {
		return nil
	}
	checker.StopScheduler()
	return nil
}
