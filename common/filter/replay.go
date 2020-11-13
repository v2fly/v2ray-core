package filter

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

// ReplayFilter filter to check for replay attacks.
type ReplayFilter struct {
	lock     sync.Mutex
	m        *cuckoo.Filter
	n        *cuckoo.Filter
	poolSwap bool
	lastSwap int64
	interval int64
}

// NewReplayFilter create a new filter with specifying the expiration time interval.
func NewReplayFilter(interval int64) *ReplayFilter {
	rf := &ReplayFilter{}
	rf.interval = interval
	return rf
}

// Interval for swap filter pool.
func (rf *ReplayFilter) Interval() int64 {
	return rf.interval
}

// Check determine if there are duplicate records in the expiration time interval.
func (rf *ReplayFilter) Check(sum []byte) bool {
	rf.lock.Lock()
	defer rf.lock.Unlock()

	now := time.Now().Unix()

	if rf.lastSwap == 0 {
		rf.lastSwap = now
		rf.m = cuckoo.NewFilter(replayFilterCapacity)
		rf.n = cuckoo.NewFilter(replayFilterCapacity)
	}

	interval := now - rf.lastSwap
	if interval >= rf.interval {
		if rf.poolSwap {
			rf.m.Reset()
		} else {
			rf.n.Reset()
		}
		rf.poolSwap = !rf.poolSwap
		rf.lastSwap = now
	}

	return rf.m.InsertUnique(sum) && rf.n.InsertUnique(sum)
}
