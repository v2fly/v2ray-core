package router

import (
	"math"
	"time"
)

// HealthPingStats is the statistics of HealthPingRTTS
type HealthPingStats struct {
	Count        int
	FailCount    int
	RTTDeviation time.Duration
	RTTAverage   time.Duration
	RTTMax       time.Duration
	RTTMin       time.Duration
}

// HealthPingRTTS holds ping rtts for health Checker
type HealthPingRTTS struct {
	idx      int
	cap      int
	validity time.Duration
	rtts     []*pingRTT

	lastUpdateAt time.Time
	stats        *HealthPingStats
}

type pingRTT struct {
	time  time.Time
	value time.Duration
}

// NewHealthPingResult returns a *HealthPingResult with specified capacity
func NewHealthPingResult(cap int, validity time.Duration) *HealthPingRTTS {
	return &HealthPingRTTS{cap: cap, validity: validity}
}

// Get gets statistics of the HealthPingRTTS
func (h *HealthPingRTTS) Get() *HealthPingStats {
	lastPutAt := h.rtts[h.idx].time
	now := time.Now()
	if h.stats == nil || h.lastUpdateAt.Before(lastPutAt) || h.findOutdated(now) >= 0 {
		h.stats = h.getStatistics()
		h.lastUpdateAt = now
	}
	return h.stats
}

// Put puts a new rtt to the HealthPingResult
func (h *HealthPingRTTS) Put(d time.Duration) {
	if h.rtts == nil {
		h.rtts = make([]*pingRTT, h.cap)
		for i := 0; i < h.cap; i++ {
			h.rtts[i] = &pingRTT{}
		}
		h.idx = -1
	}
	h.idx = h.calcIndex(1)
	now := time.Now()
	h.rtts[h.idx].time = now
	h.rtts[h.idx].value = d
}

func (h *HealthPingRTTS) calcIndex(step int) int {
	idx := h.idx
	idx += step
	if idx >= h.cap {
		idx %= h.cap
	}
	return idx
}

func (h *HealthPingRTTS) getStatistics() *HealthPingStats {
	stats := &HealthPingStats{}
	stats.FailCount = 0
	stats.RTTMax = 0
	stats.RTTMin = time.Duration(math.MaxInt64)
	sum := time.Duration(0)
	cnt := 0
	validRTTs := make([]time.Duration, 0)
	for _, rtt := range h.rtts {
		switch {
		case rtt.value == 0 || time.Since(rtt.time) > h.validity:
			continue
		case rtt.value == math.MaxInt64:
			stats.FailCount++
			continue
		}
		cnt++
		sum += rtt.value
		validRTTs = append(validRTTs, rtt.value)
		if stats.RTTMax < rtt.value {
			stats.RTTMax = rtt.value
		}
		if stats.RTTMin > rtt.value {
			stats.RTTMin = rtt.value
		}
	}
	stats.Count = cnt + stats.FailCount
	if cnt == 0 {
		stats.RTTMin = 0
	}
	if stats.FailCount > 0 || cnt == 0 {
		stats.RTTAverage = time.Duration(math.MaxInt64)
		stats.RTTDeviation = time.Duration(math.MaxInt64)
		return stats
	}
	stats.RTTAverage = time.Duration(int(sum) / cnt)
	var std float64
	if cnt < 2 {
		// no enough data for standard deviation, we assume it's half of the average rtt
		// if we don't do this, standard deviation of 1 round tested nodes is 0, will always
		// selected before 2 or more rounds tested nodes
		std = float64(stats.RTTAverage / 2)
	} else {
		variance := float64(0)
		for _, rtt := range validRTTs {
			variance += math.Pow(float64(rtt-stats.RTTAverage), 2)
		}
		std = math.Sqrt(variance / float64(cnt))
	}
	stats.RTTDeviation = time.Duration(std)
	return stats
}

func (h *HealthPingRTTS) findOutdated(now time.Time) int {
	for i := h.cap - 1; i < 2*h.cap; i++ {
		// from oldest to latest
		idx := h.calcIndex(i)
		validity := h.rtts[idx].time.Add(h.validity)
		if h.lastUpdateAt.After(validity) {
			return idx
		}
		if validity.Before(now) {
			return idx
		}
	}
	return -1
}
