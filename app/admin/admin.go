// +build !confonly

package admin

import (
	"context"
	"github.com/gin-gonic/gin"
	"v2ray.com/core/common"
)

func init() {
	gin.DefaultWriter = ErrorLoggerWriter
	gin.DefaultErrorWriter = ErrorLoggerWriter
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		return NewAdminServer(ctx, cfg.(*Config))
	}))
}
