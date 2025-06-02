package grpc

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
	Host          string                 `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	ServiceName   string                 `protobuf:"bytes,2,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_grpc_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_grpc_config_proto_msgTypes[0]
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
	return file_transport_internet_grpc_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Config) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

var File_transport_internet_grpc_config_proto protoreflect.FileDescriptor

const file_transport_internet_grpc_config_proto_rawDesc = "" +
	"\n" +
	"$transport/internet/grpc/config.proto\x12+v2ray.core.transport.internet.grpc.encoding\x1a common/protoext/extensions.proto\"a\n" +
	"\x06Config\x12\x12\n" +
	"\x04host\x18\x01 \x01(\tR\x04host\x12!\n" +
	"\fservice_name\x18\x02 \x01(\tR\vserviceName: \x82\xb5\x18\x1c\n" +
	"\ttransport\x12\x04grpc\x8a\xff)\x03gun\x90\xff)\x01B\x85\x01\n" +
	"&com.v2ray.core.transport.internet.grpcZ6github.com/v2fly/v2ray-core/v5/transport/internet/grpc\xaa\x02\"V2Ray.Core.Transport.Internet.Grpcb\x06proto3"

var (
	file_transport_internet_grpc_config_proto_rawDescOnce sync.Once
	file_transport_internet_grpc_config_proto_rawDescData []byte
)

func file_transport_internet_grpc_config_proto_rawDescGZIP() []byte {
	file_transport_internet_grpc_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_grpc_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_grpc_config_proto_rawDesc), len(file_transport_internet_grpc_config_proto_rawDesc)))
	})
	return file_transport_internet_grpc_config_proto_rawDescData
}

var file_transport_internet_grpc_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_grpc_config_proto_goTypes = []any{
	(*Config)(nil), // 0: v2ray.core.transport.internet.grpc.encoding.Config
}
var file_transport_internet_grpc_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transport_internet_grpc_config_proto_init() }
func file_transport_internet_grpc_config_proto_init() {
	if File_transport_internet_grpc_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_grpc_config_proto_rawDesc), len(file_transport_internet_grpc_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_grpc_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_grpc_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_grpc_config_proto_msgTypes,
	}.Build()
	File_transport_internet_grpc_config_proto = out.File
	file_transport_internet_grpc_config_proto_goTypes = nil
	file_transport_internet_grpc_config_proto_depIdxs = nil
}
