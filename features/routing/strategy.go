package routing

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	// PickOutbound pick one outbound from candidates.
	PickOutbound(candidates []string) string
	// GetInfo get information of the strategy
	GetInfo(tags []string) *StrategyInfo
}
