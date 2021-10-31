package strmatcher_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	. "github.com/v2fly/v2ray-core/v4/common/strmatcher"
)

func TestACAutomatonMatcherGroup(t *testing.T) {
	cases1 := []struct {
		pattern string
		mType   Type
		input   string
		output  bool
	}{
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "www.v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "www.v3fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "2fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Domain,
			input:   "xv2fly.org",
			output:  false,
		},
		{
			pattern: "v2fly.org",
			mType:   Full,
			input:   "v2fly.org",
			output:  true,
		},
		{
			pattern: "v2fly.org",
			mType:   Full,
			input:   "xv2fly.org",
			output:  false,
		},
	}
	for _, test := range cases1 {
		ac := NewACAutomatonMatcherGroup()
		matcher, err := test.mType.New(test.pattern)
		common.Must(err)
		common.Must(AddMatcherToGroup(ac, matcher, 0))
		ac.Build()
		if m := ac.MatchAny(test.input); m != test.output {
			t.Error("unexpected output: ", m, " for test case ", test)
		}
	}
	{
		cases2Input := []struct {
			pattern string
			mType   Type
		}{
			{
				pattern: "163.com",
				mType:   Domain,
			},
			{
				pattern: "m.126.com",
				mType:   Full,
			},
			{
				pattern: "3.com",
				mType:   Full,
			},
			{
				pattern: "google.com",
				mType:   Substr,
			},
			{
				pattern: "vgoogle.com",
				mType:   Substr,
			},
		}
		ac := NewACAutomatonMatcherGroup()
		for _, test := range cases2Input {
			matcher, err := test.mType.New(test.pattern)
			common.Must(err)
			common.Must(AddMatcherToGroup(ac, matcher, 0))
		}
		ac.Build()
		cases2Output := []struct {
			pattern string
			res     bool
		}{
			{
				pattern: "126.com",
				res:     false,
			},
			{
				pattern: "m.163.com",
				res:     true,
			},
			{
				pattern: "mm163.com",
				res:     false,
			},
			{
				pattern: "m.126.com",
				res:     true,
			},
			{
				pattern: "163.com",
				res:     true,
			},
			{
				pattern: "63.com",
				res:     false,
			},
			{
				pattern: "oogle.com",
				res:     false,
			},
			{
				pattern: "vvgoogle.com",
				res:     true,
			},
		}
		for _, test := range cases2Output {
			if m := ac.MatchAny(test.pattern); m != test.res {
				t.Error("unexpected output: ", m, " for test case ", test)
			}
		}
	}

	{
		cases3Input := []struct {
			pattern string
			mType   Type
		}{
			{
				pattern: "video.google.com",
				mType:   Domain,
			},
			{
				pattern: "gle.com",
				mType:   Domain,
			},
		}
		ac := NewACAutomatonMatcherGroup()
		for _, test := range cases3Input {
			matcher, err := test.mType.New(test.pattern)
			common.Must(err)
			common.Must(AddMatcherToGroup(ac, matcher, 0))
		}
		ac.Build()
		cases3Output := []struct {
			pattern string
			res     bool
		}{
			{
				pattern: "google.com",
				res:     false,
			},
		}
		for _, test := range cases3Output {
			if m := ac.MatchAny(test.pattern); m != test.res {
				t.Error("unexpected output: ", m, " for test case ", test)
			}
		}
	}
}
