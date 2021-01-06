package routing

import (
	"sync"
	"time"
)

// HealthChecker is able to check health of outbound hanlders of its balancers.
type HealthChecker interface {

	// HealthCheck performs a health check for outbound hanlders of its balancers
	HealthCheck(tags []string)
}

// ThrottledChecker run Health Checks Throttled
type ThrottledChecker struct {
	mux  sync.Mutex
	tags []string
	prev *time.Timer

	Checker HealthChecker
	Delay   time.Duration
}

// Run runs a check for give tag
func (t *ThrottledChecker) Run(tag string) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.tags = append(t.tags, tag)
	if t.prev != nil {
		t.prev.Stop()
	}
	t.prev = time.AfterFunc(t.Delay, func() {
		t.mux.Lock()
		tags := t.tags
		t.tags = make([]string, 0)
		t.mux.Unlock()
		// newError("#", idx, "running").AtDebug().WriteToLog()
		t.Checker.HealthCheck(tags)
	})
}
