package router

import (
	"testing"
)

func TestSelectLeastExpected(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: nil,
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 100},
		{Tag: "b", AverageRTT: 200},
		{Tag: "c", AverageRTT: 300},
		{Tag: "d", AverageRTT: 350},
	}
	expected := 3
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpected2(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: nil,
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 100},
		{Tag: "b", AverageRTT: 200},
	}
	expected := 2
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpectedAndBaselines(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 300, 400},
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 100},
		{Tag: "b", AverageRTT: 200},
		{Tag: "c", AverageRTT: 250},
		{Tag: "d", AverageRTT: 300},
		{Tag: "e", AverageRTT: 310},
	}
	expected := 4
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpectedAndBaselines2(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 300, 400},
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 500},
		{Tag: "b", AverageRTT: 600},
		{Tag: "c", AverageRTT: 700},
		{Tag: "d", AverageRTT: 800},
		{Tag: "e", AverageRTT: 900},
	}
	expected := 3
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastLoadBaselines(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 400, 600},
			Expected:  0,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 100},
		{Tag: "b", AverageRTT: 200},
		{Tag: "c", AverageRTT: 300},
	}
	expected := 2
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastLoadBaselinesNoQualified(t *testing.T) {
	strategy := &LeastLoadStrategy{
		balancer: nil,
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 400, 600},
			Expected:  0,
		},
	}
	nodes := []*node{
		{Tag: "a", AverageRTT: 800},
		{Tag: "b", AverageRTT: 1000},
	}
	expected := 0
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
