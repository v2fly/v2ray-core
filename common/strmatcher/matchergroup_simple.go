package strmatcher

type matcherEntry struct {
	matcher Matcher
	value   uint32
}

type SimpleMatcherGroup struct {
	matchers []matcherEntry
}

func (g *SimpleMatcherGroup) AddMatcher(matcher Matcher, value uint32) {
	g.matchers = append(g.matchers, matcherEntry{
		matcher: matcher,
		value:   value,
	})
}

func (g *SimpleMatcherGroup) Match(str string) []uint32 {
	result := []uint32{}
	for _, e := range g.matchers {
		if e.matcher.Match(str) {
			result = append(result, e.value)
		}
	}
	return result
}
