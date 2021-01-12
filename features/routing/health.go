package routing

import (
	"sync"
	"time"
)

// HealthCheckSettings holds settings for health Checker
type HealthCheckSettings struct {
	Enabled     bool
	Destination string
	Interval    time.Duration
	Rounds      int
	Timeout     time.Duration
}

// HealthCheckResult holds result for health Checker
type HealthCheckResult struct {
	Count      int
	FailCount  int
	AverageRTT time.Duration
	MaxRTT     time.Duration
	MinRTT     time.Duration
	RTTs       []time.Duration
}

// OutboundInfo holds information of an outbound, like health stats
type OutboundInfo struct {
	Tag    string
	Values []string
}

// StrategyInfo hold strategy running infomations, like
// selected and other handlers, which contains RTT etc.
type StrategyInfo struct {
	Name        string
	ValueTitles []string
	Selects     []*OutboundInfo
	Others      []*OutboundInfo
}

// BalancerInfo represents a health stats of a balancers
type BalancerInfo struct {
	Tag         string
	Strategy    *StrategyInfo
	HealthCheck *HealthCheckSettings
}

// HealthChecker represents a health checker
type HealthChecker interface {
	// CheckHanlders performs a health check for specified outbound hanlders.
	// Set distributed to make it not check all tags at same time, checks
	// are distributed randomly in the timeline
	CheckHanlders(tags []string, distributed bool) error
}

// RouterChecker is a router able to perform health check and stats for outbound hanlders.
type RouterChecker interface {
	HealthChecker
	// BalancerHealthCheck performs health checks for specified balancers,
	// if not specified, check them all.
	// Set distributed to make it not check all tags at same time, checks
	// are distributed randomly in the timeline
	CheckBalancers(tags []string, distributed bool) error
	// GetBalancersInfo get health info of specific balancer, if balancer not
	//  specified, get all
	GetBalancersInfo(tags []string) ([]*BalancerInfo, error)
}

// ThrottledChecker run Health Checks Throttled
type ThrottledChecker struct {
	mux  sync.Mutex
	tags []string
	prev *time.Timer

	Checker RouterChecker
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
		t.Checker.CheckHanlders(tags, true)
	})
}
