package webcommander

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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	WebRoot       []byte                 `protobuf:"bytes,2,opt,name=web_root,json=webRoot,proto3" json:"web_root,omitempty"`
	WebRootFile   string                 `protobuf:"bytes,96002,opt,name=web_root_file,json=webRootFile,proto3" json:"web_root_file,omitempty"`
	ApiMountpoint string                 `protobuf:"bytes,3,opt,name=api_mountpoint,json=apiMountpoint,proto3" json:"api_mountpoint,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_commander_webcommander_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_commander_webcommander_config_proto_msgTypes[0]
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
	return file_app_commander_webcommander_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Config) GetWebRoot() []byte {
	if x != nil {
		return x.WebRoot
	}
	return nil
}

func (x *Config) GetWebRootFile() string {
	if x != nil {
		return x.WebRootFile
	}
	return ""
}

func (x *Config) GetApiMountpoint() string {
	if x != nil {
		return x.ApiMountpoint
	}
	return ""
}

var File_app_commander_webcommander_config_proto protoreflect.FileDescriptor

const file_app_commander_webcommander_config_proto_rawDesc = "" +
	"\n" +
	"'app/commander/webcommander/config.proto\x12%v2ray.core.app.commander.webcommander\x1a common/protoext/extensions.proto\"\xaf\x01\n" +
	"\x06Config\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x19\n" +
	"\bweb_root\x18\x02 \x01(\fR\awebRoot\x124\n" +
	"\rweb_root_file\x18\x82\xee\x05 \x01(\tB\x0e\x82\xb5\x18\n" +
	"\"\bweb_rootR\vwebRootFile\x12%\n" +
	"\x0eapi_mountpoint\x18\x03 \x01(\tR\rapiMountpoint:\x1b\x82\xb5\x18\x17\n" +
	"\aservice\x12\fwebcommanderB\x90\x01\n" +
	")com.v2ray.core.app.commander.webcommanderP\x01Z9github.com/v2fly/v2ray-core/v5/app/commander/webcommander\xaa\x02%V2Ray.Core.App.Commander.WebCommanderb\x06proto3"

var (
	file_app_commander_webcommander_config_proto_rawDescOnce sync.Once
	file_app_commander_webcommander_config_proto_rawDescData []byte
)

func file_app_commander_webcommander_config_proto_rawDescGZIP() []byte {
	file_app_commander_webcommander_config_proto_rawDescOnce.Do(func() {
		file_app_commander_webcommander_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_commander_webcommander_config_proto_rawDesc), len(file_app_commander_webcommander_config_proto_rawDesc)))
	})
	return file_app_commander_webcommander_config_proto_rawDescData
}

var file_app_commander_webcommander_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_app_commander_webcommander_config_proto_goTypes = []any{
	(*Config)(nil), // 0: v2ray.core.app.commander.webcommander.Config
}
var file_app_commander_webcommander_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_commander_webcommander_config_proto_init() }
func file_app_commander_webcommander_config_proto_init() {
	if File_app_commander_webcommander_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_commander_webcommander_config_proto_rawDesc), len(file_app_commander_webcommander_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_commander_webcommander_config_proto_goTypes,
		DependencyIndexes: file_app_commander_webcommander_config_proto_depIdxs,
		MessageInfos:      file_app_commander_webcommander_config_proto_msgTypes,
	}.Build()
	File_app_commander_webcommander_config_proto = out.File
	file_app_commander_webcommander_config_proto_goTypes = nil
	file_app_commander_webcommander_config_proto_depIdxs = nil
}
