package stats

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

// Counter is an implementation of stats.Counter.
type Counter struct {
	gauge prometheus.Gauge
	value int64
}

// Value implements stats.Counter.
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Set implements stats.Counter.
func (c *Counter) Set(newValue int64) int64 {
	c.gauge.Set(float64(newValue))
	return atomic.SwapInt64(&c.value, newValue)
}

// Add implements stats.Counter.
func (c *Counter) Add(delta int64) int64 {
	c.gauge.Add(float64(delta))
	return atomic.AddInt64(&c.value, delta)
}
