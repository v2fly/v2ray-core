package routing

import (
	"sync"
	"time"
)

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
	Tag      string
	Strategy *StrategyInfo
}

// HealthChecker is able to perform health check and stats for outbound hanlders.
type HealthChecker interface {
	// CheckHanlders performs a health check for specified outbound hanlders
	CheckHanlders(tags []string) error
	// BalancerHealthCheck performs health checks for specified balancers,
	// if not specified, check them all
	CheckBalancers(tags []string) error
	// GetBalancersInfo get health info of specific balancer, if balancer not
	//  specified, get all
	GetBalancersInfo(tags []string) ([]*BalancerInfo, error)
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
		t.Checker.CheckHanlders(tags)
	})
}
