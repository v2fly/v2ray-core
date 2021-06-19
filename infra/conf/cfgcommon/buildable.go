package cfgcommon

import "github.com/golang/protobuf/proto"

type Buildable interface {
	Build() (proto.Message, error)
}
