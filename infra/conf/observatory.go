package conf

import (
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/app/observatory/burst"

	"github.com/v2fly/v2ray-core/v4/app/observatory"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon/duration"
)

type ObservatoryConfig struct {
	SubjectSelector []string          `json:"subjectSelector"`
	ProbeURL        string            `json:"probeURL"`
	ProbeInterval   duration.Duration `json:"probeInterval"`
}

func (o *ObservatoryConfig) Build() (proto.Message, error) {
	return &observatory.Config{SubjectSelector: o.SubjectSelector, ProbeUrl: o.ProbeURL, ProbeInterval: int64(o.ProbeInterval)}, nil
}

type BurstObservatoryConfig struct {
	SubjectSelector []string `json:"subjectSelector"`
	// health check settings
	HealthCheck *healthCheckSettings `json:"pingConfig,omitempty"`
}

func (b BurstObservatoryConfig) Build() (proto.Message, error) {
	if result, err := b.HealthCheck.Build(); err != nil {
		return &burst.Config{SubjectSelector: b.SubjectSelector, PingConfig: result.(*burst.HealthPingConfig)}, nil
	} else {
		return nil, err
	}
}
