package v2jsonpb

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

type resolver2 struct {
	backgroundResolver jsonpb.AnyResolver
}

func (r resolver2) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	panic("implement me")
}

func (r resolver2) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	msg, err := r.backgroundResolver.Resolve(url)
	if err != nil {
		return nil, err
	}
	return msg.(proto.Message).ProtoReflect().Type(), nil
}

func (r resolver2) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	panic("implement me")
}

func (r resolver2) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	panic("implement me")
}

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

type V2JsonProtobufAnyType struct {
	originalType        protoreflect.MessageType
	syntheticDescriptor V2JsonProtobufAnyTypeDescriptor
}

func (v V2JsonProtobufAnyType) New() protoreflect.Message {
	return dynamicpb.NewMessage(v.syntheticDescriptor)
}

func (v V2JsonProtobufAnyType) Zero() protoreflect.Message {
	return dynamicpb.NewMessage(v.syntheticDescriptor)
}

func (v V2JsonProtobufAnyType) Descriptor() protoreflect.MessageDescriptor {
	return v.syntheticDescriptor
}

type V2JsonProtobufFollowerFieldDescriptor struct {
	protoreflect.FieldDescriptor
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

type V2JsonProtobufAnyFollower struct {
	protoreflect.Message
}

func (v *V2JsonProtobufAnyFollower) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	v.Message.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		return followValue(descriptor, value, f)
	})
}

type V2JsonProtobufFollower struct {
	protoreflect.Message
}
type V2JsonProtobufListFollower struct {
	protoreflect.List
}

func (v V2JsonProtobufListFollower) Len() int {
	return v.List.Len()
}

func (v V2JsonProtobufListFollower) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage(&V2JsonProtobufFollower{v.List.Get(i).Message()})
}

func (v V2JsonProtobufListFollower) Set(i int, value protoreflect.Value) {
	panic("implement me")
}

func (v V2JsonProtobufListFollower) Append(value protoreflect.Value) {
	v.List.Append(value)
}

func (v V2JsonProtobufListFollower) AppendMutable() protoreflect.Value {
	panic("implement me")
}

func (v V2JsonProtobufListFollower) Truncate(i int) {
	panic("implement me")
}

func (v V2JsonProtobufListFollower) NewElement() protoreflect.Value {
	newelement := v.List.NewElement()
	return protoreflect.ValueOfMessage(&V2JsonProtobufFollower{newelement.Message()})
}

func (v V2JsonProtobufListFollower) IsValid() bool {
	panic("implement me")
}

type V2JsonProtobufMapFollower struct {
	protoreflect.Map
	ValueKind protoreflect.FieldDescriptor
}

func (v V2JsonProtobufMapFollower) Len() int {
	panic("implement me")
}

func (v V2JsonProtobufMapFollower) Range(f func(protoreflect.MapKey, protoreflect.Value) bool) {
	v.Map.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		return followMapValue(v.ValueKind, value, key, f)
	})
}

func (v V2JsonProtobufMapFollower) Has(key protoreflect.MapKey) bool {
	return v.Map.Has(key)
}

func (v V2JsonProtobufMapFollower) Clear(key protoreflect.MapKey) {
	panic("implement me")
}

func (v V2JsonProtobufMapFollower) Get(key protoreflect.MapKey) protoreflect.Value {
	panic("implement me")
}

func (v V2JsonProtobufMapFollower) Set(key protoreflect.MapKey, value protoreflect.Value) {
	v.Map.Set(key, value)
}

func (v V2JsonProtobufMapFollower) Mutable(key protoreflect.MapKey) protoreflect.Value {
	panic("implement me")
}

func (v V2JsonProtobufMapFollower) NewValue() protoreflect.Value {
	newelement := v.Map.NewValue()
	return protoreflect.ValueOfMessage(&V2JsonProtobufFollower{newelement.Message()})
}

