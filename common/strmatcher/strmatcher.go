package strmatcher

// Type is the type of the matcher.
type Type byte

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

// Matcher is the interface to determine a string matches a pattern.
//  * This is a basic matcher to represent a certain kind of match semantic(full, substr, domain or regex).
type Matcher interface {
	// Type returns the matcher's type.
	Type() Type

	// Pattern returns the matcher's raw string representation
	Pattern() string

	// String returns a string representation of the matcher containing its type and pattern.
	String() string

	// Match returns true if the given string matches a predefined pattern.
	//  * This method is seldom used for performance reason
	//    and is generally taken over by their corresponding MatcherGroup.
	Match(str string) bool
}

// MatcherGroup is an advanced type of matcher to accept a bunch of basic Matchers (of certain type, not all matcher types).
// For example:
//   * FullMatcherGroup accepts FullMatcher and uses a hash table to facilitate lookup.
//   * DomainMatcherGroup accepts DomainMatcher and uses a trie to optimize both memory consumption and lookup speed.
type MatcherGroup interface {
	// Match returns all matched matchers with their corresponding values.
	Match(input string) []uint32
	// MatchAny returns true as soon as one matching matcher is found.
	MatchAny(input string) bool
}

// IndexMatcher is the interface for matching with a group of matchers.
type IndexMatcher interface {
	Add(matcher Matcher) uint32
	// Match returns the index of a matcher that matches the input. It returns empty array if no such matcher exists.
	Match(input string) []uint32
	// Size() uint32
}
