package routing

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	// PickOutbound pick one outbound from candidates.
	PickOutbound(candidates []string) string
	// GetInformation gets information of the strategy
	GetInformation(tags []string) *StrategyInfo
}
