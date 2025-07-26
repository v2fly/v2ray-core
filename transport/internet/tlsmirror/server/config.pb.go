package server

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	mirrorenrollment "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment"
	tlstrafficgen "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/tlstrafficgen"
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

type TimeSpec struct {
	state                              protoimpl.MessageState `protogen:"open.v1"`
	BaseNanoseconds                    uint64                 `protobuf:"varint,1,opt,name=base_nanoseconds,json=baseNanoseconds,proto3" json:"base_nanoseconds,omitempty"`
	UniformRandomMultiplierNanoseconds uint64                 `protobuf:"varint,2,opt,name=uniform_random_multiplier_nanoseconds,json=uniformRandomMultiplierNanoseconds,proto3" json:"uniform_random_multiplier_nanoseconds,omitempty"`
	unknownFields                      protoimpl.UnknownFields
	sizeCache                          protoimpl.SizeCache
}

func (x *TimeSpec) Reset() {
	*x = TimeSpec{}
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TimeSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimeSpec) ProtoMessage() {}

func (x *TimeSpec) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimeSpec.ProtoReflect.Descriptor instead.
func (*TimeSpec) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP(), []int{0}
}

func (x *TimeSpec) GetBaseNanoseconds() uint64 {
	if x != nil {
		return x.BaseNanoseconds
	}
	return 0
}

func (x *TimeSpec) GetUniformRandomMultiplierNanoseconds() uint64 {
	if x != nil {
		return x.UniformRandomMultiplierNanoseconds
	}
	return 0
}

