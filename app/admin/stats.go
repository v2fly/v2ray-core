// +build !confonly

package admin

//go:generate errorgen

import (
	"sync"
	"sync/atomic"
	"time"
	"v2ray.com/core/app/stats"
	"v2ray.com/core/common/task"
	featureStats "v2ray.com/core/features/stats"
)

// CounterRate is an implementation of stats.CounterRate.
type CounterRate struct {
	value int64
	rate int64
}

// Value implements stats.CounterRate.
func (c *CounterRate) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

func (c *CounterRate) Rate() int64 {
	return atomic.LoadInt64(&c.rate)
}

// Set implements stats.CounterRate.
func (c *CounterRate) Set(newValue int64) int64 {
	return atomic.SwapInt64(&c.value, newValue)
}
func (c *CounterRate) SetRate(newValue int64) int64 {
	return atomic.SwapInt64(&c.rate, newValue)
}

// Add implements stats.CounterRate.
func (c *CounterRate) Add(delta int64) int64 {
	return atomic.AddInt64(&c.value, delta)
}

// RateManager is an implementation of stats.RateManager.
type RateManager struct {
	access   sync.RWMutex
	counters map[string]*CounterRate
	rateTask *task.Periodic
	sm *stats.Manager
}



func (m *RateManager) ResetCounter(name string)  {
	if _,ok := m.counters[name]; !ok {
		return
	}
	m.sm.GetCounter(name).Set(0)
	m.counters[name].Set(0)
	m.counters[name].SetRate(0)
}

func (m *RateManager) Visit(visitor func(string, CounterRate) bool) {
	m.access.RLock()
	defer m.access.RUnlock()

	for name, c := range m.counters {
		if !visitor(name, *c) {
			break
		}
	}
}

// Start implements common.Runnable.
func (m *RateManager) Start() error {
	m.rateTask = &task.Periodic{
		Execute:  m.rateMonitor,
		Interval: time.Second * 1,
	}
	m.rateTask.Start()
	return nil
}

func (m *RateManager) rateMonitor() error {
	m.access.Lock()
	defer m.access.Unlock()
	m.sm.VisitCounters(func(s string, counter featureStats.Counter) bool{
		if _,ok := m.counters[s]; !ok {
			m.counters[s] = &CounterRate{}
		}
		rateCount := m.counters[s]

		rate := counter.Value()
		rateCount.Add(rate)
		rateCount.SetRate(rate)
		counter.Set(0)
		return true
	})
	return nil
}

// Close implement common.Closable.
func (m *RateManager) Close() error {
	m.rateTask.Close()
	return nil
}
