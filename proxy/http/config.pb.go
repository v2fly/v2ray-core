package http

import (
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
	Username      string                 `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Account) Reset() {
	*x = Account{}
	mi := &file_proxy_http_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_http_config_proto_msgTypes[0]
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
	return file_proxy_http_config_proto_rawDescGZIP(), []int{0}
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

// Config for HTTP proxy server.
type ServerConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Deprecated: Marked as deprecated in proxy/http/config.proto.
	Timeout          uint32            `protobuf:"varint,1,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Accounts         map[string]string `protobuf:"bytes,2,rep,name=accounts,proto3" json:"accounts,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	AllowTransparent bool              `protobuf:"varint,3,opt,name=allow_transparent,json=allowTransparent,proto3" json:"allow_transparent,omitempty"`
	UserLevel        uint32            `protobuf:"varint,4,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_proxy_http_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_http_config_proto_msgTypes[1]
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
	return file_proxy_http_config_proto_rawDescGZIP(), []int{1}
}

// Deprecated: Marked as deprecated in proxy/http/config.proto.
func (x *ServerConfig) GetTimeout() uint32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *ServerConfig) GetAccounts() map[string]string {
	if x != nil {
		return x.Accounts
	}
	return nil
}

func (x *ServerConfig) GetAllowTransparent() bool {
	if x != nil {
		return x.AllowTransparent
	}
	return false
}

func (x *ServerConfig) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

// ClientConfig is the protobuf config for HTTP proxy client.
type ClientConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Sever is a list of HTTP server addresses.
	Server             []*protocol.ServerEndpoint `protobuf:"bytes,1,rep,name=server,proto3" json:"server,omitempty"`
	H1SkipWaitForReply bool                       `protobuf:"varint,2,opt,name=h1_skip_wait_for_reply,json=h1SkipWaitForReply,proto3" json:"h1_skip_wait_for_reply,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_proxy_http_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_http_config_proto_msgTypes[2]
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
	return file_proxy_http_config_proto_rawDescGZIP(), []int{2}
}

func (x *ClientConfig) GetServer() []*protocol.ServerEndpoint {
	if x != nil {
		return x.Server
	}
	return nil
}

func (x *ClientConfig) GetH1SkipWaitForReply() bool {
	if x != nil {
		return x.H1SkipWaitForReply
	}
	return false
}

var File_proxy_http_config_proto protoreflect.FileDescriptor

const file_proxy_http_config_proto_rawDesc = "" +
	"\n" +
	"\x17proxy/http/config.proto\x12\x15v2ray.core.proxy.http\x1a!common/protocol/server_spec.proto\"A\n" +
	"\aAccount\x12\x1a\n" +
	"\busername\x18\x01 \x01(\tR\busername\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\"\x84\x02\n" +
	"\fServerConfig\x12\x1c\n" +
	"\atimeout\x18\x01 \x01(\rB\x02\x18\x01R\atimeout\x12M\n" +
	"\baccounts\x18\x02 \x03(\v21.v2ray.core.proxy.http.ServerConfig.AccountsEntryR\baccounts\x12+\n" +
	"\x11allow_transparent\x18\x03 \x01(\bR\x10allowTransparent\x12\x1d\n" +
	"\n" +
	"user_level\x18\x04 \x01(\rR\tuserLevel\x1a;\n" +
	"\rAccountsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x86\x01\n" +
	"\fClientConfig\x12B\n" +
	"\x06server\x18\x01 \x03(\v2*.v2ray.core.common.protocol.ServerEndpointR\x06server\x122\n" +
	"\x16h1_skip_wait_for_reply\x18\x02 \x01(\bR\x12h1SkipWaitForReplyB`\n" +
	"\x19com.v2ray.core.proxy.httpP\x01Z)github.com/v2fly/v2ray-core/v5/proxy/http\xaa\x02\x15V2Ray.Core.Proxy.Httpb\x06proto3"

var (
	file_proxy_http_config_proto_rawDescOnce sync.Once
	file_proxy_http_config_proto_rawDescData []byte
)

func file_proxy_http_config_proto_rawDescGZIP() []byte {
	file_proxy_http_config_proto_rawDescOnce.Do(func() {
		file_proxy_http_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_http_config_proto_rawDesc), len(file_proxy_http_config_proto_rawDesc)))
	})
	return file_proxy_http_config_proto_rawDescData
}

var file_proxy_http_config_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proxy_http_config_proto_goTypes = []any{
	(*Account)(nil),                 // 0: v2ray.core.proxy.http.Account
	(*ServerConfig)(nil),            // 1: v2ray.core.proxy.http.ServerConfig
	(*ClientConfig)(nil),            // 2: v2ray.core.proxy.http.ClientConfig
	nil,                             // 3: v2ray.core.proxy.http.ServerConfig.AccountsEntry
	(*protocol.ServerEndpoint)(nil), // 4: v2ray.core.common.protocol.ServerEndpoint
}
var file_proxy_http_config_proto_depIdxs = []int32{
	3, // 0: v2ray.core.proxy.http.ServerConfig.accounts:type_name -> v2ray.core.proxy.http.ServerConfig.AccountsEntry
	4, // 1: v2ray.core.proxy.http.ClientConfig.server:type_name -> v2ray.core.common.protocol.ServerEndpoint
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proxy_http_config_proto_init() }
func file_proxy_http_config_proto_init() {
	if File_proxy_http_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_http_config_proto_rawDesc), len(file_proxy_http_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_http_config_proto_goTypes,
		DependencyIndexes: file_proxy_http_config_proto_depIdxs,
		MessageInfos:      file_proxy_http_config_proto_msgTypes,
	}.Build()
	File_proxy_http_config_proto = out.File
	file_proxy_http_config_proto_goTypes = nil
	file_proxy_http_config_proto_depIdxs = nil
}
