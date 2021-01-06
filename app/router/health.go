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
		tags, err := b.SelectOutbounds()
		if err != nil {
			newError("HealthCheckScheduler: error select outbounds: ", err).AtWarning().WriteToLog()
			return
		}
		b.cleanupResults(tags)
		go b.doHealthCheck(tags)
	}
}

// StopHealthCheckScheduler stop the health checker
func (b *Balancer) StopHealthCheckScheduler() error {
	b.healthChecker.ticker.Stop()
	return nil
}

// HealthCheck start the health checking for given tags.
// it validates tags and checks those only in current balancer
func (b *Balancer) HealthCheck(tags []string) {
	ts, err := b.getOneTimeCheckTags(tags)
	if err != nil {
		newError("HealthCheck: error select outbounds: ", err).AtWarning().WriteToLog()
		return
	}
	if len(ts) == 0 {
		return
	}
	newError("HealthCheck: Perform one-time check for tags ", ts).AtInfo().WriteToLog()
	b.doHealthCheck(ts)
}

// doHealthCheck do check for tags, you should make
// sure all tags are valid for current balancer
func (b *Balancer) doHealthCheck(tags []string) {
	if !b.healthChecker.Settings.Enabled || len(tags) == 0 {
		return
	}

	channels := make(map[string]chan time.Duration)
	rtts := make(map[string][]time.Duration)

	for _, tag := range tags {
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

func (b *Balancer) getOneTimeCheckTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	all, err := b.SelectOutbounds()
	if err != nil {
		return nil, err
	}
	ts := make([]string, 0)
	for _, t1 := range tags {
		for _, t2 := range all {
			if t1 == t2 {
				ts = append(ts, t1)
				break
			}
		}
	}
	return ts, nil
}

// cleanupResults removes results of removed handlers,
// tags is all valid tags for the Balancer now
func (b *Balancer) cleanupResults(tags []string) {
	b.healthChecker.access.Lock()
	defer b.healthChecker.access.Unlock()

	for tag := range b.healthChecker.Results {
		found := false
		for _, v := range tags {
			if tag == v {
				found = true
				break
			}
		}
		if !found {
			// newError("healthChecker: remove tag ", tag).AtDebug().WriteToLog()
			delete(b.healthChecker.Results, tag)
		}
	}
}
