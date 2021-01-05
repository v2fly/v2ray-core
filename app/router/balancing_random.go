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
func (s *RandomStrategy) PickOutbound() (string, error) {
	var (
		tags []string
		err  error
	)

	if s.balancer.healthChecker.Settings.Enabled {
		tags, err = s.balancer.SelectOutboundsAlive()
		newError("alive tags: ", tags).AtDebug().WriteToLog()
	} else {
		tags, err = s.balancer.SelectOutbounds()
	}
	if err != nil {
		return "", err
	}

	n := len(tags)
	if n == 0 {
		return "", newError("no available outbounds")
	}
	return tags[dice.Roll(n)], nil
}
