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
	Round       uint
	Timeout     time.Duration
}

// HealthCheckResult holds result for health Checker
type HealthCheckResult struct {
	Count        int
	SuccessCount int
	AverageRTT   time.Duration
	MaxRTT       time.Duration
	MinRTT       time.Duration
	RTTs         []time.Duration
}

// HealthChecker is the health checker for balancers
type HealthChecker struct {
	access     sync.Mutex
	ticker     *time.Ticker
	dispatcher routing.Dispatcher

	Settings *HealthCheckSettings
	Results  map[string]*HealthCheckResult
}

// StartHealthCheck start the health checker
func (b *Balancer) StartHealthCheck() {
	if !b.healthChecker.Settings.Enabled {
		return
	}
	if b.healthChecker.ticker != nil {
		return
	}
	ticker := time.NewTicker(b.healthChecker.Settings.Interval)
	b.healthChecker.ticker = ticker
	for {
		go b.doHealthCheck()
		_, ok := <-ticker.C
		if !ok {
			break
		}
	}
}

// StopHealthCheck stop the health checker
func (b *Balancer) StopHealthCheck() error {
	b.healthChecker.ticker.Stop()
	return nil
}

// StopHealthCheck stop the health checker
func (b *Balancer) doHealthCheck() {
	tags, err := b.SelectOutbounds()
	if err != nil {
		newError("error select balancer outbounds: ", err).AtWarning().WriteToLog()
		return
	}
	channels := make(map[string]chan time.Duration)
	rtts := make(map[string][]time.Duration)
	client := &pingClient{
		Dispatcher:  b.healthChecker.dispatcher,
		Destination: b.healthChecker.Settings.Destination,
		Timeout:     b.healthChecker.Settings.Timeout,
	}

	// make sure other go routines don't check them again
	b.healthChecker.access.Lock()
	for _, tag := range tags {
		_, ok := b.healthChecker.Results[tag]
		if !ok {
			b.healthChecker.Results[tag] = &HealthCheckResult{}
		}
	}
	b.healthChecker.access.Unlock()

	for _, tag := range tags {
		ch := make(chan time.Duration, int(b.healthChecker.Settings.Round))
		channels[tag] = ch
		client.Handler = tag
		for i := 0; i < int(b.healthChecker.Settings.Round); i++ {
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
		for i := 0; i < int(b.healthChecker.Settings.Round); i++ {
			rtt := <-ch
			newError("ping rtt of '", tag, "'=", rtt).AtDebug().WriteToLog()
			rtts[tag] = append(rtts[tag], rtt)
		}
	}
	b.healthChecker.access.Lock()
	for tag, r := range rtts {
		result, _ := b.healthChecker.Results[tag]
		sum := time.Duration(0)
		result.Count = len(r)
		result.SuccessCount = 0
		result.MaxRTT = 0
		result.MinRTT = r[0]
		for _, rtt := range r {
			if rtt < 0 {
				continue
			}
			sum += rtt
			result.SuccessCount++
			if result.MaxRTT < rtt {
				result.MaxRTT = rtt
			}
			if result.MinRTT > rtt {
				result.MinRTT = rtt
			}
		}
		result.AverageRTT = time.Duration(int(sum) / result.Count)
		newError(fmt.Sprintf(
			"health checker '%s': %d of %d success, rtt min/avg/max = %s/%s/%s",
			tag,
			result.SuccessCount,
			result.Count,
			result.MinRTT,
			result.AverageRTT,
			result.MaxRTT,
		)).AtInfo().WriteToLog()
	}
	b.healthChecker.access.Unlock()
}
