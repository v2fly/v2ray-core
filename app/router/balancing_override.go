package router

import (
	sync "sync"
	"time"

	"github.com/v2fly/v2ray-core/v4/features/outbound"
)

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

type overrideSettings struct {
	selects []string
	until   time.Time
}

type override struct {
	access   sync.RWMutex
	settings overrideSettings
}

// Get gets the override settings
func (o *override) Get() *overrideSettings {
	o.access.RLock()
	defer o.access.RUnlock()
	if len(o.settings.selects) == 0 || time.Now().After(o.settings.until) {
		return nil
	}
	return &overrideSettings{
		selects: o.settings.selects,
		until:   o.settings.until,
	}
}

// Put updates the override settings
func (o *override) Put(selects []string, until time.Time) {
	o.access.Lock()
	defer o.access.Unlock()
	o.settings.selects = selects
	o.settings.until = until
}

// Clear clears the override settings
func (o *override) Clear() {
	o.access.Lock()
	defer o.access.Unlock()
	o.settings.selects = nil
	o.settings.until = time.Time{}
}
