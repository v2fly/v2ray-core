package burst

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common/dice"
)

// HealthPingSettings holds settings for health Checker
type HealthPingSettings struct {
	Destination   string        `json:"destination"`
	Connectivity  string        `json:"connectivity"`
	Interval      time.Duration `json:"interval"`
	SamplingCount int           `json:"sampling"`
	Timeout       time.Duration `json:"timeout"`
}

// HealthPing is the health checker for balancers
type HealthPing struct {
	ctx         context.Context
	access      sync.Mutex
	ticker      *time.Ticker
	tickerClose chan struct{}

	Settings *HealthPingSettings
	Results  map[string]*HealthPingRTTS
}

// NewHealthPing creates a new HealthPing with settings
func NewHealthPing(ctx context.Context, config *HealthPingConfig) *HealthPing {
	settings := &HealthPingSettings{}
	if config != nil {
		settings = &HealthPingSettings{
			Connectivity:  strings.TrimSpace(config.Connectivity),
			Destination:   strings.TrimSpace(config.Destination),
			Interval:      time.Duration(config.Interval),
			SamplingCount: int(config.SamplingCount),
			Timeout:       time.Duration(config.Timeout),
		}
	}
	if settings.Destination == "" {
		// Destination URL, need 204 for success return default to chromium
		// https://github.com/chromium/chromium/blob/main/components/safety_check/url_constants.cc#L10
		// https://chromium.googlesource.com/chromium/src/+/refs/heads/main/components/safety_check/url_constants.cc#10
		settings.Destination = "https://connectivitycheck.gstatic.com/generate_204"
	}
	if settings.Interval == 0 {
		settings.Interval = time.Duration(1) * time.Minute
	} else if settings.Interval < 10 {
		newError("health check interval is too small, 10s is applied").AtWarning().WriteToLog()
		settings.Interval = time.Duration(10) * time.Second
	}
	if settings.SamplingCount <= 0 {
		settings.SamplingCount = 10
	}
	if settings.Timeout <= 0 {
		// results are saved after all health pings finish,
		// a larger timeout could possibly makes checks run longer
		settings.Timeout = time.Duration(5) * time.Second
	}
	return &HealthPing{
		ctx:      ctx,
		Settings: settings,
		Results:  nil,
	}
}

// StartScheduler implements the HealthChecker
func (h *HealthPing) StartScheduler(selector func() ([]string, error)) {
	if h.ticker != nil {
		return
	}
	interval := h.Settings.Interval * time.Duration(h.Settings.SamplingCount)
	ticker := time.NewTicker(interval)
	tickerClose := make(chan struct{})
	h.ticker = ticker
	h.tickerClose = tickerClose
	go func() {
		for {
			go func() {
				tags, err := selector()
				if err != nil {
					newError("error select outbounds for scheduled health check: ", err).AtWarning().WriteToLog()
					return
				}
				h.doCheck(tags, interval, h.Settings.SamplingCount)
				h.Cleanup(tags)
			}()
			select {
			case <-ticker.C:
				continue
			case <-tickerClose:
				return
			}
		}
	}()
}

// StopScheduler implements the HealthChecker
func (h *HealthPing) StopScheduler() {
	if h.ticker == nil {
		return
	}
	h.ticker.Stop()
	h.ticker = nil
	close(h.tickerClose)
	h.tickerClose = nil
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
	handler string
	value   time.Duration
}

// doCheck performs the 'rounds' amount checks in given 'duration'. You should make
// sure all tags are valid for current balancer
func (h *HealthPing) doCheck(tags []string, duration time.Duration, rounds int) {
	count := len(tags) * rounds
	if count == 0 {
		return
	}
	ch := make(chan *rtt, count)

	for _, tag := range tags {
		handler := tag
		client := newPingClient(
			h.ctx,
			h.Settings.Destination,
			h.Settings.Timeout,
			handler,
		)
		for i := 0; i < rounds; i++ {
			delay := time.Duration(0)
			if duration > 0 {
				delay = time.Duration(dice.Roll(int(duration)))
			}
			time.AfterFunc(delay, func() {
				newError("checking ", handler).AtDebug().WriteToLog()
				delay, err := client.MeasureDelay()
				if err == nil {
					ch <- &rtt{
						handler: handler,
						value:   delay,
					}
					return
				}
				if !h.checkConnectivity() {
					newError("network is down").AtWarning().WriteToLog()
					ch <- &rtt{
						handler: handler,
						value:   0,
					}
					return
				}
				newError(fmt.Sprintf(
					"error ping %s with %s: %s",
					h.Settings.Destination,
					handler,
					err,
				)).AtWarning().WriteToLog()
				ch <- &rtt{
					handler: handler,
					value:   rttFailed,
				}
			})
		}
	}
	for i := 0; i < count; i++ {
		rtt := <-ch
		if rtt.value > 0 {
			// should not put results when network is down
			h.PutResult(rtt.handler, rtt.value)
		}
	}
}

// PutResult puts a ping rtt to results
func (h *HealthPing) PutResult(tag string, rtt time.Duration) {
	h.access.Lock()
	defer h.access.Unlock()
	if h.Results == nil {
		h.Results = make(map[string]*HealthPingRTTS)
	}
	r, ok := h.Results[tag]
	if !ok {
		// validity is 2 times to sampling period, since the check are
		// distributed in the time line randomly, in extreme cases,
		// previous checks are distributed on the left, and latters
		// on the right
		validity := h.Settings.Interval * time.Duration(h.Settings.SamplingCount) * 2
		r = NewHealthPingResult(h.Settings.SamplingCount, validity)
		h.Results[tag] = r
	}
	r.Put(rtt)
}

// Cleanup removes results of removed handlers,
// tags should be all valid tags of the Balancer now
func (h *HealthPing) Cleanup(tags []string) {
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

// checkConnectivity checks the network connectivity, it returns
// true if network is good or "connectivity check url" not set
func (h *HealthPing) checkConnectivity() bool {
	if h.Settings.Connectivity == "" {
		return true
	}
	tester := newDirectPingClient(
		h.Settings.Connectivity,
		h.Settings.Timeout,
	)
	if _, err := tester.MeasureDelay(); err != nil {
		return false
	}
	return true
}
