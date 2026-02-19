package gvisorstack

import (
	routercommon "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	internet "github.com/v2fly/v2ray-core/v5/transport/internet"
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
	state                 protoimpl.MessageState `protogen:"open.v1"`
	Mtu                   uint32                 `protobuf:"varint,2,opt,name=mtu,proto3" json:"mtu,omitempty"`
	UserLevel             uint32                 `protobuf:"varint,3,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	Ips                   []*routercommon.CIDR   `protobuf:"bytes,6,rep,name=ips,proto3" json:"ips,omitempty"`
	Routes                []*routercommon.CIDR   `protobuf:"bytes,7,rep,name=routes,proto3" json:"routes,omitempty"`
	EnablePromiscuousMode bool                   `protobuf:"varint,8,opt,name=enable_promiscuous_mode,json=enablePromiscuousMode,proto3" json:"enable_promiscuous_mode,omitempty"`
	EnableSpoofing        bool                   `protobuf:"varint,9,opt,name=enable_spoofing,json=enableSpoofing,proto3" json:"enable_spoofing,omitempty"`
	SocketSettings        *internet.SocketConfig `protobuf:"bytes,10,opt,name=socket_settings,json=socketSettings,proto3" json:"socket_settings,omitempty"`
	PreferIpv6ForUdp      bool                   `protobuf:"varint,11,opt,name=prefer_ipv6_for_udp,json=preferIpv6ForUdp,proto3" json:"prefer_ipv6_for_udp,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_common_packetswitch_gvisorstack_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_common_packetswitch_gvisorstack_config_proto_msgTypes[0]
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
	return file_common_packetswitch_gvisorstack_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *Config) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

func (x *Config) GetIps() []*routercommon.CIDR {
	if x != nil {
		return x.Ips
	}
	return nil
}

func (x *Config) GetRoutes() []*routercommon.CIDR {
	if x != nil {
		return x.Routes
	}
	return nil
}

func (x *Config) GetEnablePromiscuousMode() bool {
	if x != nil {
		return x.EnablePromiscuousMode
	}
	return false
}

func (x *Config) GetEnableSpoofing() bool {
	if x != nil {
		return x.EnableSpoofing
	}
	return false
}

func (x *Config) GetSocketSettings() *internet.SocketConfig {
	if x != nil {
		return x.SocketSettings
	}
	return nil
}

func (x *Config) GetPreferIpv6ForUdp() bool {
	if x != nil {
		return x.PreferIpv6ForUdp
	}
	return false
}

var File_common_packetswitch_gvisorstack_config_proto protoreflect.FileDescriptor

const file_common_packetswitch_gvisorstack_config_proto_rawDesc = "" +
	"\n" +
	",common/packetswitch/gvisorstack/config.proto\x12*v2ray.core.common.packetswitch.gvisorstack\x1a$app/router/routercommon/common.proto\x1a\x1ftransport/internet/config.proto\x1a common/protoext/extensions.proto\"\x9d\x03\n" +
	"\x06Config\x12\x10\n" +
	"\x03mtu\x18\x02 \x01(\rR\x03mtu\x12\x1d\n" +
	"\n" +
	"user_level\x18\x03 \x01(\rR\tuserLevel\x12:\n" +
	"\x03ips\x18\x06 \x03(\v2(.v2ray.core.app.router.routercommon.CIDRR\x03ips\x12@\n" +
	"\x06routes\x18\a \x03(\v2(.v2ray.core.app.router.routercommon.CIDRR\x06routes\x126\n" +
	"\x17enable_promiscuous_mode\x18\b \x01(\bR\x15enablePromiscuousMode\x12'\n" +
	"\x0fenable_spoofing\x18\t \x01(\bR\x0eenableSpoofing\x12T\n" +
	"\x0fsocket_settings\x18\n" +
	" \x01(\v2+.v2ray.core.transport.internet.SocketConfigR\x0esocketSettings\x12-\n" +
	"\x13prefer_ipv6_for_udp\x18\v \x01(\bR\x10preferIpv6ForUdpB\x9f\x01\n" +
	".com.v2ray.core.common.packetswitch.gvisorstackP\x01Z>github.com/v2fly/v2ray-core/v5/common/packetswitch/gvisorstack\xaa\x02*V2Ray.Core.Common.Packetswitch.Gvisorstackb\x06proto3"

var (
	file_common_packetswitch_gvisorstack_config_proto_rawDescOnce sync.Once
	file_common_packetswitch_gvisorstack_config_proto_rawDescData []byte
)

func file_common_packetswitch_gvisorstack_config_proto_rawDescGZIP() []byte {
	file_common_packetswitch_gvisorstack_config_proto_rawDescOnce.Do(func() {
		file_common_packetswitch_gvisorstack_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_packetswitch_gvisorstack_config_proto_rawDesc), len(file_common_packetswitch_gvisorstack_config_proto_rawDesc)))
	})
	return file_common_packetswitch_gvisorstack_config_proto_rawDescData
}

var file_common_packetswitch_gvisorstack_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_packetswitch_gvisorstack_config_proto_goTypes = []any{
	(*Config)(nil),                // 0: v2ray.core.common.packetswitch.gvisorstack.Config
	(*routercommon.CIDR)(nil),     // 1: v2ray.core.app.router.routercommon.CIDR
	(*internet.SocketConfig)(nil), // 2: v2ray.core.transport.internet.SocketConfig
}
var file_common_packetswitch_gvisorstack_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.common.packetswitch.gvisorstack.Config.ips:type_name -> v2ray.core.app.router.routercommon.CIDR
	1, // 1: v2ray.core.common.packetswitch.gvisorstack.Config.routes:type_name -> v2ray.core.app.router.routercommon.CIDR
	2, // 2: v2ray.core.common.packetswitch.gvisorstack.Config.socket_settings:type_name -> v2ray.core.transport.internet.SocketConfig
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_common_packetswitch_gvisorstack_config_proto_init() }
func file_common_packetswitch_gvisorstack_config_proto_init() {
	if File_common_packetswitch_gvisorstack_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_packetswitch_gvisorstack_config_proto_rawDesc), len(file_common_packetswitch_gvisorstack_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_packetswitch_gvisorstack_config_proto_goTypes,
		DependencyIndexes: file_common_packetswitch_gvisorstack_config_proto_depIdxs,
		MessageInfos:      file_common_packetswitch_gvisorstack_config_proto_msgTypes,
	}.Build()
	File_common_packetswitch_gvisorstack_config_proto = out.File
	file_common_packetswitch_gvisorstack_config_proto_goTypes = nil
	file_common_packetswitch_gvisorstack_config_proto_depIdxs = nil
}
