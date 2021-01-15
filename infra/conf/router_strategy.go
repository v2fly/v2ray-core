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
	// ping rtt baselines (ms)
	Baselines []int `json:"baselines"`
	// expected nodes count to select
	Expected int32 `json:"expected"`
}

// Build implements Buildable.
func (v *strategyLeastLoadConfig) Build() (proto.Message, error) {
	config := new(router.StrategyLeastLoadConfig)
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
