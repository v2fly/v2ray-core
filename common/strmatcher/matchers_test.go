package strmatcher_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	. "github.com/v2fly/v2ray-core/v4/common/strmatcher"
)

func TestMatcher(t *testing.T) {
	cases := []struct {
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
		{
			pattern: "v2fly.org",
			mType:   Regex,
			input:   "v2flyxorg",
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