type TransportLayerPadding struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enabled       bool                   `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransportLayerPadding) Reset() {
	*x = TransportLayerPadding{}
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransportLayerPadding) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransportLayerPadding) ProtoMessage() {}

func (x *TransportLayerPadding) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransportLayerPadding.ProtoReflect.Descriptor instead.
func (*TransportLayerPadding) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP(), []int{1}
}

func (x *TransportLayerPadding) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

type Config struct {
	state                         protoimpl.MessageState   `protogen:"open.v1"`
	ForwardAddress                string                   `protobuf:"bytes,1,opt,name=forward_address,json=forwardAddress,proto3" json:"forward_address,omitempty"`
	ForwardPort                   uint32                   `protobuf:"varint,2,opt,name=forward_port,json=forwardPort,proto3" json:"forward_port,omitempty"`
	ForwardTag                    string                   `protobuf:"bytes,3,opt,name=forward_tag,json=forwardTag,proto3" json:"forward_tag,omitempty"`
	CarrierConnectionTag          string                   `protobuf:"bytes,4,opt,name=carrier_connection_tag,json=carrierConnectionTag,proto3" json:"carrier_connection_tag,omitempty"`
	EmbeddedTrafficGenerator      *tlstrafficgen.Config    `protobuf:"bytes,5,opt,name=embedded_traffic_generator,json=embeddedTrafficGenerator,proto3" json:"embedded_traffic_generator,omitempty"`
	PrimaryKey                    []byte                   `protobuf:"bytes,6,opt,name=primary_key,json=primaryKey,proto3" json:"primary_key,omitempty"`
	ExplicitNonceCiphersuites     []uint32                 `protobuf:"varint,7,rep,packed,name=explicit_nonce_ciphersuites,json=explicitNonceCiphersuites,proto3" json:"explicit_nonce_ciphersuites,omitempty"`
	DeferInstanceDerivedWriteTime *TimeSpec                `protobuf:"bytes,8,opt,name=defer_instance_derived_write_time,json=deferInstanceDerivedWriteTime,proto3" json:"defer_instance_derived_write_time,omitempty"`
	TransportLayerPadding         *TransportLayerPadding   `protobuf:"bytes,9,opt,name=transport_layer_padding,json=transportLayerPadding,proto3" json:"transport_layer_padding,omitempty"`
	ConnectionEnrolment           *mirrorenrollment.Config `protobuf:"bytes,10,opt,name=connection_enrolment,json=connectionEnrolment,proto3" json:"connection_enrolment,omitempty"`
	SequenceWatermarkingEnabled   bool                     `protobuf:"varint,11,opt,name=sequence_watermarking_enabled,json=sequenceWatermarkingEnabled,proto3" json:"sequence_watermarking_enabled,omitempty"`
	unknownFields                 protoimpl.UnknownFields
	sizeCache                     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_server_config_proto_msgTypes[2]
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
	return file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP(), []int{2}
}

func (x *Config) GetForwardAddress() string {
	if x != nil {
		return x.ForwardAddress
	}
	return ""
}

func (x *Config) GetForwardPort() uint32 {
	if x != nil {
		return x.ForwardPort
	}
	return 0
}

func (x *Config) GetForwardTag() string {
	if x != nil {
		return x.ForwardTag
	}
	return ""
}

func (x *Config) GetCarrierConnectionTag() string {
	if x != nil {
		return x.CarrierConnectionTag
	}
	return ""
}

func (x *Config) GetEmbeddedTrafficGenerator() *tlstrafficgen.Config {
	if x != nil {
		return x.EmbeddedTrafficGenerator
	}
	return nil
}

func (x *Config) GetPrimaryKey() []byte {
	if x != nil {
		return x.PrimaryKey
	}
	return nil
}

func (x *Config) GetExplicitNonceCiphersuites() []uint32 {
	if x != nil {
		return x.ExplicitNonceCiphersuites
	}
	return nil
}

func (x *Config) GetDeferInstanceDerivedWriteTime() *TimeSpec {
	if x != nil {
		return x.DeferInstanceDerivedWriteTime
	}
	return nil
}

func (x *Config) GetTransportLayerPadding() *TransportLayerPadding {
	if x != nil {
		return x.TransportLayerPadding
	}
	return nil
}

func (x *Config) GetConnectionEnrolment() *mirrorenrollment.Config {
	if x != nil {
		return x.ConnectionEnrolment
	}
	return nil
}

func (x *Config) GetSequenceWatermarkingEnabled() bool {
	if x != nil {
		return x.SequenceWatermarkingEnabled
	}
	return false
}

var File_transport_internet_tlsmirror_server_config_proto protoreflect.FileDescriptor

const file_transport_internet_tlsmirror_server_config_proto_rawDesc = "" +
	"\n" +
	"0transport/internet/tlsmirror/server/config.proto\x12.v2ray.core.transport.internet.tlsmirror.server\x1a common/protoext/extensions.proto\x1a7transport/internet/tlsmirror/tlstrafficgen/config.proto\x1a:transport/internet/tlsmirror/mirrorenrollment/config.proto\"\x88\x01\n" +
	"\bTimeSpec\x12)\n" +
	"\x10base_nanoseconds\x18\x01 \x01(\x04R\x0fbaseNanoseconds\x12Q\n" +
	"%uniform_random_multiplier_nanoseconds\x18\x02 \x01(\x04R\"uniformRandomMultiplierNanoseconds\"1\n" +
	"\x15TransportLayerPadding\x12\x18\n" +
	"\aenabled\x18\x01 \x01(\bR\aenabled\"\xef\x06\n" +
	"\x06Config\x12'\n" +
	"\x0fforward_address\x18\x01 \x01(\tR\x0eforwardAddress\x12!\n" +
	"\fforward_port\x18\x02 \x01(\rR\vforwardPort\x12\x1f\n" +
	"\vforward_tag\x18\x03 \x01(\tR\n" +
	"forwardTag\x124\n" +
	"\x16carrier_connection_tag\x18\x04 \x01(\tR\x14carrierConnectionTag\x12{\n" +
	"\x1aembedded_traffic_generator\x18\x05 \x01(\v2=.v2ray.core.transport.internet.tlsmirror.tlstrafficgen.ConfigR\x18embeddedTrafficGenerator\x12\x1f\n" +
	"\vprimary_key\x18\x06 \x01(\fR\n" +
	"primaryKey\x12>\n" +
	"\x1bexplicit_nonce_ciphersuites\x18\a \x03(\rR\x19explicitNonceCiphersuites\x12\x82\x01\n" +
	"!defer_instance_derived_write_time\x18\b \x01(\v28.v2ray.core.transport.internet.tlsmirror.server.TimeSpecR\x1ddeferInstanceDerivedWriteTime\x12}\n" +
	"\x17transport_layer_padding\x18\t \x01(\v2E.v2ray.core.transport.internet.tlsmirror.server.TransportLayerPaddingR\x15transportLayerPadding\x12s\n" +
	"\x14connection_enrolment\x18\n" +
	" \x01(\v2@.v2ray.core.transport.internet.tlsmirror.mirrorenrollment.ConfigR\x13connectionEnrolment\x12B\n" +
	"\x1dsequence_watermarking_enabled\x18\v \x01(\bR\x1bsequenceWatermarkingEnabled:'\x82\xb5\x18#\n" +
	"\ttransport\x12\ttlsmirror\x8a\xff)\ttlsmirrorB\xab\x01\n" +
	"2com.v2ray.core.transport.internet.tlsmirror.serverP\x01ZBgithub.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/server\xaa\x02.V2Ray.Core.Transport.Internet.Tlsmirror.Serverb\x06proto3"

var (
	file_transport_internet_tlsmirror_server_config_proto_rawDescOnce sync.Once
	file_transport_internet_tlsmirror_server_config_proto_rawDescData []byte
)

func file_transport_internet_tlsmirror_server_config_proto_rawDescGZIP() []byte {
	file_transport_internet_tlsmirror_server_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_tlsmirror_server_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_server_config_proto_rawDesc), len(file_transport_internet_tlsmirror_server_config_proto_rawDesc)))
	})
	return file_transport_internet_tlsmirror_server_config_proto_rawDescData
}

var file_transport_internet_tlsmirror_server_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_transport_internet_tlsmirror_server_config_proto_goTypes = []any{
	(*TimeSpec)(nil),                // 0: v2ray.core.transport.internet.tlsmirror.server.TimeSpec
	(*TransportLayerPadding)(nil),   // 1: v2ray.core.transport.internet.tlsmirror.server.TransportLayerPadding
	(*Config)(nil),                  // 2: v2ray.core.transport.internet.tlsmirror.server.Config
	(*tlstrafficgen.Config)(nil),    // 3: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Config
	(*mirrorenrollment.Config)(nil), // 4: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.Config
}
var file_transport_internet_tlsmirror_server_config_proto_depIdxs = []int32{
	3, // 0: v2ray.core.transport.internet.tlsmirror.server.Config.embedded_traffic_generator:type_name -> v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Config
	0, // 1: v2ray.core.transport.internet.tlsmirror.server.Config.defer_instance_derived_write_time:type_name -> v2ray.core.transport.internet.tlsmirror.server.TimeSpec
	1, // 2: v2ray.core.transport.internet.tlsmirror.server.Config.transport_layer_padding:type_name -> v2ray.core.transport.internet.tlsmirror.server.TransportLayerPadding
	4, // 3: v2ray.core.transport.internet.tlsmirror.server.Config.connection_enrolment:type_name -> v2ray.core.transport.internet.tlsmirror.mirrorenrollment.Config
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_transport_internet_tlsmirror_server_config_proto_init() }
func file_transport_internet_tlsmirror_server_config_proto_init() {
	if File_transport_internet_tlsmirror_server_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_server_config_proto_rawDesc), len(file_transport_internet_tlsmirror_server_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tlsmirror_server_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_tlsmirror_server_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_tlsmirror_server_config_proto_msgTypes,
	}.Build()
	File_transport_internet_tlsmirror_server_config_proto = out.File
	file_transport_internet_tlsmirror_server_config_proto_goTypes = nil
	file_transport_internet_tlsmirror_server_config_proto_depIdxs = nil
}
