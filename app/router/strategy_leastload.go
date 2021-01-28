package router

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/routing"
)

// LeastLoadStrategy represents a random balancing strategy
type LeastLoadStrategy struct {
	*HealthPing

	settings *StrategyLeastLoadConfig
	costs    *WeightManager
}

// NewLeastLoadStrategy creates a new LeastLoadStrategy with settings
func NewLeastLoadStrategy(settings *StrategyLeastLoadConfig, dispatcher routing.Dispatcher) *LeastLoadStrategy {
	return &LeastLoadStrategy{
		HealthPing: NewHealthPing(settings.HealthCheck, dispatcher),
		settings:   settings,
		costs: NewWeightManager(
			settings.Costs, 1,
			func(value, cost float64) float64 {
				return value * math.Pow(cost, 0.5)
			},
		),
	}
}

// node is a minimal copy of HealthCheckResult
// we don't use HealthCheckResult directly because
// it may change by health checker during routing
type node struct {
	Tag              string
	Count            int
	Fail             int
	RTTAverage       time.Duration
	RTTDeviation     time.Duration
	RTTDeviationCost time.Duration
}

// GetInformation implements the routing.BalancingStrategy.
func (s *LeastLoadStrategy) GetInformation(tags []string) *routing.StrategyInfo {
	s.HealthPing.access.Lock()
	defer s.HealthPing.access.Unlock()
	qualified, others := s.getNodes(tags, s.HealthPing.Results, time.Duration(s.settings.MaxRTT))
	selects := s.selectLeastLoad(qualified)
	// append qualified but not selected outbounds to others
	others = append(others, qualified[len(selects):]...)
	leastloadSort(others)
	titles, sl := s.getNodesInfo(selects)
	_, ot := s.getNodesInfo(others)
	return &routing.StrategyInfo{
		Settings:    s.getSettings(),
		ValueTitles: titles,
		Selects:     sl,
		Others:      ot,
	}
}

// SelectAndPick implements the routing.BalancingStrategy.
func (s *LeastLoadStrategy) SelectAndPick(candidates []string) string {
	s.HealthPing.access.Lock()
	defer s.HealthPing.access.Unlock()
	qualified, _ := s.getNodes(candidates, s.HealthPing.Results, time.Duration(s.settings.MaxRTT))
	selects := s.selectLeastLoad(qualified)
	count := len(selects)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return selects[dice.Roll(count)].Tag
}

