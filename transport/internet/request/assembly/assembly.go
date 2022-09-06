package assembly

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const protocolName = "request"

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return nil, newError("request is a transport protocol.")
	}))
}
