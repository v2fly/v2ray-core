package router

import (
	"fmt"
	"sort"
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

// StartScheduler start the health checker scheduler
func (b *Balancer) StartScheduler() {
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
		go func() {
			_, err := b.CheckAll()
			if err != nil {
				newError("HealthCheckScheduler: ", err).AtWarning().WriteToLog()
				return
			}
		}()
	}
}

// CheckAll checks all outbounds, and return their tags
func (b *Balancer) CheckAll() ([]string, error) {
	tags, err := b.SelectOutbounds()
	if err != nil {
		return nil, newError("error select outbounds: ", err)
	}
	b.cleanupResults(tags)
	b.doCheck(tags)
	return tags, nil
}

// StopHealthCheckScheduler stop the health checker
func (b *Balancer) StopHealthCheckScheduler() error {
	b.healthChecker.ticker.Stop()
	return nil
}

// Check start the health checking for given tags.
// it validates tags and checks those only in current balancer
func (b *Balancer) Check(tags []string) error {
	ts, err := b.getCheckTags(tags)
	if err != nil {
		return err
	}
	if len(ts) == 0 {
		return nil
	}
	newError("HealthCheck: Perform one-time check for tags ", ts).AtInfo().WriteToLog()
	b.doCheck(ts)
	return nil
}

// doCheck do check for tags, you should make
// sure all tags are valid for current balancer
func (b *Balancer) doCheck(tags []string) {
	if !b.healthChecker.Settings.Enabled || len(tags) == 0 {
		return
	}

	channels := make(map[string]chan time.Duration)
	rtts := make(map[string][]time.Duration)

	// TODO: too many concurrent health check ping?
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

func (b *Balancer) getCheckTags(tags []string) ([]string, error) {
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

func (b *Balancer) makeHealthStatItems(tags []string) []*routing.OutboundHealth {
	b.healthChecker.access.Lock()
	defer b.healthChecker.access.Unlock()
	items := make([]*routing.OutboundHealth, 0)
	for _, tag := range tags {
		item := &routing.OutboundHealth{
			Outbound: tag,
			RTT:      0,
		}
		result, ok := b.healthChecker.Results[tag]
		if ok {
			if result.FailCount > 0 {
				item.RTT = -1
			} else {
				item.RTT = result.AverageRTT
			}
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		iRTT := items[i].RTT
		jRTT := items[j].RTT
		// 0 rtt means not checked or failed, sort in the tail
		if iRTT <= 0 && jRTT > 0 {
			return false
		}
		if iRTT > 0 && jRTT <= 0 {
			return true
		}
		return iRTT < jRTT
	})
	return items
}
