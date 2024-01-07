package serial

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type AnyResolver interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

type serialResolver struct{}

func (s serialResolver) FindMessageByName(messageName protoreflect.FullName) (protoreflect.MessageType, error) {
	return protoregistry.GlobalTypes.FindMessageByName(messageName)
}

func (s serialResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	return protoregistry.GlobalTypes.FindMessageByURL(url)
}

func (s serialResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	return protoregistry.GlobalTypes.FindExtensionByName(field)
}

func (s serialResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	return protoregistry.GlobalTypes.FindExtensionByNumber(message, field)
}

func GetResolver() AnyResolver {
	return &serialResolver{}
}
