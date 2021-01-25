package router

import (
	"testing"
)

func TestSelectLeastExpected(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: nil,
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 100},
		{Tag: "b", RTTDeviation: 200},
		{Tag: "c", RTTDeviation: 300},
		{Tag: "d", RTTDeviation: 350},
	}
	expected := 3
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpected2(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: nil,
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 100},
		{Tag: "b", RTTDeviation: 200},
	}
	expected := 2
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpectedAndBaselines(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 300, 400},
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 100},
		{Tag: "b", RTTDeviation: 200},
		{Tag: "c", RTTDeviation: 250},
		{Tag: "d", RTTDeviation: 300},
		{Tag: "e", RTTDeviation: 310},
	}
	expected := 4
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastExpectedAndBaselines2(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 300, 400},
			Expected:  3,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 500},
		{Tag: "b", RTTDeviation: 600},
		{Tag: "c", RTTDeviation: 700},
		{Tag: "d", RTTDeviation: 800},
		{Tag: "e", RTTDeviation: 900},
	}
	expected := 3
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastLoadBaselines(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 400, 600},
			Expected:  0,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 100},
		{Tag: "b", RTTDeviation: 200},
		{Tag: "c", RTTDeviation: 300},
	}
	expected := 2
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
func TestSelectLeastLoadBaselinesNoQualified(t *testing.T) {
	strategy := &LeastLoadStrategy{
		settings: &StrategyLeastLoadConfig{
			Baselines: []int64{200, 400, 600},
			Expected:  0,
		},
	}
	nodes := []*node{
		{Tag: "a", RTTDeviation: 800},
		{Tag: "b", RTTDeviation: 1000},
	}
	expected := 0
	ns := strategy.selectLeastLoad(nodes)
	if len(ns) != expected {
		t.Errorf("expected: %v, actual: %v", expected, len(ns))
	}
}
