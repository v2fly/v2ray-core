package simplified

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
	packetaddr "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	shadowsocks "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks"
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
	Method         *CipherTypeWrapper        `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Password       string                    `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	Networks       *net.NetworkList          `protobuf:"bytes,3,opt,name=networks,proto3" json:"networks,omitempty"`
	PacketEncoding packetaddr.PacketAddrType `protobuf:"varint,4,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[0]
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
	return file_proxy_shadowsocks_simplified_config_proto_rawDescGZIP(), []int{0}
}

func (x *ServerConfig) GetMethod() *CipherTypeWrapper {
	if x != nil {
		return x.Method
	}
	return nil
}

func (x *ServerConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *ServerConfig) GetNetworks() *net.NetworkList {
	if x != nil {
		return x.Networks
	}
	return nil
}

func (x *ServerConfig) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

type ClientConfig struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	Address                        *net.IPOrDomain        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port                           uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Method                         *CipherTypeWrapper     `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Password                       string                 `protobuf:"bytes,4,opt,name=password,proto3" json:"password,omitempty"`
	ExperimentReducedIvHeadEntropy bool                   `protobuf:"varint,90001,opt,name=experiment_reduced_iv_head_entropy,json=experimentReducedIvHeadEntropy,proto3" json:"experiment_reduced_iv_head_entropy,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[1]
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
	return file_proxy_shadowsocks_simplified_config_proto_rawDescGZIP(), []int{1}
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

func (x *ClientConfig) GetMethod() *CipherTypeWrapper {
	if x != nil {
		return x.Method
	}
	return nil
}

func (x *ClientConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *ClientConfig) GetExperimentReducedIvHeadEntropy() bool {
	if x != nil {
		return x.ExperimentReducedIvHeadEntropy
	}
	return false
}

type CipherTypeWrapper struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         shadowsocks.CipherType `protobuf:"varint,1,opt,name=value,proto3,enum=v2ray.core.proxy.shadowsocks.CipherType" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CipherTypeWrapper) Reset() {
	*x = CipherTypeWrapper{}
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CipherTypeWrapper) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CipherTypeWrapper) ProtoMessage() {}

func (x *CipherTypeWrapper) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_simplified_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CipherTypeWrapper.ProtoReflect.Descriptor instead.
func (*CipherTypeWrapper) Descriptor() ([]byte, []int) {
	return file_proxy_shadowsocks_simplified_config_proto_rawDescGZIP(), []int{2}
}

func (x *CipherTypeWrapper) GetValue() shadowsocks.CipherType {
	if x != nil {
		return x.Value
	}
	return shadowsocks.CipherType(0)
}

var File_proxy_shadowsocks_simplified_config_proto protoreflect.FileDescriptor

const file_proxy_shadowsocks_simplified_config_proto_rawDesc = "" +
	"\n" +
	")proxy/shadowsocks/simplified/config.proto\x12'v2ray.core.proxy.shadowsocks.simplified\x1a common/protoext/extensions.proto\x1a\x18common/net/address.proto\x1a\x18common/net/network.proto\x1a\"common/net/packetaddr/config.proto\x1a\x1eproxy/shadowsocks/config.proto\"\xae\x02\n" +
	"\fServerConfig\x12R\n" +
	"\x06method\x18\x01 \x01(\v2:.v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapperR\x06method\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\x12>\n" +
	"\bnetworks\x18\x03 \x01(\v2\".v2ray.core.common.net.NetworkListR\bnetworks\x12R\n" +
	"\x0fpacket_encoding\x18\x04 \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding:\x1a\x82\xb5\x18\x16\n" +
	"\ainbound\x12\vshadowsocks\"\xbe\x02\n" +
	"\fClientConfig\x12;\n" +
	"\aaddress\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12R\n" +
	"\x06method\x18\x03 \x01(\v2:.v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapperR\x06method\x12\x1a\n" +
	"\bpassword\x18\x04 \x01(\tR\bpassword\x12L\n" +
	"\"experiment_reduced_iv_head_entropy\x18\x91\xbf\x05 \x01(\bR\x1eexperimentReducedIvHeadEntropy:\x1f\x82\xb5\x18\x1b\n" +
	"\boutbound\x12\vshadowsocks\x90\xff)\x01\"S\n" +
	"\x11CipherTypeWrapper\x12>\n" +
	"\x05value\x18\x01 \x01(\x0e2(.v2ray.core.proxy.shadowsocks.CipherTypeR\x05valueB\x96\x01\n" +
	"+com.v2ray.core.proxy.shadowsocks.simplifiedP\x01Z;github.com/v2fly/v2ray-core/v5/proxy/shadowsocks/simplified\xaa\x02'V2Ray.Core.Proxy.Shadowsocks.Simplifiedb\x06proto3"

