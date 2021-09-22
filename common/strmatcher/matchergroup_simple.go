package strmatcher

type matcherEntry struct {
	m  Matcher
	id uint32
}

type SimpleMatcherGroup struct {
	matchers []matcherEntry
}
