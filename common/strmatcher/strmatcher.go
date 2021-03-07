package strmatcher

import (
	"regexp"
	"strings"
)

// PrimeRK is the prime base used in Rabin-Karp algorithm.
const PrimeRK = 16777619

// Matcher is the interface to determine a string matches a pattern.
type Matcher interface {
	// Match returns true if the given string matches a predefined pattern.
	Match(string) bool
	String() string
}

// Type is the type of the matcher.
type Type byte

// RollingHashType is the type of rolling hash used by Rabin-Karp.
type RollingHashType uint32

const (
	// Full is the type of matcher that the input string must exactly equal to the pattern.
	Full Type = iota
	// Substr is the type of matcher that the input string must contain the pattern as a sub-string.
	Substr
	// Domain is the type of matcher that the input string must be a sub-domain or itself of the pattern.
	Domain
	// Regex is the type of matcher that the input string must matches the regular-expression pattern.
	Regex
)

// New creates a new Matcher based on the given pattern.
func (t Type) New(pattern string) (Matcher, error) {
	// 1. regex matching is case-sensitive
	switch t {
	case Full:
		pattern = strings.ToLower(pattern)
		return fullMatcher(pattern), nil
	case Substr:
		pattern = strings.ToLower(pattern)
		return substrMatcher(pattern), nil
	case Domain:
		pattern = strings.ToLower(pattern)
		return domainMatcher(pattern), nil
	case Regex:
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return &regexMatcher{
			pattern: r,
		}, nil
	default:
		panic("Unknown type")
	}
}

// IndexMatcher is the interface for matching with a group of matchers.
type IndexMatcher interface {
	// Match returns the index of a matcher that matches the input. It returns empty array if no such matcher exists.
	Match(input string) []uint32
}

type matcherEntry struct {
	m  Matcher
	id uint32
}

// The HybridMatcherGroup is divided into three parts:
// 1. `full` and `domain` patterns are matched by Rabin-Karp algorithm;
// 2. `substr` patterns are matched by ac automaton;
// 3. `regex` patterns are matched with the regex library.
type HybridMatcherGroup struct {
	count          uint32
	ac             *ACAutomaton
	rollingHashMap map[RollingHashType][]string
	otherMatchers  []matcherEntry
}

func NewHybridMatcherGroup() *HybridMatcherGroup {
	var g = new(HybridMatcherGroup)
	g.count = 1
	g.rollingHashMap = map[RollingHashType][]string{}
	return g
}

func contains(chain []string, pattern string) bool {
	for _, v := range chain {
		if v == pattern {
			return true
		}
	}
	return false
}

func (g *HybridMatcherGroup) insert(h RollingHashType, pattern string) {
	if chain, ok := g.rollingHashMap[h]; ok {
		if !contains(chain, pattern) {
			chain = append(chain, pattern) // hash collision, open hashing
		}
	} else {
		g.rollingHashMap[h] = []string{pattern}
	}
}

// Add `full` or `domain` pattern to hashmap
func (g *HybridMatcherGroup) AddFullOrDomainPattern(pattern string, t Type) {
	h := RollingHashType(0)
	for i := len(pattern) - 1; i >= 0; i-- {
		h = h*PrimeRK + RollingHashType(pattern[i])
	}
	switch t {
	case Full:
		g.insert(h, pattern)
	case Domain:
		g.insert(h, pattern)
		g.insert(h*PrimeRK+RollingHashType('.'), "."+pattern)
	default:
	}
}

func (g *HybridMatcherGroup) AddPattern(pattern string, t Type) (uint32, error) {
	// 1. AC automaton is a case-insensitive matcher.
	// 2. regex matching is case-sensitive.
	// 3. Rabin-Karp algorithm is case-sensitive. (Full or Domain)
	switch t {
	case Substr:
		if g.ac == nil {
			g.ac = NewACAutomaton()
		}
		g.ac.Add(pattern, t)
	case Full, Domain:
		pattern = strings.ToLower(pattern)
		g.AddFullOrDomainPattern(pattern, t)
	case Regex:
		g.count++
		r, err := regexp.Compile(pattern)
		if err != nil {
			return 0, err
		}
		g.otherMatchers = append(g.otherMatchers, matcherEntry{
			m:  &regexMatcher{pattern: r},
			id: g.count,
		})
	default:
		panic("Unknown type")
	}
	return g.count, nil
}

func (g *HybridMatcherGroup) Build() {
	if g.ac != nil {
		g.ac.Build()
	}
}

// Match implements IndexMatcher.Match.
func (g *HybridMatcherGroup) Match(pattern string) []uint32 {
	result := []uint32{}
	hash := RollingHashType(0)
	for i := len(pattern) - 1; i >= 0; i-- {
		hash = hash*PrimeRK + RollingHashType(pattern[i])
		if pattern[i] == '.' {
			if chain, ok := g.rollingHashMap[hash]; ok && contains(chain, pattern[i:]) {
				result = append(result, 1)
				return result
			}
		}
	}
	if chain, ok := g.rollingHashMap[hash]; ok && contains(chain, pattern) {
		result = append(result, 1)
		return result
	}
	if g.ac != nil && g.ac.Match(pattern) {
		result = append(result, 1)
		return result
	}
	for _, e := range g.otherMatchers {
		if e.m.Match(pattern) {
			result = append(result, e.id)
			return result
		}
	}
	return nil
}

// MatcherGroup is an implementation of IndexMatcher.
// Empty initialization works.
type MatcherGroup struct {
	count         uint32
	fullMatcher   FullMatcherGroup
	domainMatcher DomainMatcherGroup
	otherMatchers []matcherEntry
}

// Add adds a new Matcher into the MatcherGroup, and returns its index. The index will never be 0.
func (g *MatcherGroup) Add(m Matcher) uint32 {
	g.count++
	c := g.count

	switch tm := m.(type) {
	case fullMatcher:
		g.fullMatcher.addMatcher(tm, c)
	case domainMatcher:
		g.domainMatcher.addMatcher(tm, c)
	default:
		g.otherMatchers = append(g.otherMatchers, matcherEntry{
			m:  m,
			id: c,
		})
	}

	return c
}

// Match implements IndexMatcher.Match.
func (g *MatcherGroup) Match(pattern string) []uint32 {
	result := []uint32{}
	result = append(result, g.fullMatcher.Match(pattern)...)
	result = append(result, g.domainMatcher.Match(pattern)...)
	for _, e := range g.otherMatchers {
		if e.m.Match(pattern) {
			result = append(result, e.id)
		}
	}
	return result
}

// Size returns the number of matchers in the MatcherGroup.
func (g *MatcherGroup) Size() uint32 {
	return g.count
}
