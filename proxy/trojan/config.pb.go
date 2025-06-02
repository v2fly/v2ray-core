package trojan

import (
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

type Account struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Password      string                 `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Account) Reset() {
	*x = Account{}
	mi := &file_proxy_trojan_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_trojan_config_proto_msgTypes[0]
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
	return file_proxy_trojan_config_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type Fallback struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Alpn          string                 `protobuf:"bytes,1,opt,name=alpn,proto3" json:"alpn,omitempty"`
	Path          string                 `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Type          string                 `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	Dest          string                 `protobuf:"bytes,4,opt,name=dest,proto3" json:"dest,omitempty"`
	Xver          uint64                 `protobuf:"varint,5,opt,name=xver,proto3" json:"xver,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Fallback) Reset() {
	*x = Fallback{}
	mi := &file_proxy_trojan_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Fallback) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fallback) ProtoMessage() {}

func (x *Fallback) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_trojan_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fallback.ProtoReflect.Descriptor instead.
func (*Fallback) Descriptor() ([]byte, []int) {
	return file_proxy_trojan_config_proto_rawDescGZIP(), []int{1}
}

func (x *Fallback) GetAlpn() string {
	if x != nil {
		return x.Alpn
	}
	return ""
}

func (x *Fallback) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Fallback) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Fallback) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

func (x *Fallback) GetXver() uint64 {
	if x != nil {
		return x.Xver
	}
	return 0
}

type ClientConfig struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Server        []*protocol.ServerEndpoint `protobuf:"bytes,1,rep,name=server,proto3" json:"server,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_trojan_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_trojan_config_proto_msgTypes[2]
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
	return file_proxy_trojan_config_proto_rawDescGZIP(), []int{2}
}

func (x *ClientConfig) GetServer() []*protocol.ServerEndpoint {
	if x != nil {
		return x.Server
	}
	return nil
}

type ServerConfig struct {
	state          protoimpl.MessageState    `protogen:"open.v1"`
	Users          []*protocol.User          `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	Fallbacks      []*Fallback               `protobuf:"bytes,3,rep,name=fallbacks,proto3" json:"fallbacks,omitempty"`
	PacketEncoding packetaddr.PacketAddrType `protobuf:"varint,4,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_trojan_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_trojan_config_proto_msgTypes[3]
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
	return file_proxy_trojan_config_proto_rawDescGZIP(), []int{3}
}

func (x *ServerConfig) GetUsers() []*protocol.User {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *ServerConfig) GetFallbacks() []*Fallback {
	if x != nil {
		return x.Fallbacks
	}
	return nil
}

func (x *ServerConfig) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

var File_proxy_trojan_config_proto protoreflect.FileDescriptor

const file_proxy_trojan_config_proto_rawDesc = "" +
	"\n" +
	"\x19proxy/trojan/config.proto\x12\x17v2ray.core.proxy.trojan\x1a\x1acommon/protocol/user.proto\x1a!common/protocol/server_spec.proto\x1a\"common/net/packetaddr/config.proto\"%\n" +
	"\aAccount\x12\x1a\n" +
	"\bpassword\x18\x01 \x01(\tR\bpassword\"n\n" +
	"\bFallback\x12\x12\n" +
	"\x04alpn\x18\x01 \x01(\tR\x04alpn\x12\x12\n" +
	"\x04path\x18\x02 \x01(\tR\x04path\x12\x12\n" +
	"\x04type\x18\x03 \x01(\tR\x04type\x12\x12\n" +
	"\x04dest\x18\x04 \x01(\tR\x04dest\x12\x12\n" +
	"\x04xver\x18\x05 \x01(\x04R\x04xver\"R\n" +
	"\fClientConfig\x12B\n" +
	"\x06server\x18\x01 \x03(\v2*.v2ray.core.common.protocol.ServerEndpointR\x06server\"\xdb\x01\n" +
	"\fServerConfig\x126\n" +
	"\x05users\x18\x01 \x03(\v2 .v2ray.core.common.protocol.UserR\x05users\x12?\n" +
	"\tfallbacks\x18\x03 \x03(\v2!.v2ray.core.proxy.trojan.FallbackR\tfallbacks\x12R\n" +
	"\x0fpacket_encoding\x18\x04 \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncodingBf\n" +
	"\x1bcom.v2ray.core.proxy.trojanP\x01Z+github.com/v2fly/v2ray-core/v5/proxy/trojan\xaa\x02\x17V2Ray.Core.Proxy.Trojanb\x06proto3"

var (
	file_proxy_trojan_config_proto_rawDescOnce sync.Once
	file_proxy_trojan_config_proto_rawDescData []byte
)

func file_proxy_trojan_config_proto_rawDescGZIP() []byte {
	file_proxy_trojan_config_proto_rawDescOnce.Do(func() {
		file_proxy_trojan_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_trojan_config_proto_rawDesc), len(file_proxy_trojan_config_proto_rawDesc)))
	})
	return file_proxy_trojan_config_proto_rawDescData
}

var file_proxy_trojan_config_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proxy_trojan_config_proto_goTypes = []any{
	(*Account)(nil),                 // 0: v2ray.core.proxy.trojan.Account
	(*Fallback)(nil),                // 1: v2ray.core.proxy.trojan.Fallback
	(*ClientConfig)(nil),            // 2: v2ray.core.proxy.trojan.ClientConfig
	(*ServerConfig)(nil),            // 3: v2ray.core.proxy.trojan.ServerConfig
	(*protocol.ServerEndpoint)(nil), // 4: v2ray.core.common.protocol.ServerEndpoint
	(*protocol.User)(nil),           // 5: v2ray.core.common.protocol.User
	(packetaddr.PacketAddrType)(0),  // 6: v2ray.core.net.packetaddr.PacketAddrType
}
var file_proxy_trojan_config_proto_depIdxs = []int32{
	4, // 0: v2ray.core.proxy.trojan.ClientConfig.server:type_name -> v2ray.core.common.protocol.ServerEndpoint
	5, // 1: v2ray.core.proxy.trojan.ServerConfig.users:type_name -> v2ray.core.common.protocol.User
	1, // 2: v2ray.core.proxy.trojan.ServerConfig.fallbacks:type_name -> v2ray.core.proxy.trojan.Fallback
	6, // 3: v2ray.core.proxy.trojan.ServerConfig.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proxy_trojan_config_proto_init() }
func file_proxy_trojan_config_proto_init() {
	if File_proxy_trojan_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_trojan_config_proto_rawDesc), len(file_proxy_trojan_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_trojan_config_proto_goTypes,
		DependencyIndexes: file_proxy_trojan_config_proto_depIdxs,
		MessageInfos:      file_proxy_trojan_config_proto_msgTypes,
	}.Build()
	File_proxy_trojan_config_proto = out.File
	file_proxy_trojan_config_proto_goTypes = nil
	file_proxy_trojan_config_proto_depIdxs = nil
}
