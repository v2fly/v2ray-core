package dtls

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

type DTLSMode int32

const (
	DTLSMode_INVALID DTLSMode = 0
	DTLSMode_PSK     DTLSMode = 1
)

// Enum value maps for DTLSMode.
var (
	DTLSMode_name = map[int32]string{
		0: "INVALID",
		1: "PSK",
	}
	DTLSMode_value = map[string]int32{
		"INVALID": 0,
		"PSK":     1,
	}
)

func (x DTLSMode) Enum() *DTLSMode {
	p := new(DTLSMode)
	*p = x
	return p
}

func (x DTLSMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DTLSMode) Descriptor() protoreflect.EnumDescriptor {
	return file_transport_internet_dtls_config_proto_enumTypes[0].Descriptor()
}

func (DTLSMode) Type() protoreflect.EnumType {
	return &file_transport_internet_dtls_config_proto_enumTypes[0]
}

func (x DTLSMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DTLSMode.Descriptor instead.
func (DTLSMode) EnumDescriptor() ([]byte, []int) {
	return file_transport_internet_dtls_config_proto_rawDescGZIP(), []int{0}
}

type Config struct {
	state                  protoimpl.MessageState `protogen:"open.v1"`
	Mode                   DTLSMode               `protobuf:"varint,1,opt,name=mode,proto3,enum=v2ray.core.transport.internet.dtls.DTLSMode" json:"mode,omitempty"`
	Psk                    []byte                 `protobuf:"bytes,2,opt,name=psk,proto3" json:"psk,omitempty"`
	Mtu                    uint32                 `protobuf:"varint,3,opt,name=mtu,proto3" json:"mtu,omitempty"`
	ReplayProtectionWindow uint32                 `protobuf:"varint,4,opt,name=replay_protection_window,json=replayProtectionWindow,proto3" json:"replay_protection_window,omitempty"`
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_dtls_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_dtls_config_proto_msgTypes[0]
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
	return file_transport_internet_dtls_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetMode() DTLSMode {
	if x != nil {
		return x.Mode
	}
	return DTLSMode_INVALID
}

func (x *Config) GetPsk() []byte {
	if x != nil {
		return x.Psk
	}
	return nil
}

func (x *Config) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *Config) GetReplayProtectionWindow() uint32 {
	if x != nil {
		return x.ReplayProtectionWindow
	}
	return 0
}

var File_transport_internet_dtls_config_proto protoreflect.FileDescriptor

const file_transport_internet_dtls_config_proto_rawDesc = "" +
	"\n" +
	"$transport/internet/dtls/config.proto\x12\"v2ray.core.transport.internet.dtls\x1a common/protoext/extensions.proto\"\xbf\x01\n" +
	"\x06Config\x12@\n" +
	"\x04mode\x18\x01 \x01(\x0e2,.v2ray.core.transport.internet.dtls.DTLSModeR\x04mode\x12\x10\n" +
	"\x03psk\x18\x02 \x01(\fR\x03psk\x12\x10\n" +
	"\x03mtu\x18\x03 \x01(\rR\x03mtu\x128\n" +
	"\x18replay_protection_window\x18\x04 \x01(\rR\x16replayProtectionWindow:\x15\x82\xb5\x18\x11\n" +
	"\ttransport\x12\x04dtls* \n" +
	"\bDTLSMode\x12\v\n" +
	"\aINVALID\x10\x00\x12\a\n" +
	"\x03PSK\x10\x01B\x87\x01\n" +
	"&com.v2ray.core.transport.internet.dtlsP\x01Z6github.com/v2fly/v2ray-core/v5/transport/internet/dtls\xaa\x02\"V2Ray.Core.Transport.Internet.Dtlsb\x06proto3"

var (
	file_transport_internet_dtls_config_proto_rawDescOnce sync.Once
	file_transport_internet_dtls_config_proto_rawDescData []byte
)

func file_transport_internet_dtls_config_proto_rawDescGZIP() []byte {
	file_transport_internet_dtls_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_dtls_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_dtls_config_proto_rawDesc), len(file_transport_internet_dtls_config_proto_rawDesc)))
	})
	return file_transport_internet_dtls_config_proto_rawDescData
}

var file_transport_internet_dtls_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_transport_internet_dtls_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_dtls_config_proto_goTypes = []any{
	(DTLSMode)(0),  // 0: v2ray.core.transport.internet.dtls.DTLSMode
	(*Config)(nil), // 1: v2ray.core.transport.internet.dtls.Config
}
var file_transport_internet_dtls_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.transport.internet.dtls.Config.mode:type_name -> v2ray.core.transport.internet.dtls.DTLSMode
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_dtls_config_proto_init() }
func file_transport_internet_dtls_config_proto_init() {
	if File_transport_internet_dtls_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_dtls_config_proto_rawDesc), len(file_transport_internet_dtls_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_dtls_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_dtls_config_proto_depIdxs,
		EnumInfos:         file_transport_internet_dtls_config_proto_enumTypes,
		MessageInfos:      file_transport_internet_dtls_config_proto_msgTypes,
	}.Build()
	File_transport_internet_dtls_config_proto = out.File
	file_transport_internet_dtls_config_proto_goTypes = nil
	file_transport_internet_dtls_config_proto_depIdxs = nil
}
