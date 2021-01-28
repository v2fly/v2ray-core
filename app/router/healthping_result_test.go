package router_test

import (
	"math"
	reflect "reflect"
	"testing"
	"time"

	"v2ray.com/core/app/router"
)

func TestHealthPingResults(t *testing.T) {
	rtts := []int64{60, 140, 60, 140, 60, 60, 140, 60, 140}
	hr := router.NewHealthPingResult(4, time.Hour)
	for _, rtt := range rtts {
		hr.Put(time.Duration(rtt))
	}
	rttFailed := time.Duration(math.MaxInt64)
	expected := &router.HealthPingStats{
		Count:        4,
		FailCount:    0,
		RTTDeviation: 40,
		RTTAverage:   100,
		RTTMax:       140,
		RTTMin:       60,
	}
	actual := hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
	hr.Put(rttFailed)
	hr.Put(rttFailed)
	expected.FailCount = 2
	actual = hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("failed half-failures test, expected: %v, actual: %v", expected, actual)
	}
	hr.Put(rttFailed)
	hr.Put(rttFailed)
	expected = &router.HealthPingStats{
		Count:        4,
		FailCount:    4,
		RTTDeviation: 0,
		RTTAverage:   0,
		RTTMax:       0,
		RTTMin:       0,
	}
	actual = hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("failed all-failures test, expected: %v, actual: %v", expected, actual)
	}
}

func TestHealthPingResultsIgnoreOutdated(t *testing.T) {
	rtts := []int64{60, 140, 60, 140}
	hr := router.NewHealthPingResult(4, time.Duration(10)*time.Millisecond)
	for i, rtt := range rtts {
		if i == 2 {
			// wait for previous 2 outdated
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
		hr.Put(time.Duration(rtt))
	}
	hr.Get()
	expected := &router.HealthPingStats{
		Count:        2,
		FailCount:    0,
		RTTDeviation: 40,
		RTTAverage:   100,
		RTTMax:       140,
		RTTMin:       60,
	}
	actual := hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("failed 'half-outdated' test, expected: %v, actual: %v", expected, actual)
	}
	// wait for all outdated
	time.Sleep(time.Duration(10) * time.Millisecond)
	expected = &router.HealthPingStats{
		Count:        0,
		FailCount:    0,
		RTTDeviation: 0,
		RTTAverage:   0,
		RTTMax:       0,
		RTTMin:       0,
	}
	actual = hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("failed 'outdated / not-tested' test, expected: %v, actual: %v", expected, actual)
	}

	hr.Put(time.Duration(60))
	expected = &router.HealthPingStats{
		Count:     1,
		FailCount: 0,
		// 1 sample, std=0.5rtt
		RTTDeviation: 30,
		RTTAverage:   60,
		RTTMax:       60,
		RTTMin:       60,
	}
	actual = hr.Get()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
}
