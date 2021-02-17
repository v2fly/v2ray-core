package api

import (
	"testing"

	statsService "github.com/v2fly/v2ray-core/v4/app/stats/command"
)

func TestEmptyResponese_0(t *testing.T) {
	r := &statsService.QueryStatsResponse{
		Stat: []*statsService.Stat{
			{
				Name:  "1>>2",
				Value: 1,
			},
			{
				Name:  "1>>2>>3",
				Value: 2,
			},
		},
	}
	assert(t, isEmpty(r), false)
}

func TestEmptyResponese_1(t *testing.T) {
	r := (*statsService.QueryStatsResponse)(nil)
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_2(t *testing.T) {
	r := &statsService.QueryStatsResponse{
		Stat: nil,
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_3(t *testing.T) {
	r := &statsService.QueryStatsResponse{
		Stat: []*statsService.Stat{},
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_4(t *testing.T) {
	r := &statsService.QueryStatsResponse{
		Stat: []*statsService.Stat{
			{
				Name:  "",
				Value: 0,
			},
		},
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_5(t *testing.T) {
	type test struct {
		Value *statsService.QueryStatsResponse
	}
	r := &test{
		Value: &statsService.QueryStatsResponse{
			Stat: []*statsService.Stat{
				{
					Name: "",
				},
			},
		},
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_6(t *testing.T) {
	type test struct {
		Value *statsService.QueryStatsResponse
	}
	r := &test{
		Value: &statsService.QueryStatsResponse{
			Stat: []*statsService.Stat{
				{
					Value: 1,
				},
			},
		},
	}
	assert(t, isEmpty(r), false)
}

func TestEmptyResponese_7(t *testing.T) {
	type test struct {
		Value *int
	}
	v := 1
	r := &test{
		Value: &v,
	}
	assert(t, isEmpty(r), false)
}

func TestEmptyResponese_8(t *testing.T) {
	type test struct {
		Value *int
	}
	v := 0
	r := &test{
		Value: &v,
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_9(t *testing.T) {
	assert(t, isEmpty(0), true)
}

func TestEmptyResponese_10(t *testing.T) {
	assert(t, isEmpty(1), false)
}

func TestEmptyResponese_11(t *testing.T) {
	r := []*statsService.Stat{
		{
			Name: "",
		},
	}
	assert(t, isEmpty(r), true)
}

func TestEmptyResponese_12(t *testing.T) {
	r := []*statsService.Stat{
		{
			Value: 1,
		},
	}
	assert(t, isEmpty(r), false)
}

func assert(t *testing.T, value, expected bool) {
	if value != expected {
		t.Fatalf("Expected: %v, actual: %v", expected, value)
	}
}
