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

const file_app_log_command_config_proto_rawDesc = "" +
	"\n" +
	"\x1capp/log/command/config.proto\x12\x1av2ray.core.app.log.command\"\b\n" +
	"\x06Config\"\x16\n" +
	"\x14RestartLoggerRequest\"\x17\n" +
	"\x15RestartLoggerResponse\"\x12\n" +
	"\x10FollowLogRequest\"-\n" +
	"\x11FollowLogResponse\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage2\xf5\x01\n" +
	"\rLoggerService\x12v\n" +
	"\rRestartLogger\x120.v2ray.core.app.log.command.RestartLoggerRequest\x1a1.v2ray.core.app.log.command.RestartLoggerResponse\"\x00\x12l\n" +
	"\tFollowLog\x12,.v2ray.core.app.log.command.FollowLogRequest\x1a-.v2ray.core.app.log.command.FollowLogResponse\"\x000\x01Bo\n" +
	"\x1ecom.v2ray.core.app.log.commandP\x01Z.github.com/v2fly/v2ray-core/v5/app/log/command\xaa\x02\x1aV2Ray.Core.App.Log.Commandb\x06proto3"

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
