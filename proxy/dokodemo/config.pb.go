package dokodemo

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
	state   protoimpl.MessageState `protogen:"open.v1"`
	Address *net.IPOrDomain        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port    uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// List of networks that the Dokodemo accepts.
	// Deprecated. Use networks.
	//
	// Deprecated: Marked as deprecated in proxy/dokodemo/config.proto.
	NetworkList *net.NetworkList `protobuf:"bytes,3,opt,name=network_list,json=networkList,proto3" json:"network_list,omitempty"`
	// List of networks that the Dokodemo accepts.
	Networks []net.Network `protobuf:"varint,7,rep,packed,name=networks,proto3,enum=v2ray.core.common.net.Network" json:"networks,omitempty"`
	// Deprecated: Marked as deprecated in proxy/dokodemo/config.proto.
	Timeout        uint32 `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
	FollowRedirect bool   `protobuf:"varint,5,opt,name=follow_redirect,json=followRedirect,proto3" json:"follow_redirect,omitempty"`
	UserLevel      uint32 `protobuf:"varint,6,opt,name=user_level,json=userLevel,proto3" json:"user_level,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_proxy_dokodemo_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_dokodemo_config_proto_msgTypes[0]
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
	return file_proxy_dokodemo_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetAddress() *net.IPOrDomain {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *Config) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

// Deprecated: Marked as deprecated in proxy/dokodemo/config.proto.
func (x *Config) GetNetworkList() *net.NetworkList {
	if x != nil {
		return x.NetworkList
	}
	return nil
}

func (x *Config) GetNetworks() []net.Network {
	if x != nil {
		return x.Networks
	}
	return nil
}

// Deprecated: Marked as deprecated in proxy/dokodemo/config.proto.
func (x *Config) GetTimeout() uint32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *Config) GetFollowRedirect() bool {
	if x != nil {
		return x.FollowRedirect
	}
	return false
}

func (x *Config) GetUserLevel() uint32 {
	if x != nil {
		return x.UserLevel
	}
	return 0
}

type SimplifiedConfig struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Address        *net.IPOrDomain        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port           uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Networks       *net.NetworkList       `protobuf:"bytes,3,opt,name=networks,proto3" json:"networks,omitempty"`
	FollowRedirect bool                   `protobuf:"varint,4,opt,name=follow_redirect,json=followRedirect,proto3" json:"follow_redirect,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SimplifiedConfig) Reset() {
	*x = SimplifiedConfig{}
	mi := &file_proxy_dokodemo_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedConfig) ProtoMessage() {}

func (x *SimplifiedConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_dokodemo_config_proto_msgTypes[1]
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
	return file_proxy_dokodemo_config_proto_rawDescGZIP(), []int{1}
}

func (x *SimplifiedConfig) GetAddress() *net.IPOrDomain {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *SimplifiedConfig) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *SimplifiedConfig) GetNetworks() *net.NetworkList {
	if x != nil {
		return x.Networks
	}
	return nil
}

func (x *SimplifiedConfig) GetFollowRedirect() bool {
	if x != nil {
		return x.FollowRedirect
	}
	return false
}

var File_proxy_dokodemo_config_proto protoreflect.FileDescriptor

const file_proxy_dokodemo_config_proto_rawDesc = "" +
	"\n" +
	"\x1bproxy/dokodemo/config.proto\x12\x19v2ray.core.proxy.dokodemo\x1a\x18common/net/address.proto\x1a\x18common/net/network.proto\x1a common/protoext/extensions.proto\"\xc6\x02\n" +
	"\x06Config\x12;\n" +
	"\aaddress\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12I\n" +
	"\fnetwork_list\x18\x03 \x01(\v2\".v2ray.core.common.net.NetworkListB\x02\x18\x01R\vnetworkList\x12:\n" +
	"\bnetworks\x18\a \x03(\x0e2\x1e.v2ray.core.common.net.NetworkR\bnetworks\x12\x1c\n" +
	"\atimeout\x18\x04 \x01(\rB\x02\x18\x01R\atimeout\x12'\n" +
	"\x0ffollow_redirect\x18\x05 \x01(\bR\x0efollowRedirect\x12\x1d\n" +
	"\n" +
	"user_level\x18\x06 \x01(\rR\tuserLevel\"\xea\x01\n" +
	"\x10SimplifiedConfig\x12;\n" +
	"\aaddress\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12>\n" +
	"\bnetworks\x18\x03 \x01(\v2\".v2ray.core.common.net.NetworkListR\bnetworks\x12'\n" +
	"\x0ffollow_redirect\x18\x04 \x01(\bR\x0efollowRedirect:\x1c\x82\xb5\x18\x18\n" +
	"\ainbound\x12\rdokodemo-doorBl\n" +
	"\x1dcom.v2ray.core.proxy.dokodemoP\x01Z-github.com/v2fly/v2ray-core/v5/proxy/dokodemo\xaa\x02\x19V2Ray.Core.Proxy.Dokodemob\x06proto3"

var (
	file_proxy_dokodemo_config_proto_rawDescOnce sync.Once
	file_proxy_dokodemo_config_proto_rawDescData []byte
)

func file_proxy_dokodemo_config_proto_rawDescGZIP() []byte {
	file_proxy_dokodemo_config_proto_rawDescOnce.Do(func() {
		file_proxy_dokodemo_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_dokodemo_config_proto_rawDesc), len(file_proxy_dokodemo_config_proto_rawDesc)))
	})
	return file_proxy_dokodemo_config_proto_rawDescData
}

var file_proxy_dokodemo_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proxy_dokodemo_config_proto_goTypes = []any{
	(*Config)(nil),           // 0: v2ray.core.proxy.dokodemo.Config
	(*SimplifiedConfig)(nil), // 1: v2ray.core.proxy.dokodemo.SimplifiedConfig
	(*net.IPOrDomain)(nil),   // 2: v2ray.core.common.net.IPOrDomain
	(*net.NetworkList)(nil),  // 3: v2ray.core.common.net.NetworkList
	(net.Network)(0),         // 4: v2ray.core.common.net.Network
}
var file_proxy_dokodemo_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.proxy.dokodemo.Config.address:type_name -> v2ray.core.common.net.IPOrDomain
	3, // 1: v2ray.core.proxy.dokodemo.Config.network_list:type_name -> v2ray.core.common.net.NetworkList
	4, // 2: v2ray.core.proxy.dokodemo.Config.networks:type_name -> v2ray.core.common.net.Network
	2, // 3: v2ray.core.proxy.dokodemo.SimplifiedConfig.address:type_name -> v2ray.core.common.net.IPOrDomain
	3, // 4: v2ray.core.proxy.dokodemo.SimplifiedConfig.networks:type_name -> v2ray.core.common.net.NetworkList
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proxy_dokodemo_config_proto_init() }
func file_proxy_dokodemo_config_proto_init() {
	if File_proxy_dokodemo_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_dokodemo_config_proto_rawDesc), len(file_proxy_dokodemo_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_dokodemo_config_proto_goTypes,
		DependencyIndexes: file_proxy_dokodemo_config_proto_depIdxs,
		MessageInfos:      file_proxy_dokodemo_config_proto_msgTypes,
	}.Build()
	File_proxy_dokodemo_config_proto = out.File
	file_proxy_dokodemo_config_proto_goTypes = nil
	file_proxy_dokodemo_config_proto_depIdxs = nil
}
