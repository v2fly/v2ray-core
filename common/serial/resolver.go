package serial

import (
	"github.com/golang/protobuf/proto"
)

type AnyResolver interface {
	Resolve(typeURL string) (proto.Message, error)
}

type serialResolver struct{}

func (s serialResolver) Resolve(typeURL string) (proto.Message, error) {
	instance, err := GetInstance(typeURL)
	if err != nil {
		return nil, err
	}
	return instance.(proto.Message), nil
}

func GetResolver() AnyResolver {
	return &serialResolver{}
}
