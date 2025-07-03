package log

import (
	log "github.com/v2fly/v2ray-core/v5/common/log"
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

type LogType int32

const (
	LogType_None    LogType = 0
	LogType_Console LogType = 1
	LogType_File    LogType = 2
	LogType_Event   LogType = 3
)

// Enum value maps for LogType.
var (
	LogType_name = map[int32]string{
		0: "None",
		1: "Console",
		2: "File",
		3: "Event",
	}
	LogType_value = map[string]int32{
		"None":    0,
		"Console": 1,
		"File":    2,
		"Event":   3,
	}
)

func (x LogType) Enum() *LogType {
	p := new(LogType)
	*p = x
	return p
}

func (x LogType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LogType) Descriptor() protoreflect.EnumDescriptor {
	return file_app_log_config_proto_enumTypes[0].Descriptor()
}

func (LogType) Type() protoreflect.EnumType {
	return &file_app_log_config_proto_enumTypes[0]
}

func (x LogType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LogType.Descriptor instead.
func (LogType) EnumDescriptor() ([]byte, []int) {
	return file_app_log_config_proto_rawDescGZIP(), []int{0}
}

type LogSpecification struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          LogType                `protobuf:"varint,1,opt,name=type,proto3,enum=v2ray.core.app.log.LogType" json:"type,omitempty"`
	Level         log.Severity           `protobuf:"varint,2,opt,name=level,proto3,enum=v2ray.core.common.log.Severity" json:"level,omitempty"`
	Path          string                 `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LogSpecification) Reset() {
	*x = LogSpecification{}
	mi := &file_app_log_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LogSpecification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogSpecification) ProtoMessage() {}

func (x *LogSpecification) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogSpecification.ProtoReflect.Descriptor instead.
func (*LogSpecification) Descriptor() ([]byte, []int) {
	return file_app_log_config_proto_rawDescGZIP(), []int{0}
}

func (x *LogSpecification) GetType() LogType {
	if x != nil {
		return x.Type
	}
	return LogType_None
}

func (x *LogSpecification) GetLevel() log.Severity {
	if x != nil {
		return x.Level
	}
	return log.Severity(0)
}

func (x *LogSpecification) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *LogSpecification      `protobuf:"bytes,6,opt,name=error,proto3" json:"error,omitempty"`
	Access        *LogSpecification      `protobuf:"bytes,7,opt,name=access,proto3" json:"access,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_log_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_log_config_proto_msgTypes[1]
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
	return file_app_log_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetError() *LogSpecification {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *Config) GetAccess() *LogSpecification {
	if x != nil {
		return x.Access
	}
	return nil
}

var File_app_log_config_proto protoreflect.FileDescriptor

const file_app_log_config_proto_rawDesc = "" +
	"\n" +
	"\x14app/log/config.proto\x12\x12v2ray.core.app.log\x1a\x14common/log/log.proto\x1a common/protoext/extensions.proto\"\x8e\x01\n" +
	"\x10LogSpecification\x12/\n" +
	"\x04type\x18\x01 \x01(\x0e2\x1b.v2ray.core.app.log.LogTypeR\x04type\x125\n" +
	"\x05level\x18\x02 \x01(\x0e2\x1f.v2ray.core.common.log.SeverityR\x05level\x12\x12\n" +
	"\x04path\x18\x03 \x01(\tR\x04path\"\xb4\x01\n" +
	"\x06Config\x12:\n" +
	"\x05error\x18\x06 \x01(\v2$.v2ray.core.app.log.LogSpecificationR\x05error\x12<\n" +
	"\x06access\x18\a \x01(\v2$.v2ray.core.app.log.LogSpecificationR\x06access:\x12\x82\xb5\x18\x0e\n" +
	"\aservice\x12\x03logJ\x04\b\x01\x10\x02J\x04\b\x02\x10\x03J\x04\b\x03\x10\x04J\x04\b\x04\x10\x05J\x04\b\x05\x10\x06*5\n" +
	"\aLogType\x12\b\n" +
	"\x04None\x10\x00\x12\v\n" +
	"\aConsole\x10\x01\x12\b\n" +
	"\x04File\x10\x02\x12\t\n" +
	"\x05Event\x10\x03BW\n" +
	"\x16com.v2ray.core.app.logP\x01Z&github.com/v2fly/v2ray-core/v5/app/log\xaa\x02\x12V2Ray.Core.App.Logb\x06proto3"

var (
	file_app_log_config_proto_rawDescOnce sync.Once
	file_app_log_config_proto_rawDescData []byte
)

func file_app_log_config_proto_rawDescGZIP() []byte {
	file_app_log_config_proto_rawDescOnce.Do(func() {
		file_app_log_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_log_config_proto_rawDesc), len(file_app_log_config_proto_rawDesc)))
	})
	return file_app_log_config_proto_rawDescData
}

var file_app_log_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_app_log_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_app_log_config_proto_goTypes = []any{
	(LogType)(0),             // 0: v2ray.core.app.log.LogType
	(*LogSpecification)(nil), // 1: v2ray.core.app.log.LogSpecification
	(*Config)(nil),           // 2: v2ray.core.app.log.Config
	(log.Severity)(0),        // 3: v2ray.core.common.log.Severity
}
var file_app_log_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.app.log.LogSpecification.type:type_name -> v2ray.core.app.log.LogType
	3, // 1: v2ray.core.app.log.LogSpecification.level:type_name -> v2ray.core.common.log.Severity
	1, // 2: v2ray.core.app.log.Config.error:type_name -> v2ray.core.app.log.LogSpecification
	1, // 3: v2ray.core.app.log.Config.access:type_name -> v2ray.core.app.log.LogSpecification
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_app_log_config_proto_init() }
func file_app_log_config_proto_init() {
	if File_app_log_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_log_config_proto_rawDesc), len(file_app_log_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_log_config_proto_goTypes,
		DependencyIndexes: file_app_log_config_proto_depIdxs,
		EnumInfos:         file_app_log_config_proto_enumTypes,
		MessageInfos:      file_app_log_config_proto_msgTypes,
	}.Build()
	File_app_log_config_proto = out.File
	file_app_log_config_proto_goTypes = nil
	file_app_log_config_proto_depIdxs = nil
}