func (v V2JsonProtobufMapFollower) IsValid() bool {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) Type() protoreflect.MessageType {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) New() protoreflect.Message {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) Interface() protoreflect.ProtoMessage {
	return v.Message.Interface()
}

func (v *V2JsonProtobufFollower) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	v.Message.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		name := descriptor.FullName()
		fullname := v.Message.Descriptor().FullName()
		if fullname == "google.protobuf.Any" {

			switch name {
			case "google.protobuf.Any.type_url":
				fd := V2JsonProtobufAnyTypeFieldDescriptor{descriptor}
				return f(fd, value)
			case "google.protobuf.Any.value":
				url := v.Message.Get(v.Message.Descriptor().Fields().ByName("type_url")).String()
				fd := &V2JsonProtobufAnyValueField{descriptor, url}

				bytesout := v.Message.Get(v.Message.Descriptor().Fields().Get(1)).Bytes()
				v2type := serial.V2TypeFromURL(url)
				instance, err := serial.GetInstance(v2type)
				if err != nil {
					panic(err)
				}
				unmarshaler := proto.UnmarshalOptions{AllowPartial: true, Resolver: resolver2{backgroundResolver: serial.GetResolver()}}
				err = unmarshaler.Unmarshal(bytesout, instance.(proto.Message))
				if err != nil {
					panic(err)
				}

				return f(fd, protoreflect.ValueOfMessage(&V2JsonProtobufFollower{instance.(proto.Message).ProtoReflect()}))
			default:
				panic("unexpected any value")
			}

		}
		return followValue(descriptor, value, f)
	})
}

func followValue(descriptor protoreflect.FieldDescriptor, value protoreflect.Value, f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) bool {
	fd := V2JsonProtobufFollowerFieldDescriptor{descriptor}
	if descriptor.Kind() == protoreflect.MessageKind {
		if descriptor.IsList() {
			value2 := protoreflect.ValueOfList(V2JsonProtobufListFollower{value.List()})
			return f(fd, value2)
		}
		if descriptor.IsMap() {
			value2 := protoreflect.ValueOfMap(V2JsonProtobufMapFollower{value.Map(), descriptor.MapValue()})
			return f(fd, value2)
		}
		value2 := protoreflect.ValueOfMessage(&V2JsonProtobufFollower{value.Message()})
		return f(fd, value2)
	}

	return f(fd, value)
}

func followMapValue(descriptor protoreflect.FieldDescriptor, value protoreflect.Value, mapkey protoreflect.MapKey, f func(protoreflect.MapKey, protoreflect.Value) bool) bool {
	if descriptor.Kind() == protoreflect.MessageKind {
		if descriptor.IsList() {
			value2 := protoreflect.ValueOfList(V2JsonProtobufListFollower{value.List()})
			return f(mapkey, value2)
		}
		if descriptor.IsMap() {
			value2 := protoreflect.ValueOfMap(V2JsonProtobufMapFollower{value.Map(), descriptor.MapValue()})
			return f(mapkey, value2)
		}
		value2 := protoreflect.ValueOfMessage(&V2JsonProtobufFollower{value.Message()})
		return f(mapkey, value2)
	}

	return f(mapkey, value)
}

func (v *V2JsonProtobufFollower) Has(descriptor protoreflect.FieldDescriptor) bool {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) Clear(descriptor protoreflect.FieldDescriptor) {
	v.Message.Clear(descriptor)
}

