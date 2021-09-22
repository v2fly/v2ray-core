package strmatcher

// LinearIndexMatcher is an implementation of IndexMatcher.
// Empty initialization works.
type LinearIndexMatcher struct {
	count         uint32
	fullMatcher   FullMatcherGroup
	domainMatcher DomainMatcherGroup
	otherMatchers []matcherEntry
}

// Add adds a new Matcher into the MatcherGroup, and returns its index. The index will never be 0.
func (g *LinearIndexMatcher) Add(m Matcher) uint32 {
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
func (g *LinearIndexMatcher) Match(pattern string) []uint32 {
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
func (g *LinearIndexMatcher) Size() uint32 {
	return g.count
}
