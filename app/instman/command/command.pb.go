package command

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ListInstanceReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListInstanceReq) Reset() {
	*x = ListInstanceReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListInstanceReq) ProtoMessage() {}

func (x *ListInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListInstanceReq.ProtoReflect.Descriptor instead.
func (*ListInstanceReq) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{0}
}

type ListInstanceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name []string `protobuf:"bytes,1,rep,name=name,proto3" json:"name,omitempty"`
}

func (x *ListInstanceResp) Reset() {
	*x = ListInstanceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListInstanceResp) ProtoMessage() {}

func (x *ListInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListInstanceResp.ProtoReflect.Descriptor instead.
func (*ListInstanceResp) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{1}
}

func (x *ListInstanceResp) GetName() []string {
	if x != nil {
		return x.Name
	}
	return nil
}

type AddInstanceReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name             string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ConfigType       string `protobuf:"bytes,2,opt,name=configType,proto3" json:"configType,omitempty"`
	ConfigContentB64 string `protobuf:"bytes,3,opt,name=configContentB64,proto3" json:"configContentB64,omitempty"`
}

func (x *AddInstanceReq) Reset() {
	*x = AddInstanceReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInstanceReq) ProtoMessage() {}

func (x *AddInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddInstanceReq.ProtoReflect.Descriptor instead.
func (*AddInstanceReq) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{2}
}

func (x *AddInstanceReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AddInstanceReq) GetConfigType() string {
	if x != nil {
		return x.ConfigType
	}
	return ""
}

func (x *AddInstanceReq) GetConfigContentB64() string {
	if x != nil {
		return x.ConfigContentB64
	}
	return ""
}

type AddInstanceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AddInstanceResp) Reset() {
	*x = AddInstanceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInstanceResp) ProtoMessage() {}

func (x *AddInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddInstanceResp.ProtoReflect.Descriptor instead.
func (*AddInstanceResp) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{3}
}

type StartInstanceReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *StartInstanceReq) Reset() {
	*x = StartInstanceReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartInstanceReq) ProtoMessage() {}

func (x *StartInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartInstanceReq.ProtoReflect.Descriptor instead.
func (*StartInstanceReq) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{4}
}

func (x *StartInstanceReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type StartInstanceResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StartInstanceResp) Reset() {
	*x = StartInstanceResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartInstanceResp) ProtoMessage() {}

func (x *StartInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartInstanceResp.ProtoReflect.Descriptor instead.
func (*StartInstanceResp) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{5}
}

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_instman_command_command_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_app_instman_command_command_proto_rawDescGZIP(), []int{6}
}

var File_app_instman_command_command_proto protoreflect.FileDescriptor

var file_app_instman_command_command_proto_rawDesc = []byte{
	0x0a, 0x21, 0x61, 0x70, 0x70, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x61, 0x70, 0x70, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x1a, 0x20, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x65, 0x78, 0x74, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x11, 0x0a, 0x0f, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x22, 0x26, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x22, 0x70, 0x0a, 0x0e, 0x41, 0x64, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52,
	0x65, 0x71, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x54, 0x79, 0x70, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x42, 0x36, 0x34, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x42,
	0x36, 0x34, 0x22, 0x11, 0x0a, 0x0f, 0x41, 0x64, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x22, 0x26, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x72, 0x74, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x13, 0x0a,
	0x11, 0x53, 0x74, 0x61, 0x72, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x22, 0x28, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x3a, 0x1e, 0x82, 0xb5,
	0x18, 0x0d, 0x0a, 0x0b, 0x67, 0x72, 0x70, 0x63, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x82,
	0xb5, 0x18, 0x09, 0x12, 0x07, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x32, 0xf4, 0x02, 0x0a,
	0x19, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d,
	0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x71, 0x0a, 0x0c, 0x4c, 0x69,
	0x73, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2f, 0x2e, 0x76, 0x32, 0x72,
	0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x69, 0x6e, 0x73, 0x74,
	0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x30, 0x2e, 0x76, 0x32,
	0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x69, 0x6e, 0x73,
	0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x6e, 0x0a,
	0x0b, 0x41, 0x64, 0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2e, 0x2e, 0x76,
	0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x69, 0x6e,
	0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x41, 0x64,
	0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x1a, 0x2f, 0x2e, 0x76,
	0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x69, 0x6e,
	0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x41, 0x64,
	0x64, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x74, 0x0a,
	0x0d, 0x53, 0x74, 0x61, 0x72, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x30,
	0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e,
	0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e,
	0x53, 0x74, 0x61, 0x72, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71,
	0x1a, 0x31, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70,
	0x70, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x42, 0x7f, 0x0a, 0x26, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
	0x61, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x50, 0x01, 0x5a,
	0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x32, 0x66, 0x6c,
	0x79, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x35, 0x2f,
	0x61, 0x70, 0x70, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0xaa, 0x02, 0x1e, 0x56, 0x32, 0x52, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72, 0x65,
	0x2e, 0x41, 0x70, 0x70, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x6d, 0x61, 0x6e, 0x2e, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_app_instman_command_command_proto_rawDescOnce sync.Once
	file_app_instman_command_command_proto_rawDescData = file_app_instman_command_command_proto_rawDesc
)

