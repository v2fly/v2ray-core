package routing

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	// String return a human readable string of strategy name
	// and its settings
	String() string
	// PickOutbound pick one outbound from the results of SelectOutbound()
	PickOutbound(candidates []string, results map[string]*HealthCheckResult) string
	// SelectOutbound selects outbounds before the final pick
	SelectOutbounds(candidates []string, results map[string]*HealthCheckResult) []string
	// GetInfo get running information of the strategy
	GetInfo(tags []string, results map[string]*HealthCheckResult) *StrategyInfo
}