var (
	file_proxy_shadowsocks_simplified_config_proto_rawDescOnce sync.Once
	file_proxy_shadowsocks_simplified_config_proto_rawDescData []byte
)

func file_proxy_shadowsocks_simplified_config_proto_rawDescGZIP() []byte {
	file_proxy_shadowsocks_simplified_config_proto_rawDescOnce.Do(func() {
		file_proxy_shadowsocks_simplified_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_shadowsocks_simplified_config_proto_rawDesc), len(file_proxy_shadowsocks_simplified_config_proto_rawDesc)))
	})
	return file_proxy_shadowsocks_simplified_config_proto_rawDescData
}

var file_proxy_shadowsocks_simplified_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proxy_shadowsocks_simplified_config_proto_goTypes = []any{
	(*ServerConfig)(nil),           // 0: v2ray.core.proxy.shadowsocks.simplified.ServerConfig
	(*ClientConfig)(nil),           // 1: v2ray.core.proxy.shadowsocks.simplified.ClientConfig
	(*CipherTypeWrapper)(nil),      // 2: v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapper
	(*net.NetworkList)(nil),        // 3: v2ray.core.common.net.NetworkList
	(packetaddr.PacketAddrType)(0), // 4: v2ray.core.net.packetaddr.PacketAddrType
	(*net.IPOrDomain)(nil),         // 5: v2ray.core.common.net.IPOrDomain
	(shadowsocks.CipherType)(0),    // 6: v2ray.core.proxy.shadowsocks.CipherType
}
var file_proxy_shadowsocks_simplified_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.proxy.shadowsocks.simplified.ServerConfig.method:type_name -> v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapper
	3, // 1: v2ray.core.proxy.shadowsocks.simplified.ServerConfig.networks:type_name -> v2ray.core.common.net.NetworkList
	4, // 2: v2ray.core.proxy.shadowsocks.simplified.ServerConfig.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	5, // 3: v2ray.core.proxy.shadowsocks.simplified.ClientConfig.address:type_name -> v2ray.core.common.net.IPOrDomain
	2, // 4: v2ray.core.proxy.shadowsocks.simplified.ClientConfig.method:type_name -> v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapper
	6, // 5: v2ray.core.proxy.shadowsocks.simplified.CipherTypeWrapper.value:type_name -> v2ray.core.proxy.shadowsocks.CipherType
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_proxy_shadowsocks_simplified_config_proto_init() }
func file_proxy_shadowsocks_simplified_config_proto_init() {
	if File_proxy_shadowsocks_simplified_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_shadowsocks_simplified_config_proto_rawDesc), len(file_proxy_shadowsocks_simplified_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_shadowsocks_simplified_config_proto_goTypes,
		DependencyIndexes: file_proxy_shadowsocks_simplified_config_proto_depIdxs,
		MessageInfos:      file_proxy_shadowsocks_simplified_config_proto_msgTypes,
	}.Build()
	File_proxy_shadowsocks_simplified_config_proto = out.File
	file_proxy_shadowsocks_simplified_config_proto_goTypes = nil
	file_proxy_shadowsocks_simplified_config_proto_depIdxs = nil
}
