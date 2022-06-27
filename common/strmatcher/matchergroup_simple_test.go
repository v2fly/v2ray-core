package strmatcher_test

import (
	"reflect"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common"
	. "github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

func TestSimpleMatcherGroup(t *testing.T) {
	patterns := []struct {
		pattern string
		mType   Type
	}{
		{
			pattern: "v2fly.org",
			mType:   Domain,
		},
		{
			pattern: "v2fly.org",
			mType:   Full,
		},
		{
			pattern: "v2fly.org",
			mType:   Regex,
		},
	}
	cases := []struct {
		input  string
		output []uint32
	}{
		{
			input:  "www.v2fly.org",
			output: []uint32{0, 2},
		},
		{
			input:  "v2fly.org",
			output: []uint32{0, 1, 2},
		},
		{
			input:  "www.v3fly.org",
			output: []uint32{},
		},
		{
			input:  "2fly.org",
			output: []uint32{},
		},
		{
			input:  "xv2fly.org",
			output: []uint32{2},
		},
		{
			input:  "v2flyxorg",
			output: []uint32{2},
		},
	}
	matcherGroup := &SimpleMatcherGroup{}
	for id, entry := range patterns {
		matcher, err := entry.mType.New(entry.pattern)
		common.Must(err)
		common.Must(AddMatcherToGroup(matcherGroup, matcher, uint32(id)))
	}
	for _, test := range cases {
		if r := matcherGroup.Match(test.input); !reflect.DeepEqual(r, test.output) {
			t.Error("unexpected output: ", r, " for test case ", test)
		}
	}
}
