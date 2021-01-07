package router

import (
	"errors"
	"strings"

	"v2ray.com/core/features/routing"
)

// CheckHanlders implements routing.HealthChecker.
func (r *Router) CheckHanlders(tags []string) error {
	errs := make([]error, 0)
	for _, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		err := b.Check(tags)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return getCollectError(errs)
}

// CheckBalancers implements routing.HealthChecker.
func (r *Router) CheckBalancers(tags []string) error {
	errs := make([]error, 0)
	for _, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		_, err := b.CheckAll()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return getCollectError(errs)
}

func getCollectError(errs []error) error {
	sb := new(strings.Builder)
	sb.WriteString("collect errors:\n")
	for _, err := range errs {
		sb.WriteString("    * ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return errors.New(sb.String())
}

// GetHealthInfo implements routing.HealthChecker.
func (r *Router) GetHealthInfo(tags []string) ([]*routing.BalancerHealth, error) {
	resp := make([]*routing.BalancerHealth, 0)
	for t, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		if len(tags) > 0 {
			found := false
			for _, v := range tags {
				if t == v {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		stat := &routing.BalancerHealth{
			Balancer:  t,
			Selects:   make([]*routing.OutboundHealth, 0),
			Outbounds: make([]*routing.OutboundHealth, 0),
		}
		selects, err := b.strategy.SelectOutbounds()
		if err != nil {
			stat.Selects = append(stat.Selects, &routing.OutboundHealth{
				Outbound: err.Error(),
			})
		} else {
			stat.Selects = b.makeHealthStatItems(selects)
		}
		all, err := b.SelectOutbounds()
		if err != nil {
			stat.Outbounds = append(stat.Outbounds, &routing.OutboundHealth{
				Outbound: err.Error(),
			})
		} else {
			stat.Outbounds = b.makeHealthStatItems(all)
		}
		resp = append(resp, stat)
	}
	return resp, nil
}
