package strmatcher_test

import (
	"reflect"
	"testing"

	. "github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

func TestFullMatcherGroup(t *testing.T) {
	g := new(FullMatcherGroup)
	g.Add("v2fly.org", 1)
	g.Add("google.com", 2)
	g.Add("x.a.com", 3)
	g.Add("x.y.com", 4)
	g.Add("x.y.com", 6)

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

	for _, testCase := range testCases {
		r := g.Match(testCase.Domain)
		if !reflect.DeepEqual(r, testCase.Result) {
			t.Error("Failed to match domain: ", testCase.Domain, ", expect ", testCase.Result, ", but got ", r)
		}
	}
}

func TestEmptyFullMatcherGroup(t *testing.T) {
	g := new(FullMatcherGroup)
	r := g.Match("v2fly.org")
	if len(r) != 0 {
		t.Error("Expect [], but ", r)
	}
}
