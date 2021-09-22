package strmatcher

import (
	"regexp"
	"strings"
)

type fullMatcher string

func (m fullMatcher) Match(s string) bool {
	return string(m) == s
}

func (m fullMatcher) String() string {
	return "full:" + string(m)
}

type substrMatcher string

func (m substrMatcher) Match(s string) bool {
	return strings.Contains(s, string(m))
}

func (m substrMatcher) String() string {
	return "keyword:" + string(m)
}

type domainMatcher string

func (m domainMatcher) Match(s string) bool {
	pattern := string(m)
	if !strings.HasSuffix(s, pattern) {
		return false
	}
	return len(s) == len(pattern) || s[len(s)-len(pattern)-1] == '.'
}

func (m domainMatcher) String() string {
	return "domain:" + string(m)
}

type regexMatcher struct {
	pattern *regexp.Regexp
}

func (m *regexMatcher) Match(s string) bool {
	return m.pattern.MatchString(s)
}

func (m *regexMatcher) String() string {
	return "regexp:" + m.pattern.String()
}

// New creates a new Matcher based on the given pattern.
func (t Type) New(pattern string) (Matcher, error) {
	// 1. regex matching is case-sensitive
	switch t {
	case Full:
		return fullMatcher(pattern), nil
	case Substr:
		return substrMatcher(pattern), nil
	case Domain:
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
