package strmatcher

import (
	"errors"
	"regexp"
	"strings"
)

// FullMatcher is an implementation of Matcher.
type FullMatcher string

func (FullMatcher) Type() Type {
	return Full
}

func (m FullMatcher) Pattern() string {
	return string(m)
}

func (m FullMatcher) String() string {
	return "full:" + m.Pattern()
}

func (m FullMatcher) Match(s string) bool {
	return string(m) == s
}

// DomainMatcher is an implementation of Matcher.
type DomainMatcher string

func (DomainMatcher) Type() Type {
	return Domain
}

func (m DomainMatcher) Pattern() string {
	return string(m)
}

func (m DomainMatcher) String() string {
	return "domain:" + m.Pattern()
}

func (m DomainMatcher) Match(s string) bool {
	pattern := m.Pattern()
	if !strings.HasSuffix(s, pattern) {
		return false
	}
	return len(s) == len(pattern) || s[len(s)-len(pattern)-1] == '.'
}

// SubstrMatcher is an implementation of Matcher.
type SubstrMatcher string

func (SubstrMatcher) Type() Type {
	return Substr
}

func (m SubstrMatcher) Pattern() string {
	return string(m)
}

func (m SubstrMatcher) String() string {
	return "keyword:" + m.Pattern()
}

func (m SubstrMatcher) Match(s string) bool {
	return strings.Contains(s, m.Pattern())
}

// RegexMatcher is an implementation of Matcher.
type RegexMatcher struct {
	pattern *regexp.Regexp
}

func (*RegexMatcher) Type() Type {
	return Regex
}

func (m *RegexMatcher) Pattern() string {
	return m.pattern.String()
}

func (m *RegexMatcher) String() string {
	return "regexp:" + m.Pattern()
}

func (m *RegexMatcher) Match(s string) bool {
	return m.pattern.MatchString(s)
}

// New creates a new Matcher based on the given pattern.
func (t Type) New(pattern string) (Matcher, error) {
	switch t {
	case Full:
		return FullMatcher(pattern), nil
	case Substr:
		return SubstrMatcher(pattern), nil
	case Domain:
		return DomainMatcher(pattern), nil
	case Regex: // 1. regex matching is case-sensitive
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return &RegexMatcher{pattern: regex}, nil
	default:
		panic("Unknown type")
	}
}

// MatcherGroupForAll is an interface indicating a MatcherGroup could accept all types of matchers.
type MatcherGroupForAll interface {
	AddMatcher(matcher Matcher, value uint32)
}

// MatcherGroupForFull is an interface indicating a MatcherGroup could accept FullMatchers.
type MatcherGroupForFull interface {
	AddFullMatcher(matcher FullMatcher, value uint32)
}

// MatcherGroupForDomain is an interface indicating a MatcherGroup could accept DomainMatchers.
type MatcherGroupForDomain interface {
	AddDomainMatcher(matcher DomainMatcher, value uint32)
}

// MatcherGroupForSubstr is an interface indicating a MatcherGroup could accept SubstrMatchers.
type MatcherGroupForSubstr interface {
	AddSubstrMatcher(matcher SubstrMatcher, value uint32)
}

// MatcherGroupForRegex is an interface indicating a MatcherGroup could accept RegexMatchers.
type MatcherGroupForRegex interface {
	AddRegexMatcher(matcher *RegexMatcher, value uint32)
}

// AddMatcherGroup is a helper function to try to add a Matcher to any kind of MatcherGroup.
// It returns error if the MatcherGroup does not accept the provided Matcher's type.
// This function is provided to help writing code to test a MatcherGroup.
func AddMatcherToGroup(g MatcherGroup, matcher Matcher, value uint32) error {
	if g, ok := g.(MatcherGroupForAll); ok {
		g.AddMatcher(matcher, value)
		return nil
	}
	switch matcher := matcher.(type) {
	case FullMatcher:
		if g, ok := g.(MatcherGroupForFull); ok {
			g.AddFullMatcher(matcher, value)
			return nil
		}
	case DomainMatcher:
		if g, ok := g.(MatcherGroupForDomain); ok {
			g.AddDomainMatcher(matcher, value)
			return nil
		}
	case SubstrMatcher:
		if g, ok := g.(MatcherGroupForSubstr); ok {
			g.AddSubstrMatcher(matcher, value)
			return nil
		}
	case *RegexMatcher:
		if g, ok := g.(MatcherGroupForRegex); ok {
			g.AddRegexMatcher(matcher, value)
			return nil
		}
	}
	return errors.New("cannot add matcher to matcher group")
}
