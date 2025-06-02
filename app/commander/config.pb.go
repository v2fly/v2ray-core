package commander

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

// Config is the settings for Commander.
type Config struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Tag of the outbound handler that handles grpc connections.
	Tag string `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	// Services that supported by this server. All services must implement Service
	// interface.
	Service       []*anypb.Any `protobuf:"bytes,2,rep,name=service,proto3" json:"service,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_commander_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_commander_config_proto_msgTypes[0]
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
	return file_app_commander_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Config) GetService() []*anypb.Any {
	if x != nil {
		return x.Service
	}
	return nil
}

// ReflectionConfig is the placeholder config for ReflectionService.
type ReflectionConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReflectionConfig) Reset() {
	*x = ReflectionConfig{}
	mi := &file_app_commander_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReflectionConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReflectionConfig) ProtoMessage() {}

func (x *ReflectionConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_commander_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReflectionConfig.ProtoReflect.Descriptor instead.
func (*ReflectionConfig) Descriptor() ([]byte, []int) {
	return file_app_commander_config_proto_rawDescGZIP(), []int{1}
}

type SimplifiedConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Name          []string               `protobuf:"bytes,2,rep,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SimplifiedConfig) Reset() {
	*x = SimplifiedConfig{}
	mi := &file_app_commander_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedConfig) ProtoMessage() {}

func (x *SimplifiedConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_commander_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimplifiedConfig.ProtoReflect.Descriptor instead.
func (*SimplifiedConfig) Descriptor() ([]byte, []int) {
	return file_app_commander_config_proto_rawDescGZIP(), []int{2}
}

func (x *SimplifiedConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *SimplifiedConfig) GetName() []string {
	if x != nil {
		return x.Name
	}
	return nil
}

var File_app_commander_config_proto protoreflect.FileDescriptor

const file_app_commander_config_proto_rawDesc = "" +
	"\n" +
	"\x1aapp/commander/config.proto\x12\x18v2ray.core.app.commander\x1a\x19google/protobuf/any.proto\x1a common/protoext/extensions.proto\"J\n" +
	"\x06Config\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12.\n" +
	"\aservice\x18\x02 \x03(\v2\x14.google.protobuf.AnyR\aservice\"1\n" +
	"\x10ReflectionConfig:\x1d\x82\xb5\x18\x19\n" +
	"\vgrpcservice\x12\n" +
	"reflection\"R\n" +
	"\x10SimplifiedConfig\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x12\n" +
	"\x04name\x18\x02 \x03(\tR\x04name:\x18\x82\xb5\x18\x14\n" +
	"\aservice\x12\tcommanderBi\n" +
	"\x1ccom.v2ray.core.app.commanderP\x01Z,github.com/v2fly/v2ray-core/v5/app/commander\xaa\x02\x18V2Ray.Core.App.Commanderb\x06proto3"

var (
	file_app_commander_config_proto_rawDescOnce sync.Once
	file_app_commander_config_proto_rawDescData []byte
)

func file_app_commander_config_proto_rawDescGZIP() []byte {
	file_app_commander_config_proto_rawDescOnce.Do(func() {
		file_app_commander_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_commander_config_proto_rawDesc), len(file_app_commander_config_proto_rawDesc)))
	})
	return file_app_commander_config_proto_rawDescData
}

var file_app_commander_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_app_commander_config_proto_goTypes = []any{
	(*Config)(nil),           // 0: v2ray.core.app.commander.Config
	(*ReflectionConfig)(nil), // 1: v2ray.core.app.commander.ReflectionConfig
	(*SimplifiedConfig)(nil), // 2: v2ray.core.app.commander.SimplifiedConfig
	(*anypb.Any)(nil),        // 3: google.protobuf.Any
}
var file_app_commander_config_proto_depIdxs = []int32{
	3, // 0: v2ray.core.app.commander.Config.service:type_name -> google.protobuf.Any
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_app_commander_config_proto_init() }
func file_app_commander_config_proto_init() {
	if File_app_commander_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_commander_config_proto_rawDesc), len(file_app_commander_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_commander_config_proto_goTypes,
		DependencyIndexes: file_app_commander_config_proto_depIdxs,
		MessageInfos:      file_app_commander_config_proto_msgTypes,
	}.Build()
	File_app_commander_config_proto = out.File
	file_app_commander_config_proto_goTypes = nil
	file_app_commander_config_proto_depIdxs = nil
}
