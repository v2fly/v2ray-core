package v5cfg

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/grpc"
)

func TestStreamConfigBuildV5GrpcParseXForwardedFor(t *testing.T) {
	config, err := (StreamConfig{
		Transport:         "grpc",
		TransportSettings: []byte(`{"serviceName":"svc","parseXForwardedFor":true}`),
	}).BuildV5(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	streamConfig := config.(*internet.StreamConfig)
	if len(streamConfig.TransportSettings) != 1 {
		t.Fatalf("unexpected transport settings count: %d", len(streamConfig.TransportSettings))
	}
	grpcConfig, err := serial.GetInstanceOf(streamConfig.TransportSettings[0].Settings)
	if err != nil {
		t.Fatal(err)
	}
	got := grpcConfig.(*grpc.Config)
	if !got.ParseXForwardedFor {
		t.Fatal("expected parseXForwardedFor to be enabled")
	}
}
