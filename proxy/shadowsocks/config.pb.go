package shadowsocks

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
	packetaddr "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	protocol "github.com/v2fly/v2ray-core/v5/common/protocol"
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

type CipherType int32

const (
	CipherType_UNKNOWN           CipherType = 0
	CipherType_AES_128_GCM       CipherType = 1
	CipherType_AES_256_GCM       CipherType = 2
	CipherType_CHACHA20_POLY1305 CipherType = 3
	CipherType_NONE              CipherType = 4
)

// Enum value maps for CipherType.
var (
	CipherType_name = map[int32]string{
		0: "UNKNOWN",
		1: "AES_128_GCM",
		2: "AES_256_GCM",
		3: "CHACHA20_POLY1305",
		4: "NONE",
	}
	CipherType_value = map[string]int32{
		"UNKNOWN":           0,
		"AES_128_GCM":       1,
		"AES_256_GCM":       2,
		"CHACHA20_POLY1305": 3,
		"NONE":              4,
	}
)

func (x CipherType) Enum() *CipherType {
	p := new(CipherType)
	*p = x
	return p
}

func (x CipherType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CipherType) Descriptor() protoreflect.EnumDescriptor {
	return file_proxy_shadowsocks_config_proto_enumTypes[0].Descriptor()
}

func (CipherType) Type() protoreflect.EnumType {
	return &file_proxy_shadowsocks_config_proto_enumTypes[0]
}

func (x CipherType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CipherType.Descriptor instead.
func (CipherType) EnumDescriptor() ([]byte, []int) {
	return file_proxy_shadowsocks_config_proto_rawDescGZIP(), []int{0}
}

type Account struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	Password                       string                 `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
	CipherType                     CipherType             `protobuf:"varint,2,opt,name=cipher_type,json=cipherType,proto3,enum=v2ray.core.proxy.shadowsocks.CipherType" json:"cipher_type,omitempty"`
	IvCheck                        bool                   `protobuf:"varint,3,opt,name=iv_check,json=ivCheck,proto3" json:"iv_check,omitempty"`
	ExperimentReducedIvHeadEntropy bool                   `protobuf:"varint,90001,opt,name=experiment_reduced_iv_head_entropy,json=experimentReducedIvHeadEntropy,proto3" json:"experiment_reduced_iv_head_entropy,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *Account) Reset() {
	*x = Account{}
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Account.ProtoReflect.Descriptor instead.
func (*Account) Descriptor() ([]byte, []int) {
	return file_proxy_shadowsocks_config_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *Account) GetCipherType() CipherType {
	if x != nil {
		return x.CipherType
	}
	return CipherType_UNKNOWN
}

func (x *Account) GetIvCheck() bool {
	if x != nil {
		return x.IvCheck
	}
	return false
}

func (x *Account) GetExperimentReducedIvHeadEntropy() bool {
	if x != nil {
		return x.ExperimentReducedIvHeadEntropy
	}
	return false
}

type ServerConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// UdpEnabled specified whether or not to enable UDP for Shadowsocks.
	// Deprecated. Use 'network' field.
	//
	// Deprecated: Marked as deprecated in proxy/shadowsocks/config.proto.
	UdpEnabled     bool                      `protobuf:"varint,1,opt,name=udp_enabled,json=udpEnabled,proto3" json:"udp_enabled,omitempty"`
	User           *protocol.User            `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
	Network        []net.Network             `protobuf:"varint,3,rep,packed,name=network,proto3,enum=v2ray.core.common.net.Network" json:"network,omitempty"`
	PacketEncoding packetaddr.PacketAddrType `protobuf:"varint,4,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[1]
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
	return file_proxy_shadowsocks_config_proto_rawDescGZIP(), []int{1}
}

// Deprecated: Marked as deprecated in proxy/shadowsocks/config.proto.
func (x *ServerConfig) GetUdpEnabled() bool {
	if x != nil {
		return x.UdpEnabled
	}
	return false
}

