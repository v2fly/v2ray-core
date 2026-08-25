package registry

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

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
	return enforceRestrictionWithOptions(
		config,
		msgOpts.GetAllowRestrictedModeLoad(),
		msgOpts.GetAllowRestrictedModeLoadIfSet(),
	)
}

func enforceRestrictionWithOptions(config proto.Message, allowRestrictedModeLoad bool, allowRestrictedModeLoadIfSet string) error {
	if allowRestrictedModeLoad {
		return nil
	}
	if allowRestrictedModeLoadIfSet == "" {
		return newError("component has not opted in for load in restricted mode")
	}

	message := config.ProtoReflect()
	field := message.Descriptor().Fields().ByName(protoreflect.Name(allowRestrictedModeLoadIfSet))
	if field == nil {
		return newError("allow_restricted_mode_load_if_set references unknown field: ", allowRestrictedModeLoadIfSet)
	}
	if field.Cardinality() == protoreflect.Repeated || field.Kind() != protoreflect.BoolKind {
		return newError("allow_restricted_mode_load_if_set must reference a singular bool field: ", allowRestrictedModeLoadIfSet)
	}
	if !message.Has(field) || !message.Get(field).Bool() {
		return newError("component has not opted in for load in restricted mode")
	}
	return nil
}
