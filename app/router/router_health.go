package router

import (
	"errors"
	"strings"

	"v2ray.com/core/features/routing"
)

// CheckHanlders implements routing.HealthChecker.
func (r *Router) CheckHanlders(tags []string, distributed bool) error {
	errs := make([]error, 0)
	for _, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		err := b.Check(tags, distributed)
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
func (r *Router) CheckBalancers(tags []string, distributed bool) error {
	errs := make([]error, 0)
	for _, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		_, err := b.CheckAll(distributed)
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

// GetBalancersInfo implements routing.HealthChecker.
func (r *Router) GetBalancersInfo(tags []string) (resp []*routing.BalancerInfo, err error) {
	resp = make([]*routing.BalancerInfo, 0)
	for t, b := range r.balancers {
		if len(tags) > 0 && findSliceIndex(tags, t) < 0 {
			continue
		}
		all, err := b.SelectOutbounds()
		if err != nil {
			return nil, err
		}
		var s *routing.StrategyInfo
		if !b.healthChecker.Settings.Enabled {
			// if not enabled, should ignore the results, since they could be outdated
			s = b.strategy.GetInfo(all, nil)
		} else {
			b.healthChecker.access.Lock()
			s = b.strategy.GetInfo(all, b.healthChecker.Results)
			b.healthChecker.access.Unlock()
		}
		stat := &routing.BalancerInfo{
			Tag:         t,
			Strategy:    s,
			HealthCheck: b.healthChecker.Settings,
		}
		resp = append(resp, stat)
	}
	return resp, nil
}

func findSliceIndex(slice []string, find string) int {
	if len(slice) == 0 {
		return -1
	}
	index := -1
	for i, v := range slice {
		if find == v {
			index = i
			break
		}
	}
	return index
}
