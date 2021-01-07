package router

import "v2ray.com/core/features/routing"

// HealthCheck implements routing.HealthChecker.
func (r *Router) HealthCheck(tags []string) {
	for _, b := range r.balancers {
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		b.HealthCheck(tags)
	}
}

// GetHealthStats implements routing.HealthChecker.
func (r *Router) GetHealthStats(tag string) ([]*routing.HealthStats, error) {
	resp := make([]*routing.HealthStats, 0)
	for t, b := range r.balancers {
		if tag != "" && t != tag {
			continue
		}
		if !b.healthChecker.Settings.Enabled {
			continue
		}
		stat := &routing.HealthStats{
			Balancer:  t,
			Selects:   make([]*routing.HealthStatItem, 0),
			Outbounds: make([]*routing.HealthStatItem, 0),
		}
		selects, err := b.strategy.SelectOutbounds()
		if err != nil {
			stat.Selects = append(stat.Selects, &routing.HealthStatItem{
				Outbound: err.Error(),
			})
		} else {
			stat.Selects = b.makeHealthStatItems(selects)
		}
		all, err := b.SelectOutbounds()
		if err != nil {
			stat.Outbounds = append(stat.Outbounds, &routing.HealthStatItem{
				Outbound: err.Error(),
			})
		} else {
			stat.Outbounds = b.makeHealthStatItems(all)
		}
		resp = append(resp, stat)
	}
	return resp, nil
}
