package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/observatory"
)

type ObservatoryConfig struct {
	SubjectSelector []string `json:"subjectSelector"`
	ProbeURL        string   `json:"probeURL"`
}

func (o *ObservatoryConfig) Build() (proto.Message, error) {
	return &observatory.Config{SubjectSelector: o.SubjectSelector, ProbeUrl: o.ProbeURL}, nil
}
