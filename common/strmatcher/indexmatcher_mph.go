package strmatcher

// A MphIndexMatcher is divided into three parts:
// 1. `full` and `domain` patterns are matched by Rabin-Karp algorithm and minimal perfect hash table;
// 2. `substr` patterns are matched by ac automaton;
// 3. `regex` patterns are matched with the regex library.
type MphIndexMatcher struct {
	count uint32
	mph   *MphMatcherGroup
	ac    *ACAutomatonMatcherGroup
	regex SimpleMatcherGroup
}

func NewMphIndexMatcher() *MphIndexMatcher {
	return &MphIndexMatcher{
		mph:   nil,
		ac:    nil,
		regex: SimpleMatcherGroup{},
	}
}

// Add implements IndexMatcher.Add.
func (g *MphIndexMatcher) Add(matcher Matcher) uint32 {
	g.count++
	index := g.count

	switch matcher := matcher.(type) {
	case FullMatcher:
		if g.mph == nil {
			g.mph = NewMphMatcherGroup()
		}
		g.mph.AddFullMatcher(matcher)
	case DomainMatcher:
		if g.mph == nil {
			g.mph = NewMphMatcherGroup()
		}
		g.mph.AddDomainMatcher(matcher)
	case SubstrMatcher:
		if g.ac == nil {
			g.ac = NewACAutomatonMatcherGroup()
		}
		g.ac.Add(string(matcher), Substr)
	case *RegexMatcher:
		g.regex.AddMatcher(matcher, index)
	}

	return index
}

func (g *MphIndexMatcher) Build() {
	if g.mph != nil {
		g.mph.Build()
	}
	if g.ac != nil {
		g.ac.Build()
	}
}

// Match implements IndexMatcher.Match.
func (g *MphIndexMatcher) Match(pattern string) []uint32 {
	result := []uint32{}
	if len(g.mph.Match(pattern)) > 0 {
		result = append(result, 1)
		return result
	}
	if g.ac != nil && g.ac.Match(pattern) {
		result = append(result, 1)
		return result
	}
	result = append(result, g.regex.Match(pattern)...)
	return result
}
