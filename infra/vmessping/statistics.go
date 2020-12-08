package vmessping

import (
	"fmt"
	"time"
)

// PingStat represents a ping statistics
type PingStat struct {
	StartTime  time.Time
	SumMs      uint
	MaxMs      uint
	MinMs      uint
	AvgMs      uint
	Delays     []int64
	ReqCounter uint
	ErrCounter uint
}

// CalStats calculates ping statistics
func (p *PingStat) CalStats() {
	for _, v := range p.Delays {
		p.SumMs += uint(v)
		if p.MaxMs == 0 || p.MinMs == 0 {
			p.MaxMs = uint(v)
			p.MinMs = uint(v)
		}
		if uv := uint(v); uv > p.MaxMs {
			p.MaxMs = uv
		}
		if uv := uint(v); uv < p.MinMs {
			p.MinMs = uv
		}
	}
	if len(p.Delays) > 0 {
		p.AvgMs = uint(float64(p.SumMs) / float64(len(p.Delays)))
	}
}

// PrintStats prints ping statistics
func (p PingStat) PrintStats() {
	fmt.Println("\n--- vmess ping statistics ---")
	fmt.Printf("%d requests made, %d success, total time %v\n", p.ReqCounter, len(p.Delays), time.Since(p.StartTime))
	fmt.Printf("rtt min/avg/max = %d/%d/%d ms\n", p.MinMs, p.AvgMs, p.MaxMs)
}

// IsErr returns true if no delay records
func (p PingStat) IsErr() bool {
	return len(p.Delays) == 0
}
