package conf

import (
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/proxy/lo"
)

type LoConfig struct {
	InboundTag string `json:"inboundTag"`
}

func (l LoConfig) Build() (proto.Message, error) {
	return &lo.Config{InboundTag: l.InboundTag}, nil
}
