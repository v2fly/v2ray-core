package hysteria2

import (
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

type Congestion struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          string                 `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	UpMbps        uint64                 `protobuf:"varint,2,opt,name=up_mbps,json=upMbps,proto3" json:"up_mbps,omitempty"`
	DownMbps      uint64                 `protobuf:"varint,3,opt,name=down_mbps,json=downMbps,proto3" json:"down_mbps,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Congestion) Reset() {
	*x = Congestion{}
	mi := &file_transport_internet_hysteria2_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Congestion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Congestion) ProtoMessage() {}

func (x *Congestion) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_hysteria2_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Congestion.ProtoReflect.Descriptor instead.
func (*Congestion) Descriptor() ([]byte, []int) {
	return file_transport_internet_hysteria2_config_proto_rawDescGZIP(), []int{0}
}

func (x *Congestion) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Congestion) GetUpMbps() uint64 {
	if x != nil {
		return x.UpMbps
	}
	return 0
}

func (x *Congestion) GetDownMbps() uint64 {
	if x != nil {
		return x.DownMbps
	}
	return 0
}

type Config struct {
	state                 protoimpl.MessageState `protogen:"open.v1"`
	Password              string                 `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	Congestion            *Congestion            `protobuf:"bytes,4,opt,name=congestion,proto3" json:"congestion,omitempty"`
	IgnoreClientBandwidth bool                   `protobuf:"varint,5,opt,name=ignore_client_bandwidth,json=ignoreClientBandwidth,proto3" json:"ignore_client_bandwidth,omitempty"`
	UseUdpExtension       bool                   `protobuf:"varint,6,opt,name=use_udp_extension,json=useUdpExtension,proto3" json:"use_udp_extension,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_hysteria2_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_hysteria2_config_proto_msgTypes[1]
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
	return file_transport_internet_hysteria2_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *Config) GetCongestion() *Congestion {
	if x != nil {
		return x.Congestion
	}
	return nil
}

func (x *Config) GetIgnoreClientBandwidth() bool {
	if x != nil {
		return x.IgnoreClientBandwidth
	}
	return false
}

func (x *Config) GetUseUdpExtension() bool {
	if x != nil {
		return x.UseUdpExtension
	}
	return false
}

var File_transport_internet_hysteria2_config_proto protoreflect.FileDescriptor

const file_transport_internet_hysteria2_config_proto_rawDesc = "" +
	"\n" +
	")transport/internet/hysteria2/config.proto\x12'v2ray.core.transport.internet.hysteria2\x1a common/protoext/extensions.proto\"V\n" +
	"\n" +
	"Congestion\x12\x12\n" +
	"\x04type\x18\x01 \x01(\tR\x04type\x12\x17\n" +
	"\aup_mbps\x18\x02 \x01(\x04R\x06upMbps\x12\x1b\n" +
	"\tdown_mbps\x18\x03 \x01(\x04R\bdownMbps\"\xf9\x01\n" +
	"\x06Config\x12\x1a\n" +
	"\bpassword\x18\x03 \x01(\tR\bpassword\x12S\n" +
	"\n" +
	"congestion\x18\x04 \x01(\v23.v2ray.core.transport.internet.hysteria2.CongestionR\n" +
	"congestion\x126\n" +
	"\x17ignore_client_bandwidth\x18\x05 \x01(\bR\x15ignoreClientBandwidth\x12*\n" +
	"\x11use_udp_extension\x18\x06 \x01(\bR\x0fuseUdpExtension:\x1a\x82\xb5\x18\x16\n" +
	"\ttransport\x12\thysteria2B\x96\x01\n" +
	"+com.v2ray.core.transport.internet.hysteria2P\x01Z;github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2\xaa\x02'V2Ray.Core.Transport.Internet.Hysteria2b\x06proto3"

var (
	file_transport_internet_hysteria2_config_proto_rawDescOnce sync.Once
	file_transport_internet_hysteria2_config_proto_rawDescData []byte
)

func file_transport_internet_hysteria2_config_proto_rawDescGZIP() []byte {
	file_transport_internet_hysteria2_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_hysteria2_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_hysteria2_config_proto_rawDesc), len(file_transport_internet_hysteria2_config_proto_rawDesc)))
	})
	return file_transport_internet_hysteria2_config_proto_rawDescData
}

var file_transport_internet_hysteria2_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_hysteria2_config_proto_goTypes = []any{
	(*Congestion)(nil), // 0: v2ray.core.transport.internet.hysteria2.Congestion
	(*Config)(nil),     // 1: v2ray.core.transport.internet.hysteria2.Config
}
var file_transport_internet_hysteria2_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.transport.internet.hysteria2.Config.congestion:type_name -> v2ray.core.transport.internet.hysteria2.Congestion
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_hysteria2_config_proto_init() }
func file_transport_internet_hysteria2_config_proto_init() {
	if File_transport_internet_hysteria2_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_hysteria2_config_proto_rawDesc), len(file_transport_internet_hysteria2_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_hysteria2_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_hysteria2_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_hysteria2_config_proto_msgTypes,
	}.Build()
	File_transport_internet_hysteria2_config_proto = out.File
	file_transport_internet_hysteria2_config_proto_goTypes = nil
	file_transport_internet_hysteria2_config_proto_depIdxs = nil
}
