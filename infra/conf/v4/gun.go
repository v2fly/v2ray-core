package v4

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/transport/internet/grpc"
)

type GunConfig struct {
	ServiceName         string `json:"serviceName"`
	IdleTimeout         int32  `json:"idle_timeout"`
	HealthCheckTimeout  int32  `json:"health_check_timeout"`
	PermitWithoutStream bool   `json:"permit_without_stream"`
	InitialWindowsSize  int32  `json:"initial_windows_size"`
}

func (g GunConfig) Build() (proto.Message, error) {
	if g.IdleTimeout <= 0 {
		g.IdleTimeout = 0
	}
	if g.HealthCheckTimeout <= 0 {
		g.HealthCheckTimeout = 0
	}
	if g.InitialWindowsSize < 0 {
		g.InitialWindowsSize = 0
	}
	return &grpc.Config{
		ServiceName:         g.ServiceName,
		IdleTimeout:         g.IdleTimeout,
		HealthCheckTimeout:  g.HealthCheckTimeout,
		PermitWithoutStream: g.PermitWithoutStream,
		InitialWindowsSize:  g.InitialWindowsSize,
	}, nil
}
