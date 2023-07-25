// modified from github.com/oleiade/lane

package pq

import (
	"golang.org/x/exp/constraints"
)

// PriorityQueue is a heap-based priority-queue data structure implementation.
//
// It can either be min (ascending) or max (descending)
// oriented/ordered. Its type parameters `T` and `P`, respectively
// specify the underlying value type and the underlying priority type.
type PriorityQueue[T any, P constraints.Ordered] struct {
	items      []*priorityQueueItem[T, P]
	itemCount  uint
	comparator func(lhs, rhs P) bool
}

// NewPriorityQueue instantiates a new PriorityQueue with the provided comparison heuristic.
// The package defines the `Max` and `Min` heuristic to define a max-oriented or
// min-oriented heuristics, respectively.
func NewPriorityQueue[T any, P constraints.Ordered](heuristic func(lhs, rhs P) bool) *PriorityQueue[T, P] {
	items := make([]*priorityQueueItem[T, P], 1)
	items[0] = nil

	return &PriorityQueue[T, P]{
		items:      items,
		itemCount:  0,
		comparator: heuristic,
	}
}

// NewMaxPriorityQueue instantiates a new maximum oriented PriorityQueue.
func NewMaxPriorityQueue[T any, P constraints.Ordered]() *PriorityQueue[T, P] {
	return NewPriorityQueue[T](Maximum[P])
}

// NewMinPriorityQueue instantiates a new minimum oriented PriorityQueue.
func NewMinPriorityQueue[T any, P constraints.Ordered]() *PriorityQueue[T, P] {
	return NewPriorityQueue[T](Minimum[P])
}

// Maximum returns whether `rhs` is greater than `lhs`.
//
// Use it as a comparison heuristic during a PriorityQueue's
// instantiation.
func Maximum[T constraints.Ordered](lhs, rhs T) bool {
	return lhs < rhs
}

// Minimum returns whether `rhs` is less than `lhs`.
//
// Use it as a comparison heuristic during a PriorityQueue's
// instantiation.
func Minimum[T constraints.Ordered](lhs, rhs T) bool {
	return lhs > rhs
}

// Push inserts the value in the PriorityQueue with the provided priority
// in at most *O(log n)* time complexity.
func (pq *PriorityQueue[T, P]) Push(value T, priority P) {
	item := newPriorityQueueItem(value, priority)

	pq.items = append(pq.items, item)
	pq.itemCount++
	pq.swim(pq.size())
}

// Pop and return the highest or lowest priority item (depending on the
// comparison heuristic of your PriorityQueue) from the PriorityQueue in
// at most *O(log n)* complexity.
func (pq *PriorityQueue[T, P]) Pop() (value T, priority P, ok bool) {

	if pq.size() < 1 {
		ok = false
		return
	}

	max := pq.items[1]
	pq.exch(1, pq.size())
	pq.items = pq.items[0:pq.size()]
	pq.itemCount--
	pq.sink(1)

	value = max.value
	priority = max.priority
	ok = true

	return
}

// Head returns the highest or lowest priority item (depending on
// the comparison heuristic of your PriorityQueue) from the PriorityQueue
// in *O(1)* complexity.
func (pq *PriorityQueue[T, P]) Head() (value T, priority P, ok bool) {

	if pq.size() < 1 {
		ok = false
		return
	}

	value = pq.items[1].value
	priority = pq.items[1].priority
	ok = true

	return
}

// Size returns the number of elements present in the PriorityQueue.
func (pq *PriorityQueue[T, P]) Size() uint {

	return pq.size()
}

// Empty returns whether the PriorityQueue is empty.
func (pq *PriorityQueue[T, P]) Empty() bool {

	return pq.size() == 0
}

// Clone returns a copy of the PriorityQueue.
func (pq *PriorityQueue[T, P]) Clone() *PriorityQueue[T, P] {
	items := make([]*priorityQueueItem[T, P], len(pq.items))
	copy(items, pq.items)

	return &PriorityQueue[T, P]{
		items:      items,
		itemCount:  pq.itemCount,
		comparator: pq.comparator,
	}
}

func (pq *PriorityQueue[T, P]) swim(k uint) {
	for k > 1 && pq.less(k/2, k) {
		pq.exch(k/2, k)
		k /= 2
	}
}

func (pq *PriorityQueue[T, P]) sink(k uint) {
	for 2*k <= pq.size() {
		j := 2 * k

		if j < pq.size() && pq.less(j, j+1) {
			j++
		}

		if !pq.less(k, j) {
			break
		}

		pq.exch(k, j)
		k = j
	}
}

func (pq *PriorityQueue[T, P]) size() uint {
	return pq.itemCount
}

func (pq *PriorityQueue[T, P]) less(lhs, rhs uint) bool {
	return pq.comparator(pq.items[lhs].priority, pq.items[rhs].priority)
}

func (pq *PriorityQueue[T, P]) exch(lhs, rhs uint) {
	pq.items[lhs], pq.items[rhs] = pq.items[rhs], pq.items[lhs]
}

// priorityQueueItem is the underlying PriorityQueue item container.
type priorityQueueItem[T any, P constraints.Ordered] struct {
	value    T
	priority P
}

// newPriorityQueue instantiates a new priorityQueueItem.
func newPriorityQueueItem[T any, P constraints.Ordered](value T, priority P) *priorityQueueItem[T, P] {
	return &priorityQueueItem[T, P]{
		value:    value,
		priority: priority,
	}
}
