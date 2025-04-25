package restfulapi

import (
	"context"

	"github.com/ghxhy/v2ray-core/v5/common"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return newRestfulService(ctx, config.(*Config))
	}))
}
