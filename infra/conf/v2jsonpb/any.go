package v2jsonpb

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/serial"
)

type AnyHolder struct {
	proto.Message
}

type resolver struct {
	backgroundResolver jsonpb.AnyResolver
}

func (r resolver) Resolve(typeURL string) (proto.Message, error) {
	obj, err := r.backgroundResolver.Resolve(typeURL)
	if err != nil {
		return nil, err
	}
	return AnyHolder{obj}, nil

}

func NewV2JsonPBResolver() jsonpb.AnyResolver {
	return &resolver{backgroundResolver: serial.GetResolver()}
}
