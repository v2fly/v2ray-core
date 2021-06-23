package protoext

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func GetMessageOptions(msgDesc protoreflect.MessageDescriptor) (*MessageOpt, error) {
	msgOpt := msgDesc.Options().(*descriptorpb.MessageOptions)
	var V2MessageOption *MessageOpt
	if msgOptRet, err := proto.GetExtension(msgOpt, E_MessageOpt); err != nil {
		return nil, newError("unable to parse extension from message").Base(err)
	} else {
		V2MessageOption = msgOptRet.(*MessageOpt)
	}
	return V2MessageOption, nil
}

func GetFieldOptions(fieldDesc protoreflect.FieldDescriptor) (*FieldOpt, error) {
	fieldOpt := fieldDesc.Options().(*descriptorpb.FieldOptions)
	var V2FieldOption *FieldOpt
	if msgOptRet, err := proto.GetExtension(fieldOpt, E_FieldOpt); err != nil {
		return nil, newError("unable to parse extension from message").Base(err)
	} else {
		V2FieldOption = msgOptRet.(*FieldOpt)
	}
	return V2FieldOption, nil
}
