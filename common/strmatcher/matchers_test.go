package strmatcher_test

import (
	"testing"

	"v2ray.com/core/common"
	. "v2ray.com/core/common/strmatcher"
)

func TestMatcher(t *testing.T) {
	cases := []struct {
		pattern string
		mType   Type
		input   string
		output  bool
	}{
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "www.v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "www.v3ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "2ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "xv2ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Full,
			input:   "v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Full,
			input:   "xv2ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Regex,
			input:   "v2rayxcom",
			output:  true,
		},
	}
	for _, test := range cases {
		matcher, err := test.mType.New(test.pattern)
		common.Must(err)
		if m := matcher.Match(test.input); m != test.output {
			t.Error("unexpected output: ", m, " for test case ", test)
		}
	}
}
func TestACAutomaton(t *testing.T) {
	cases1 := []struct {
		pattern string
		mType   Type
		input   string
		output  bool
	}{
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "www.v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "www.v3ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "2ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Domain,
			input:   "xv2ray.com",
			output:  false,
		},
		{
			pattern: "v2ray.com",
			mType:   Full,
			input:   "v2ray.com",
			output:  true,
		},
		{
			pattern: "v2ray.com",
			mType:   Full,
			input:   "xv2ray.com",
			output:  false,
		},
	}
	for _, test := range cases1 {
		var ac = NewACAutomaton()
		ac.Add(test.pattern, test.mType)
		ac.Build()
		if m := ac.Match(test.input); m != test.output {
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
		var ac = NewACAutomaton()
		for _, test := range cases2Input {
			ac.Add(test.pattern, test.mType)
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
			if m := ac.Match(test.pattern); m != test.res {
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
		var ac = NewACAutomaton()
		for _, test := range cases3Input {
			ac.Add(test.pattern, test.mType)
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
			if m := ac.Match(test.pattern); m != test.res {
				t.Error("unexpected output: ", m, " for test case ", test)
			}
		}

	}
}
