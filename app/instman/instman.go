package instman

import (
	"context"
	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/features/extension"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type InstanceMgr struct {
	config    *Config
	instances map[string]*core.Instance
}

func (i InstanceMgr) Type() interface{} {
	return extension.InstanceManagementType()
}

func (i InstanceMgr) Start() error {
	return nil
}

func (i InstanceMgr) Close() error {
	return nil
}

func (i InstanceMgr) ListInstance(ctx context.Context) ([]string, error) {
	var instanceNames []string
	for k, _ := range i.instances {
		instanceNames = append(instanceNames, k)
	}
	return instanceNames, nil
}

func (i InstanceMgr) AddInstance(ctx context.Context, name string, config []byte, configType string) error {
	coreConfig, err := core.LoadConfig(configType, config)
	if err != nil {
		return newError("unable to load config").Base(err)
	}
	instance, err := core.New(coreConfig)
	if err != nil {
		return newError("unable to create instance").Base(err)
	}
	i.instances[name] = instance
	return nil
}

func (i InstanceMgr) StartInstance(ctx context.Context, name string) error {
	err := i.instances[name].Start()
	if err != nil {
		return newError("failed to start instance").Base(err)
	}
	return nil
}

func (i InstanceMgr) StopInstance(ctx context.Context, name string) error {
	err := i.instances[name].Close()
	if err != nil {
		return newError("failed to stop instance").Base(err)
	}
	return nil
}

func (i InstanceMgr) UntrackInstance(ctx context.Context, name string) error {
	delete(i.instances, name)
	return nil
}
