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
	}, "name", "settings")
)

type strategyEmptyConfig struct {
}

func (v *strategyEmptyConfig) Build() (proto.Message, error) {
	return nil, nil
}

type strategyLeastLoadConfig struct {
	// ping rtt baselines (ms)
	Baselines []int `json:"baselines"`
	// minimal nodes count to select
	MinNodes int32 `json:"minNodes"`
}

// Build implements Buildable.
func (v *strategyLeastLoadConfig) Build() (proto.Message, error) {
	config := new(router.StrategyLeastLoadConfig)
	config.MinNodes = v.MinNodes
	config.Baselines = make([]int64, 0)
	for _, b := range v.Baselines {
		config.Baselines = append(config.Baselines, int64(time.Duration(b)*time.Millisecond))
	}
	return config, nil
}