func (x *ServerConfig) GetUser() *protocol.User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *ServerConfig) GetNetwork() []net.Network {
	if x != nil {
		return x.Network
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
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Server        []*protocol.ServerEndpoint `protobuf:"bytes,1,rep,name=server,proto3" json:"server,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_shadowsocks_config_proto_msgTypes[2]
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
	return file_proxy_shadowsocks_config_proto_rawDescGZIP(), []int{2}
}

func (x *ClientConfig) GetServer() []*protocol.ServerEndpoint {
	if x != nil {
		return x.Server
	}
	return nil
}

var File_proxy_shadowsocks_config_proto protoreflect.FileDescriptor

const file_proxy_shadowsocks_config_proto_rawDesc = "" +
	"\n" +
	"\x1eproxy/shadowsocks/config.proto\x12\x1cv2ray.core.proxy.shadowsocks\x1a\x18common/net/network.proto\x1a\x1acommon/protocol/user.proto\x1a!common/protocol/server_spec.proto\x1a\"common/net/packetaddr/config.proto\"\xd9\x01\n" +
	"\aAccount\x12\x1a\n" +
	"\bpassword\x18\x01 \x01(\tR\bpassword\x12I\n" +
	"\vcipher_type\x18\x02 \x01(\x0e2(.v2ray.core.proxy.shadowsocks.CipherTypeR\n" +
	"cipherType\x12\x19\n" +
	"\biv_check\x18\x03 \x01(\bR\aivCheck\x12L\n" +
	"\"experiment_reduced_iv_head_entropy\x18\x91\xbf\x05 \x01(\bR\x1eexperimentReducedIvHeadEntropy\"\xf7\x01\n" +
	"\fServerConfig\x12#\n" +
	"\vudp_enabled\x18\x01 \x01(\bB\x02\x18\x01R\n" +
	"udpEnabled\x124\n" +
	"\x04user\x18\x02 \x01(\v2 .v2ray.core.common.protocol.UserR\x04user\x128\n" +
	"\anetwork\x18\x03 \x03(\x0e2\x1e.v2ray.core.common.net.NetworkR\anetwork\x12R\n" +
	"\x0fpacket_encoding\x18\x04 \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding\"R\n" +
	"\fClientConfig\x12B\n" +
	"\x06server\x18\x01 \x03(\v2*.v2ray.core.common.protocol.ServerEndpointR\x06server*\\\n" +
	"\n" +
	"CipherType\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\x0f\n" +
	"\vAES_128_GCM\x10\x01\x12\x0f\n" +
	"\vAES_256_GCM\x10\x02\x12\x15\n" +
	"\x11CHACHA20_POLY1305\x10\x03\x12\b\n" +
	"\x04NONE\x10\x04Bu\n" +
	" com.v2ray.core.proxy.shadowsocksP\x01Z0github.com/v2fly/v2ray-core/v5/proxy/shadowsocks\xaa\x02\x1cV2Ray.Core.Proxy.Shadowsocksb\x06proto3"

var (
	file_proxy_shadowsocks_config_proto_rawDescOnce sync.Once
	file_proxy_shadowsocks_config_proto_rawDescData []byte
)

func file_proxy_shadowsocks_config_proto_rawDescGZIP() []byte {
	file_proxy_shadowsocks_config_proto_rawDescOnce.Do(func() {
		file_proxy_shadowsocks_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_shadowsocks_config_proto_rawDesc), len(file_proxy_shadowsocks_config_proto_rawDesc)))
	})
	return file_proxy_shadowsocks_config_proto_rawDescData
}

var file_proxy_shadowsocks_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proxy_shadowsocks_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proxy_shadowsocks_config_proto_goTypes = []any{
	(CipherType)(0),                 // 0: v2ray.core.proxy.shadowsocks.CipherType
	(*Account)(nil),                 // 1: v2ray.core.proxy.shadowsocks.Account
	(*ServerConfig)(nil),            // 2: v2ray.core.proxy.shadowsocks.ServerConfig
	(*ClientConfig)(nil),            // 3: v2ray.core.proxy.shadowsocks.ClientConfig
	(*protocol.User)(nil),           // 4: v2ray.core.common.protocol.User
	(net.Network)(0),                // 5: v2ray.core.common.net.Network
	(packetaddr.PacketAddrType)(0),  // 6: v2ray.core.net.packetaddr.PacketAddrType
	(*protocol.ServerEndpoint)(nil), // 7: v2ray.core.common.protocol.ServerEndpoint
}
var file_proxy_shadowsocks_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.proxy.shadowsocks.Account.cipher_type:type_name -> v2ray.core.proxy.shadowsocks.CipherType
	4, // 1: v2ray.core.proxy.shadowsocks.ServerConfig.user:type_name -> v2ray.core.common.protocol.User
	5, // 2: v2ray.core.proxy.shadowsocks.ServerConfig.network:type_name -> v2ray.core.common.net.Network
	6, // 3: v2ray.core.proxy.shadowsocks.ServerConfig.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	7, // 4: v2ray.core.proxy.shadowsocks.ClientConfig.server:type_name -> v2ray.core.common.protocol.ServerEndpoint
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proxy_shadowsocks_config_proto_init() }
func file_proxy_shadowsocks_config_proto_init() {
	if File_proxy_shadowsocks_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_shadowsocks_config_proto_rawDesc), len(file_proxy_shadowsocks_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_shadowsocks_config_proto_goTypes,
		DependencyIndexes: file_proxy_shadowsocks_config_proto_depIdxs,
		EnumInfos:         file_proxy_shadowsocks_config_proto_enumTypes,
		MessageInfos:      file_proxy_shadowsocks_config_proto_msgTypes,
	}.Build()
	File_proxy_shadowsocks_config_proto = out.File
	file_proxy_shadowsocks_config_proto_goTypes = nil
	file_proxy_shadowsocks_config_proto_depIdxs = nil
}
