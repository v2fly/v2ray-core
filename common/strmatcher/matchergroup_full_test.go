package strmatcher_test

import (
	"reflect"
	"testing"

	. "github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

func TestFullMatcherGroup(t *testing.T) {
	patterns := []struct {
		Pattern string
		Value   uint32
	}{
		{
			Pattern: "v2fly.org",
			Value:   1,
		},
		{
			Pattern: "google.com",
			Value:   2,
		},
		{
			Pattern: "x.a.com",
			Value:   3,
		},
		{
			Pattern: "x.y.com",
			Value:   4,
		},
		{
			Pattern: "x.y.com",
			Value:   6,
		},
	}
	testCases := []struct {
		Domain string
		Result []uint32
	}{
		{
			Domain: "v2fly.org",
			Result: []uint32{1},
		},
		{
			Domain: "y.com",
			Result: nil,
		},
		{
			Domain: "x.y.com",
			Result: []uint32{4, 6},
		},
	}
	g := NewFullMatcherGroup()
	for _, pattern := range patterns {
		AddMatcherToGroup(g, FullMatcher(pattern.Pattern), pattern.Value)
	}
	for _, testCase := range testCases {
		r := g.Match(testCase.Domain)
		if !reflect.DeepEqual(r, testCase.Result) {
			t.Error("Failed to match domain: ", testCase.Domain, ", expect ", testCase.Result, ", but got ", r)
		}
	}
}

func TestEmptyFullMatcherGroup(t *testing.T) {
	g := NewFullMatcherGroup()
	r := g.Match("v2fly.org")
	if len(r) != 0 {
		t.Error("Expect [], but ", r)
	}
}
