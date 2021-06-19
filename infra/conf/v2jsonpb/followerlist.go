package v2jsonpb

import "google.golang.org/protobuf/reflect/protoreflect"

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
