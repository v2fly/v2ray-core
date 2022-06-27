package strmatcher

import (
	"container/list"
)

const (
	acValidCharCount = 38 // aA-zZ (26), 0-9 (10), - (1), . (1)
	acMatchTypeCount = 3  // Full, Domain and Substr
)

type acEdge byte

const (
	acTrieEdge acEdge = 1
	acFailEdge acEdge = 0
)

type acNode struct {
	next  [acValidCharCount]uint32 // EdgeIdx -> Next NodeIdx (Next trie node or fail node)
	edge  [acValidCharCount]acEdge // EdgeIdx -> Trie Edge / Fail Edge
	fail  uint32                   // NodeIdx of *next matched* Substr Pattern on its fail path
	match uint32                   // MatchIdx of matchers registered on this node, 0 indicates no match
} // Sizeof acNode: (4+1)*acValidCharCount + <padding> + 4 + 4

type acValue [acMatchTypeCount][]uint32 // MatcherType -> Registered Matcher Values

// ACAutoMationMatcherGroup is an implementation of MatcherGroup.
// It uses an AC Automata to provide support for Full, Domain and Substr matcher. Trie node is char based.
//
// NOTICE: ACAutomatonMatcherGroup currently uses a restricted charset (LDH Subset),
// upstream should manually in a way to ensure all patterns and inputs passed to it to be in this charset.
type ACAutomatonMatcherGroup struct {
	nodes  []acNode  // NodeIdx -> acNode
	values []acValue // MatchIdx -> acValue
}

func NewACAutomatonMatcherGroup() *ACAutomatonMatcherGroup {
	ac := new(ACAutomatonMatcherGroup)
	ac.addNode()       // Create root node (NodeIdx 0)
	ac.addMatchEntry() // Create sentinel match entry (MatchIdx 0)
	return ac
}

// AddFullMatcher implements MatcherGroupForFull.AddFullMatcher.
func (ac *ACAutomatonMatcherGroup) AddFullMatcher(matcher FullMatcher, value uint32) {
	ac.addPattern(0, matcher.Pattern(), matcher.Type(), value)
}

// AddDomainMatcher implements MatcherGroupForDomain.AddDomainMatcher.
func (ac *ACAutomatonMatcherGroup) AddDomainMatcher(matcher DomainMatcher, value uint32) {
	node := ac.addPattern(0, matcher.Pattern(), matcher.Type(), value) // For full domain match
	ac.addPattern(node, ".", matcher.Type(), value)                    // For partial domain match
}

// AddSubstrMatcher implements MatcherGroupForSubstr.AddSubstrMatcher.
func (ac *ACAutomatonMatcherGroup) AddSubstrMatcher(matcher SubstrMatcher, value uint32) {
	ac.addPattern(0, matcher.Pattern(), matcher.Type(), value)
}

func (ac *ACAutomatonMatcherGroup) addPattern(nodeIdx uint32, pattern string, matcherType Type, value uint32) uint32 {
	node := &ac.nodes[nodeIdx]
	for i := len(pattern) - 1; i >= 0; i-- {
		edgeIdx := acCharset[pattern[i]]
		nextIdx := node.next[edgeIdx]
		if nextIdx == 0 { // Add new Trie Edge
			nextIdx = ac.addNode()
			ac.nodes[nodeIdx].next[edgeIdx] = nextIdx
			ac.nodes[nodeIdx].edge[edgeIdx] = acTrieEdge
		}
		nodeIdx = nextIdx
		node = &ac.nodes[nodeIdx]
	}
	if node.match == 0 { // Add new match entry
		node.match = ac.addMatchEntry()
	}
	ac.values[node.match][matcherType] = append(ac.values[node.match][matcherType], value)
	return nodeIdx
}

func (ac *ACAutomatonMatcherGroup) addNode() uint32 {
	ac.nodes = append(ac.nodes, acNode{})
	return uint32(len(ac.nodes) - 1)
}

func (ac *ACAutomatonMatcherGroup) addMatchEntry() uint32 {
	ac.values = append(ac.values, acValue{})
	return uint32(len(ac.values) - 1)
}

func (ac *ACAutomatonMatcherGroup) Build() error {
	fail := make([]uint32, len(ac.nodes))
	queue := list.New()
	for edgeIdx := 0; edgeIdx < acValidCharCount; edgeIdx++ {
		if nextIdx := ac.nodes[0].next[edgeIdx]; nextIdx != 0 {
			queue.PushBack(nextIdx)
		}
	}
	for {
		front := queue.Front()
		if front == nil {
			break
		}
		queue.Remove(front)
		nodeIdx := front.Value.(uint32)
		node := &ac.nodes[nodeIdx]           // Current node
		failNode := &ac.nodes[fail[nodeIdx]] // Fail node of currrent node
		for edgeIdx := 0; edgeIdx < acValidCharCount; edgeIdx++ {
			nodeIdx := node.next[edgeIdx]     // Next node through trie edge
			failIdx := failNode.next[edgeIdx] // Next node through fail edge
			if nodeIdx != 0 {
				queue.PushBack(nodeIdx)
				fail[nodeIdx] = failIdx
				if match := ac.nodes[failIdx].match; match != 0 && len(ac.values[match][Substr]) > 0 { // Fail node is a Substr match node
					ac.nodes[nodeIdx].fail = failIdx
				} else { // Use path compression to reduce fail path to only contain match nodes
					ac.nodes[nodeIdx].fail = ac.nodes[failIdx].fail
				}
			} else { // Add new fail edge
				node.next[edgeIdx] = failIdx
				node.edge[edgeIdx] = acFailEdge
			}
		}
	}
	return nil
}

