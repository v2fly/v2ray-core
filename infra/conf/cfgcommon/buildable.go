package cfgcommon

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Buildable interface {
	Build() (proto.Message, error)
}

type BuildableV5 interface {
	BuildV5(ctx context.Context) (proto.Message, error)
}
