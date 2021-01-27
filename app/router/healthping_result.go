package router

import (
	"math"
	"time"
)

// HealthPingResult holds result for health Checker
type HealthPingResult struct {
	Count        int
	FailCount    int
	RTTDeviation time.Duration
	RTTAverage   time.Duration
	RTTMax       time.Duration
	RTTMin       time.Duration

	idx      int
	cap      int
	validity time.Duration
	rtts     []*pingRTT
}

type pingRTT struct {
	time  time.Time
	value time.Duration
}

// NewHealthPingResult returns a *HealthPingResult with specified capacity
func NewHealthPingResult(cap int, validity time.Duration) *HealthPingResult {
	return &HealthPingResult{cap: cap, validity: validity}
}

// Put puts a new rtt to the HealthPingResult
func (h *HealthPingResult) Put(d time.Duration) {
	if h.rtts == nil {
		h.rtts = make([]*pingRTT, h.cap)
		for i := 0; i < h.cap; i++ {
			h.rtts[i] = &pingRTT{}
		}
		h.idx = -1
	}
	h.moveIndex()
	h.rtts[h.idx].time = time.Now()
	h.rtts[h.idx].value = d
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
	h.RTTMin = h.rtts[0].value
	cnt := 0
	for _, rtt := range h.rtts {
		switch {
		case rtt.value == 0 || time.Since(rtt.time) > h.validity:
			continue
		case rtt.value == math.MaxInt64:
			h.FailCount++
			continue
		}
		cnt++
		sum += rtt.value
		if h.RTTMax < rtt.value {
			h.RTTMax = rtt.value
		}
		if h.RTTMin > rtt.value {
			h.RTTMin = rtt.value
		}
	}
	if h.RTTMin < 0 {
		// all failed
		h.RTTMin = 0
	}
	h.Count = cnt + h.FailCount
	if h.FailCount > 0 {
		h.RTTAverage = time.Duration(math.MaxInt64)
		h.RTTDeviation = time.Duration(math.MaxInt64)
		return
	}
	h.RTTAverage = time.Duration(int(sum) / cnt)
	variance := float64(0)
	cnt = 0
	for _, rtt := range h.rtts {
		if rtt.value <= 0 {
			continue
		}
		cnt++
		variance += math.Pow(float64(rtt.value-h.RTTAverage), 2)
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
