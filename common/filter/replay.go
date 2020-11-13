package filter

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

type ReplayFilter struct {
	lock     sync.Mutex
	m        *cuckoo.Filter
	n        *cuckoo.Filter
	poolSwap bool
	lastSwap int64
	interval int64
}

func NewReplayFilter(interval int64) *ReplayFilter {
	rf := &ReplayFilter{}
	rf.interval = interval
	return rf
}

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
