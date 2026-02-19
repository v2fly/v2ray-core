package wgcommon

import (
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

type PeerConfig struct {
	state                       protoimpl.MessageState `protogen:"open.v1"`
	PublicKey                   []byte                 `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	PresharedKey                []byte                 `protobuf:"bytes,2,opt,name=preshared_key,json=presharedKey,proto3" json:"preshared_key,omitempty"`
	AllowedIps                  []string               `protobuf:"bytes,3,rep,name=allowed_ips,json=allowedIps,proto3" json:"allowed_ips,omitempty"`
	Endpoint                    string                 `protobuf:"bytes,4,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	PersistentKeepaliveInterval int64                  `protobuf:"varint,5,opt,name=persistent_keepalive_interval,json=persistentKeepaliveInterval,proto3" json:"persistent_keepalive_interval,omitempty"`
	unknownFields               protoimpl.UnknownFields
	sizeCache                   protoimpl.SizeCache
}

func (x *PeerConfig) Reset() {
	*x = PeerConfig{}
	mi := &file_proxy_wireguard_wgcommon_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PeerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerConfig) ProtoMessage() {}

func (x *PeerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_wireguard_wgcommon_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerConfig.ProtoReflect.Descriptor instead.
func (*PeerConfig) Descriptor() ([]byte, []int) {
	return file_proxy_wireguard_wgcommon_config_proto_rawDescGZIP(), []int{0}
}

func (x *PeerConfig) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *PeerConfig) GetPresharedKey() []byte {
	if x != nil {
		return x.PresharedKey
	}
	return nil
}

func (x *PeerConfig) GetAllowedIps() []string {
	if x != nil {
		return x.AllowedIps
	}
	return nil
}

func (x *PeerConfig) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *PeerConfig) GetPersistentKeepaliveInterval() int64 {
	if x != nil {
		return x.PersistentKeepaliveInterval
	}
	return 0
}

type DeviceConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PrivateKey    []byte                 `protobuf:"bytes,1,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	ListenPort    uint32                 `protobuf:"varint,3,opt,name=listen_port,json=listenPort,proto3" json:"listen_port,omitempty"`
	Peers         []*PeerConfig          `protobuf:"bytes,4,rep,name=peers,proto3" json:"peers,omitempty"`
	Mtu           uint32                 `protobuf:"varint,5,opt,name=mtu,proto3" json:"mtu,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeviceConfig) Reset() {
	*x = DeviceConfig{}
	mi := &file_proxy_wireguard_wgcommon_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeviceConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceConfig) ProtoMessage() {}

func (x *DeviceConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_wireguard_wgcommon_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceConfig.ProtoReflect.Descriptor instead.
func (*DeviceConfig) Descriptor() ([]byte, []int) {
	return file_proxy_wireguard_wgcommon_config_proto_rawDescGZIP(), []int{1}
}

func (x *DeviceConfig) GetPrivateKey() []byte {
	if x != nil {
		return x.PrivateKey
	}
	return nil
}

func (x *DeviceConfig) GetListenPort() uint32 {
	if x != nil {
		return x.ListenPort
	}
	return 0
}

func (x *DeviceConfig) GetPeers() []*PeerConfig {
	if x != nil {
		return x.Peers
	}
	return nil
}

func (x *DeviceConfig) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

var File_proxy_wireguard_wgcommon_config_proto protoreflect.FileDescriptor

const file_proxy_wireguard_wgcommon_config_proto_rawDesc = "" +
	"\n" +
	"%proxy/wireguard/wgcommon/config.proto\x12#v2ray.core.proxy.wireguard.wgcommon\"\xd1\x01\n" +
	"\n" +
	"PeerConfig\x12\x1d\n" +
	"\n" +
	"public_key\x18\x01 \x01(\fR\tpublicKey\x12#\n" +
	"\rpreshared_key\x18\x02 \x01(\fR\fpresharedKey\x12\x1f\n" +
	"\vallowed_ips\x18\x03 \x03(\tR\n" +
	"allowedIps\x12\x1a\n" +
	"\bendpoint\x18\x04 \x01(\tR\bendpoint\x12B\n" +
	"\x1dpersistent_keepalive_interval\x18\x05 \x01(\x03R\x1bpersistentKeepaliveInterval\"\xa9\x01\n" +
	"\fDeviceConfig\x12\x1f\n" +
	"\vprivate_key\x18\x01 \x01(\fR\n" +
	"privateKey\x12\x1f\n" +
	"\vlisten_port\x18\x03 \x01(\rR\n" +
	"listenPort\x12E\n" +
	"\x05peers\x18\x04 \x03(\v2/.v2ray.core.proxy.wireguard.wgcommon.PeerConfigR\x05peers\x12\x10\n" +
	"\x03mtu\x18\x05 \x01(\rR\x03mtuB\x8a\x01\n" +
	"'com.v2ray.core.proxy.wireguard.wgcommonP\x01Z7github.com/v2fly/v2ray-core/v5/proxy/wireguard/wgcommon\xaa\x02#V2Ray.Core.Proxy.Wireguard.Wgcommonb\x06proto3"

var (
	file_proxy_wireguard_wgcommon_config_proto_rawDescOnce sync.Once
	file_proxy_wireguard_wgcommon_config_proto_rawDescData []byte
)

func file_proxy_wireguard_wgcommon_config_proto_rawDescGZIP() []byte {
	file_proxy_wireguard_wgcommon_config_proto_rawDescOnce.Do(func() {
		file_proxy_wireguard_wgcommon_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_wireguard_wgcommon_config_proto_rawDesc), len(file_proxy_wireguard_wgcommon_config_proto_rawDesc)))
	})
	return file_proxy_wireguard_wgcommon_config_proto_rawDescData
}

var file_proxy_wireguard_wgcommon_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proxy_wireguard_wgcommon_config_proto_goTypes = []any{
	(*PeerConfig)(nil),   // 0: v2ray.core.proxy.wireguard.wgcommon.PeerConfig
	(*DeviceConfig)(nil), // 1: v2ray.core.proxy.wireguard.wgcommon.DeviceConfig
}
var file_proxy_wireguard_wgcommon_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.proxy.wireguard.wgcommon.DeviceConfig.peers:type_name -> v2ray.core.proxy.wireguard.wgcommon.PeerConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proxy_wireguard_wgcommon_config_proto_init() }
func file_proxy_wireguard_wgcommon_config_proto_init() {
	if File_proxy_wireguard_wgcommon_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_wireguard_wgcommon_config_proto_rawDesc), len(file_proxy_wireguard_wgcommon_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_wireguard_wgcommon_config_proto_goTypes,
		DependencyIndexes: file_proxy_wireguard_wgcommon_config_proto_depIdxs,
		MessageInfos:      file_proxy_wireguard_wgcommon_config_proto_msgTypes,
	}.Build()
	File_proxy_wireguard_wgcommon_config_proto = out.File
	file_proxy_wireguard_wgcommon_config_proto_goTypes = nil
	file_proxy_wireguard_wgcommon_config_proto_depIdxs = nil
}
