package router

import (
	"fmt"
	"math"
	sync "sync"
	"time"

	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/routing"
)

// HealthPingSettings holds settings for health Checker
type HealthPingSettings struct {
	Destination string        `json:"destination"`
	Interval    time.Duration `json:"interval"`
	Rounds      int           `json:"rounds"`
	Timeout     time.Duration `json:"timeout"`
}

// HealthPingResult holds result for health Checker
type HealthPingResult struct {
	Count        int
	FailCount    int
	RTTDeviation time.Duration
	RTTAverage   time.Duration
	RTTMax       time.Duration
	RTTMin       time.Duration
}

// HealthPing is the health checker for balancers
type HealthPing struct {
	access     sync.Mutex
	ticker     *time.Ticker
	dispatcher routing.Dispatcher

	Settings *HealthPingSettings
	Results  map[string]*HealthPingResult
}

// StartScheduler implements the HealthChecker
func (h *HealthPing) StartScheduler(selector func() ([]string, error)) {
	if h.ticker != nil {
		return
	}
	ticker := time.NewTicker(h.Settings.Interval)
	h.ticker = ticker
	for {
		go func() {
			tags, err := selector()
			if err != nil {
				newError("error select outbounds for scheduled health check: ", err).AtWarning().WriteToLog()
				return
			}
			h.doCheck(tags, true)
			h.cleanupResults(tags)
		}()
		_, ok := <-ticker.C
		if !ok {
			break
		}
	}
}

// StopScheduler implements the HealthChecker
func (h *HealthPing) StopScheduler() {
	h.ticker.Stop()
}

// Check implements the HealthChecker
func (h *HealthPing) Check(tags []string, distributed bool) error {
	if len(tags) == 0 {
		return nil
	}
	newError("perform one-time health check for tags ", tags).AtInfo().WriteToLog()
	h.doCheck(tags, distributed)
	return nil
}

// doCheck do check for tags, you should make
// sure all tags are valid for current balancer
func (h *HealthPing) doCheck(tags []string, distributed bool) {
	if len(tags) == 0 {
		return
	}

	channels := make(map[string]chan time.Duration)
	rtts := make(map[string][]time.Duration)

	rounds := h.Settings.Rounds
	if !distributed {
		// if not distributed, multiple rounds has no practical meaning,
		// forced to use 1
		rounds = 1
	}
	for _, tag := range tags {
		ch := make(chan time.Duration, rounds)
		channels[tag] = ch
		client := &pingClient{
			Dispatcher:  h.dispatcher,
			Handler:     tag,
			Destination: h.Settings.Destination,
			Timeout:     h.Settings.Timeout,
		}
		for i := 0; i < rounds; i++ {
			delay := time.Duration(0)
			if distributed {
				delay = time.Duration(dice.Roll(int(h.Settings.Interval)))
			}
			time.AfterFunc(delay, func() {
				newError("checking ", client.Handler).AtDebug().WriteToLog()
				delay, err := client.MeasureDelay()
				if err != nil {
					newError(fmt.Sprintf(
						"error ping %s with %s: %s",
						h.Settings.Destination,
						client.Handler,
						err,
					)).AtWarning().WriteToLog()
					delay = -1
				}
				ch <- delay
			})
		}
	}
	for tag, ch := range channels {
		for i := 0; i < rounds; i++ {
			rtt := <-ch
			// newError("ping rtt of '", tag, "'=", rtt).AtDebug().WriteToLog()
			rtts[tag] = append(rtts[tag], rtt)
		}
	}
	h.access.Lock()
	defer h.access.Unlock()
	if h.Results == nil {
		h.Results = make(map[string]*HealthPingResult)
	}
	for tag, r := range rtts {
		result, ok := h.Results[tag]
		if !ok {
			result = &HealthPingResult{}
			h.Results[tag] = result
		}
		sum := time.Duration(0)
		result.Count = len(r)
		result.FailCount = 0
		result.RTTMax = 0
		result.RTTMin = r[0]
		for _, rtt := range r {
			if rtt < 0 {
				result.FailCount++
				continue
			}
			sum += rtt
			if result.RTTMax < rtt {
				result.RTTMax = rtt
			}
			if result.RTTMin > rtt {
				result.RTTMin = rtt
			}
		}
		if result.RTTMin < 0 {
			// all failed
			result.RTTMin = 0
		}
		result.RTTAverage = time.Duration(int(sum) / result.Count)

		variance := float64(0)
		for _, rtt := range r {
			variance += math.Pow(float64(rtt-result.RTTAverage), 2)
		}
		std := math.Sqrt(variance / float64(result.Count-result.FailCount))
		result.RTTDeviation = time.Duration(std)
		newError(fmt.Sprintf(
			"check '%s': %d of %d success, rtt min/avg/max = %s/%s/%s",
			tag,
			result.Count-result.FailCount,
			result.Count,
			result.RTTMin,
			result.RTTAverage,
			result.RTTMax,
		)).AtInfo().WriteToLog()
	}
}

// cleanupResults removes results of removed handlers,
// tags is all valid tags for the Balancer now
func (h *HealthPing) cleanupResults(tags []string) {
	h.access.Lock()
	defer h.access.Unlock()
	for tag := range h.Results {
		found := false
		for _, v := range tags {
			if tag == v {
				found = true
				break
			}
		}
		if !found {
			delete(h.Results, tag)
		}
	}
}
