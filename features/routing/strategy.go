package routing

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	// PickOutbound pick one outbound from the results of SelectOutbound().
	// Note the results can be nil if health not enabled.
	PickOutbound(candidates []string, results map[string]*HealthCheckResult) string
	// SelectOutbound selects outbounds before the final pick
	// Note the results can be nil if health not enabled.
	SelectOutbounds(candidates []string, results map[string]*HealthCheckResult) []string
	// GetInfo get running information of the strategy
	// Note the results can be nil if health not enabled.
	GetInfo(tags []string, results map[string]*HealthCheckResult) *StrategyInfo
}