// Match implements MatcherGroup.Match.
func (ac *ACAutomatonMatcherGroup) Match(input string) []uint32 {
	var suffixMatches [][]uint32
	var substrMatches [][]uint32
	fullMatch := true    // fullMatch indicates no fail edge traversed so far.
	node := &ac.nodes[0] // start from root node.
	// 1. the match string is all through trie edge. FULL MATCH or DOMAIN
	// 2. the match string is through a fail edge. NOT FULL MATCH
	// 2.1 Through a fail edge, but there exists a valid node. SUBSTR
	for i := len(input) - 1; i >= 0; i-- {
		edge := acCharset[input[i]]
		fullMatch = fullMatch && (node.edge[edge] == acTrieEdge)
		node = &ac.nodes[node.next[edge]] // Advance to next node
		// When entering a new node, traverse the fail path to find all possible Substr patterns:
		//   1. The fail path is compressed to only contains match nodes and root node (for terminate condition).
		//   2. node.fail != 0 is added here for better performance (as shown by benchmark), possibly it helps branch prediction.
		if node.fail != 0 {
			for failIdx, failNode := node.fail, &ac.nodes[node.fail]; failIdx != 0; failIdx, failNode = failNode.fail, &ac.nodes[failIdx] {
				substrMatches = append(substrMatches, ac.values[failNode.match][Substr])
			}
		}
		// When entering a new node, check whether this node is a match.
		// For Substr matchers:
		//   1. Matched in any situation, whether a failNode edge is traversed or not.
		// For Domain matchers:
		//   1. Should not traverse any fail edge (fullMatch).
		//   2. Only check on dot separator (input[i] == '.').
		if node.match != 0 {
			values := ac.values[node.match]
			if len(values[Substr]) > 0 {
				substrMatches = append(substrMatches, values[Substr])
			}
			if fullMatch && input[i] == '.' && len(values[Domain]) > 0 {
				suffixMatches = append(suffixMatches, values[Domain])
			}
		}
	}
	// At the end of input, check if the whole string matches a pattern.
	// For Domain matchers:
	//   1. Exact match on Domain Matcher works like Full Match. e.g. foo.com is a full match for domain:foo.com.
	// For Full matchers:
	//   1. Only when no fail edge is traversed (fullMatch).
	//   2. Takes the highest priority (added at last).
	if fullMatch && node.match != 0 {
		values := ac.values[node.match]
		if len(values[Domain]) > 0 {
			suffixMatches = append(suffixMatches, values[Domain])
		}
		if len(values[Full]) > 0 {
			suffixMatches = append(suffixMatches, values[Full])
		}
	}
	switch matches := append(substrMatches, suffixMatches...); len(matches) { // nolint: gocritic
	case 0:
		return nil
	case 1:
		return matches[0]
	default:
		result := []uint32{}
		for i := len(matches) - 1; i >= 0; i-- {
			result = append(result, matches[i]...)
		}
		return result
	}
}

// MatchAny implements MatcherGroup.MatchAny.
func (ac *ACAutomatonMatcherGroup) MatchAny(input string) bool {
	fullMatch := true
	node := &ac.nodes[0]
	for i := len(input) - 1; i >= 0; i-- {
		edge := acCharset[input[i]]
		fullMatch = fullMatch && (node.edge[edge] == acTrieEdge)
		node = &ac.nodes[node.next[edge]]
		if node.fail != 0 { // There is a match on this node's fail path
			return true
		}
		if node.match != 0 { // There is a match on this node
			values := ac.values[node.match]
			if len(values[Substr]) > 0 { // Substr match succeeds unconditionally
				return true
			}
			if fullMatch && input[i] == '.' && len(values[Domain]) > 0 { // Domain match only succeeds with dot separator on trie path
				return true
			}
		}
	}
	return fullMatch && node.match != 0 // At the end of input, Domain and Full match will succeed if no fail edge is traversed
}

// Letter-Digit-Hyphen (LDH) subset (https://tools.ietf.org/html/rfc952):
//   * Letters A to Z (no distinction is made between uppercase and lowercase)
//   * Digits 0 to 9
//   * Hyphens(-) and Periods(.)
//
// If for future the strmatcher are used for other scenarios than domain,
// we could add a new Charset interface to represent variable charsets.
var acCharset = []int{
	'A': 0,
	'a': 0,
	'B': 1,
	'b': 1,
	'C': 2,
	'c': 2,
	'D': 3,
	'd': 3,
	'E': 4,
	'e': 4,
	'F': 5,
	'f': 5,
	'G': 6,
	'g': 6,
	'H': 7,
	'h': 7,
	'I': 8,
	'i': 8,
	'J': 9,
	'j': 9,
	'K': 10,
	'k': 10,
	'L': 11,
	'l': 11,
	'M': 12,
	'm': 12,
	'N': 13,
	'n': 13,
	'O': 14,
	'o': 14,
	'P': 15,
	'p': 15,
	'Q': 16,
	'q': 16,
	'R': 17,
	'r': 17,
	'S': 18,
	's': 18,
	'T': 19,
	't': 19,
	'U': 20,
	'u': 20,
	'V': 21,
	'v': 21,
	'W': 22,
	'w': 22,
	'X': 23,
	'x': 23,
	'Y': 24,
	'y': 24,
	'Z': 25,
	'z': 25,
	'-': 26,
	'.': 27,
	'0': 28,
	'1': 29,
	'2': 30,
	'3': 31,
	'4': 32,
	'5': 33,
	'6': 34,
	'7': 35,
	'8': 36,
	'9': 37,
}
