package socks

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

// AuthType is the authentication type of Socks proxy.
type AuthType int32

const (
	// NO_AUTH is for anonymous authentication.
	AuthType_NO_AUTH AuthType = 0
	// PASSWORD is for username/password authentication.
	AuthType_PASSWORD AuthType = 1
)

// Enum value maps for AuthType.
var (
	AuthType_name = map[int32]string{
		0: "NO_AUTH",
		1: "PASSWORD",
	}
	AuthType_value = map[string]int32{
		"NO_AUTH":  0,
		"PASSWORD": 1,
	}
)

func (x AuthType) Enum() *AuthType {
	p := new(AuthType)
	*p = x
	return p
}

func (x AuthType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthType) Descriptor() protoreflect.EnumDescriptor {
	return file_proxy_socks_config_proto_enumTypes[0].Descriptor()
}

func (AuthType) Type() protoreflect.EnumType {
	return &file_proxy_socks_config_proto_enumTypes[0]
}

func (x AuthType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuthType.Descriptor instead.
func (AuthType) EnumDescriptor() ([]byte, []int) {
	return file_proxy_socks_config_proto_rawDescGZIP(), []int{0}
}

type Version int32

const (
	Version_SOCKS5  Version = 0
	Version_SOCKS4  Version = 1
	Version_SOCKS4A Version = 2
)

// Enum value maps for Version.
var (
	Version_name = map[int32]string{
		0: "SOCKS5",
		1: "SOCKS4",
		2: "SOCKS4A",
	}
	Version_value = map[string]int32{
		"SOCKS5":  0,
		"SOCKS4":  1,
		"SOCKS4A": 2,
	}
)

func (x Version) Enum() *Version {
	p := new(Version)
	*p = x
	return p
}

func (x Version) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Version) Descriptor() protoreflect.EnumDescriptor {
	return file_proxy_socks_config_proto_enumTypes[1].Descriptor()
}

func (Version) Type() protoreflect.EnumType {
	return &file_proxy_socks_config_proto_enumTypes[1]
}

func (x Version) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Version.Descriptor instead.
func (Version) EnumDescriptor() ([]byte, []int) {
	return file_proxy_socks_config_proto_rawDescGZIP(), []int{1}
}

// Account represents a Socks account.
type Account struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Username      string                 `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Account) Reset() {
	*x = Account{}
	mi := &file_proxy_socks_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_socks_config_proto_msgTypes[0]
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
	return file_proxy_socks_config_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Account) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

// ServerConfig is the protobuf config for Socks server.
type ServerConfig struct {
	state      protoimpl.MessageState `protogen:"open.v1"`
	AuthType   AuthType               `protobuf:"varint,1,opt,name=auth_type,json=authType,proto3,enum=v2ray.core.proxy.socks.AuthType" json:"auth_type,omitempty"`
	Accounts   map[string]string      `protobuf:"bytes,2,rep,name=accounts,proto3" json:"accounts,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Address    *net.IPOrDomain        `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	UdpEnabled bool                   `protobuf:"varint,4,opt,name=udp_enabled,json=udpEnabled,proto3" json:"udp_enabled,omitempty"`
	// Deprecated: Marked as deprecated in proxy/socks/config.proto.
	Timeout        uint32                    `protobuf:"varint,5,opt,name=timeout,proto3" json:"timeout,omitempty"`
	UserLevel      uint32                    `protobuf:"varint,6,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	PacketEncoding packetaddr.PacketAddrType `protobuf:"varint,7,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_socks_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_socks_config_proto_msgTypes[1]
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
	return file_proxy_socks_config_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetAuthType() AuthType {
	if x != nil {
		return x.AuthType
	}
	return AuthType_NO_AUTH
}

func (x *ServerConfig) GetAccounts() map[string]string {
	if x != nil {
		return x.Accounts
	}
	return nil
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

// Deprecated: Marked as deprecated in proxy/socks/config.proto.
func (x *ServerConfig) GetTimeout() uint32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *ServerConfig) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

func (x *ServerConfig) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

