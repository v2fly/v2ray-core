package strmatcher

type FullMatcherGroup struct {
	matchers map[string][]uint32
}

func (g *FullMatcherGroup) AddFullMatcher(matcher FullMatcher, value uint32) {
	if g.matchers == nil {
		g.matchers = make(map[string][]uint32)
	}

	domain := matcher.Pattern()
	g.matchers[domain] = append(g.matchers[domain], value)
}

func (g *FullMatcherGroup) Match(str string) []uint32 {
	if g.matchers == nil {
		return nil
	}

	return g.matchers[str]
}
