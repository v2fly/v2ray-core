package routing

import (
	"sync"
	"time"
)

// HealthStatItem represents a health stats of an outbound
type HealthStatItem struct {
	Outbound string
	RTT      time.Duration
}

// HealthStats represents a health stats of a balancers
type HealthStats struct {
	Balancer  string
	Selects   []*HealthStatItem
	Outbounds []*HealthStatItem
}

// HealthChecker is able to perform health check and stats for outbound hanlders.
type HealthChecker interface {

	// HealthCheck performs a health check for outbound hanlders
	HealthCheck(tags []string)
	// GetHealthStats get health info of specific balancer, if balancer not specified, get all
	GetHealthStats(tag string) ([]*HealthStats, error)
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
