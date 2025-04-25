package router

import (
	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/app/observatory/burst"
	"github.com/ghxhy/v2ray-core/v5/app/router"
	"github.com/ghxhy/v2ray-core/v5/infra/conf/cfgcommon/duration"
	"github.com/ghxhy/v2ray-core/v5/infra/conf/cfgcommon/loader"
)

const (
	strategyRandom    string = "random"
	strategyLeastLoad string = "leastload"
	strategyLeastPing string = "leastping"
)

var strategyConfigLoader = loader.NewJSONConfigLoader(loader.ConfigCreatorCache{
	strategyRandom:    func() interface{} { return new(strategyRandomConfig) },
	strategyLeastLoad: func() interface{} { return new(strategyLeastLoadConfig) },
	strategyLeastPing: func() interface{} { return new(strategyLeastPingConfig) },
}, "type", "settings")

type strategyEmptyConfig struct{}

func (v *strategyEmptyConfig) Build() (proto.Message, error) {
	return nil, nil
}

type strategyLeastLoadConfig struct {
	// weight settings
	Costs []*router.StrategyWeight `json:"costs,omitempty"`
	// ping rtt baselines
	Baselines []duration.Duration `json:"baselines,omitempty"`
	// expected nodes count to select
	Expected int32 `json:"expected,omitempty"`
	// max acceptable rtt, filter away high delay nodes. default 0
	MaxRTT duration.Duration `json:"maxRTT,omitempty"`
	// acceptable failure rate
	Tolerance float64 `json:"tolerance,omitempty"`

	ObserverTag string `json:"observerTag,omitempty"`
}

// HealthCheckSettings holds settings for health Checker
type HealthCheckSettings struct {
	Destination   string            `json:"destination"`
	Connectivity  string            `json:"connectivity"`
	Interval      duration.Duration `json:"interval"`
	SamplingCount int               `json:"sampling"`
	Timeout       duration.Duration `json:"timeout"`
}

func (h HealthCheckSettings) Build() (proto.Message, error) {
	return &burst.HealthPingConfig{
		Destination:   h.Destination,
		Connectivity:  h.Connectivity,
		Interval:      int64(h.Interval),
		Timeout:       int64(h.Timeout),
		SamplingCount: int32(h.SamplingCount),
	}, nil
}

// Build implements Buildable.
func (v *strategyLeastLoadConfig) Build() (proto.Message, error) {
	config := &router.StrategyLeastLoadConfig{}
	config.Costs = v.Costs
	config.Tolerance = float32(v.Tolerance)
	config.ObserverTag = v.ObserverTag
	if config.Tolerance < 0 {
		config.Tolerance = 0
	}
	if config.Tolerance > 1 {
		config.Tolerance = 1
	}
	config.Expected = v.Expected
	if config.Expected < 0 {
		config.Expected = 0
	}
	config.MaxRTT = int64(v.MaxRTT)
	if config.MaxRTT < 0 {
		config.MaxRTT = 0
	}
	config.Baselines = make([]int64, 0)
	for _, b := range v.Baselines {
		if b <= 0 {
			continue
		}
		config.Baselines = append(config.Baselines, int64(b))
	}
	return config, nil
}

type strategyLeastPingConfig struct {
	ObserverTag string `json:"observerTag,omitempty"`
}

func (s strategyLeastPingConfig) Build() (proto.Message, error) {
	return &router.StrategyLeastPingConfig{ObserverTag: s.ObserverTag}, nil
}

type strategyRandomConfig struct {
	AliveOnly   bool   `json:"aliveOnly,omitempty"`
	ObserverTag string `json:"observerTag,omitempty"`
}

func (s strategyRandomConfig) Build() (proto.Message, error) {
	return &router.StrategyRandomConfig{ObserverTag: s.ObserverTag, AliveOnly: s.AliveOnly}, nil
}
