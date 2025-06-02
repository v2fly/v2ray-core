package simplified

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
	packetaddr "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
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

type ServerConfig struct {
	state          protoimpl.MessageState    `protogen:"open.v1"`
	Address        *net.IPOrDomain           `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	UdpEnabled     bool                      `protobuf:"varint,4,opt,name=udp_enabled,json=udpEnabled,proto3" json:"udp_enabled,omitempty"`
	PacketEncoding packetaddr.PacketAddrType `protobuf:"varint,7,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_socks_simplified_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_socks_simplified_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_proxy_socks_simplified_config_proto_rawDescGZIP(), []int{0}
}

func (x *ServerConfig) GetAddress() *net.IPOrDomain {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *ServerConfig) GetUdpEnabled() bool {
	if x != nil {
		return x.UdpEnabled
	}
	return false
}

func (x *ServerConfig) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

type ClientConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       *net.IPOrDomain        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_socks_simplified_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_socks_simplified_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientConfig.ProtoReflect.Descriptor instead.
func (*ClientConfig) Descriptor() ([]byte, []int) {
	return file_proxy_socks_simplified_config_proto_rawDescGZIP(), []int{1}
}

func (x *ClientConfig) GetAddress() *net.IPOrDomain {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *ClientConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

var File_proxy_socks_simplified_config_proto protoreflect.FileDescriptor

const file_proxy_socks_simplified_config_proto_rawDesc = "" +
	"\n" +
	"#proxy/socks/simplified/config.proto\x12!v2ray.core.proxy.socks.simplified\x1a common/protoext/extensions.proto\x1a\x18common/net/address.proto\x1a\"common/net/packetaddr/config.proto\"\xd6\x01\n" +
	"\fServerConfig\x12;\n" +
	"\aaddress\x18\x03 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x1f\n" +
	"\vudp_enabled\x18\x04 \x01(\bR\n" +
	"udpEnabled\x12R\n" +
	"\x0fpacket_encoding\x18\a \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding:\x14\x82\xb5\x18\x10\n" +
	"\ainbound\x12\x05socks\"v\n" +
	"\fClientConfig\x12;\n" +
	"\aaddress\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port:\x15\x82\xb5\x18\x11\n" +
	"\boutbound\x12\x05socksB\x84\x01\n" +
	"%com.v2ray.core.proxy.socks.simplifiedP\x01Z5github.com/v2fly/v2ray-core/v5/proxy/socks/simplified\xaa\x02!V2Ray.Core.Proxy.Socks.Simplifiedb\x06proto3"

var (
	file_proxy_socks_simplified_config_proto_rawDescOnce sync.Once
	file_proxy_socks_simplified_config_proto_rawDescData []byte
)

func file_proxy_socks_simplified_config_proto_rawDescGZIP() []byte {
	file_proxy_socks_simplified_config_proto_rawDescOnce.Do(func() {
		file_proxy_socks_simplified_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_socks_simplified_config_proto_rawDesc), len(file_proxy_socks_simplified_config_proto_rawDesc)))
	})
	return file_proxy_socks_simplified_config_proto_rawDescData
}

var file_proxy_socks_simplified_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proxy_socks_simplified_config_proto_goTypes = []any{
	(*ServerConfig)(nil),           // 0: v2ray.core.proxy.socks.simplified.ServerConfig
	(*ClientConfig)(nil),           // 1: v2ray.core.proxy.socks.simplified.ClientConfig
	(*net.IPOrDomain)(nil),         // 2: v2ray.core.common.net.IPOrDomain
	(packetaddr.PacketAddrType)(0), // 3: v2ray.core.net.packetaddr.PacketAddrType
}
var file_proxy_socks_simplified_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.proxy.socks.simplified.ServerConfig.address:type_name -> v2ray.core.common.net.IPOrDomain
	3, // 1: v2ray.core.proxy.socks.simplified.ServerConfig.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	2, // 2: v2ray.core.proxy.socks.simplified.ClientConfig.address:type_name -> v2ray.core.common.net.IPOrDomain
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proxy_socks_simplified_config_proto_init() }
func file_proxy_socks_simplified_config_proto_init() {
	if File_proxy_socks_simplified_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_socks_simplified_config_proto_rawDesc), len(file_proxy_socks_simplified_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_socks_simplified_config_proto_goTypes,
		DependencyIndexes: file_proxy_socks_simplified_config_proto_depIdxs,
		MessageInfos:      file_proxy_socks_simplified_config_proto_msgTypes,
	}.Build()
	File_proxy_socks_simplified_config_proto = out.File
	file_proxy_socks_simplified_config_proto_goTypes = nil
	file_proxy_socks_simplified_config_proto_depIdxs = nil
}
