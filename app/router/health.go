package router

import (
	"fmt"
	"time"
)

// HealthChecker is the health checker for balancers
type HealthChecker struct {
	Enabled     bool
	Destination string
	Round       int
	Timeout     time.Duration
	// TODO: checkers
}

// StartHealthCheck start the health checker
func (b *Balancer) StartHealthCheck() error {
	if !b.healthChecker.Enabled {
		return nil
	}
	// TODO: start cheker
	fmt.Println("todo: start checker", b.healthChecker)
	return nil
}

// StopHealthCheck stop the health checker
func (b *Balancer) StopHealthCheck() error {
	// TODO: stop cheker
	fmt.Println("todo: stop checker", b.healthChecker)
	return nil
}
