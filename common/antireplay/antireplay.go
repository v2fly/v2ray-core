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
	lastSwapTime   int64
	poolSwap       bool
	AntiReplayTime int64
}

func NewAntiReplayWindow(antiReplayTime int64) *AntiReplayWindow {
	arw := &AntiReplayWindow{}
	arw.AntiReplayTime = antiReplayTime
	return arw
}

func (aw *AntiReplayWindow) Check(sum []byte) bool {
	aw.lock.Lock()
	defer aw.lock.Unlock()

	if aw.lastSwapTime == 0 {
		aw.lastSwapTime = time.Now().Unix()
		aw.m = cuckoo.NewFilter(replayFilterCapacity)
		aw.n = cuckoo.NewFilter(replayFilterCapacity)
	}

	tnow := time.Now().Unix()
	timediff := tnow - aw.lastSwapTime

	if timediff >= aw.AntiReplayTime {
		if aw.poolSwap {
			aw.poolSwap = false
			aw.m.Reset()
		} else {
			aw.poolSwap = true
			aw.n.Reset()
		}
		aw.lastSwapTime = tnow
	}

	return aw.m.InsertUnique(sum) && aw.n.InsertUnique(sum)
}
