package antireplay

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

// ReplayFilter check for replay attacks.
type ReplayFilter struct {
	lock     sync.Mutex
	m        *cuckoo.Filter
	n        *cuckoo.Filter
	lastSwap int64
	poolSwap bool
	interval int64
}

// NewReplayFilter create a new filter with specifying the expiration time interval in seconds.
func NewReplayFilter(interval int64) *ReplayFilter {
	filter := &ReplayFilter{}
	filter.interval = interval
	return filter
}

// Interval in second for expiration time for duplicate records.
func (filter *ReplayFilter) Interval() int64 {
	return filter.interval
}

// Check determine if there are duplicate records.
func (filter *ReplayFilter) Check(sum []byte) bool {
	filter.lock.Lock()
	defer filter.lock.Unlock()

	now := time.Now().Unix()
	if filter.lastSwap == 0 {
		filter.lastSwap = now
		filter.m = cuckoo.NewFilter(replayFilterCapacity)
		filter.n = cuckoo.NewFilter(replayFilterCapacity)
	}

	elapsed := now - filter.lastSwap
	if elapsed >= filter.Interval() {
		if filter.poolSwap {
			filter.m.Reset()
		} else {
			filter.n.Reset()
		}
		filter.poolSwap = !filter.poolSwap
		filter.lastSwap = now
	}

	return filter.m.InsertUnique(sum) && filter.n.InsertUnique(sum)
}
