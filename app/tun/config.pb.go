package tun

import (
	proxyman "github.com/v2fly/v2ray-core/v5/app/proxyman"
	routercommon "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	packetaddr "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
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
	state                 protoimpl.MessageState    `protogen:"open.v1"`
	Name                  string                    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Mtu                   uint32                    `protobuf:"varint,2,opt,name=mtu,proto3" json:"mtu,omitempty"`
	UserLevel             uint32                    `protobuf:"varint,3,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	PacketEncoding        packetaddr.PacketAddrType `protobuf:"varint,4,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	Tag                   string                    `protobuf:"bytes,5,opt,name=tag,proto3" json:"tag,omitempty"`
	Ips                   []*routercommon.CIDR      `protobuf:"bytes,6,rep,name=ips,proto3" json:"ips,omitempty"`
	Routes                []*routercommon.CIDR      `protobuf:"bytes,7,rep,name=routes,proto3" json:"routes,omitempty"`
	EnablePromiscuousMode bool                      `protobuf:"varint,8,opt,name=enable_promiscuous_mode,json=enablePromiscuousMode,proto3" json:"enable_promiscuous_mode,omitempty"`
	EnableSpoofing        bool                      `protobuf:"varint,9,opt,name=enable_spoofing,json=enableSpoofing,proto3" json:"enable_spoofing,omitempty"`
	SocketSettings        *internet.SocketConfig    `protobuf:"bytes,10,opt,name=socket_settings,json=socketSettings,proto3" json:"socket_settings,omitempty"`
	SniffingSettings      *proxyman.SniffingConfig  `protobuf:"bytes,11,opt,name=sniffing_settings,json=sniffingSettings,proto3" json:"sniffing_settings,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_tun_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_tun_config_proto_msgTypes[0]
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
	return file_app_tun_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
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

func (x *Config) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

func (x *Config) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
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

func (x *Config) GetSniffingSettings() *proxyman.SniffingConfig {
	if x != nil {
		return x.SniffingSettings
	}
	return nil
}

var File_app_tun_config_proto protoreflect.FileDescriptor

const file_app_tun_config_proto_rawDesc = "" +
	"\n" +
	"\x14app/tun/config.proto\x12\x12v2ray.core.app.tun\x1a\x19app/proxyman/config.proto\x1a$app/router/routercommon/common.proto\x1a common/protoext/extensions.proto\x1a\"common/net/packetaddr/config.proto\x1a\x1ftransport/internet/config.proto\"\xd2\x04\n" +
	"\x06Config\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x10\n" +
	"\x03mtu\x18\x02 \x01(\rR\x03mtu\x12\x1d\n" +
	"\n" +
	"user_level\x18\x03 \x01(\rR\tuserLevel\x12R\n" +
	"\x0fpacket_encoding\x18\x04 \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding\x12\x10\n" +
	"\x03tag\x18\x05 \x01(\tR\x03tag\x12:\n" +
	"\x03ips\x18\x06 \x03(\v2(.v2ray.core.app.router.routercommon.CIDRR\x03ips\x12@\n" +
	"\x06routes\x18\a \x03(\v2(.v2ray.core.app.router.routercommon.CIDRR\x06routes\x126\n" +
	"\x17enable_promiscuous_mode\x18\b \x01(\bR\x15enablePromiscuousMode\x12'\n" +
	"\x0fenable_spoofing\x18\t \x01(\bR\x0eenableSpoofing\x12T\n" +
	"\x0fsocket_settings\x18\n" +
	" \x01(\v2+.v2ray.core.transport.internet.SocketConfigR\x0esocketSettings\x12T\n" +
	"\x11sniffing_settings\x18\v \x01(\v2'.v2ray.core.app.proxyman.SniffingConfigR\x10sniffingSettings:\x12\x82\xb5\x18\x0e\n" +
	"\aservice\x12\x03tunBW\n" +
	"\x16com.v2ray.core.app.tunP\x01Z&github.com/v2fly/v2ray-core/v5/app/tun\xaa\x02\x12V2Ray.Core.App.Tunb\x06proto3"

var (
	file_app_tun_config_proto_rawDescOnce sync.Once
	file_app_tun_config_proto_rawDescData []byte
)

func file_app_tun_config_proto_rawDescGZIP() []byte {
	file_app_tun_config_proto_rawDescOnce.Do(func() {
		file_app_tun_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_tun_config_proto_rawDesc), len(file_app_tun_config_proto_rawDesc)))
	})
	return file_app_tun_config_proto_rawDescData
}

var file_app_tun_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_app_tun_config_proto_goTypes = []any{
	(*Config)(nil),                  // 0: v2ray.core.app.tun.Config
	(packetaddr.PacketAddrType)(0),  // 1: v2ray.core.net.packetaddr.PacketAddrType
	(*routercommon.CIDR)(nil),       // 2: v2ray.core.app.router.routercommon.CIDR
	(*internet.SocketConfig)(nil),   // 3: v2ray.core.transport.internet.SocketConfig
	(*proxyman.SniffingConfig)(nil), // 4: v2ray.core.app.proxyman.SniffingConfig
}
var file_app_tun_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.app.tun.Config.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	2, // 1: v2ray.core.app.tun.Config.ips:type_name -> v2ray.core.app.router.routercommon.CIDR
	2, // 2: v2ray.core.app.tun.Config.routes:type_name -> v2ray.core.app.router.routercommon.CIDR
	3, // 3: v2ray.core.app.tun.Config.socket_settings:type_name -> v2ray.core.transport.internet.SocketConfig
	4, // 4: v2ray.core.app.tun.Config.sniffing_settings:type_name -> v2ray.core.app.proxyman.SniffingConfig
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_app_tun_config_proto_init() }
func file_app_tun_config_proto_init() {
	if File_app_tun_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_tun_config_proto_rawDesc), len(file_app_tun_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_tun_config_proto_goTypes,
		DependencyIndexes: file_app_tun_config_proto_depIdxs,
		MessageInfos:      file_app_tun_config_proto_msgTypes,
	}.Build()
	File_app_tun_config_proto = out.File
	file_app_tun_config_proto_goTypes = nil
	file_app_tun_config_proto_depIdxs = nil
}
