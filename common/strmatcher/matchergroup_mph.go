package strmatcher

import (
	"math/bits"
	"sort"
	"strings"
	"unsafe"
)

// PrimeRK is the prime base used in Rabin-Karp algorithm.
const PrimeRK = 16777619

// calculate the rolling murmurHash of given string
func RollingHash(s string) uint32 {
	h := uint32(0)
	for i := len(s) - 1; i >= 0; i-- {
		h = h*PrimeRK + uint32(s[i])
	}
	return h
}

// MphMatcherGroup is an implementation of MatcherGroup.
// It implements Rabin-Karp algorithm and minimal perfect hash table for Full and Domain matcher.
type MphMatcherGroup struct {
	rules      []string
	level0     []uint32
	level0Mask int
	level1     []uint32
	level1Mask int
	ruleMap    *map[string]uint32
}

func NewMphMatcherGroup() *MphMatcherGroup {
	return &MphMatcherGroup{
		rules:      nil,
		level0:     nil,
		level0Mask: 0,
		level1:     nil,
		level1Mask: 0,
		ruleMap:    &map[string]uint32{},
	}
}

// AddFullMatcher implements MatcherGroupForFull.
func (g *MphMatcherGroup) AddFullMatcher(matcher FullMatcher, _ uint32) {
	pattern := strings.ToLower(matcher.Pattern())
	(*g.ruleMap)[pattern] = RollingHash(pattern)
}

// AddDomainMatcher implements MatcherGroupForDomain.
func (g *MphMatcherGroup) AddDomainMatcher(matcher DomainMatcher, _ uint32) {
	pattern := strings.ToLower(matcher.Pattern())
	h := RollingHash(pattern)
	(*g.ruleMap)[pattern] = h
	(*g.ruleMap)["."+pattern] = h*PrimeRK + uint32('.')
}

// Build builds a minimal perfect hash table for insert rules.
func (g *MphMatcherGroup) Build() {
	keyLen := len(*g.ruleMap)
	if keyLen == 0 {
		keyLen = 1
		(*g.ruleMap)["empty___"] = RollingHash("empty___")
	}
	g.level0 = make([]uint32, nextPow2(keyLen/4))
	g.level0Mask = len(g.level0) - 1
	g.level1 = make([]uint32, nextPow2(keyLen))
	g.level1Mask = len(g.level1) - 1
	sparseBuckets := make([][]int, len(g.level0))
	var ruleIdx int
	for rule, hash := range *g.ruleMap {
		n := int(hash) & g.level0Mask
		g.rules = append(g.rules, rule)
		sparseBuckets[n] = append(sparseBuckets[n], ruleIdx)
		ruleIdx++
	}
	g.ruleMap = nil
	var buckets []indexBucket
	for n, vals := range sparseBuckets {
		if len(vals) > 0 {
			buckets = append(buckets, indexBucket{n, vals})
		}
	}
	sort.Sort(bySize(buckets))

	occ := make([]bool, len(g.level1))
	var tmpOcc []int
	for _, bucket := range buckets {
		seed := uint32(0)
		for {
			findSeed := true
			tmpOcc = tmpOcc[:0]
			for _, i := range bucket.vals {
				n := int(strhashFallback(unsafe.Pointer(&g.rules[i]), uintptr(seed))) & g.level1Mask // nosemgrep
				if occ[n] {
					for _, n := range tmpOcc {
						occ[n] = false
					}
					seed++
					findSeed = false
					break
				}
				occ[n] = true
				tmpOcc = append(tmpOcc, n)
				g.level1[n] = uint32(i)
			}
			if findSeed {
				g.level0[bucket.n] = seed
				break
			}
		}
	}
}

// Lookup searches for s in t and returns its index and whether it was found.
func (g *MphMatcherGroup) Lookup(h uint32, s string) bool {
	i0 := int(h) & g.level0Mask
	seed := g.level0[i0]
	i1 := int(strhashFallback(unsafe.Pointer(&s), uintptr(seed))) & g.level1Mask // nosemgrep
	n := g.level1[i1]
	return s == g.rules[int(n)]
}

