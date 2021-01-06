package router

import (
	"fmt"
	sync "sync"
	"time"

	"v2ray.com/core/features/routing"
)

// HealthCheckSettings holds settings for health Checker
type HealthCheckSettings struct {
	Enabled     bool
	Destination string
	Interval    time.Duration
	Rounds      int
	Timeout     time.Duration
}

// HealthCheckResult holds result for health Checker
type HealthCheckResult struct {
	Count      int
	FailCount  int
	AverageRTT time.Duration
	MaxRTT     time.Duration
	MinRTT     time.Duration
	RTTs       []time.Duration
}

// HealthChecker is the health checker for balancers
type HealthChecker struct {
	access     sync.Mutex
	ticker     *time.Ticker
	dispatcher routing.Dispatcher

	Settings *HealthCheckSettings
	Results  map[string]*HealthCheckResult
}

// StartHealthCheckScheduler start the health checker
func (b *Balancer) StartHealthCheckScheduler() {
	if !b.healthChecker.Settings.Enabled {
		return
	}
	if b.healthChecker.ticker != nil {
		return
	}
	ticker := time.NewTicker(b.healthChecker.Settings.Interval)
	b.healthChecker.ticker = ticker
	for {
		_, ok := <-ticker.C
		if !ok {
			break
		}
		go b.HealthCheck(nil)
	}
}

// StopHealthCheckScheduler stop the health checker
func (b *Balancer) StopHealthCheckScheduler() error {
	b.healthChecker.ticker.Stop()
	return nil
}

// HealthCheck start the health checking for given tags. if tags not provided,
// it checks only those outbounds not yet checed.
func (b *Balancer) HealthCheck(tags []string) {
	if !b.healthChecker.Settings.Enabled {
		return
	}
	all, err := b.SelectOutbounds()
	if err != nil {
		newError("error select balancer outbounds: ", err).AtWarning().WriteToLog()
		return
	}
	if len(all) == 0 {
		return
	}
	checkingTags := all
	if len(tags) > 0 {
		checkingTags = make([]string, 0)
		for _, t1 := range tags {
			for _, t2 := range all {
				if t1 == t2 {
					checkingTags = append(checkingTags, t1)
					break
				}
			}
		}
		newError("Perform one-time health check for tags ", checkingTags).AtInfo().WriteToLog()
	}

	if len(checkingTags) == 0 {
		// newError("no outbound check needed.").AtInfo().WriteToLog()
		return
	}

	channels := make(map[string]chan time.Duration)
	rtts := make(map[string][]time.Duration)

	for _, tag := range checkingTags {
		ch := make(chan time.Duration, b.healthChecker.Settings.Rounds)
		channels[tag] = ch
		client := &pingClient{
			Dispatcher:  b.healthChecker.dispatcher,
			Handler:     tag,
			Destination: b.healthChecker.Settings.Destination,
			Timeout:     b.healthChecker.Settings.Timeout,
		}
		for i := 0; i < b.healthChecker.Settings.Rounds; i++ {
			// newError("health checker ping ", tag, "#", i).AtDebug().WriteToLog()
			go func() {
				delay, err := client.MeasureDelay()
				if err != nil {
					newError(fmt.Sprintf(
						"error ping %s with %s: %s",
						b.healthChecker.Settings.Destination,
						tag,
						err,
					)).AtWarning().WriteToLog()
				}
				ch <- delay
			}()
		}
	}
	for tag, ch := range channels {
		for i := 0; i < b.healthChecker.Settings.Rounds; i++ {
			rtt := <-ch
			// newError("ping rtt of '", tag, "'=", rtt).AtDebug().WriteToLog()
			rtts[tag] = append(rtts[tag], rtt)
		}
	}
	b.healthChecker.access.Lock()
	defer b.healthChecker.access.Unlock()
	for tag, r := range rtts {
		result, ok := b.healthChecker.Results[tag]
		if !ok {
			result = &HealthCheckResult{}
			b.healthChecker.Results[tag] = result
		}
		sum := time.Duration(0)
		result.Count = len(r)
		result.FailCount = 0
		result.MaxRTT = 0
		result.MinRTT = r[0]
		for _, rtt := range r {
			if rtt < 0 {
				result.FailCount++
				continue
			}
			sum += rtt
			if result.MaxRTT < rtt {
				result.MaxRTT = rtt
			}
			if result.MinRTT > rtt {
				result.MinRTT = rtt
			}
		}
		if result.MinRTT < 0 {
			// all failed
			result.MinRTT = 0
		}
		result.AverageRTT = time.Duration(int(sum) / result.Count)
		newError(fmt.Sprintf(
			"health checker '%s': %d of %d success, rtt min/avg/max = %s/%s/%s",
			tag,
			result.Count-result.FailCount,
			result.Count,
			result.MinRTT,
			result.AverageRTT,
			result.MaxRTT,
		)).AtInfo().WriteToLog()
	}
}