func (v *V2JsonProtobufFollower) Set(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch descriptor.(type) {
	case V2JsonProtobufFollowerFieldDescriptor:
		v.Message.Set(descriptor.(V2JsonProtobufFollowerFieldDescriptor).FieldDescriptor, value)
	case *V2JsonProtobufFollowerFieldDescriptor:
		v.Message.Set(descriptor.(*V2JsonProtobufFollowerFieldDescriptor).FieldDescriptor, value)
	case *V2JsonProtobufAnyValueField:
		protodata := value.Message()
		bytesw, err := proto.MarshalOptions{AllowPartial: true}.Marshal(&V2JsonProtobufAnyValueFieldReturn{protodata})
		if err != nil {
			panic(err)
		}
		v.Message.Set(descriptor.(*V2JsonProtobufAnyValueField).FieldDescriptor, protoreflect.ValueOfBytes(bytesw))
	default:
		v.Message.Set(descriptor, value)
	}

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

func (v *V2JsonProtobufFollower) Mutable(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	value := v.Message.Mutable(descriptor.(V2JsonProtobufFollowerFieldDescriptor).FieldDescriptor)
	if descriptor.IsList() {
		return protoreflect.ValueOfList(&V2JsonProtobufListFollower{value.List()})
	}
	if descriptor.IsMap() {
		return protoreflect.ValueOfMap(&V2JsonProtobufMapFollower{value.Map(), descriptor})
	}
	if descriptor.Kind() == protoreflect.MessageKind {
		return protoreflect.ValueOfMessage(&V2JsonProtobufFollower{value.Message()})
	}
	return value
}

func (v *V2JsonProtobufFollower) NewField(descriptor protoreflect.FieldDescriptor) protoreflect.Value {

	if _, ok := descriptor.(*V2JsonProtobufAnyValueField); ok {

		url := v.Message.Get(v.Message.Descriptor().Fields().ByName("type_url")).String()

		v2type := serial.V2TypeFromURL(url)
		instance, err := serial.GetInstance(v2type)
		if err != nil {
			panic(err)
		}
		newvalue := protoreflect.ValueOfMessage(&V2JsonProtobufFollower{instance.(proto.Message).ProtoReflect()})
		return newvalue
	}

	value := v.Message.NewField(descriptor.(V2JsonProtobufFollowerFieldDescriptor).FieldDescriptor)
	newvalue := protoreflect.ValueOfMessage(&V2JsonProtobufFollower{value.Message()})
	return newvalue
}

func (v *V2JsonProtobufFollower) WhichOneof(descriptor protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) GetUnknown() protoreflect.RawFields {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) SetUnknown(fields protoreflect.RawFields) {
	v.Message.SetUnknown(fields)
}

func (v *V2JsonProtobufFollower) IsValid() bool {
	return v.Message.IsValid()
}

func (v *V2JsonProtobufFollower) ProtoReflect() protoreflect.Message {
	return v
}

func (v *V2JsonProtobufFollower) Descriptor() protoreflect.MessageDescriptor {
	fullname := v.Message.Descriptor().FullName()
	if fullname == "google.protobuf.Any" {
		//desc := &V2JsonProtobufAnyType{v.Message.Type(), V2JsonProtobufAnyTypeDescriptor{(&anypb.Any{}).ProtoReflect().Descriptor()}}
		desc := &V2JsonProtobufAnyTypeDescriptor{(&anypb.Any{}).ProtoReflect().Descriptor()}
		return desc
	}
	return &V2JsonProtobufFollowerDescriptor{v.Message.Descriptor()}
}

func (v *V2JsonProtobufFollower) Get(fd protoreflect.FieldDescriptor) protoreflect.Value {
	panic("implement me")
}

type V2JsonProtobufFollowerDescriptor struct {
	protoreflect.MessageDescriptor
}

func (v *V2JsonProtobufFollowerDescriptor) Fields() protoreflect.FieldDescriptors {
	return &V2JsonProtobufFollowerFields{v.MessageDescriptor.Fields()}
}

type V2JsonProtobufFollowerFields struct {
	protoreflect.FieldDescriptors
}

func (v *V2JsonProtobufFollowerFields) ByJSONName(s string) protoreflect.FieldDescriptor {
	return V2JsonProtobufFollowerFieldDescriptor{v.FieldDescriptors.ByJSONName(s)}
}

func (v *V2JsonProtobufFollowerFields) ByTextName(s string) protoreflect.FieldDescriptor {
	return V2JsonProtobufFollowerFieldDescriptor{v.FieldDescriptors.ByTextName(s)}
}