func file_app_instman_command_command_proto_rawDescGZIP() []byte {
	file_app_instman_command_command_proto_rawDescOnce.Do(func() {
		file_app_instman_command_command_proto_rawDescData = protoimpl.X.CompressGZIP(file_app_instman_command_command_proto_rawDescData)
	})
	return file_app_instman_command_command_proto_rawDescData
}

var file_app_instman_command_command_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_app_instman_command_command_proto_goTypes = []interface{}{
	(*ListInstanceReq)(nil),   // 0: v2ray.core.app.instman.command.ListInstanceReq
	(*ListInstanceResp)(nil),  // 1: v2ray.core.app.instman.command.ListInstanceResp
	(*AddInstanceReq)(nil),    // 2: v2ray.core.app.instman.command.AddInstanceReq
	(*AddInstanceResp)(nil),   // 3: v2ray.core.app.instman.command.AddInstanceResp
	(*StartInstanceReq)(nil),  // 4: v2ray.core.app.instman.command.StartInstanceReq
	(*StartInstanceResp)(nil), // 5: v2ray.core.app.instman.command.StartInstanceResp
	(*Config)(nil),            // 6: v2ray.core.app.instman.command.Config
}
var file_app_instman_command_command_proto_depIdxs = []int32{
	0, // 0: v2ray.core.app.instman.command.InstanceManagementService.ListInstance:input_type -> v2ray.core.app.instman.command.ListInstanceReq
	2, // 1: v2ray.core.app.instman.command.InstanceManagementService.AddInstance:input_type -> v2ray.core.app.instman.command.AddInstanceReq
	4, // 2: v2ray.core.app.instman.command.InstanceManagementService.StartInstance:input_type -> v2ray.core.app.instman.command.StartInstanceReq
	1, // 3: v2ray.core.app.instman.command.InstanceManagementService.ListInstance:output_type -> v2ray.core.app.instman.command.ListInstanceResp
	3, // 4: v2ray.core.app.instman.command.InstanceManagementService.AddInstance:output_type -> v2ray.core.app.instman.command.AddInstanceResp
	5, // 5: v2ray.core.app.instman.command.InstanceManagementService.StartInstance:output_type -> v2ray.core.app.instman.command.StartInstanceResp
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_instman_command_command_proto_init() }
func file_app_instman_command_command_proto_init() {
	if File_app_instman_command_command_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_app_instman_command_command_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListInstanceReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListInstanceResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddInstanceReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddInstanceResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartInstanceReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartInstanceResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_app_instman_command_command_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_app_instman_command_command_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_instman_command_command_proto_goTypes,
		DependencyIndexes: file_app_instman_command_command_proto_depIdxs,
		MessageInfos:      file_app_instman_command_command_proto_msgTypes,
	}.Build()
	File_app_instman_command_command_proto = out.File
	file_app_instman_command_command_proto_rawDesc = nil
	file_app_instman_command_command_proto_goTypes = nil
	file_app_instman_command_command_proto_depIdxs = nil
}
