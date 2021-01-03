package router

import (
	"v2ray.com/core/common/dice"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct {
	balancer *Balancer
}

// PickOutbound implements the BalancingStrategy.
// It picks an outbound from tags randomly and respects the health check settings
func (s *RandomStrategy) PickOutbound() string {
	n := len(s.balancer.selectors)
	if n == 0 {
		panic("0 tags")
	}
	// TODO: filter alive outbounds
	return s.balancer.selectors[dice.Roll(n)]
}
