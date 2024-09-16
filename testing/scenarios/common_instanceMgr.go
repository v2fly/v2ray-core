package scenarios

import (
	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/instman"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

func NewInstanceManagerInstanceConfig() *core.Config {
	config := &core.Config{}
	config.App = append(config.App, serial.ToTypedMessage(&instman.Config{}))
	return config
}

func NewInstanceManagerCoreInstance() (*core.Instance, extension.InstanceManagement) {
	coreConfig := NewInstanceManagerInstanceConfig()
	instance, err := core.New(coreConfig)
	if err != nil {
		panic(err)
	}
	common.Must(instance.Start())
	instanceMgr := instance.GetFeature(extension.InstanceManagementType())
	InstanceMgrIfce := instanceMgr.(extension.InstanceManagement)
	return instance, InstanceMgrIfce
}
