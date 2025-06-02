package multiobservatory

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	taggedfeatures "github.com/v2fly/v2ray-core/v5/common/taggedfeatures"
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
	Holders       *taggedfeatures.Config `protobuf:"bytes,1,opt,name=holders,proto3" json:"holders,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[0]
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
	return file_app_observatory_multiobservatory_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetHolders() *taggedfeatures.Config {
	if x != nil {
		return x.Holders
	}
	return nil
}

var File_app_observatory_multiobservatory_config_proto protoreflect.FileDescriptor

const file_app_observatory_multiobservatory_config_proto_rawDesc = "" +
	"\n" +
	"-app/observatory/multiobservatory/config.proto\x12+v2ray.core.app.observatory.multiobservatory\x1a$common/taggedfeatures/skeleton.proto\x1a common/protoext/extensions.proto\"m\n" +
	"\x06Config\x12B\n" +
	"\aholders\x18\x01 \x01(\v2(.v2ray.core.common.taggedfeatures.ConfigR\aholders:\x1f\x82\xb5\x18\x1b\n" +
	"\aservice\x12\x10multiobservatoryB\xa2\x01\n" +
	"/com.v2ray.core.app.observatory.multiObservatoryP\x01Z?github.com/v2fly/v2ray-core/v5/app/observatory/multiobservatory\xaa\x02+V2Ray.Core.App.Observatory.MultiObservatoryb\x06proto3"

var (
	file_app_observatory_multiobservatory_config_proto_rawDescOnce sync.Once
	file_app_observatory_multiobservatory_config_proto_rawDescData []byte
)

func file_app_observatory_multiobservatory_config_proto_rawDescGZIP() []byte {
	file_app_observatory_multiobservatory_config_proto_rawDescOnce.Do(func() {
		file_app_observatory_multiobservatory_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_observatory_multiobservatory_config_proto_rawDesc), len(file_app_observatory_multiobservatory_config_proto_rawDesc)))
	})
	return file_app_observatory_multiobservatory_config_proto_rawDescData
}

var file_app_observatory_multiobservatory_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_app_observatory_multiobservatory_config_proto_goTypes = []any{
	(*Config)(nil),                // 0: v2ray.core.app.observatory.multiobservatory.Config
	(*taggedfeatures.Config)(nil), // 1: v2ray.core.common.taggedfeatures.Config
}
var file_app_observatory_multiobservatory_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.app.observatory.multiobservatory.Config.holders:type_name -> v2ray.core.common.taggedfeatures.Config
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_app_observatory_multiobservatory_config_proto_init() }
func file_app_observatory_multiobservatory_config_proto_init() {
	if File_app_observatory_multiobservatory_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_observatory_multiobservatory_config_proto_rawDesc), len(file_app_observatory_multiobservatory_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_observatory_multiobservatory_config_proto_goTypes,
		DependencyIndexes: file_app_observatory_multiobservatory_config_proto_depIdxs,
		MessageInfos:      file_app_observatory_multiobservatory_config_proto_msgTypes,
	}.Build()
	File_app_observatory_multiobservatory_config_proto = out.File
	file_app_observatory_multiobservatory_config_proto_goTypes = nil
	file_app_observatory_multiobservatory_config_proto_depIdxs = nil
}
