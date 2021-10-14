package strmatcher_test

import (
	"strconv"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	. "github.com/v2fly/v2ray-core/v4/common/strmatcher"
)

// Benchmark Domain Matcher Groups

func BenchmarkSimpleMatcherGroupForDomain(b *testing.B) {
	g := new(SimpleMatcherGroup)

	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(g, DomainMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkDomainMatcherGroup(b *testing.B) {
	g := new(DomainMatcherGroup)

	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(g, DomainMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkACAutomatonMatcherGroupForDomain(b *testing.B) {
	ac := NewACAutomatonMatcherGroup()
	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(ac, DomainMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}
	ac.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.MatchAny("0.v2fly.org")
	}
}

func BenchmarkMphMatcherGroupForDomain(b *testing.B) {
	mph := NewMphMatcherGroup()
	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(mph, DomainMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}
	mph.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mph.MatchAny("0.v2fly.org")
	}
}

// Benchmark Full Matcher Groups

func BenchmarkSimpleMatcherGroupForFull(b *testing.B) {
	g := new(SimpleMatcherGroup)

	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(g, FullMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkFullMatcherGroup(b *testing.B) {
	g := new(FullMatcherGroup)

	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(g, FullMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkACAutomatonMatcherGroupForFull(b *testing.B) {
	ac := NewACAutomatonMatcherGroup()
	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(ac, FullMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}
	ac.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.MatchAny("0.v2fly.org")
	}
}

func BenchmarkMphMatcherGroupFull(b *testing.B) {
	mph := NewMphMatcherGroup()
	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(mph, FullMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}
	mph.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mph.MatchAny("0.v2fly.org")
	}
}

// Benchmark Substr Matcher Groups

func BenchmarkSimpleMatcherGroupForSubstr(b *testing.B) {
	g := new(SimpleMatcherGroup)

	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(g, SubstrMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Match("0.v2fly.org")
	}
}

func BenchmarkACAutomatonMatcherGroupForSubstr(b *testing.B) {
	ac := NewACAutomatonMatcherGroup()
	for i := 1; i <= 1024; i++ {
		AddMatcherToGroup(ac, SubstrMatcher(strconv.Itoa(i)+".v2fly.org"), uint32(i))
	}
	ac.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.MatchAny("0.v2fly.org")
	}
}

// Benchmark Index Matchers

func BenchmarkLinearIndexMatcher(b *testing.B) {
	g := new(LinearIndexMatcher)
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
