package dns

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
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
	state protoimpl.MessageState `protogen:"open.v1"`
	// Server is the DNS server address. If specified, this address overrides the
	// original one.
	Server              *net.Endpoint `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
	UserLevel           uint32        `protobuf:"varint,2,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	OverrideResponseTtl bool          `protobuf:"varint,4,opt,name=override_response_ttl,json=overrideResponseTtl,proto3" json:"override_response_ttl,omitempty"`
	ResponseTtl         uint32        `protobuf:"varint,3,opt,name=response_ttl,json=responseTtl,proto3" json:"response_ttl,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_proxy_dns_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_dns_config_proto_msgTypes[0]
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
	return file_proxy_dns_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetServer() *net.Endpoint {
	if x != nil {
		return x.Server
	}
	return nil
}

func (x *Config) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

func (x *Config) GetOverrideResponseTtl() bool {
	if x != nil {
		return x.OverrideResponseTtl
	}
	return false
}

func (x *Config) GetResponseTtl() uint32 {
	if x != nil {
		return x.ResponseTtl
	}
	return 0
}

type SimplifiedConfig struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	OverrideResponseTtl bool                   `protobuf:"varint,4,opt,name=override_response_ttl,json=overrideResponseTtl,proto3" json:"override_response_ttl,omitempty"`
	ResponseTtl         uint32                 `protobuf:"varint,3,opt,name=response_ttl,json=responseTtl,proto3" json:"response_ttl,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *SimplifiedConfig) Reset() {
	*x = SimplifiedConfig{}
	mi := &file_proxy_dns_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedConfig) ProtoMessage() {}

func (x *SimplifiedConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_dns_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimplifiedConfig.ProtoReflect.Descriptor instead.
func (*SimplifiedConfig) Descriptor() ([]byte, []int) {
	return file_proxy_dns_config_proto_rawDescGZIP(), []int{1}
}

func (x *SimplifiedConfig) GetOverrideResponseTtl() bool {
	if x != nil {
		return x.OverrideResponseTtl
	}
	return false
}

func (x *SimplifiedConfig) GetResponseTtl() uint32 {
	if x != nil {
		return x.ResponseTtl
	}
	return 0
}

var File_proxy_dns_config_proto protoreflect.FileDescriptor

const file_proxy_dns_config_proto_rawDesc = "" +
	"\n" +
	"\x16proxy/dns/config.proto\x12\x14v2ray.core.proxy.dns\x1a\x1ccommon/net/destination.proto\x1a common/protoext/extensions.proto\"\xb7\x01\n" +
	"\x06Config\x127\n" +
	"\x06server\x18\x01 \x01(\v2\x1f.v2ray.core.common.net.EndpointR\x06server\x12\x1d\n" +
	"\n" +
	"user_level\x18\x02 \x01(\rR\tuserLevel\x122\n" +
	"\x15override_response_ttl\x18\x04 \x01(\bR\x13overrideResponseTtl\x12!\n" +
	"\fresponse_ttl\x18\x03 \x01(\rR\vresponseTtl\"~\n" +
	"\x10SimplifiedConfig\x122\n" +
	"\x15override_response_ttl\x18\x04 \x01(\bR\x13overrideResponseTtl\x12!\n" +
	"\fresponse_ttl\x18\x03 \x01(\rR\vresponseTtl:\x13\x82\xb5\x18\x0f\n" +
	"\boutbound\x12\x03dnsB]\n" +
	"\x18com.v2ray.core.proxy.dnsP\x01Z(github.com/v2fly/v2ray-core/v5/proxy/dns\xaa\x02\x14V2Ray.Core.Proxy.Dnsb\x06proto3"

var (
	file_proxy_dns_config_proto_rawDescOnce sync.Once
	file_proxy_dns_config_proto_rawDescData []byte
)

func file_proxy_dns_config_proto_rawDescGZIP() []byte {
	file_proxy_dns_config_proto_rawDescOnce.Do(func() {
		file_proxy_dns_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_dns_config_proto_rawDesc), len(file_proxy_dns_config_proto_rawDesc)))
	})
	return file_proxy_dns_config_proto_rawDescData
}

var file_proxy_dns_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proxy_dns_config_proto_goTypes = []any{
	(*Config)(nil),           // 0: v2ray.core.proxy.dns.Config
	(*SimplifiedConfig)(nil), // 1: v2ray.core.proxy.dns.SimplifiedConfig
	(*net.Endpoint)(nil),     // 2: v2ray.core.common.net.Endpoint
}
var file_proxy_dns_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.proxy.dns.Config.server:type_name -> v2ray.core.common.net.Endpoint
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proxy_dns_config_proto_init() }
func file_proxy_dns_config_proto_init() {
	if File_proxy_dns_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_dns_config_proto_rawDesc), len(file_proxy_dns_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_dns_config_proto_goTypes,
		DependencyIndexes: file_proxy_dns_config_proto_depIdxs,
		MessageInfos:      file_proxy_dns_config_proto_msgTypes,
	}.Build()
	File_proxy_dns_config_proto = out.File
	file_proxy_dns_config_proto_goTypes = nil
	file_proxy_dns_config_proto_depIdxs = nil
}
