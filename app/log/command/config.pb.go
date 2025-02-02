package command

import (
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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_log_command_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_command_config_proto_msgTypes[0]
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
	return file_app_log_command_config_proto_rawDescGZIP(), []int{0}
}

type RestartLoggerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RestartLoggerRequest) Reset() {
	*x = RestartLoggerRequest{}
	mi := &file_app_log_command_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RestartLoggerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RestartLoggerRequest) ProtoMessage() {}

func (x *RestartLoggerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_command_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RestartLoggerRequest.ProtoReflect.Descriptor instead.
func (*RestartLoggerRequest) Descriptor() ([]byte, []int) {
	return file_app_log_command_config_proto_rawDescGZIP(), []int{1}
}

type RestartLoggerResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RestartLoggerResponse) Reset() {
	*x = RestartLoggerResponse{}
	mi := &file_app_log_command_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RestartLoggerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RestartLoggerResponse) ProtoMessage() {}

func (x *RestartLoggerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_command_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RestartLoggerResponse.ProtoReflect.Descriptor instead.
func (*RestartLoggerResponse) Descriptor() ([]byte, []int) {
	return file_app_log_command_config_proto_rawDescGZIP(), []int{2}
}

type FollowLogRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowLogRequest) Reset() {
	*x = FollowLogRequest{}
	mi := &file_app_log_command_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowLogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowLogRequest) ProtoMessage() {}

func (x *FollowLogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_command_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowLogRequest.ProtoReflect.Descriptor instead.
func (*FollowLogRequest) Descriptor() ([]byte, []int) {
	return file_app_log_command_config_proto_rawDescGZIP(), []int{3}
}

type FollowLogResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FollowLogResponse) Reset() {
	*x = FollowLogResponse{}
	mi := &file_app_log_command_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FollowLogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FollowLogResponse) ProtoMessage() {}

func (x *FollowLogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_command_config_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FollowLogResponse.ProtoReflect.Descriptor instead.
func (*FollowLogResponse) Descriptor() ([]byte, []int) {
	return file_app_log_command_config_proto_rawDescGZIP(), []int{4}
}

func (x *FollowLogResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_app_log_command_config_proto protoreflect.FileDescriptor

var file_app_log_command_config_proto_rawDesc = string([]byte{
	0x0a, 0x1c, 0x61, 0x70, 0x70, 0x2f, 0x6c, 0x6f, 0x67, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a,
	0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6c,
	0x6f, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0x08, 0x0a, 0x06, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x22, 0x16, 0x0a, 0x14, 0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4c,
	0x6f, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x17, 0x0a, 0x15,
	0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x12, 0x0a, 0x10, 0x46, 0x6f, 0x6c, 0x6c, 0x6f, 0x77, 0x4c,
	0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x2d, 0x0a, 0x11, 0x46, 0x6f, 0x6c,
	0x6c, 0x6f, 0x77, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xf5, 0x01, 0x0a, 0x0d, 0x4c, 0x6f, 0x67,
	0x67, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x76, 0x0a, 0x0d, 0x52, 0x65,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x12, 0x30, 0x2e, 0x76, 0x32,
	0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6c, 0x6f, 0x67,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e,
	0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6c,
	0x6f, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x52, 0x65, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x6c, 0x0a, 0x09, 0x46, 0x6f, 0x6c, 0x6c, 0x6f, 0x77, 0x4c, 0x6f, 0x67, 0x12,
	0x2c, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70,
	0x2e, 0x6c, 0x6f, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x46, 0x6f, 0x6c,
	0x6c, 0x6f, 0x77, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e,
	0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6c,
	0x6f, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x46, 0x6f, 0x6c, 0x6c, 0x6f,
	0x77, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01,
	0x42, 0x6f, 0x0a, 0x1e, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6c, 0x6f, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x50, 0x01, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x76, 0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72,
	0x65, 0x2f, 0x76, 0x35, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x6c, 0x6f, 0x67, 0x2f, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0xaa, 0x02, 0x1a, 0x56, 0x32, 0x52, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72,
	0x65, 0x2e, 0x41, 0x70, 0x70, 0x2e, 0x4c, 0x6f, 0x67, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_app_log_command_config_proto_rawDescOnce sync.Once
	file_app_log_command_config_proto_rawDescData []byte
)

func file_app_log_command_config_proto_rawDescGZIP() []byte {
	file_app_log_command_config_proto_rawDescOnce.Do(func() {
		file_app_log_command_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_log_command_config_proto_rawDesc), len(file_app_log_command_config_proto_rawDesc)))
	})
	return file_app_log_command_config_proto_rawDescData
}

var file_app_log_command_config_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_app_log_command_config_proto_goTypes = []any{
	(*Config)(nil),                // 0: v2ray.core.app.log.command.Config
	(*RestartLoggerRequest)(nil),  // 1: v2ray.core.app.log.command.RestartLoggerRequest
	(*RestartLoggerResponse)(nil), // 2: v2ray.core.app.log.command.RestartLoggerResponse
	(*FollowLogRequest)(nil),      // 3: v2ray.core.app.log.command.FollowLogRequest
	(*FollowLogResponse)(nil),     // 4: v2ray.core.app.log.command.FollowLogResponse
}
var file_app_log_command_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.app.log.command.LoggerService.RestartLogger:input_type -> v2ray.core.app.log.command.RestartLoggerRequest
	3, // 1: v2ray.core.app.log.command.LoggerService.FollowLog:input_type -> v2ray.core.app.log.command.FollowLogRequest
	2, // 2: v2ray.core.app.log.command.LoggerService.RestartLogger:output_type -> v2ray.core.app.log.command.RestartLoggerResponse
	4, // 3: v2ray.core.app.log.command.LoggerService.FollowLog:output_type -> v2ray.core.app.log.command.FollowLogResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_log_command_config_proto_init() }
func file_app_log_command_config_proto_init() {
	if File_app_log_command_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_log_command_config_proto_rawDesc), len(file_app_log_command_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_log_command_config_proto_goTypes,
		DependencyIndexes: file_app_log_command_config_proto_depIdxs,
		MessageInfos:      file_app_log_command_config_proto_msgTypes,
	}.Build()
	File_app_log_command_config_proto = out.File
	file_app_log_command_config_proto_goTypes = nil
	file_app_log_command_config_proto_depIdxs = nil
}
