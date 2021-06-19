package serial

import (
	"errors"
	"google.golang.org/protobuf/types/known/anypb"
	"reflect"

	"github.com/golang/protobuf/proto"
)

// ToTypedMessage converts a proto Message into TypedMessage.
func ToTypedMessage(message proto.Message) *anypb.Any {
	if message == nil {
		return nil
	}
	settings, _ := proto.Marshal(message)
	return &anypb.Any{
		TypeUrl: GetMessageType(message),
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
	instance, err := GetInstance(v.TypeUrl)
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
	return v.TypeUrl
}
