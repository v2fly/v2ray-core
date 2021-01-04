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

// HealthChecker is the health checker for balancers
type HealthChecker struct {
	access     sync.Mutex
	ticker     *time.Ticker
	dispatcher routing.Dispatcher

	Settings *HealthCheckSettings
	Results  map[string]time.Duration
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
			newError("health checker rtt of '", tag, "'=", rtt).AtDebug().WriteToLog()
			rtts[tag] = append(rtts[tag], rtt)
		}
	}
	for tag, r := range rtts {
		sum := time.Duration(0)
		for _, rtt := range r {
			sum += rtt
		}
		avg := time.Duration(int(sum) / len(r))
		newError("health checker average rtt of '", tag, "'=", avg).AtInfo().WriteToLog()
		b.healthChecker.Results[tag] = avg
	}
}
