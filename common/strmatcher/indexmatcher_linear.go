package strmatcher

// LinearIndexMatcher is an implementation of IndexMatcher.
// Empty initialization works.
type LinearIndexMatcher struct {
	count         uint32
	fullMatcher   FullMatcherGroup
	domainMatcher DomainMatcherGroup
	otherMatchers SimpleMatcherGroup
}

// Add implements IndexMatcher.Add.
func (g *LinearIndexMatcher) Add(matcher Matcher) uint32 {
	g.count++
	index := g.count

	switch matcher := matcher.(type) {
	case FullMatcher:
		g.fullMatcher.AddFullMatcher(matcher, index)
	case DomainMatcher:
		g.domainMatcher.AddDomainMatcher(matcher, index)
	default:
		g.otherMatchers.AddMatcher(matcher, index)
	}

	return index
}

// Match implements IndexMatcher.Match.
func (g *LinearIndexMatcher) Match(input string) []uint32 {
	result := []uint32{}
	result = append(result, g.fullMatcher.Match(input)...)
	result = append(result, g.domainMatcher.Match(input)...)
	result = append(result, g.otherMatchers.Match(input)...)
	return result
}

// Size implements IndexMatcher.Size.
func (g *LinearIndexMatcher) Size() uint32 {
	return g.count
}
