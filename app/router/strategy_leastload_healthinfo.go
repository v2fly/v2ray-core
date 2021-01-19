package router

import (
	"fmt"
	"sort"
	"time"

	"v2ray.com/core/features/routing"
)

type healthExt struct {
	*routing.OutboundInfo
	rtt time.Duration
}

// getHealthPingInfo is the shared health info maker
// for HealthPing based strategies, like leastload
func getHealthPingInfo(tags []string, results map[string]*HealthPingResult) ([]string, []*routing.OutboundInfo) {
	failed := []string{"failed"}
	notTested := []string{"not tested"}
	items := make([]*healthExt, 0)
	for _, tag := range tags {
		item := &healthExt{
			OutboundInfo: &routing.OutboundInfo{
				Tag: tag,
			},
		}
		result, ok := results[tag]
		switch {
		case !ok || result.Count == 0:
			item.Values = notTested
			item.rtt = 0
		case result.FailCount > 0:
			item.Values = failed
			item.rtt = -1
		default:
			item.Values = []string{fmt.Sprint(result.AverageRTT)}
			item.rtt = result.AverageRTT
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		iRTT := items[i].rtt
		jRTT := items[j].rtt
		// 0 rtt means not checked or failed, sort in the tail
		if iRTT <= 0 && jRTT > 0 {
			return false
		}
		if iRTT > 0 && jRTT <= 0 {
			return true
		}
		return iRTT < jRTT
	})
	hs := make([]*routing.OutboundInfo, 0)
	for _, h := range items {
		hs = append(hs, h.OutboundInfo)
	}
	return []string{"RTT"}, hs
}
