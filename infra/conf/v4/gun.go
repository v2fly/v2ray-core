package v4

import (
	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/transport/internet/grpc"
)

type GunConfig struct {
	ServiceName string `json:"serviceName"`
}

func (g GunConfig) Build() (proto.Message, error) {
	return &grpc.Config{ServiceName: g.ServiceName}, nil
}
