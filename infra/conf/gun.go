package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/transport/internet/grpc"
)

type GunConfig struct {
	ServiceName         string `json:"serviceName"`
	Mode                string `json:"mode"`
	IdleTimeout         int32  `json:"idle_timeout"`
	HealthCheckTimeout  int32  `json:"health_check_timeout"`
	PermitWithoutStream bool   `json:"permit_without_stream"`
}

func (g GunConfig) Build() (proto.Message, error) {
	var mode grpc.Mode
	switch g.Mode {
	case "", "gun":
		mode = grpc.Mode_Gun
	case "multi":
		mode = grpc.Mode_Multi
	case "raw":
		mode = grpc.Mode_Raw
	default:
		return nil, newError("undefined grpc mode: ", g.Mode)
	}

	if g.IdleTimeout <= 0 {
		g.IdleTimeout = 0
	}
	if g.HealthCheckTimeout <= 0 {
		g.HealthCheckTimeout = 0
	}

	return &grpc.Config{
		ServiceName:         g.ServiceName,
		Mode:                mode,
		IdleTimeout:         g.IdleTimeout,
		HealthCheckTimeout:  g.HealthCheckTimeout,
		PermitWithoutStream: g.PermitWithoutStream,
	}, nil
}
