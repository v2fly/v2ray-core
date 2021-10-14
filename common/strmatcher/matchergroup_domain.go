package strmatcher

import "strings"

func breakDomain(domain string) []string {
	return strings.Split(domain, ".")
}

type node struct {
	values []uint32
	sub    map[string]*node
}

// DomainMatcherGroup is an implementation of MatcherGroup.
// It uses trie to optimize both memory consumption and lookup speed. Trie node is domain label based.
type DomainMatcherGroup struct {
	root *node
}

// AddDomainMatcher implements MatcherGroupForDomain.AddDomainMatcher.
func (g *DomainMatcherGroup) AddDomainMatcher(matcher DomainMatcher, value uint32) {
	if g.root == nil {
		g.root = new(node)
	}

	current := g.root
	parts := breakDomain(matcher.Pattern())
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.sub == nil {
			current.sub = make(map[string]*node)
		}
		next := current.sub[part]
		if next == nil {
			next = new(node)
			current.sub[part] = next
		}
		current = next
	}

	current.values = append(current.values, value)
}

// Match implements MatcherGroup.Match.
func (g *DomainMatcherGroup) Match(domain string) []uint32 {
	if domain == "" {
		return nil
	}

	current := g.root
	if current == nil {
		return nil
	}

	nextPart := func(idx int) int {
		for i := idx - 1; i >= 0; i-- {
			if domain[i] == '.' {
				return i
			}
		}
		return -1
	}

	matches := [][]uint32{}
	idx := len(domain)
	for {
		if idx == -1 || current.sub == nil {
			break
		}

		nidx := nextPart(idx)
		part := domain[nidx+1 : idx]
		next := current.sub[part]
		if next == nil {
			break
		}
		current = next
		idx = nidx
		if len(current.values) > 0 {
			matches = append(matches, current.values)
		}
	}
	switch len(matches) {
	case 0:
		return nil
	case 1:
		return matches[0]
	default:
		result := []uint32{}
		for idx := range matches {
			// Insert reversely, the subdomain that matches further ranks higher
			result = append(result, matches[len(matches)-1-idx]...)
		}
		return result
	}
}

// MatchAny implements MatcherGroup.MatchAny.
func (g *DomainMatcherGroup) MatchAny(domain string) bool {
	return len(g.Match(domain)) > 0
}
