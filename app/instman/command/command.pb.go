package command

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ListInstanceReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListInstanceReq) Reset() {
	*x = ListInstanceReq{}
	mi := &file_app_instman_command_command_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListInstanceReq) ProtoMessage() {}

func (x *ListInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[0]
	if x != nil {
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          []string               `protobuf:"bytes,1,rep,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListInstanceResp) Reset() {
	*x = ListInstanceResp{}
	mi := &file_app_instman_command_command_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListInstanceResp) ProtoMessage() {}

func (x *ListInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[1]
	if x != nil {
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
	state            protoimpl.MessageState `protogen:"open.v1"`
	Name             string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ConfigType       string                 `protobuf:"bytes,2,opt,name=configType,proto3" json:"configType,omitempty"`
	ConfigContentB64 string                 `protobuf:"bytes,3,opt,name=configContentB64,proto3" json:"configContentB64,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *AddInstanceReq) Reset() {
	*x = AddInstanceReq{}
	mi := &file_app_instman_command_command_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInstanceReq) ProtoMessage() {}

func (x *AddInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[2]
	if x != nil {
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddInstanceResp) Reset() {
	*x = AddInstanceResp{}
	mi := &file_app_instman_command_command_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddInstanceResp) ProtoMessage() {}

func (x *AddInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[3]
	if x != nil {
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StartInstanceReq) Reset() {
	*x = StartInstanceReq{}
	mi := &file_app_instman_command_command_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartInstanceReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartInstanceReq) ProtoMessage() {}

func (x *StartInstanceReq) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[4]
	if x != nil {
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StartInstanceResp) Reset() {
	*x = StartInstanceResp{}
	mi := &file_app_instman_command_command_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartInstanceResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartInstanceResp) ProtoMessage() {}

func (x *StartInstanceResp) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[5]
	if x != nil {
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_instman_command_command_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_instman_command_command_proto_msgTypes[6]
	if x != nil {
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

const file_app_instman_command_command_proto_rawDesc = "" +
	"\n" +
	"!app/instman/command/command.proto\x12\x1ev2ray.core.app.instman.command\x1a common/protoext/extensions.proto\"\x11\n" +
	"\x0fListInstanceReq\"&\n" +
	"\x10ListInstanceResp\x12\x12\n" +
	"\x04name\x18\x01 \x03(\tR\x04name\"p\n" +
	"\x0eAddInstanceReq\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x1e\n" +
	"\n" +
	"configType\x18\x02 \x01(\tR\n" +
	"configType\x12*\n" +
	"\x10configContentB64\x18\x03 \x01(\tR\x10configContentB64\"\x11\n" +
	"\x0fAddInstanceResp\"&\n" +
	"\x10StartInstanceReq\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"\x13\n" +
	"\x11StartInstanceResp\"$\n" +
	"\x06Config:\x1a\x82\xb5\x18\x16\n" +
	"\vgrpcservice\x12\ainstman2\xf4\x02\n" +
	"\x19InstanceManagementService\x12q\n" +
	"\fListInstance\x12/.v2ray.core.app.instman.command.ListInstanceReq\x1a0.v2ray.core.app.instman.command.ListInstanceResp\x12n\n" +
	"\vAddInstance\x12..v2ray.core.app.instman.command.AddInstanceReq\x1a/.v2ray.core.app.instman.command.AddInstanceResp\x12t\n" +
	"\rStartInstance\x120.v2ray.core.app.instman.command.StartInstanceReq\x1a1.v2ray.core.app.instman.command.StartInstanceRespB\x7f\n" +
	"&com.v2ray.core.app.observatory.instmanP\x01Z2github.com/v2fly/v2ray-core/v5/app/instman/command\xaa\x02\x1eV2Ray.Core.App.Instman.Commandb\x06proto3"

var (
	file_app_instman_command_command_proto_rawDescOnce sync.Once
	file_app_instman_command_command_proto_rawDescData []byte
)

func file_app_instman_command_command_proto_rawDescGZIP() []byte {
	file_app_instman_command_command_proto_rawDescOnce.Do(func() {
		file_app_instman_command_command_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_instman_command_command_proto_rawDesc), len(file_app_instman_command_command_proto_rawDesc)))
	})
	return file_app_instman_command_command_proto_rawDescData
}

var file_app_instman_command_command_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_app_instman_command_command_proto_goTypes = []any{
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_instman_command_command_proto_rawDesc), len(file_app_instman_command_command_proto_rawDesc)),
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
	file_app_instman_command_command_proto_goTypes = nil
	file_app_instman_command_command_proto_depIdxs = nil
}
