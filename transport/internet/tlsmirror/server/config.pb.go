package server

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
	state                protoimpl.MessageState `protogen:"open.v1"`
	ForwardAddress       string                 `protobuf:"bytes,1,opt,name=forward_address,json=forwardAddress,proto3" json:"forward_address,omitempty"`
	ForwardPort          uint32                 `protobuf:"varint,2,opt,name=forward_port,json=forwardPort,proto3" json:"forward_port,omitempty"`
	ForwardTag           string                 `protobuf:"bytes,3,opt,name=forward_tag,json=forwardTag,proto3" json:"forward_tag,omitempty"`
	CarrierConnectionTag string                 `protobuf:"bytes,4,opt,name=carrier_connection_tag,json=carrierConnectionTag,proto3" json:"carrier_connection_tag,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[0]
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
	return file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetForwardAddress() string {
	if x != nil {
		return x.ForwardAddress
	}
	return ""
}

func (x *Config) GetForwardPort() uint32 {
	if x != nil {
		return x.ForwardPort
	}
	return 0
}

func (x *Config) GetForwardTag() string {
	if x != nil {
		return x.ForwardTag
	}
	return ""
}

func (x *Config) GetCarrierConnectionTag() string {
	if x != nil {
		return x.CarrierConnectionTag
	}
	return ""
}

var File_transport_internet_tlsmirror_server_config_proto protoreflect.FileDescriptor

const file_transport_internet_tlsmirror_server_config_proto_rawDesc = "" +
	"\n" +
	"0transport/internet/tlsmirror/server/config.proto\x12.v2ray.core.transport.internet.tlsmirror.server\x1a common/protoext/extensions.proto\"\xd8\x01\n" +
	"\x06Config\x12'\n" +
	"\x0fforward_address\x18\x01 \x01(\tR\x0eforwardAddress\x12!\n" +
	"\fforward_port\x18\x02 \x01(\rR\vforwardPort\x12\x1f\n" +
	"\vforward_tag\x18\x03 \x01(\tR\n" +
	"forwardTag\x124\n" +
	"\x16carrier_connection_tag\x18\x04 \x01(\tR\x14carrierConnectionTag:+\x82\xb5\x18'\n" +
	"\ttransport\x12\ttlsmirror\x8a\xff)\ttlsmirror\x90\xff)\x01B\xab\x01\n" +
	"2com.v2ray.core.transport.internet.tlsmirror.serverP\x01ZBgithub.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/server\xaa\x02.V2Ray.Core.Transport.Internet.Tlsmirror.Serverb\x06proto3"

var (
	file_transport_internet_tlsmirror_server_config_proto_rawDescOnce sync.Once
	file_transport_internet_tlsmirror_server_config_proto_rawDescData []byte
)

func file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP() []byte {
	file_transport_internet_tlsmirror_server_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_tlsmirror_server_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_server_config_proto_rawDesc), len(file_transport_internet_tlsmirror_server_config_proto_rawDesc)))
	})
	return file_transport_internet_tlsmirror_server_config_proto_rawDescData
}

var file_transport_internet_tlsmirror_server_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_tlsmirror_server_config_proto_goTypes = []any{
	(*Config)(nil), // 0: v2ray.core.transport.internet.tlsmirror.server.Config
}
var file_transport_internet_tlsmirror_server_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transport_internet_tlsmirror_server_config_proto_init() }
func file_transport_internet_tlsmirror_server_config_proto_init() {
	if File_transport_internet_tlsmirror_server_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_server_config_proto_rawDesc), len(file_transport_internet_tlsmirror_server_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tlsmirror_server_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_tlsmirror_server_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_tlsmirror_server_config_proto_msgTypes,
	}.Build()
	File_transport_internet_tlsmirror_server_config_proto = out.File
	file_transport_internet_tlsmirror_server_config_proto_goTypes = nil
	file_transport_internet_tlsmirror_server_config_proto_depIdxs = nil
}
