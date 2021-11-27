package v2jsonpb

import "google.golang.org/protobuf/reflect/protoreflect"

type V2JsonProtobufAnyTypeDescriptor struct {
	protoreflect.MessageDescriptor
}

func (v V2JsonProtobufAnyTypeDescriptor) FullName() protoreflect.FullName {
	return "org.v2fly.SynAny"
}

func (v V2JsonProtobufAnyTypeDescriptor) Fields() protoreflect.FieldDescriptors {
	return V2JsonProtobufAnyTypeFields{v.MessageDescriptor.Fields()}
}

type V2JsonProtobufAnyTypeFields struct {
	protoreflect.FieldDescriptors
}

func (v V2JsonProtobufAnyTypeFields) Len() int {
	panic("implement me")
}

func (v V2JsonProtobufAnyTypeFields) Get(i int) protoreflect.FieldDescriptor {
	panic("implement me")
}

func (v V2JsonProtobufAnyTypeFields) ByName(s protoreflect.Name) protoreflect.FieldDescriptor {
	panic("implement me")
}

func (v V2JsonProtobufAnyTypeFields) ByJSONName(s string) protoreflect.FieldDescriptor {
	switch s {
	case "type":
		return &V2JsonProtobufFollowerFieldDescriptor{v.FieldDescriptors.ByName("type_url")}
	default:
		return &V2JsonProtobufAnyValueField{v.FieldDescriptors.ByName("value"), "value"}
	}
}

func (v V2JsonProtobufAnyTypeFields) ByTextName(s string) protoreflect.FieldDescriptor {
	panic("implement me")
}

func (v V2JsonProtobufAnyTypeFields) ByNumber(n protoreflect.FieldNumber) protoreflect.FieldDescriptor {
	panic("implement me")
}

type V2JsonProtobufAnyTypeFieldDescriptor struct {
	protoreflect.FieldDescriptor
}

func (v V2JsonProtobufAnyTypeFieldDescriptor) JSONName() string {
	return "type"
}

func (v V2JsonProtobufAnyTypeFieldDescriptor) TextName() string {
	return "type"
}

type V2JsonProtobufAnyValueField struct {
	protoreflect.FieldDescriptor
	name string
}

func (v *V2JsonProtobufAnyValueField) Kind() protoreflect.Kind {
	return protoreflect.MessageKind
}

func (v *V2JsonProtobufAnyValueField) JSONName() string {
	return v.name
}

func (v *V2JsonProtobufAnyValueField) TextName() string {
	return v.name
}

type V2JsonProtobufAnyValueFieldReturn struct {
	protoreflect.Message
}

func (v *V2JsonProtobufAnyValueFieldReturn) ProtoReflect() protoreflect.Message {
	if bufFollow, ok := v.Message.(*V2JsonProtobufFollower); ok {
		return bufFollow.Message
	}
	return v.Message
}
