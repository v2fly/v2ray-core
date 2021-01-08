package router

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/routing"
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

// String implements the BalancingStrategy.
func (s *LeastLoadStrategy) String() string {
	sb := new(strings.Builder)
	for i, b := range s.settings.Baselines {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(time.Duration(b).String())
	}
	return fmt.Sprintf(`LeastLoad strategy, expected: %d, baselines: %s`, s.settings.Expected, sb)
}

// GetInfo implements the BalancingStrategy.
func (s *LeastLoadStrategy) GetInfo() (*routing.StrategyInfo, error) {
	tags, err := s.SelectOutbounds()
	if err != nil {
		return nil, err
	}
	selectsCount := len(tags)
	all, err := s.balancer.SelectOutbounds()
	if err != nil {
		return nil, err
	}
	// append other outbounds to selected tags
	for _, t := range all {
		if findSliceIndex(tags, t) < 0 {
			tags = append(tags, t)
		}
	}
	items := getHealthRTT(tags, s.balancer.healthChecker)
	return &routing.StrategyInfo{
		Name:        s.String(),
		ValueTitles: []string{"RTT"},
		Selects:     items[:selectsCount],
		Others:      items[selectsCount:],
	}, nil
}

// PickOutbound implements the BalancingStrategy.
// It picks an outbound from least load tags, according to the health check results
func (s *LeastLoadStrategy) PickOutbound() (string, error) {
	tags, err := s.SelectOutbounds()
	if err != nil {
		return "", err
	}
	count := len(tags)
	if count == 0 {
		// goes to fallbackTag
		return "", nil
	}
	return tags[dice.Roll(count)], nil
}

// SelectOutbounds implements BalancingStrategy
func (s *LeastLoadStrategy) SelectOutbounds() ([]string, error) {
	if !s.balancer.healthChecker.Settings.Enabled {
		newError("least load: health checher not enabled, will work like random strategy").AtWarning().WriteToLog()
	}
	tags, err := s.balancer.SelectOutbounds()
	if err != nil {
		return nil, err
	}
	cntAll := len(tags)
	if cntAll == 0 {
		return nil, nil
	}

	alive, err := s.getNodesAlive(tags)
	if err != nil {
		return nil, err
	}
	cntAlive := len(alive)
	if cntAlive == 0 {
		newError("least load: no outbound alive").AtInfo().WriteToLog()
		return nil, nil
	}

	selects := make([]string, 0)
	nodes := s.selectLeastLoad(alive)

	for _, node := range nodes {
		selects = append(selects, node.Tag)
	}
	return selects, nil
}

// selectLeastLoad selects nodes according to Baselines and Expected Count.
//
// The strategy always improves network response speed, not matter which mode below is configurated.
// But they can still have different priorities.
//
// 1. Bandwidth priority: no Baseline + Expected Count > 0.: selects `Expected Count` of nodes.
// (one if Expected Count <= 0)
//
// 2. Bandwidth priority advanced: Baselines + Expected Count > 0.
// Select `Expected Count` amount of nodes, and also those near them according to baselines.
// In other words, it selects according to different Baselines, until one of them matches
// the Expected Count, if no Baseline matches, Expected Count applied.
//
// 3. Speed priority: Baselines + `Expected Count <= 0`.
// go through all baselines until find selects, if not, select none. Used in combination
// with 'balancer.fallbackTag', it means: selects qualified nodes or use the fallback.
func (s *LeastLoadStrategy) selectLeastLoad(nodes []*node) []*node {
	expected := int(s.settings.Expected)
	availableCount := len(nodes)
	if expected > availableCount {
		return nodes
	}

	if expected <= 0 {
		expected = 1
	}
	if len(s.settings.Baselines) == 0 {
		return nodes[:expected]
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
	if s.settings.Expected > 0 && count < expected {
		count = expected
	}
	return nodes[:count]
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
