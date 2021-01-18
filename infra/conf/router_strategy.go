package conf

import (
	"time"

	"google.golang.org/protobuf/proto"
	"v2ray.com/core/app/router"
)

const (
	strategyRandom    string = "random"
	strategyLeastLoad string = "leastload"
)

var (
	strategyConfigLoader = NewJSONConfigLoader(ConfigCreatorCache{
		strategyRandom:    func() interface{} { return new(strategyEmptyConfig) },
		strategyLeastLoad: func() interface{} { return new(strategyLeastLoadConfig) },
	}, "type", "settings")
)

type strategyEmptyConfig struct {
}

func (v *strategyEmptyConfig) Build() (proto.Message, error) {
	return nil, nil
}

type strategyLeastLoadConfig struct {
	// note the time values of the HealthCheck holds is not
	// 'time.Duration' but plain number, sice they were parsed
	// directly from json
	HealthCheck *router.HealthPingSettings `json:"healthCheck"`
	// ping rtt baselines (ms)
	Baselines []int `json:"baselines"`
	// expected nodes count to select
	Expected int32 `json:"expected"`
}

// Build implements Buildable.
func (v *strategyLeastLoadConfig) Build() (proto.Message, error) {
	config := &router.StrategyLeastLoadConfig{
		HealthCheck: &router.HealthPingConfig{},
	}
	if v.HealthCheck != nil {
		config.HealthCheck = &router.HealthPingConfig{
			Destination: v.HealthCheck.Destination,
			Interval:    int64(v.HealthCheck.Interval * time.Second),
			Timeout:     int64(v.HealthCheck.Timeout * time.Second),
			Rounds:      int32(v.HealthCheck.Rounds),
		}
	}
	config.Expected = v.Expected
	if config.Expected < 0 {
		config.Expected = 0
	}
	config.Baselines = make([]int64, 0)
	for _, b := range v.Baselines {
		if b <= 0 {
			continue
		}
		config.Baselines = append(config.Baselines, int64(time.Duration(b)*time.Millisecond))
	}
	return config, nil
}
