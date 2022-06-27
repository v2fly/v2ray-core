package v2jsonpb

import (
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type anyresolverv2 struct {
	backgroundResolver jsonpb.AnyResolver
}

func (r anyresolverv2) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	panic("implement me")
}

func (r anyresolverv2) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	msg, err := r.backgroundResolver.Resolve(url)
	if err != nil {
		return nil, err
	}
	return msg.(proto.Message).ProtoReflect().Type(), nil
}

func (r anyresolverv2) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	panic("implement me")
}

func (r anyresolverv2) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	panic("implement me")
}
