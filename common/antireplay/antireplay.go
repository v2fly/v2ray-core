package antireplay

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

const replayFilterCapacity = 100000

type AntiReplayWindow struct {
	lock           sync.Mutex
	m              *cuckoo.Filter
	n              *cuckoo.Filter
	lastSwap       int64
	poolSwap       bool
	antiReplayTime int64
}

func NewAntiReplayWindow(antiReplayTime int64) *AntiReplayWindow {
	arw := &AntiReplayWindow{}
	arw.antiReplayTime = antiReplayTime
	return arw
}

func (aw *AntiReplayWindow) AntiReplayTime() int64 {
	return aw.antiReplayTime
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
	if elapsed >= aw.antiReplayTime {
		if aw.poolSwap {
			aw.poolSwap = false
			aw.m.Reset()
		} else {
			aw.poolSwap = true
			aw.n.Reset()
		}
		aw.lastSwap = now
	}

	return aw.m.InsertUnique(sum) && aw.n.InsertUnique(sum)
}
