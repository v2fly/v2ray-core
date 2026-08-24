package instman

import (
	"context"
	"sync"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type InstanceMgr struct {
	config    *Config // nolint: structcheck
	access    *sync.Mutex
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
	i.access.Lock()
	defer i.access.Unlock()

	var instanceNames []string
	for k := range i.instances {
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
	i.access.Lock()
	defer i.access.Unlock()
	i.instances[name] = instance
	return nil
}

func (i InstanceMgr) getInstance(name string) (*core.Instance, error) {
	i.access.Lock()
	defer i.access.Unlock()

	instance, found := i.instances[name]
	if !found {
		return nil, newError("instance not found: ", name)
	}
	return instance, nil
}

func (i InstanceMgr) StartInstance(ctx context.Context, name string) error {
	instance, err := i.getInstance(name)
	if err != nil {
		return err
	}
	if err := instance.Start(); err != nil {
		return newError("failed to start instance").Base(err)
	}
	return nil
}

func (i InstanceMgr) StopInstance(ctx context.Context, name string) error {
	instance, err := i.getInstance(name)
	if err != nil {
		return err
	}
	if err := instance.Close(); err != nil {
		return newError("failed to stop instance").Base(err)
	}
	return nil
}

func (i InstanceMgr) UntrackInstance(ctx context.Context, name string) error {
	i.access.Lock()
	defer i.access.Unlock()

	delete(i.instances, name)
	return nil
}

func NewInstanceMgr(ctx context.Context, config *Config) (extension.InstanceManagement, error) {
	return InstanceMgr{access: &sync.Mutex{}, instances: map[string]*core.Instance{}}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		var f extension.InstanceManagement
		var err error
		if f, err = NewInstanceMgr(ctx, config.(*Config)); err != nil {
			return nil, err
		}
		return f, nil
	}))
}
