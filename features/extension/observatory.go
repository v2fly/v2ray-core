package extension

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/features"
)

type Observatory interface {
	features.Feature
	GetObservation(ctx context.Context) (proto.Message, error)
}

func ObservatoryType() interface{} {
	return (*Observatory)(nil)
}
