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
	Destination   string        `json:"destination"`
	Interval      time.Duration `json:"interval"`
	SamplingCount int           `json:"sampling"`
	Timeout       time.Duration `json:"timeout"`
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
	interval := h.Settings.Interval * time.Duration(h.Settings.SamplingCount)
	ticker := time.NewTicker(interval)
	h.ticker = ticker
	for {
		go func() {
			tags, err := selector()
			if err != nil {
				newError("error select outbounds for scheduled health check: ", err).AtWarning().WriteToLog()
				return
			}
			h.doCheck(tags, interval, h.Settings.SamplingCount)
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
func (h *HealthPing) Check(tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	newError("perform one-time health check for tags ", tags).AtInfo().WriteToLog()
	h.doCheck(tags, 0, 1)
	return nil
}

type rtt struct {
	tag   string
	value time.Duration
}

// doCheck performs the 'rounds' amount checks in given 'duration'. You should make
// sure all tags are valid for current balancer
func (h *HealthPing) doCheck(tags []string, duration time.Duration, rounds int) {
	count := len(tags) * rounds
	if count == 0 {
		return
	}
	ch := make(chan *rtt, count)
	// rtts := make(map[string][]time.Duration)
	for _, tag := range tags {
		client := &pingClient{
			Dispatcher:  h.dispatcher,
			Handler:     tag,
			Destination: h.Settings.Destination,
			Timeout:     h.Settings.Timeout,
		}
		for i := 0; i < rounds; i++ {
			delay := time.Duration(0)
			if duration > 0 {
				delay = time.Duration(dice.Roll(int(duration)))
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
					delay = math.MaxInt64
				}
				ch <- &rtt{
					tag:   client.Handler,
					value: delay,
				}
			})
		}
	}
	for i := 0; i < count; i++ {
		rtt := <-ch
		h.putResult(rtt.tag, rtt.value)
	}
}

func (h *HealthPing) putResult(tag string, rtt time.Duration) {
	h.access.Lock()
	defer h.access.Unlock()
	if h.Results == nil {
		h.Results = make(map[string]*HealthPingResult)
	}
	r, ok := h.Results[tag]
	if !ok {
		r = newHealthPingResult(h.Settings.SamplingCount)
		h.Results[tag] = r
	}
	r.Put(rtt)
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
