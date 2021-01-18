package routing

// HealthChecker is the interface for health checkers
type HealthChecker interface {
	// StartScheduler starts the check scheduler
	StartScheduler(hs func() ([]string, error))
	// StopScheduler stops the check scheduler
	StopScheduler()
	// Check start the health checking for given tags.
	Check(tags []string, distributed bool) error
}

// OutboundInfo holds information of an outbound, like health stats
type OutboundInfo struct {
	Tag    string
	Values []string
}

// StrategyInfo hold strategy running infomations, like
// selected and other handlers, which contains RTT etc.
type StrategyInfo struct {
	Settings    []string
	ValueTitles []string
	Selects     []*OutboundInfo
	Others      []*OutboundInfo
}

// BalancerInfo represents a health stats of a balancers
type BalancerInfo struct {
	Tag      string
	Strategy *StrategyInfo
}

// RouterChecker is a router able to perform health check and stats for outbound hanlders.
type RouterChecker interface {
	// CheckHanlders performs a health check for specified outbound hanlders.
	// Set distributed to make it not check all tags at same time, checks
	// are distributed randomly in the timeline
	CheckHanlders(tags []string, distributed bool) error
	// CheckBalancers performs health checks for specified balancers,
	// if not specified, check them all.
	// Set distributed to make it not check all tags at same time, checks
	// are distributed randomly in the timeline
	CheckBalancers(tags []string, distributed bool) error
	// GetBalancersInfo get health info of specific balancer, if balancer not
	//  specified, get all
	GetBalancersInfo(tags []string) ([]*BalancerInfo, error)
}
