package antireplay

import (
	"sync"
	"time"

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

func NewAntiReplayWindow(antiReplayTime int64) *AntiReplayWindow {
	arw := &AntiReplayWindow{}
	arw.AntiReplayTime = antiReplayTime
	return arw
}

type AntiReplayWindow struct {
	lock           sync.Mutex
	poolA          *cuckoo.Filter
	poolB          *cuckoo.Filter
	lastSwapTime   int64
	poolSwap       bool
	AntiReplayTime int64
}

func (aw *AntiReplayWindow) Check(sum []byte) bool {
	aw.lock.Lock()
	defer aw.lock.Unlock()

	if aw.lastSwapTime == 0 {
		aw.lastSwapTime = time.Now().Unix()
		aw.poolA = cuckoo.NewFilter(100000)
		aw.poolB = cuckoo.NewFilter(100000)
	}

	tnow := time.Now().Unix()
	timediff := tnow - aw.lastSwapTime

	if timediff >= aw.AntiReplayTime {
		if aw.poolSwap {
			aw.poolSwap = false
			aw.poolA.Reset()
		} else {
			aw.poolSwap = true
			aw.poolB.Reset()
		}
		aw.lastSwapTime = tnow
	}

	return aw.poolA.InsertUnique(sum) && aw.poolB.InsertUnique(sum)
}