// Match implements MatcherGroup.Match.
func (*MphMatcherGroup) Match(_ string) []uint32 {
	return nil
}

// MatchAny implements MatcherGroup.MatchAny.
func (g *MphMatcherGroup) MatchAny(pattern string) bool {
	hash := uint32(0)
	for i := len(pattern) - 1; i >= 0; i-- {
		hash = hash*PrimeRK + uint32(pattern[i])
		if pattern[i] == '.' {
			if g.Lookup(hash, pattern[i:]) {
				return true
			}
		}
	}
	return g.Lookup(hash, pattern)
}

func nextPow2(v int) int {
	if v <= 1 {
		return 1
	}
	const MaxUInt = ^uint(0)
	n := (MaxUInt >> bits.LeadingZeros(uint(v))) + 1
	return int(n)
}

type indexBucket struct {
	n    int
	vals []int
}

type bySize []indexBucket

func (s bySize) Len() int           { return len(s) }
func (s bySize) Less(i, j int) bool { return len(s[i].vals) > len(s[j].vals) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type stringStruct struct {
	str unsafe.Pointer
	len int
}

func strhashFallback(a unsafe.Pointer, h uintptr) uintptr {
	x := (*stringStruct)(a)
	return memhashFallback(x.str, h, uintptr(x.len))
}

const (
	// Constants for multiplication: four random odd 64-bit numbers.
	m1 = 16877499708836156737
	m2 = 2820277070424839065
	m3 = 9497967016996688599
	m4 = 15839092249703872147
)

var hashkey = [4]uintptr{1, 1, 1, 1}

func memhashFallback(p unsafe.Pointer, seed, s uintptr) uintptr {
	h := uint64(seed + s*hashkey[0])
tail:
	switch {
	case s == 0:
	case s < 4:
		h ^= uint64(*(*byte)(p))
		h ^= uint64(*(*byte)(add(p, s>>1))) << 8
		h ^= uint64(*(*byte)(add(p, s-1))) << 16
		h = rotl31(h*m1) * m2
	case s <= 8:
		h ^= uint64(readUnaligned32(p))
		h ^= uint64(readUnaligned32(add(p, s-4))) << 32
		h = rotl31(h*m1) * m2
	case s <= 16:
		h ^= readUnaligned64(p)
		h = rotl31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-8))
		h = rotl31(h*m1) * m2
	case s <= 32:
		h ^= readUnaligned64(p)
		h = rotl31(h*m1) * m2
		h ^= readUnaligned64(add(p, 8))
		h = rotl31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-16))
		h = rotl31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-8))
		h = rotl31(h*m1) * m2
	default:
		v1 := h
		v2 := uint64(seed * hashkey[1])
		v3 := uint64(seed * hashkey[2])
		v4 := uint64(seed * hashkey[3])
		for s >= 32 {
			v1 ^= readUnaligned64(p)
			v1 = rotl31(v1*m1) * m2
			p = add(p, 8)
			v2 ^= readUnaligned64(p)
			v2 = rotl31(v2*m2) * m3
			p = add(p, 8)
			v3 ^= readUnaligned64(p)
			v3 = rotl31(v3*m3) * m4
			p = add(p, 8)
			v4 ^= readUnaligned64(p)
			v4 = rotl31(v4*m4) * m1
			p = add(p, 8)
			s -= 32
		}
		h = v1 ^ v2 ^ v3 ^ v4
		goto tail
	}

	h ^= h >> 29
	h *= m3
	h ^= h >> 32
	return uintptr(h)
}

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x) // nosemgrep
}

func readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

func rotl31(x uint64) uint64 {
	return (x << 31) | (x >> (64 - 31))
}

func readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 | uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}
