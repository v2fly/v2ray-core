package outbound

import (
	_ "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	gvisorstack "github.com/v2fly/v2ray-core/v5/common/packetswitch/gvisorstack"
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	wgcommon "github.com/v2fly/v2ray-core/v5/proxy/wireguard/wgcommon"
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

type Config_DomainStrategy int32

const (
	Config_AS_IS   Config_DomainStrategy = 0
	Config_USE_IP  Config_DomainStrategy = 1
	Config_USE_IP4 Config_DomainStrategy = 2
	Config_USE_IP6 Config_DomainStrategy = 3
)

// Enum value maps for Config_DomainStrategy.
var (
	Config_DomainStrategy_name = map[int32]string{
		0: "AS_IS",
		1: "USE_IP",
		2: "USE_IP4",
		3: "USE_IP6",
	}
	Config_DomainStrategy_value = map[string]int32{
		"AS_IS":   0,
		"USE_IP":  1,
		"USE_IP4": 2,
		"USE_IP6": 3,
	}
)

func (x Config_DomainStrategy) Enum() *Config_DomainStrategy {
	p := new(Config_DomainStrategy)
	*p = x
	return p
}

func (x Config_DomainStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Config_DomainStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_proxy_wireguard_outbound_config_proto_enumTypes[0].Descriptor()
}

func (Config_DomainStrategy) Type() protoreflect.EnumType {
	return &file_proxy_wireguard_outbound_config_proto_enumTypes[0]
}

func (x Config_DomainStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Config_DomainStrategy.Descriptor instead.
func (Config_DomainStrategy) EnumDescriptor() ([]byte, []int) {
	return file_proxy_wireguard_outbound_config_proto_rawDescGZIP(), []int{0, 0}
}

type Config struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	WgDevice *wgcommon.DeviceConfig `protobuf:"bytes,1,opt,name=wg_device,json=wgDevice,proto3" json:"wg_device,omitempty"`
	Stack    *gvisorstack.Config    `protobuf:"bytes,2,opt,name=stack,proto3" json:"stack,omitempty"`
	// v2ray.core.net.packetaddr.PacketAddrType outbound_packet_encoding = 3;
	ListenOnSystemNetwork bool                  `protobuf:"varint,4,opt,name=listen_on_system_network,json=listenOnSystemNetwork,proto3" json:"listen_on_system_network,omitempty"`
	DomainStrategy        Config_DomainStrategy `protobuf:"varint,5,opt,name=domain_strategy,json=domainStrategy,proto3,enum=v2ray.core.proxy.wireguard.outbound.Config_DomainStrategy" json:"domain_strategy,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_proxy_wireguard_outbound_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_wireguard_outbound_config_proto_msgTypes[0]
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
	return file_proxy_wireguard_outbound_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetWgDevice() *wgcommon.DeviceConfig {
	if x != nil {
		return x.WgDevice
	}
	return nil
}

func (x *Config) GetStack() *gvisorstack.Config {
	if x != nil {
		return x.Stack
	}
	return nil
}

func (x *Config) GetListenOnSystemNetwork() bool {
	if x != nil {
		return x.ListenOnSystemNetwork
	}
	return false
}

func (x *Config) GetDomainStrategy() Config_DomainStrategy {
	if x != nil {
		return x.DomainStrategy
	}
	return Config_AS_IS
}

var File_proxy_wireguard_outbound_config_proto protoreflect.FileDescriptor

const file_proxy_wireguard_outbound_config_proto_rawDesc = "" +
	"\n" +
	"%proxy/wireguard/outbound/config.proto\x12#v2ray.core.proxy.wireguard.outbound\x1a%proxy/wireguard/wgcommon/config.proto\x1a,common/packetswitch/gvisorstack/config.proto\x1a\"common/net/packetaddr/config.proto\x1a common/protoext/extensions.proto\"\x9e\x03\n" +
	"\x06Config\x12N\n" +
	"\twg_device\x18\x01 \x01(\v21.v2ray.core.proxy.wireguard.wgcommon.DeviceConfigR\bwgDevice\x12H\n" +
	"\x05stack\x18\x02 \x01(\v22.v2ray.core.common.packetswitch.gvisorstack.ConfigR\x05stack\x127\n" +
	"\x18listen_on_system_network\x18\x04 \x01(\bR\x15listenOnSystemNetwork\x12c\n" +
	"\x0fdomain_strategy\x18\x05 \x01(\x0e2:.v2ray.core.proxy.wireguard.outbound.Config.DomainStrategyR\x0edomainStrategy\"A\n" +
	"\x0eDomainStrategy\x12\t\n" +
	"\x05AS_IS\x10\x00\x12\n" +
	"\n" +
	"\x06USE_IP\x10\x01\x12\v\n" +
	"\aUSE_IP4\x10\x02\x12\v\n" +
	"\aUSE_IP6\x10\x03:\x19\x82\xb5\x18\x15\n" +
	"\boutbound\x12\twireguardB\x8a\x01\n" +
	"'com.v2ray.core.proxy.wireguard.outboundP\x01Z7github.com/v2fly/v2ray-core/v5/proxy/wireguard/outbound\xaa\x02#V2Ray.Core.Proxy.Wireguard.Outboundb\x06proto3"

var (
	file_proxy_wireguard_outbound_config_proto_rawDescOnce sync.Once
	file_proxy_wireguard_outbound_config_proto_rawDescData []byte
)

func file_proxy_wireguard_outbound_config_proto_rawDescGZIP() []byte {
	file_proxy_wireguard_outbound_config_proto_rawDescOnce.Do(func() {
		file_proxy_wireguard_outbound_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_wireguard_outbound_config_proto_rawDesc), len(file_proxy_wireguard_outbound_config_proto_rawDesc)))
	})
	return file_proxy_wireguard_outbound_config_proto_rawDescData
}

var file_proxy_wireguard_outbound_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proxy_wireguard_outbound_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proxy_wireguard_outbound_config_proto_goTypes = []any{
	(Config_DomainStrategy)(0),    // 0: v2ray.core.proxy.wireguard.outbound.Config.DomainStrategy
	(*Config)(nil),                // 1: v2ray.core.proxy.wireguard.outbound.Config
	(*wgcommon.DeviceConfig)(nil), // 2: v2ray.core.proxy.wireguard.wgcommon.DeviceConfig
	(*gvisorstack.Config)(nil),    // 3: v2ray.core.common.packetswitch.gvisorstack.Config
}
var file_proxy_wireguard_outbound_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.proxy.wireguard.outbound.Config.wg_device:type_name -> v2ray.core.proxy.wireguard.wgcommon.DeviceConfig
	3, // 1: v2ray.core.proxy.wireguard.outbound.Config.stack:type_name -> v2ray.core.common.packetswitch.gvisorstack.Config
	0, // 2: v2ray.core.proxy.wireguard.outbound.Config.domain_strategy:type_name -> v2ray.core.proxy.wireguard.outbound.Config.DomainStrategy
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proxy_wireguard_outbound_config_proto_init() }
func file_proxy_wireguard_outbound_config_proto_init() {
	if File_proxy_wireguard_outbound_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_wireguard_outbound_config_proto_rawDesc), len(file_proxy_wireguard_outbound_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_wireguard_outbound_config_proto_goTypes,
		DependencyIndexes: file_proxy_wireguard_outbound_config_proto_depIdxs,
		EnumInfos:         file_proxy_wireguard_outbound_config_proto_enumTypes,
		MessageInfos:      file_proxy_wireguard_outbound_config_proto_msgTypes,
	}.Build()
	File_proxy_wireguard_outbound_config_proto = out.File
	file_proxy_wireguard_outbound_config_proto_goTypes = nil
	file_proxy_wireguard_outbound_config_proto_depIdxs = nil
}
