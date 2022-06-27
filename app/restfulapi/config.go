package restfulapi

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return newRestfulService(ctx, config.(*Config))
	}))
}
