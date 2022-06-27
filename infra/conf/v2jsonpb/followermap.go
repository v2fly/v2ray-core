package v2jsonpb

import "google.golang.org/protobuf/reflect/protoreflect"

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
