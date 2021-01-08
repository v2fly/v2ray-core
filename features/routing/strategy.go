package routing

// BalancingStrategy is the interface of a balancing strategy
type BalancingStrategy interface {
	// String return a human readable string of strategy name
	// and its settings
	String() string
	// PickOutbound pick one outbound from the results of SelectOutbound()
	PickOutbound() (string, error)
	// SelectOutbound selects outbounds before the final pick
	SelectOutbounds() ([]string, error)
	// GetInfo get running information of the strategy
	GetInfo() (*StrategyInfo, error)
}
