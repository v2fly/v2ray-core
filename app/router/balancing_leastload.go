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
		newError("health checher not enabled, 'Least Load' strategy will work like 'Random'").AtWarning().WriteToLog()
	}
	tags, err := s.balancer.SelectOutbounds()
	if err != nil {
		return "", err
	}
	cntAll := len(tags)
	if cntAll == 0 {
		return "", newError("no available outbounds").AtWarning()
	}

	nodes, err := s.getNodesAlive(tags)
	if err != nil {
		return "", err
	}
	if len(nodes) > 0 {
		nodes, err = s.selectLeastLoad(nodes)
		if err != nil {
			return "", err
		}
	}
	cntNodes := len(nodes)
	if cntNodes == 0 {
		newError("no outbounds alive, select one whatever").AtInfo().WriteToLog()
		return tags[dice.Roll(cntAll)], nil
	}
	newError("least load tags: ", nodes).AtDebug().WriteToLog()
	return nodes[dice.Roll(cntNodes)].Tag, nil
}

// selectLeastLoad selects nodes with Baselines and Expected Count.
//
// If no baseline provided: selects first `Expected Count` amount of nodes.
// (first one if Expected Count <= 0)
//
// If Baselines provided and `Expected Count <= 0`: no minimal nodes required, selecting only according
// to Baselines[0]
//
// If Baselines provided and `Expected Count > 0`: requires a minimal amount of nodes, selecting according
// to different Baselines, until one of them matches Expected Count.
// If no Baselines match, Expected Count applied.
func (s *LeastLoadStrategy) selectLeastLoad(nodes []*node) ([]*node, error) {
	expected := int(s.settings.Expected)
	availableCount := len(nodes)
	if expected > availableCount {
		return nodes, nil
	}
	if len(s.settings.Baselines) == 0 {
		if expected <= 0 {
			return nodes[:1], nil
		}
		return nodes[:expected], nil
	}

	// no Expected Count required
	if expected == 0 {
		count := 0
		baseline := time.Duration(s.settings.Baselines[0])
		newError("applied baseline: ", baseline).AtDebug().WriteToLog()
		for i := 0; i < availableCount; i++ {
			if nodes[i].AverageRTT > baseline {
				break
			}
			count = i + 1
		}
		return nodes[:count], nil
	}
	// Expected Count required
	count := expected
	baseline := nodes[expected-1].AverageRTT
	for _, b := range s.settings.Baselines {
		tb := time.Duration(b)
		if tb > baseline {
			newError("applied baseline: ", tb).AtDebug().WriteToLog()
			baseline = tb
			for i := expected; i < availableCount; i++ {
				if nodes[i].AverageRTT > baseline {
					break
				}
				count = i + 1
			}
			break
		}
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
