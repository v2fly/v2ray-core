package mirrorcommon

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func SetLoopbackProtectionFlagForContext(ctx context.Context, enrollmentID []byte) context.Context {
	loopbackProtectionKey := tlsmirror.ConnectionLoopbackPrevention{Key: string(enrollmentID)}
	return context.WithValue(ctx, loopbackProtectionKey, true)
}

func IsLoopbackProtectionEnabled(ctx context.Context, enrollmentID []byte) bool {
	loopbackProtectionKey := tlsmirror.ConnectionLoopbackPrevention{Key: string(enrollmentID)}
	val := ctx.Value(loopbackProtectionKey)
	enabled, ok := val.(bool)
	return ok && enabled
}
