package serial

import (
	"errors"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const V2RayTypeURLHeader = "types.v2fly.org/"

// ToTypedMessage converts a proto Message into TypedMessage.
func ToTypedMessage(message proto.Message) *anypb.Any {
	if message == nil {
		return nil
	}
	settings, _ := proto.Marshal(message)
	return &anypb.Any{
		TypeUrl: V2RayTypeURLHeader + GetMessageType(message),
		Value:   settings,
	}
}

// GetMessageType returns the name of this proto Message.
func GetMessageType(message proto.Message) string {
	return proto.MessageName(message)
}

// GetInstance creates a new instance of the message with messageType.
func GetInstance(messageType string) (interface{}, error) {
	mType := proto.MessageType(messageType)
	if mType == nil || mType.Elem() == nil {
		return nil, errors.New("Serial: Unknown type: " + messageType)
	}
	return reflect.New(mType.Elem()).Interface(), nil
}

func GetInstanceOf(v *anypb.Any) (proto.Message, error) {
	instance, err := GetInstance(V2TypeFromURL(v.TypeUrl))
	if err != nil {
		return nil, err
	}
	protoMessage := instance.(proto.Message)
	if err := proto.Unmarshal(v.Value, protoMessage); err != nil {
		return nil, err
	}
	return protoMessage, nil
}

func V2Type(v *anypb.Any) string {
	return V2TypeFromURL(v.TypeUrl)
}

func V2TypeFromURL(string2 string) string {
	return strings.TrimPrefix(string2, V2RayTypeURLHeader)
}

func V2TypeHumanReadable(v *anypb.Any) string {
	return v.TypeUrl
}

func V2URLFromV2Type(readableType string) string {
	return readableType
}
