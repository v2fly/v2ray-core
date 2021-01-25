package router

import (
	"math"
	"time"
)

// HealthPingResult holds result for health Checker
type HealthPingResult struct {
	FailCount    int
	RTTDeviation time.Duration
	RTTAverage   time.Duration
	RTTMax       time.Duration
	RTTMin       time.Duration

	idx  int
	cap  int
	rtts []time.Duration
}

func newHealthPingResult(cap int) *HealthPingResult {
	return &HealthPingResult{cap: cap}
}

// Put puts a new rtt to the HealthPingResult
func (h *HealthPingResult) Put(d time.Duration) {
	if h.rtts == nil {
		h.rtts = make([]time.Duration, h.cap)
		h.idx = -1
	}
	h.moveIndex()
	h.rtts[h.idx] = d
	h.update()
}

func (h *HealthPingResult) moveIndex() {
	h.idx++
	if h.idx >= h.cap {
		h.idx %= h.cap
	}
}

func (h *HealthPingResult) update() {
	sum := time.Duration(0)
	h.FailCount = 0
	h.RTTMax = 0
	h.RTTMin = h.rtts[0]
	cnt := 0
	for _, rtt := range h.rtts {
		switch {
		case rtt == 0:
			continue
		case rtt == math.MaxInt64:
			h.FailCount++
			continue
		}
		cnt++
		sum += rtt
		if h.RTTMax < rtt {
			h.RTTMax = rtt
		}
		if h.RTTMin > rtt {
			h.RTTMin = rtt
		}
	}
	if h.RTTMin < 0 {
		// all failed
		h.RTTMin = 0
	}
	if h.FailCount > 0 {
		h.RTTAverage = time.Duration(math.MaxInt64)
		h.RTTDeviation = time.Duration(math.MaxInt64)
		return
	}
	h.RTTAverage = time.Duration(int(sum) / cnt)
	variance := float64(0)
	cnt = 0
	for _, rtt := range h.rtts {
		if rtt <= 0 {
			continue
		}
		cnt++
		variance += math.Pow(float64(rtt-h.RTTAverage), 2)
	}
	var std float64
	if cnt < 2 {
		// no enough data for standard deviation, we assume it's half of the average rtt
		// if we don't do this, standard deviation of 1 round tested nodes is 0, will always
		// selected before 2 or more rounds tested nodes
		std = float64(h.RTTAverage / 2)
	} else {
		std = math.Sqrt(variance / float64(cnt))
	}
	h.RTTDeviation = time.Duration(std)
}
