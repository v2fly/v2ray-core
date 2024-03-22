package serial

import (
	"errors"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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
	return string(message.ProtoReflect().Descriptor().FullName())
}

// GetMessageDescriptor returns the MessageDescriptor of the message with fullName.
func GetMessageDescriptor(fullName string) (protoreflect.MessageDescriptor, error) {
	mt, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(fullName))
	if err != nil {
		return nil, errors.New("Serial: Unknown message name: " + fullName)
	}

	message, ok := mt.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, errors.New("Serial: Message with name: " + fullName + " is not a MessageDescriptor")
	}
	return message, nil
}

// GetInstance creates a new instance of the message with messageType.
func GetInstance(messageType string) (proto.Message, error) {
	// mType := proto.MessageType(messageType)
	mType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(messageType))
	if err != nil {
		return nil, errors.New("Serial: Unknown type: " + messageType)
	}

	return mType.New().Interface(), nil
}

func GetInstanceOf(v *anypb.Any) (proto.Message, error) {
	instance, err := GetInstance(V2TypeFromURL(v.TypeUrl))
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(v.Value, instance); err != nil {
		return nil, err
	}
	return instance, nil
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
