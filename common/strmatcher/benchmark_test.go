package strmatcher_test

import (
	"strconv"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common"
	. "github.com/v2fly/v2ray-core/v5/common/strmatcher"
)

func BenchmarkDomainMatcherGroup(b *testing.B) {
	g := new(DomainMatcherGroup)

	for i := 1; i <= 1024; i++ {
		g.Add(strconv.Itoa(i)+".v2fly.org", uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkFullMatcherGroup(b *testing.B) {
	g := new(FullMatcherGroup)

	for i := 1; i <= 1024; i++ {
		g.Add(strconv.Itoa(i)+".v2fly.org", uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkMarchGroup(b *testing.B) {
	g := new(MatcherGroup)
	for i := 1; i <= 1024; i++ {
		m, err := Domain.New(strconv.Itoa(i) + ".v2fly.org")
		common.Must(err)
		g.Add(m)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}
