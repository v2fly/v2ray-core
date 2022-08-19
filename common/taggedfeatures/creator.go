package taggedfeatures

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features"
)

func NewHolderFromConfig(ctx context.Context, config *Config, memberType interface{}) (features.TaggedFeatures, error) {
	holder := NewHolder(ctx, memberType)
	for k, v := range config.Features {
		var err error
		instance, err := serial.GetInstanceOf(v)
		if err != nil {
			return nil, newError("unable to get instance").Base(err)
		}
		obj, err := common.CreateObject(ctx, instance)
		if err != nil {
			return nil, newError("unable to create object").Base(err)
		}

		if feat, ok := obj.(features.Feature); ok {
			err = holder.AddFeaturesByTag(k, feat)
			if err != nil {
				return nil, newError("unable to add feature").Base(err)
			}
			continue
		}
		return nil, newError("not a feature ", k)
	}
	return holder, nil
}
