package inbound

import (
	protocol "github.com/v2fly/v2ray-core/v5/common/protocol"
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

type DetourConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	To            string                 `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DetourConfig) Reset() {
	*x = DetourConfig{}
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DetourConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DetourConfig) ProtoMessage() {}

func (x *DetourConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DetourConfig.ProtoReflect.Descriptor instead.
func (*DetourConfig) Descriptor() ([]byte, []int) {
	return file_proxy_vmess_inbound_config_proto_rawDescGZIP(), []int{0}
}

func (x *DetourConfig) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

type DefaultConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AlterId       uint32                 `protobuf:"varint,1,opt,name=alter_id,json=alterId,proto3" json:"alter_id,omitempty"`
	Level         uint32                 `protobuf:"varint,2,opt,name=level,proto3" json:"level,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DefaultConfig) Reset() {
	*x = DefaultConfig{}
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DefaultConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DefaultConfig) ProtoMessage() {}

func (x *DefaultConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DefaultConfig.ProtoReflect.Descriptor instead.
func (*DefaultConfig) Descriptor() ([]byte, []int) {
	return file_proxy_vmess_inbound_config_proto_rawDescGZIP(), []int{1}
}

func (x *DefaultConfig) GetAlterId() uint32 {
	if x != nil {
		return x.AlterId
	}
	return 0
}

func (x *DefaultConfig) GetLevel() uint32 {
	if x != nil {
		return x.Level
	}
	return 0
}

type Config struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	User                 []*protocol.User       `protobuf:"bytes,1,rep,name=user,proto3" json:"user,omitempty"`
	Default              *DefaultConfig         `protobuf:"bytes,2,opt,name=default,proto3" json:"default,omitempty"`
	Detour               *DetourConfig          `protobuf:"bytes,3,opt,name=detour,proto3" json:"detour,omitempty"`
	SecureEncryptionOnly bool                   `protobuf:"varint,4,opt,name=secure_encryption_only,json=secureEncryptionOnly,proto3" json:"secure_encryption_only,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[2]
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
	return file_proxy_vmess_inbound_config_proto_rawDescGZIP(), []int{2}
}

func (x *Config) GetUser() []*protocol.User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *Config) GetDefault() *DefaultConfig {
	if x != nil {
		return x.Default
	}
	return nil
}

func (x *Config) GetDetour() *DetourConfig {
	if x != nil {
		return x.Detour
	}
	return nil
}

func (x *Config) GetSecureEncryptionOnly() bool {
	if x != nil {
		return x.SecureEncryptionOnly
	}
	return false
}

type SimplifiedConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []string               `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SimplifiedConfig) Reset() {
	*x = SimplifiedConfig{}
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedConfig) ProtoMessage() {}

func (x *SimplifiedConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vmess_inbound_config_proto_msgTypes[3]
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
	return file_proxy_vmess_inbound_config_proto_rawDescGZIP(), []int{3}
}

func (x *SimplifiedConfig) GetUsers() []string {
	if x != nil {
		return x.Users
	}
	return nil
}

var File_proxy_vmess_inbound_config_proto protoreflect.FileDescriptor

const file_proxy_vmess_inbound_config_proto_rawDesc = "" +
	"\n" +
	" proxy/vmess/inbound/config.proto\x12\x1ev2ray.core.proxy.vmess.inbound\x1a\x1acommon/protocol/user.proto\x1a common/protoext/extensions.proto\"\x1e\n" +
	"\fDetourConfig\x12\x0e\n" +
	"\x02to\x18\x01 \x01(\tR\x02to\"@\n" +
	"\rDefaultConfig\x12\x19\n" +
	"\balter_id\x18\x01 \x01(\rR\aalterId\x12\x14\n" +
	"\x05level\x18\x02 \x01(\rR\x05level\"\x83\x02\n" +
	"\x06Config\x124\n" +
	"\x04user\x18\x01 \x03(\v2 .v2ray.core.common.protocol.UserR\x04user\x12G\n" +
	"\adefault\x18\x02 \x01(\v2-.v2ray.core.proxy.vmess.inbound.DefaultConfigR\adefault\x12D\n" +
	"\x06detour\x18\x03 \x01(\v2,.v2ray.core.proxy.vmess.inbound.DetourConfigR\x06detour\x124\n" +
	"\x16secure_encryption_only\x18\x04 \x01(\bR\x14secureEncryptionOnly\">\n" +
	"\x10SimplifiedConfig\x12\x14\n" +
	"\x05users\x18\x01 \x03(\tR\x05users:\x14\x82\xb5\x18\x10\n" +
	"\ainbound\x12\x05vmessB{\n" +
	"\"com.v2ray.core.proxy.vmess.inboundP\x01Z2github.com/v2fly/v2ray-core/v5/proxy/vmess/inbound\xaa\x02\x1eV2Ray.Core.Proxy.Vmess.Inboundb\x06proto3"

var (
	file_proxy_vmess_inbound_config_proto_rawDescOnce sync.Once
	file_proxy_vmess_inbound_config_proto_rawDescData []byte
)

func file_proxy_vmess_inbound_config_proto_rawDescGZIP() []byte {
	file_proxy_vmess_inbound_config_proto_rawDescOnce.Do(func() {
		file_proxy_vmess_inbound_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_vmess_inbound_config_proto_rawDesc), len(file_proxy_vmess_inbound_config_proto_rawDesc)))
	})
	return file_proxy_vmess_inbound_config_proto_rawDescData
}

var file_proxy_vmess_inbound_config_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proxy_vmess_inbound_config_proto_goTypes = []any{
	(*DetourConfig)(nil),     // 0: v2ray.core.proxy.vmess.inbound.DetourConfig
	(*DefaultConfig)(nil),    // 1: v2ray.core.proxy.vmess.inbound.DefaultConfig
	(*Config)(nil),           // 2: v2ray.core.proxy.vmess.inbound.Config
	(*SimplifiedConfig)(nil), // 3: v2ray.core.proxy.vmess.inbound.SimplifiedConfig
	(*protocol.User)(nil),    // 4: v2ray.core.common.protocol.User
}
var file_proxy_vmess_inbound_config_proto_depIdxs = []int32{
	4, // 0: v2ray.core.proxy.vmess.inbound.Config.user:type_name -> v2ray.core.common.protocol.User
	1, // 1: v2ray.core.proxy.vmess.inbound.Config.default:type_name -> v2ray.core.proxy.vmess.inbound.DefaultConfig
	0, // 2: v2ray.core.proxy.vmess.inbound.Config.detour:type_name -> v2ray.core.proxy.vmess.inbound.DetourConfig
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proxy_vmess_inbound_config_proto_init() }
func file_proxy_vmess_inbound_config_proto_init() {
	if File_proxy_vmess_inbound_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_vmess_inbound_config_proto_rawDesc), len(file_proxy_vmess_inbound_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_vmess_inbound_config_proto_goTypes,
		DependencyIndexes: file_proxy_vmess_inbound_config_proto_depIdxs,
		MessageInfos:      file_proxy_vmess_inbound_config_proto_msgTypes,
	}.Build()
	File_proxy_vmess_inbound_config_proto = out.File
	file_proxy_vmess_inbound_config_proto_goTypes = nil
	file_proxy_vmess_inbound_config_proto_depIdxs = nil
}
