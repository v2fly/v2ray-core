package router

import (
	sync "sync"
	"time"
)

type overriddenSettings struct {
	selects []string
	until   time.Time
	pasue   bool
}

type overridden struct {
	access   sync.RWMutex
	settings overriddenSettings
}

// Get gets the overridden settings
func (o *overridden) Get() *overriddenSettings {
	o.access.RLock()
	defer o.access.RUnlock()
	if len(o.settings.selects) == 0 || time.Now().After(o.settings.until) {
		return nil
	}
	return &overriddenSettings{
		selects: o.settings.selects,
		until:   o.settings.until,
		pasue:   o.settings.pasue,
	}
}

// Put updates the overridden settings
func (o *overridden) Put(selects []string, until time.Time, pause bool) {
	o.access.Lock()
	defer o.access.Unlock()
	o.settings.selects = selects
	o.settings.until = until
	o.settings.pasue = pause
}

// Clear clears the overridden settings
func (o *overridden) Clear() {
	o.access.Lock()
	defer o.access.Unlock()
	o.settings.selects = nil
	o.settings.until = time.Time{}
	o.settings.pasue = false
}
