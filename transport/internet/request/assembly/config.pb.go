package assembly

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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Assembler     *anypb.Any             `protobuf:"bytes,1,opt,name=assembler,proto3" json:"assembler,omitempty"`
	Roundtripper  *anypb.Any             `protobuf:"bytes,2,opt,name=roundtripper,proto3" json:"roundtripper,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_request_assembly_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembly_config_proto_msgTypes[0]
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
	return file_transport_internet_request_assembly_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetAssembler() *anypb.Any {
	if x != nil {
		return x.Assembler
	}
	return nil
}

func (x *Config) GetRoundtripper() *anypb.Any {
	if x != nil {
		return x.Roundtripper
	}
	return nil
}

var File_transport_internet_request_assembly_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_assembly_config_proto_rawDesc = "" +
	"\n" +
	"0transport/internet/request/assembly/config.proto\x12.v2ray.core.transport.internet.request.assembly\x1a common/protoext/extensions.proto\x1a\x19google/protobuf/any.proto\"\x90\x01\n" +
	"\x06Config\x122\n" +
	"\tassembler\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\tassembler\x128\n" +
	"\froundtripper\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\froundtripper:\x18\x82\xb5\x18\x14\n" +
	"\ttransport\x12\arequestB\xab\x01\n" +
	"2com.v2ray.core.transport.internet.request.assemblyP\x01ZBgithub.com/v2fly/v2ray-core/v5/transport/internet/request/assembly\xaa\x02.V2Ray.Core.Transport.Internet.Request.Assemblyb\x06proto3"

var (
	file_transport_internet_request_assembly_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_assembly_config_proto_rawDescData []byte
)

func file_transport_internet_request_assembly_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_assembly_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_assembly_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembly_config_proto_rawDesc), len(file_transport_internet_request_assembly_config_proto_rawDesc)))
	})
	return file_transport_internet_request_assembly_config_proto_rawDescData
}

var file_transport_internet_request_assembly_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_request_assembly_config_proto_goTypes = []any{
	(*Config)(nil),    // 0: v2ray.core.transport.internet.request.assembly.Config
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_transport_internet_request_assembly_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.transport.internet.request.assembly.Config.assembler:type_name -> google.protobuf.Any
	1, // 1: v2ray.core.transport.internet.request.assembly.Config.roundtripper:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_internet_request_assembly_config_proto_init() }
func file_transport_internet_request_assembly_config_proto_init() {
	if File_transport_internet_request_assembly_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembly_config_proto_rawDesc), len(file_transport_internet_request_assembly_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_assembly_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_assembly_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_assembly_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_assembly_config_proto = out.File
	file_transport_internet_request_assembly_config_proto_goTypes = nil
	file_transport_internet_request_assembly_config_proto_depIdxs = nil
}
