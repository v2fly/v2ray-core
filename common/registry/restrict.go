package registry

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/protoext"
)

const restrictedLoadModeCtx = "restrictedLoadModeCtx"

func CreateRestrictedModeContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, restrictedLoadModeCtx, true) //nolint: staticcheck
}

func isRestrictedModeContext(ctx context.Context) bool {
	v := ctx.Value(restrictedLoadModeCtx)
	if v == nil {
		return false
	}
	return v.(bool)
}

func enforceRestriction(config proto.Message) error {
	configDescriptor := config.ProtoReflect().Descriptor()
	msgOpts, err := protoext.GetMessageOptions(configDescriptor)
	if err != nil {
		return newError("unable to find message options").Base(err)
	}
	if !msgOpts.AllowRestrictedModeLoad {
		return newError("component has not opted in for load in restricted mode")
	}
	return nil
}
