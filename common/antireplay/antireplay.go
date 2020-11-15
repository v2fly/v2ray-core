package antireplay

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

type AntiReplayWindow struct {
	lock     sync.Mutex
	m        *cuckoo.Filter
	n        *cuckoo.Filter
	lastSwap int64
	poolSwap bool
	interval int64
}

func NewAntiReplayWindow(interval int64) *AntiReplayWindow {
	arw := &AntiReplayWindow{}
	arw.interval = interval
	return arw
}

func (aw *AntiReplayWindow) Interval() int64 {
	return aw.interval
}

func (aw *AntiReplayWindow) Check(sum []byte) bool {
	aw.lock.Lock()
	defer aw.lock.Unlock()

	now := time.Now().Unix()
	if aw.lastSwap == 0 {
		aw.lastSwap = now
		aw.m = cuckoo.NewFilter(replayFilterCapacity)
		aw.n = cuckoo.NewFilter(replayFilterCapacity)
	}

	elapsed := now - aw.lastSwap
	if elapsed >= aw.Interval() {
		if aw.poolSwap {
			aw.m.Reset()
		} else {
			aw.n.Reset()
		}
		aw.poolSwap = !aw.poolSwap
		aw.lastSwap = now
	}

	return aw.m.InsertUnique(sum) && aw.n.InsertUnique(sum)
}
