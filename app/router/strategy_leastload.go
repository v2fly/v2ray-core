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

// FIXME: balancer unavailable if: top nodes are failed now,
// but next check not yet performed. "top but failed" nodes
// will always be selected.

// LeastLoadStrategy represents a random balancing strategy
type LeastLoadStrategy struct {
	*HealthPing

	settings *StrategyLeastLoadConfig
}

// node is a minimal copy of HealthCheckResult
// we don't use HealthCheckResult directly because
// it may change by health checker during routing
type node struct {
	Tag          string
	RTTAverage   time.Duration
	RTTDeviation time.Duration
}

// GetInformation implements the routing.BalancingStrategy.
func (s *LeastLoadStrategy) GetInformation(tags []string) *routing.StrategyInfo {
	s.HealthPing.access.Lock()
	defer s.HealthPing.access.Unlock()

	var (
		selects      []*node
		selectsCount int
	)
	if s.HealthPing.Results == nil {
		// be consistent with PickOutbound(), select all if not checked
		_, selects = s.getNodes(tags, nil, 0)
		selectsCount = len(tags)
	} else {
		qualified, others := s.getNodes(tags, s.HealthPing.Results, 0)
		// others is no sorted by getNodes()
		sort.Slice(others, func(i, j int) bool {
			return leastloadSortLess(others[i], others[j])
		})
		selects = s.selectLeastLoad(qualified)
		selectsCount = len(selects)
		// append qualified but not selected outbounds to selected tags
		for i := selectsCount; i < len(qualified); i++ {
			selects = append(selects, qualified[i])
		}
		// append other outbounds to selected tags
		for _, n := range others {
			selects = append(selects, n)
		}
	}
	titles, items := getHealthPingInfo(selects, s.HealthPing.Results)
	return &routing.StrategyInfo{
		Settings:    s.getSettings(),
		ValueTitles: titles,
		Selects:     items[:selectsCount],
		Others:      items[selectsCount:],
	}
}

// PickOutbound implements the routing.BalancingStrategy.
// It picks an outbound from least load tags, according to the health check results
func (s *LeastLoadStrategy) PickOutbound(candidates []string) string {
	s.HealthPing.access.Lock()
	defer s.HealthPing.access.Unlock()
	count := len(candidates)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	if s.HealthPing.Results == nil {
		return candidates[dice.Roll(count)]
	}
	nodes, _ := s.getNodes(candidates, s.HealthPing.Results, 0)
	selects := s.selectLeastLoad(nodes)
	count = len(selects)
	if count == 0 {
		// goes to fallbackTag
		return ""
	}
	return selects[dice.Roll(count)].Tag
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
			if nodes[i].RTTDeviation > baseline {
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

func (s *LeastLoadStrategy) getNodes(candidates []string, results map[string]*HealthPingResult, maxRTT time.Duration) ([]*node, []*node) {
	qualified := make([]*node, 0)
	others := make([]*node, 0)
	for _, tag := range candidates {
		r, ok := results[tag]
		// make sure sort less order: "normal" < "not tested" < "failed"
		switch {
		case !ok:
			others = append(others, &node{
				Tag:          tag,
				RTTDeviation: math.MaxInt64,
				RTTAverage:   math.MaxInt64 - 1,
			})
		case r.FailCount > 0:
			others = append(others, &node{
				Tag:          tag,
				RTTDeviation: math.MaxInt64,
				RTTAverage:   math.MaxInt64,
			})
		case maxRTT > 0 && r.RTTAverage > maxRTT:
			others = append(others, &node{
				Tag:          tag,
				RTTDeviation: r.RTTDeviation,
				RTTAverage:   r.RTTAverage,
			})
		default:
			qualified = append(qualified, &node{
				Tag:          tag,
				RTTDeviation: r.RTTDeviation,
				RTTAverage:   r.RTTAverage,
			})
		}
	}
	sort.Slice(qualified, func(i, j int) bool {
		return leastloadSortLess(qualified[i], qualified[j])
	})
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
	settings = append(settings, fmt.Sprintf("leastload, expected: %d, baselines: %s", s.settings.Expected, sb))
	settings = append(settings, fmt.Sprintf(
		"health ping, interval: %s, rounds: %d, timeout: %s, destination: %s",
		s.HealthPing.Settings.Interval,
		s.HealthPing.Settings.Rounds,
		s.HealthPing.Settings.Timeout,
		s.HealthPing.Settings.Destination,
	))
	return settings
}

func getHealthPingInfo(nodes []*node, results map[string]*HealthPingResult) ([]string, []*routing.OutboundInfo) {
	failed := []string{"failed", "-"}
	notTested := []string{"not tested", "-"}
	items := make([]*routing.OutboundInfo, 0)
	for _, node := range nodes {
		item := &routing.OutboundInfo{
			Tag: node.Tag,
		}
		result, ok := results[node.Tag]
		switch {
		case !ok || result.Count == 0:
			item.Values = notTested
		case result.FailCount > 0:
			item.Values = failed
		default:
			item.Values = []string{result.RTTDeviation.String(), result.RTTAverage.String()}
		}
		items = append(items, item)
	}
	return []string{"RTT STD.", "RTT Avg."}, items
}

func leastloadSortLess(i, j *node) bool {
	left := i.RTTDeviation
	right := j.RTTDeviation
	if left == right {
		return i.RTTAverage < j.RTTAverage
	}
	return left < right
}
