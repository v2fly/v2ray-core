package router

import (
	sync "sync"
	"time"

	"v2ray.com/core/common/dice"
	"v2ray.com/core/features/outbound"
)

// HealthCheckSettings holds settings for health Checker
type HealthCheckSettings struct {
	Enabled     bool
	Destination string
	Interval    uint
	Round       uint
	Timeout     time.Duration
}

// HealthChecker is the health checker for balancers
type HealthChecker struct {
	access sync.Mutex
	ticker *time.Ticker

	Settings *HealthCheckSettings
	Results  map[string]time.Duration
}

// StartHealthCheck start the health checker
func (b *Balancer) StartHealthCheck() error {
	if !b.healthChecker.Settings.Enabled {
		return nil
	}
	if b.healthChecker.ticker != nil {
		return nil
	}
	ticker := time.NewTicker(time.Duration(b.healthChecker.Settings.Interval) * time.Minute)
	b.healthChecker.ticker = ticker
	go func() {
		for {
			err := b.doHealthCheck()
			if err != nil {
				newError("healthChecker error:", err).AtWarning().WriteToLog()
			}
			<-ticker.C
		}
	}()
	return nil
}

// StopHealthCheck stop the health checker
func (b *Balancer) StopHealthCheck() error {
	b.healthChecker.ticker.Stop()
	return nil
}

// StopHealthCheck stop the health checker
func (b *Balancer) doHealthCheck() error {
	tags, err := b.SelectOutbounds()
	if err != nil {
		return err
	}
	channels := make(map[string]chan int)
	rtts := make(map[string][]int)
	for _, tag := range tags {
		h := b.ohm.GetHandler(tag)
		ch := make(chan int, int(b.healthChecker.Settings.Round))
		channels[tag] = ch
		for i := 0; i < int(b.healthChecker.Settings.Round); i++ {
			// newError("health checker ping ", tag, "#", i).AtDebug().WriteToLog()
			go b.pingOutbound(h, ch)
		}
	}
	for tag, ch := range channels {
		for i := 0; i < int(b.healthChecker.Settings.Round); i++ {
			rtt := <-ch
			newError("health checker rtt of ", tag, "=", rtt).AtDebug().WriteToLog()
			rtts[tag] = append(rtts[tag], rtt)
		}
	}
	for tag, r := range rtts {
		sum := 0
		for _, rtt := range r {
			sum += rtt
		}
		avg := time.Duration(int(sum) / len(r))
		newError("health checker average rtt of ", tag, "=", avg).AtInfo().WriteToLog()
		b.healthChecker.Results[tag] = avg
	}
	return nil
}

func (b *Balancer) pingOutbound(handler outbound.Handler, ch chan int) {
	// TODO: ping outbound
	// handler.Dispatch()
	ch <- dice.Roll(1000)
}
