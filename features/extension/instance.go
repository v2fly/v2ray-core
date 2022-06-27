package extension

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/features"
)

// InstanceManagement : unstable
type InstanceManagement interface {
	features.Feature
	ListInstance(ctx context.Context) ([]string, error)
	AddInstance(ctx context.Context, name string, config []byte, configType string) error
	StartInstance(ctx context.Context, name string) error
	StopInstance(ctx context.Context, name string) error
	UntrackInstance(ctx context.Context, name string) error
}

func InstanceManagementType() interface{} {
	return (*InstanceManagement)(nil)
}