// Pick implements the routing.BalancingStrategy.
func (s *LeastLoadStrategy) Pick(candidates []string) string {
	count := len(candidates)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return candidates[dice.Roll(count)]
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
	if len(nodes) == 0 {
		newError("least load: no qualified outbound").AtInfo().WriteToLog()
		return nil
	}
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
			if nodes[i].RTTDeviationCost > baseline {
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

func (s *LeastLoadStrategy) getNodes(candidates []string, results map[string]*HealthPingRTTS, maxRTT time.Duration) ([]*node, []*node) {
	qualified := make([]*node, 0)
	unqualified := make([]*node, 0)
	failed := make([]*node, 0)
	untested := make([]*node, 0)
	others := make([]*node, 0)
	for _, tag := range candidates {
		r, ok := results[tag]
		if !ok {
			untested = append(untested, &node{
				Tag:              tag,
				RTTDeviationCost: math.MaxInt64 - 1,
				RTTDeviation:     math.MaxInt64 - 1,
				RTTAverage:       math.MaxInt64 - 1,
			})
			continue
		}
		stats := r.Get()
		switch {
		case !ok:
			untested = append(untested, &node{
				Tag:              tag,
				RTTDeviationCost: math.MaxInt64 - 1,
				RTTDeviation:     math.MaxInt64 - 1,
				RTTAverage:       math.MaxInt64 - 1,
			})
		case stats.FailCount > 0:
			failed = append(failed, &node{
				Tag:              tag,
				RTTDeviationCost: math.MaxInt64,
				RTTDeviation:     math.MaxInt64,
				RTTAverage:       math.MaxInt64,
				Count:            stats.Count,
				Fail:             stats.FailCount,
			})
		case maxRTT > 0 && stats.RTTAverage > maxRTT:
			unqualified = append(unqualified, &node{
				Tag:              tag,
				RTTDeviationCost: time.Duration(s.costs.Apply(tag, float64(stats.RTTDeviation))),
				RTTDeviation:     stats.RTTDeviation,
				RTTAverage:       stats.RTTAverage,
				Count:            stats.Count,
				Fail:             stats.FailCount,
			})
		default:
			qualified = append(qualified, &node{
				Tag:              tag,
				RTTDeviationCost: time.Duration(s.costs.Apply(tag, float64(stats.RTTDeviation))),
				RTTDeviation:     stats.RTTDeviation,
				RTTAverage:       stats.RTTAverage,
				Count:            stats.Count,
				Fail:             stats.FailCount,
			})
		}
	}
	if len(qualified) > 0 {
		leastloadSort(qualified)
		others = append(others, unqualified...)
		others = append(others, untested...)
		others = append(others, failed...)
	} else {
		qualified = untested
		others = append(others, unqualified...)
		others = append(others, failed...)
	}
	return qualified, others
}

func (s *LeastLoadStrategy) getSettings() []string {
	settings := make([]string, 0)
	sb := new(strings.Builder)
	for i, b := range s.settings.Baselines {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(time.Duration(b).String())
	}
	baselines := sb.String()
	if baselines == "" {
		baselines = "none"
	}
	maxRTT := time.Duration(s.settings.MaxRTT).String()
	if s.settings.MaxRTT == 0 {
		maxRTT = "none"
	}
	settings = append(settings, fmt.Sprintf("leastload, expected: %d, baselines: %s, max rtt: %s", s.settings.Expected, baselines, maxRTT))
	settings = append(settings, fmt.Sprintf(
		"health ping, interval: %s, sampling: %d, timeout: %s, destination: %s",
		s.HealthPing.Settings.Interval,
		s.HealthPing.Settings.SamplingCount,
		s.HealthPing.Settings.Timeout,
		s.HealthPing.Settings.Destination,
	))
	return settings
}

func (s *LeastLoadStrategy) getNodesInfo(nodes []*node) ([]string, []*routing.OutboundInfo) {
	titles := []string{"RTT STD+C    ", "RTT STD.     ", "RTT Avg.     ", "Cost "}
	hasCost := len(s.settings.Costs) > 0
	if !hasCost {
		titles = titles[1:3]
	}
	items := make([]*routing.OutboundInfo, 0)
	for _, node := range nodes {
		item := &routing.OutboundInfo{
			Tag: node.Tag,
		}
		cost := fmt.Sprintf("%.2f", s.costs.Get(node.Tag))
		switch node.RTTAverage {
		case math.MaxInt64 - 1:
			item.Values = []string{"not tested", "-"}
			if hasCost {
				item.Values = append(item.Values, "-", cost)
			}
		case math.MaxInt64:
			item.Values = []string{fmt.Sprintf("%d/%d failed", node.Fail, node.Count), "-"}
			if hasCost {
				item.Values = append(item.Values, "-", cost)
			}
		default:
			if hasCost {
				item.Values = []string{
					node.RTTDeviationCost.String(),
					node.RTTDeviation.String(),
					node.RTTAverage.String(),
					cost,
				}
			} else {
				item.Values = []string{
					node.RTTDeviation.String(),
					node.RTTAverage.String(),
				}
			}
		}
		items = append(items, item)
	}
	return titles, items
}

func leastloadSort(nodes []*node) {
	sort.Slice(nodes, func(i, j int) bool {
		left := nodes[i]
		right := nodes[j]
		if left.RTTDeviationCost != right.RTTDeviationCost {
			return left.RTTDeviationCost < right.RTTDeviationCost
		}
		if left.RTTAverage != right.RTTAverage {
			return left.RTTAverage < right.RTTAverage
		}
		if left.Fail != right.Fail {
			return left.Fail < right.Fail
		}
		return left.Tag < right.Tag
	})
}
