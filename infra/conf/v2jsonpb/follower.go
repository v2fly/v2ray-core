package v2jsonpb

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/common/serial"
)

type V2JsonProtobufFollowerFieldDescriptor struct {
	protoreflect.FieldDescriptor
}

type V2JsonProtobufFollower struct {
	protoreflect.Message
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
				unmarshaler := proto.UnmarshalOptions{AllowPartial: true, Resolver: anyresolverv2{backgroundResolver: serial.GetResolver()}}
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

func (v *V2JsonProtobufFollower) Has(descriptor protoreflect.FieldDescriptor) bool {
	panic("implement me")
}

func (v *V2JsonProtobufFollower) Clear(descriptor protoreflect.FieldDescriptor) {
	v.Message.Clear(descriptor)
}

func (v *V2JsonProtobufFollower) Set(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch descriptor := descriptor.(type) {
	case V2JsonProtobufFollowerFieldDescriptor:
		v.Message.Set(descriptor.FieldDescriptor, value)
	case *V2JsonProtobufFollowerFieldDescriptor:
		v.Message.Set(descriptor.FieldDescriptor, value)
	case *V2JsonProtobufAnyValueField:
		protodata := value.Message()
		bytesw, err := proto.MarshalOptions{AllowPartial: true}.Marshal(&V2JsonProtobufAnyValueFieldReturn{protodata})
		if err != nil {
			panic(err)
		}
		v.Message.Set(descriptor.FieldDescriptor, protoreflect.ValueOfBytes(bytesw))
	default:
		v.Message.Set(descriptor, value)
	}
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
