package router

import (
	"fmt"
	"sort"
	"time"

	"v2ray.com/core/common/dice"
)

// LeastLoadStrategy represents a random balancing strategy
type LeastLoadStrategy struct {
	balancer *Balancer
	settings *StrategyLeastLoadConfig
}

// node is a minimal copy of HealthCheckResult
// we don't use HealthCheckResult directly because
// it may change by health checker during routing
type node struct {
	Tag        string
	AverageRTT time.Duration
}

func (n *node) String() string {
	return fmt.Sprintf("%s(%s)", n.Tag, n.AverageRTT)
}

// PickOutbound implements the BalancingStrategy.
// It picks an outbound from tags randomly and respects the health check settings
func (s *LeastLoadStrategy) PickOutbound() (string, error) {
	if !s.balancer.healthChecker.Settings.Enabled {
		newError("least load: health checher not enabled, will work like random strategy").AtWarning().WriteToLog()
	}
	tags, err := s.balancer.SelectOutbounds()
	if err != nil {
		return "", err
	}
	cntAll := len(tags)
	if cntAll == 0 {
		return "", newError("least load: no available outbounds").AtWarning()
	}

	alive, err := s.getNodesAlive(tags)
	if err != nil {
		return "", err
	}
	cntAlive := len(alive)
	if cntAlive == 0 {
		newError("least load: no outbound alive, select one whatever").AtInfo().WriteToLog()
		return tags[dice.Roll(cntAll)], nil
	}

	nodes, err := s.selectLeastLoad(alive)
	if err != nil {
		return "", err
	}
	cntNodes := len(nodes)
	if cntNodes == 0 {
		newError("least load: no outbound matches, select alive one whatever").AtInfo().WriteToLog()
		return alive[dice.Roll(cntAlive)].Tag, nil
	}
	newError("least load tags: ", nodes).AtDebug().WriteToLog()
	return nodes[dice.Roll(cntNodes)].Tag, nil
}

// TODO: test for config modes below

// selectLeastLoad selects nodes according to Baselines and Expected Count.
//
// The strategy always improves network response speed, not matter which mode below is configurated.
// But they can still have different priorities.
//
// 1. Bandwidth priority: no Baseline + Expected Count > 0.: selects `Expected Count` amount of nodes.
// (one if Expected Count <= 0)
//
// 2. Bandwidth priority advanced: Baselines + Expected Count > 0.
// Select `Expected Count` amount of nodes, and also those near them according to baselines.
// In other words, it selects according to different Baselines, until one of them matches
// the Expected Count, if no Baseline matches, Expected Count applied.
//
// 3. Speed priority: Baselines + `Expected Count <= 0`.
// go through all base line until find selects, if not, select the fastest first one
func (s *LeastLoadStrategy) selectLeastLoad(nodes []*node) ([]*node, error) {
	expected := int(s.settings.Expected)
	availableCount := len(nodes)
	if expected > availableCount {
		return nodes, nil
	}

	if expected <= 0 {
		expected = 1
	}
	if len(s.settings.Baselines) == 0 {
		return nodes[:expected], nil
	}

	count := 0
	// go through all base line until find expected selects
	for _, b := range s.settings.Baselines {
		baseline := time.Duration(b)
		for i := 0; i < availableCount; i++ {
			if nodes[i].AverageRTT > baseline {
				break
			}
			count = i + 1
		}
		// don't continue if find expected selects
		if count >= expected {
			newError("applied baseline: ", baseline).AtDebug().WriteToLog()
			break
		}
	}
	if count < expected {
		count = expected
	}
	return nodes[:count], nil
}

func (s *LeastLoadStrategy) getNodesAlive(tags []string) (nodes []*node, err error) {
	s.balancer.healthChecker.access.Lock()
	defer s.balancer.healthChecker.access.Unlock()
	// nodes := make([]*node, 0)
	for _, tag := range tags {
		r, ok := s.balancer.healthChecker.Results[tag]
		if !ok {
			// not checked, marked with AverageRTT=0
			nodes = append(nodes, &node{
				Tag:        tag,
				AverageRTT: 0,
			})
			continue
		}
		if r.FailCount > 0 {
			continue
		}
		nodes = append(nodes, &node{
			Tag:        tag,
			AverageRTT: r.AverageRTT,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		iRTT := nodes[i].AverageRTT
		jRTT := nodes[j].AverageRTT
		// 0 rtt means not checked, sort in the tail
		if iRTT == 0 && jRTT > 0 {
			return false
		}
		if iRTT > 0 && jRTT == 0 {
			return true
		}
		return iRTT < jRTT
	})
	return nodes, nil
}
