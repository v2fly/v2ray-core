package strmatcher

import (
	"regexp"
	"strings"
)

// FullMatcher is an implementation of Matcher.
type FullMatcher string

func (m FullMatcher) Type() Type {
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

// SubstrMatcher is an implementation of Matcher.
type SubstrMatcher string

func (m SubstrMatcher) Type() Type {
	return Substr
}

func (m SubstrMatcher) Pattern() string {
	return string(m)
}

func (m SubstrMatcher) String() string {
	return "keyword:" + m.Pattern()
}

func (m SubstrMatcher) Match(s string) bool {
	return strings.Contains(s, string(m))
}

// DomainMatcher is an implementation of Matcher.
type DomainMatcher string

func (m DomainMatcher) Type() Type {
	return Substr
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

// RegexMatcher is an implementation of Matcher.
type RegexMatcher struct {
	pattern *regexp.Regexp
}

func (m *RegexMatcher) Type() Type {
	return Substr
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
		if regex, err := regexp.Compile(pattern); err == nil {
			return &RegexMatcher{pattern: regex}, nil
		} else {
			return nil, err
		}
	default:
		panic("Unknown type")
	}
}
