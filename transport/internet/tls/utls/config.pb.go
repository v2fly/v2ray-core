package utls

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	tls "github.com/v2fly/v2ray-core/v5/transport/internet/tls"
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

type ForcedALPN int32

const (
	ForcedALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY ForcedALPN = 0
	ForcedALPN_NO_ALPN                            ForcedALPN = 1
	ForcedALPN_UTLS_PRESET                        ForcedALPN = 2
)

// Enum value maps for ForcedALPN.
var (
	ForcedALPN_name = map[int32]string{
		0: "TRANSPORT_PREFERENCE_TAKE_PRIORITY",
		1: "NO_ALPN",
		2: "UTLS_PRESET",
	}
	ForcedALPN_value = map[string]int32{
		"TRANSPORT_PREFERENCE_TAKE_PRIORITY": 0,
		"NO_ALPN":                            1,
		"UTLS_PRESET":                        2,
	}
)

func (x ForcedALPN) Enum() *ForcedALPN {
	p := new(ForcedALPN)
	*p = x
	return p
}

func (x ForcedALPN) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ForcedALPN) Descriptor() protoreflect.EnumDescriptor {
	return file_transport_internet_tls_utls_config_proto_enumTypes[0].Descriptor()
}

func (ForcedALPN) Type() protoreflect.EnumType {
	return &file_transport_internet_tls_utls_config_proto_enumTypes[0]
}

func (x ForcedALPN) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ForcedALPN.Descriptor instead.
func (ForcedALPN) EnumDescriptor() ([]byte, []int) {
	return file_transport_internet_tls_utls_config_proto_rawDescGZIP(), []int{0}
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TlsConfig     *tls.Config            `protobuf:"bytes,1,opt,name=tls_config,json=tlsConfig,proto3" json:"tls_config,omitempty"`
	Imitate       string                 `protobuf:"bytes,2,opt,name=imitate,proto3" json:"imitate,omitempty"`
	NoSNI         bool                   `protobuf:"varint,3,opt,name=noSNI,proto3" json:"noSNI,omitempty"`
	ForceAlpn     ForcedALPN             `protobuf:"varint,4,opt,name=force_alpn,json=forceAlpn,proto3,enum=v2ray.core.transport.internet.tls.utls.ForcedALPN" json:"force_alpn,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_tls_utls_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tls_utls_config_proto_msgTypes[0]
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
	return file_transport_internet_tls_utls_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetTlsConfig() *tls.Config {
	if x != nil {
		return x.TlsConfig
	}
	return nil
}

func (x *Config) GetImitate() string {
	if x != nil {
		return x.Imitate
	}
	return ""
}

func (x *Config) GetNoSNI() bool {
	if x != nil {
		return x.NoSNI
	}
	return false
}

func (x *Config) GetForceAlpn() ForcedALPN {
	if x != nil {
		return x.ForceAlpn
	}
	return ForcedALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY
}

var File_transport_internet_tls_utls_config_proto protoreflect.FileDescriptor

const file_transport_internet_tls_utls_config_proto_rawDesc = "" +
	"\n" +
	"(transport/internet/tls/utls/config.proto\x12&v2ray.core.transport.internet.tls.utls\x1a common/protoext/extensions.proto\x1a#transport/internet/tls/config.proto\"\xef\x01\n" +
	"\x06Config\x12H\n" +
	"\n" +
	"tls_config\x18\x01 \x01(\v2).v2ray.core.transport.internet.tls.ConfigR\ttlsConfig\x12\x18\n" +
	"\aimitate\x18\x02 \x01(\tR\aimitate\x12\x14\n" +
	"\x05noSNI\x18\x03 \x01(\bR\x05noSNI\x12Q\n" +
	"\n" +
	"force_alpn\x18\x04 \x01(\x0e22.v2ray.core.transport.internet.tls.utls.ForcedALPNR\tforceAlpn:\x18\x82\xb5\x18\x14\n" +
	"\bsecurity\x12\x04utls\x90\xff)\x01*R\n" +
	"\n" +
	"ForcedALPN\x12&\n" +
	"\"TRANSPORT_PREFERENCE_TAKE_PRIORITY\x10\x00\x12\v\n" +
	"\aNO_ALPN\x10\x01\x12\x0f\n" +
	"\vUTLS_PRESET\x10\x02B\x93\x01\n" +
	"*com.v2ray.core.transport.internet.tls.utlsP\x01Z:github.com/v2fly/v2ray-core/v5/transport/internet/tls/utls\xaa\x02&V2Ray.Core.Transport.Internet.Tls.UTlsb\x06proto3"

var (
	file_transport_internet_tls_utls_config_proto_rawDescOnce sync.Once
	file_transport_internet_tls_utls_config_proto_rawDescData []byte
)

func file_transport_internet_tls_utls_config_proto_rawDescGZIP() []byte {
	file_transport_internet_tls_utls_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_tls_utls_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tls_utls_config_proto_rawDesc), len(file_transport_internet_tls_utls_config_proto_rawDesc)))
	})
	return file_transport_internet_tls_utls_config_proto_rawDescData
}

var file_transport_internet_tls_utls_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_transport_internet_tls_utls_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_tls_utls_config_proto_goTypes = []any{
	(ForcedALPN)(0),    // 0: v2ray.core.transport.internet.tls.utls.ForcedALPN
	(*Config)(nil),     // 1: v2ray.core.transport.internet.tls.utls.Config
	(*tls.Config)(nil), // 2: v2ray.core.transport.internet.tls.Config
}
var file_transport_internet_tls_utls_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.tls.utls.Config.tls_config:type_name -> v2ray.core.transport.internet.tls.Config
	0, // 1: v2ray.core.transport.internet.tls.utls.Config.force_alpn:type_name -> v2ray.core.transport.internet.tls.utls.ForcedALPN
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_internet_tls_utls_config_proto_init() }
func file_transport_internet_tls_utls_config_proto_init() {
	if File_transport_internet_tls_utls_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tls_utls_config_proto_rawDesc), len(file_transport_internet_tls_utls_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tls_utls_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_tls_utls_config_proto_depIdxs,
		EnumInfos:         file_transport_internet_tls_utls_config_proto_enumTypes,
		MessageInfos:      file_transport_internet_tls_utls_config_proto_msgTypes,
	}.Build()
	File_transport_internet_tls_utls_config_proto = out.File
	file_transport_internet_tls_utls_config_proto_goTypes = nil
	file_transport_internet_tls_utls_config_proto_depIdxs = nil
}