// ClientConfig is the protobuf config for Socks client.
type ClientConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Sever is a list of Socks server addresses.
	Server         []*protocol.ServerEndpoint `protobuf:"bytes,1,rep,name=server,proto3" json:"server,omitempty"`
	Version        Version                    `protobuf:"varint,2,opt,name=version,proto3,enum=v2ray.core.proxy.socks.Version" json:"version,omitempty"`
	DelayAuthWrite bool                       `protobuf:"varint,3,opt,name=delay_auth_write,json=delayAuthWrite,proto3" json:"delay_auth_write,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_socks_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_socks_config_proto_msgTypes[2]
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
	return file_proxy_socks_config_proto_rawDescGZIP(), []int{2}
}

func (x *ClientConfig) GetServer() []*protocol.ServerEndpoint {
	if x != nil {
		return x.Server
	}
	return nil
}

func (x *ClientConfig) GetVersion() Version {
	if x != nil {
		return x.Version
	}
	return Version_SOCKS5
}

func (x *ClientConfig) GetDelayAuthWrite() bool {
	if x != nil {
		return x.DelayAuthWrite
	}
	return false
}

var File_proxy_socks_config_proto protoreflect.FileDescriptor

const file_proxy_socks_config_proto_rawDesc = "" +
	"\n" +
	"\x18proxy/socks/config.proto\x12\x16v2ray.core.proxy.socks\x1a\x18common/net/address.proto\x1a\"common/net/packetaddr/config.proto\x1a!common/protocol/server_spec.proto\"A\n" +
	"\aAccount\x12\x1a\n" +
	"\busername\x18\x01 \x01(\tR\busername\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\"\xc9\x03\n" +
	"\fServerConfig\x12=\n" +
	"\tauth_type\x18\x01 \x01(\x0e2 .v2ray.core.proxy.socks.AuthTypeR\bauthType\x12N\n" +
	"\baccounts\x18\x02 \x03(\v22.v2ray.core.proxy.socks.ServerConfig.AccountsEntryR\baccounts\x12;\n" +
	"\aaddress\x18\x03 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x1f\n" +
	"\vudp_enabled\x18\x04 \x01(\bR\n" +
	"udpEnabled\x12\x1c\n" +
	"\atimeout\x18\x05 \x01(\rB\x02\x18\x01R\atimeout\x12\x1d\n" +
	"\n" +
	"user_level\x18\x06 \x01(\rR\tuserLevel\x12R\n" +
	"\x0fpacket_encoding\x18\a \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding\x1a;\n" +
	"\rAccountsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\xb7\x01\n" +
	"\fClientConfig\x12B\n" +
	"\x06server\x18\x01 \x03(\v2*.v2ray.core.common.protocol.ServerEndpointR\x06server\x129\n" +
	"\aversion\x18\x02 \x01(\x0e2\x1f.v2ray.core.proxy.socks.VersionR\aversion\x12(\n" +
	"\x10delay_auth_write\x18\x03 \x01(\bR\x0edelayAuthWrite*%\n" +
	"\bAuthType\x12\v\n" +
	"\aNO_AUTH\x10\x00\x12\f\n" +
	"\bPASSWORD\x10\x01*.\n" +
	"\aVersion\x12\n" +
	"\n" +
	"\x06SOCKS5\x10\x00\x12\n" +
	"\n" +
	"\x06SOCKS4\x10\x01\x12\v\n" +
	"\aSOCKS4A\x10\x02Bc\n" +
	"\x1acom.v2ray.core.proxy.socksP\x01Z*github.com/v2fly/v2ray-core/v5/proxy/socks\xaa\x02\x16V2Ray.Core.Proxy.Socksb\x06proto3"

var (
	file_proxy_socks_config_proto_rawDescOnce sync.Once
	file_proxy_socks_config_proto_rawDescData []byte
)

func file_proxy_socks_config_proto_rawDescGZIP() []byte {
	file_proxy_socks_config_proto_rawDescOnce.Do(func() {
		file_proxy_socks_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_socks_config_proto_rawDesc), len(file_proxy_socks_config_proto_rawDesc)))
	})
	return file_proxy_socks_config_proto_rawDescData
}

var file_proxy_socks_config_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proxy_socks_config_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proxy_socks_config_proto_goTypes = []any{
	(AuthType)(0),                   // 0: v2ray.core.proxy.socks.AuthType
	(Version)(0),                    // 1: v2ray.core.proxy.socks.Version
	(*Account)(nil),                 // 2: v2ray.core.proxy.socks.Account
	(*ServerConfig)(nil),            // 3: v2ray.core.proxy.socks.ServerConfig
	(*ClientConfig)(nil),            // 4: v2ray.core.proxy.socks.ClientConfig
	nil,                             // 5: v2ray.core.proxy.socks.ServerConfig.AccountsEntry
	(*net.IPOrDomain)(nil),          // 6: v2ray.core.common.net.IPOrDomain
	(packetaddr.PacketAddrType)(0),  // 7: v2ray.core.net.packetaddr.PacketAddrType
	(*protocol.ServerEndpoint)(nil), // 8: v2ray.core.common.protocol.ServerEndpoint
}
var file_proxy_socks_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.proxy.socks.ServerConfig.auth_type:type_name -> v2ray.core.proxy.socks.AuthType
	5, // 1: v2ray.core.proxy.socks.ServerConfig.accounts:type_name -> v2ray.core.proxy.socks.ServerConfig.AccountsEntry
	6, // 2: v2ray.core.proxy.socks.ServerConfig.address:type_name -> v2ray.core.common.net.IPOrDomain
	7, // 3: v2ray.core.proxy.socks.ServerConfig.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	8, // 4: v2ray.core.proxy.socks.ClientConfig.server:type_name -> v2ray.core.common.protocol.ServerEndpoint
	1, // 5: v2ray.core.proxy.socks.ClientConfig.version:type_name -> v2ray.core.proxy.socks.Version
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_proxy_socks_config_proto_init() }
func file_proxy_socks_config_proto_init() {
	if File_proxy_socks_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_socks_config_proto_rawDesc), len(file_proxy_socks_config_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_socks_config_proto_goTypes,
		DependencyIndexes: file_proxy_socks_config_proto_depIdxs,
		EnumInfos:         file_proxy_socks_config_proto_enumTypes,
		MessageInfos:      file_proxy_socks_config_proto_msgTypes,
	}.Build()
	File_proxy_socks_config_proto = out.File
	file_proxy_socks_config_proto_goTypes = nil
	file_proxy_socks_config_proto_depIdxs = nil
}
