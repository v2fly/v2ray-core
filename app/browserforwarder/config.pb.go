package browserforwarder

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

// Config is the settings for BrowserForwarder.
type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ListenAddr    string                 `protobuf:"bytes,1,opt,name=listen_addr,json=listenAddr,proto3" json:"listen_addr,omitempty"`
	ListenPort    int32                  `protobuf:"varint,2,opt,name=listen_port,json=listenPort,proto3" json:"listen_port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_browserforwarder_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_browserforwarder_config_proto_msgTypes[0]
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
	return file_app_browserforwarder_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetListenAddr() string {
	if x != nil {
		return x.ListenAddr
	}
	return ""
}

func (x *Config) GetListenPort() int32 {
	if x != nil {
		return x.ListenPort
	}
	return 0
}

var File_app_browserforwarder_config_proto protoreflect.FileDescriptor

const file_app_browserforwarder_config_proto_rawDesc = "" +
	"\n" +
	"!app/browserforwarder/config.proto\x12\x1fv2ray.core.app.browserforwarder\x1a common/protoext/extensions.proto\"b\n" +
	"\x06Config\x12\x1f\n" +
	"\vlisten_addr\x18\x01 \x01(\tR\n" +
	"listenAddr\x12\x1f\n" +
	"\vlisten_port\x18\x02 \x01(\x05R\n" +
	"listenPort:\x16\x82\xb5\x18\x12\n" +
	"\aservice\x12\abrowserB~\n" +
	"#com.v2ray.core.app.browserforwarderP\x01Z3github.com/v2fly/v2ray-core/v5/app/browserforwarder\xaa\x02\x1fV2Ray.Core.App.Browserforwarderb\x06proto3"

var (
	file_app_browserforwarder_config_proto_rawDescOnce sync.Once
	file_app_browserforwarder_config_proto_rawDescData []byte
)

func file_app_browserforwarder_config_proto_rawDescGZIP() []byte {
	file_app_browserforwarder_config_proto_rawDescOnce.Do(func() {
		file_app_browserforwarder_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_browserforwarder_config_proto_rawDesc), len(file_app_browserforwarder_config_proto_rawDesc)))
	})
	return file_app_browserforwarder_config_proto_rawDescData
}

var file_app_browserforwarder_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_app_browserforwarder_config_proto_goTypes = []any{
	(*Config)(nil), // 0: v2ray.core.app.browserforwarder.Config
}
var file_app_browserforwarder_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_browserforwarder_config_proto_init() }
func file_app_browserforwarder_config_proto_init() {
	if File_app_browserforwarder_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_browserforwarder_config_proto_rawDesc), len(file_app_browserforwarder_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_browserforwarder_config_proto_goTypes,
		DependencyIndexes: file_app_browserforwarder_config_proto_depIdxs,
		MessageInfos:      file_app_browserforwarder_config_proto_msgTypes,
	}.Build()
	File_app_browserforwarder_config_proto = out.File
	file_app_browserforwarder_config_proto_goTypes = nil
	file_app_browserforwarder_config_proto_depIdxs = nil
}
