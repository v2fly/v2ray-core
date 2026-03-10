package v2binding

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/v2fly/v2ray-core/v5/features/extension"
)

func (b *bindingInstance) setBaseDir(baseDir string) {
	b.baseDir = baseDir
	attentionFile := path.Join(baseDir, "place your config file here.txt")
	{
		f, err := os.OpenFile(attentionFile, os.O_RDONLY|os.O_CREATE, 0o666)
		if err != nil {
			return
		}
		_ = f.Close()
	}
	_ = os.Chdir(baseDir)
}

func (b *bindingInstance) loadDefaultConfigIfExists() error {
	config, err := os.ReadFile(path.Join(b.baseDir, "config.json"))
	if err != nil {
		return err
	}

	instanceManagement := b.instance.GetFeature(extension.InstanceManagementType())
	if instanceManagement == nil {
		return fmt.Errorf("instance management type not found")
	}
	instance, ok := instanceManagement.(extension.InstanceManagement)
	if !ok {
		return fmt.Errorf("instance management instance is invalid")
	}
	ctx := context.TODO()
	err = instance.AddInstance(ctx, "default", config, "jsonv5")
	if err != nil {
		return err
	}
	err = instance.StartInstance(ctx, "default")
	if err != nil {
		return err
	}
	return nil
}
