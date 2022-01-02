package v4

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/app/observatory"
	"github.com/v2fly/v2ray-core/v5/app/observatory/burst"
	"github.com/v2fly/v2ray-core/v5/app/observatory/multiobservatory"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/taggedfeatures"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/duration"
	"github.com/v2fly/v2ray-core/v5/infra/conf/synthetic/router"
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
	HealthCheck *router.HealthCheckSettings `json:"pingConfig,omitempty"`
}

func (b BurstObservatoryConfig) Build() (proto.Message, error) {
	result, err := b.HealthCheck.Build()
	if err == nil {
		return &burst.Config{SubjectSelector: b.SubjectSelector, PingConfig: result.(*burst.HealthPingConfig)}, nil
	}
	return nil, err
}

type MultiObservatoryItem struct {
	MemberType string          `json:"type"`
	Tag        string          `json:"tag"`
	Value      json.RawMessage `json:"settings"`
}

type MultiObservatoryConfig struct {
	Observers []MultiObservatoryItem `json:"observers"`
}

func (o *MultiObservatoryConfig) Build() (proto.Message, error) {
	ret := &multiobservatory.Config{Holders: &taggedfeatures.Config{Features: make(map[string]*anypb.Any)}}
	for _, v := range o.Observers {
		switch v.MemberType {
		case "burst":
			var burstObservatoryConfig BurstObservatoryConfig
			err := json.Unmarshal(v.Value, &burstObservatoryConfig)
			if err != nil {
				return nil, err
			}
			burstObservatoryConfigPb, err := burstObservatoryConfig.Build()
			if err != nil {
				return nil, err
			}
			ret.Holders.Features[v.Tag] = serial.ToTypedMessage(burstObservatoryConfigPb)
		case "default":
			fallthrough
		default:
			var observatoryConfig ObservatoryConfig
			err := json.Unmarshal(v.Value, &observatoryConfig)
			if err != nil {
				return nil, err
			}
			observatoryConfigPb, err := observatoryConfig.Build()
			if err != nil {
				return nil, err
			}
			ret.Holders.Features[v.Tag] = serial.ToTypedMessage(observatoryConfigPb)
		}
	}
	return ret, nil
}
