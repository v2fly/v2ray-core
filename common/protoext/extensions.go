package protoext

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func GetMessageOptions(msgDesc protoreflect.MessageDescriptor) (*MessageOpt, error) {
	msgOpt := msgDesc.Options().(*descriptorpb.MessageOptions)
	msgOptRet := proto.GetExtension(msgOpt, E_MessageOpt)
	return msgOptRet.(*MessageOpt), nil
}

func GetFieldOptions(fieldDesc protoreflect.FieldDescriptor) (*FieldOpt, error) {
	fieldOpt := fieldDesc.Options().(*descriptorpb.FieldOptions)
	msgOptRet := proto.GetExtension(fieldOpt, E_FieldOpt)
	return msgOptRet.(*FieldOpt), nil
}
