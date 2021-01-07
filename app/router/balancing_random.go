package router

import (
	"v2ray.com/core/common/dice"
)

// RandomStrategy represents a random balancing strategy
type RandomStrategy struct {
	balancer *Balancer
}

// PickOutbound implements the BalancingStrategy.
// It picks an outbound from all tags (or alive tags if health check enabled) randomly
func (s *RandomStrategy) PickOutbound() (string, error) {
	tags, err := s.balancer.SelectOutbounds()
	if err != nil {
		return "", err
	}
	cntAll := len(tags)
	if cntAll == 0 {
		return "", nil
	}
	if !s.balancer.healthChecker.Settings.Enabled {
		return tags[dice.Roll(cntAll)], nil
	}

	alive := s.selectOutboundsAlive(tags)
	cntAlive := len(alive)
	if cntAll == 0 {
		newError("random: no outbound alive").AtInfo().WriteToLog()
		// goes to fallbackTag
		return "", nil
	}
	return alive[dice.Roll(cntAlive)], nil
}

// selectOutboundsAlive select alive outbounds.
func (s *RandomStrategy) selectOutboundsAlive(tags []string) []string {
	aliveTags := make([]string, 0)
	s.balancer.healthChecker.access.Lock()
	defer s.balancer.healthChecker.access.Unlock()
	for _, tag := range tags {
		r, ok := s.balancer.healthChecker.Results[tag]
		if !ok {
			continue
		}
		if r.AverageRTT <= 0 {
			continue
		}
		aliveTags = append(aliveTags, tag)
	}
	return aliveTags
}
