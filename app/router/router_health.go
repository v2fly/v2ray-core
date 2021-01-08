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

// GetBalancersInfo implements routing.HealthChecker.
func (r *Router) GetBalancersInfo(tags []string) (resp []*routing.BalancerInfo, err error) {
	resp = make([]*routing.BalancerInfo, 0)
	for t, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		if len(tags) > 0 && findSliceIndex(tags, t) < 0 {
			continue
		}
		s, err := b.strategy.GetInfo()
		if err != nil {
			return nil, err
		}
		stat := &routing.BalancerInfo{
			Tag:      t,
			Strategy: s,
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
