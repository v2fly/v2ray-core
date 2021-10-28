package cfgcommon

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type Buildable interface {
	Build() (proto.Message, error)
}

type BuildableV5 interface {
	BuildV5(ctx context.Context) (proto.Message, error)
}
