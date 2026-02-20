package happyEyeball

import (
	"context"
	"fmt"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/common/task/taskDerive"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type Dialer func(ctx context.Context, domainDestination net.Destination, ips net.IP) (internet.Connection, error)

func RacingDialer(ctx context.Context, domainDestination net.Destination, ips []net.IP, dialer Dialer, preferIPv6 bool, preferredHeadStart time.Duration) (internet.Connection, error) {
	// check if they are of a single family, if so no one have head start
	hasIPv4 := false
	hasIPv6 := false
	for _, a := range ips {
		if a == nil {
			continue
		}
		switch a.To4() != nil {
		case true:
			hasIPv4 = true
		case false:
			hasIPv6 = true
		}
	}

	// If there is only one family present, there is no head start
	if !hasIPv4 || !hasIPv6 {
		preferredHeadStart = 0
	}
	if preferredHeadStart < 0 {
		preferredHeadStart = 0
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses to dial")
	}

	connCh := make(chan internet.Connection, 1)
	finished := done.New()

	tasks := make([]func() error, 0, len(ips))
	for _, a := range ips {
		addr := a
		// determine delay for this address
		delay := time.Duration(0)
		if preferredHeadStart > 0 && addr != nil {
			isIPv6 := a.To4() == nil
			// if preferIPv6 is true, IPv4 gets the head start delay (i.e. IPv6 starts earlier)
			if preferIPv6 {
				if !isIPv6 {
					delay = preferredHeadStart
				}
			} else {
				// prefer IPv4, so IPv6 gets delayed
				if isIPv6 {
					delay = preferredHeadStart
				}
			}
		}

		// capture addr and delay in closure
		tasks = append(tasks, func() error {
			if delay > 0 {
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if finished.Done() {
				return fmt.Errorf("dial attempt cancelled due to another successful dial")
			}
			c, err := dialer(ctx, domainDestination, addr)
			if err != nil {
				return err
			}
			if c == nil {
				return fmt.Errorf("dialer returned nil connection")
			}
			// send the successful conn (non-blocking due to buffer)
			select {
			case connCh <- c:
				// stored
			default:
				// channel already has a conn, close this extra one
				_ = c.Close()
			}
			_ = finished.Close()
			return nil
		})
	}

	errs := taskDerive.RunTryAll(ctx, tasks...)
	if errs == nil {
		// at least one dial succeeded; return the conn
		select {
		case c := <-connCh:
			return c, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// theoretically should not happen, but handle defensively
			return nil, fmt.Errorf("dial succeeded but no connection available")
		}
	}
	// all attempts failed, return the first non-nil error
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return nil, fmt.Errorf("all dial attempts failed")
}
