// +build !confonly

package stats

import (
	"context"

	"github.com/v2fly/v2ray-core/common"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewManager(ctx, config.(*Config))
	}))
}
