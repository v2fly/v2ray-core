package taggedfeatures

import (
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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Features      map[string]*anypb.Any  `protobuf:"bytes,1,rep,name=features,proto3" json:"features,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_common_taggedfeatures_skeleton_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_common_taggedfeatures_skeleton_proto_msgTypes[0]
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
	return file_common_taggedfeatures_skeleton_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetFeatures() map[string]*anypb.Any {
	if x != nil {
		return x.Features
	}
	return nil
}

var File_common_taggedfeatures_skeleton_proto protoreflect.FileDescriptor

const file_common_taggedfeatures_skeleton_proto_rawDesc = "" +
	"\n" +
	"$common/taggedfeatures/skeleton.proto\x12 v2ray.core.common.taggedfeatures\x1a\x19google/protobuf/any.proto\"\xaf\x01\n" +
	"\x06Config\x12R\n" +
	"\bfeatures\x18\x01 \x03(\v26.v2ray.core.common.taggedfeatures.Config.FeaturesEntryR\bfeatures\x1aQ\n" +
	"\rFeaturesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12*\n" +
	"\x05value\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x05value:\x028\x01B\x81\x01\n" +
	"$com.v2ray.core.common.taggedfeaturesP\x01Z4github.com/v2fly/v2ray-core/v5/common/taggedfeatures\xaa\x02 V2Ray.Core.Common.Taggedfeaturesb\x06proto3"

var (
	file_common_taggedfeatures_skeleton_proto_rawDescOnce sync.Once
	file_common_taggedfeatures_skeleton_proto_rawDescData []byte
)

func file_common_taggedfeatures_skeleton_proto_rawDescGZIP() []byte {
	file_common_taggedfeatures_skeleton_proto_rawDescOnce.Do(func() {
		file_common_taggedfeatures_skeleton_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_taggedfeatures_skeleton_proto_rawDesc), len(file_common_taggedfeatures_skeleton_proto_rawDesc)))
	})
	return file_common_taggedfeatures_skeleton_proto_rawDescData
}

var file_common_taggedfeatures_skeleton_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_common_taggedfeatures_skeleton_proto_goTypes = []any{
	(*Config)(nil),    // 0: v2ray.core.common.taggedfeatures.Config
	nil,               // 1: v2ray.core.common.taggedfeatures.Config.FeaturesEntry
	(*anypb.Any)(nil), // 2: google.protobuf.Any
}
var file_common_taggedfeatures_skeleton_proto_depIdxs = []int32{
	1, // 0: v2ray.core.common.taggedfeatures.Config.features:type_name -> v2ray.core.common.taggedfeatures.Config.FeaturesEntry
	2, // 1: v2ray.core.common.taggedfeatures.Config.FeaturesEntry.value:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_common_taggedfeatures_skeleton_proto_init() }
func file_common_taggedfeatures_skeleton_proto_init() {
	if File_common_taggedfeatures_skeleton_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_taggedfeatures_skeleton_proto_rawDesc), len(file_common_taggedfeatures_skeleton_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_taggedfeatures_skeleton_proto_goTypes,
		DependencyIndexes: file_common_taggedfeatures_skeleton_proto_depIdxs,
		MessageInfos:      file_common_taggedfeatures_skeleton_proto_msgTypes,
	}.Build()
	File_common_taggedfeatures_skeleton_proto = out.File
	file_common_taggedfeatures_skeleton_proto_goTypes = nil
	file_common_taggedfeatures_skeleton_proto_depIdxs = nil
}
