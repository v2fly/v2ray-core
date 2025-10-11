package roundtripperenrollmentconfirmation

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type ClientConfig struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	RoundTripperClient *anypb.Any             `protobuf:"bytes,1,opt,name=round_tripper_client,json=roundTripperClient,proto3" json:"round_tripper_client,omitempty"`
	SecurityConfig     *anypb.Any             `protobuf:"bytes,2,opt,name=security_config,json=securityConfig,proto3" json:"security_config,omitempty"`
	Dest               string                 `protobuf:"bytes,3,opt,name=dest,proto3" json:"dest,omitempty"`
	OutboundTag        string                 `protobuf:"bytes,4,opt,name=outbound_tag,json=outboundTag,proto3" json:"outbound_tag,omitempty"`
	ServerIdentity     []byte                 `protobuf:"bytes,5,opt,name=server_identity,json=serverIdentity,proto3" json:"server_identity,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientConfig.ProtoReflect.Descriptor instead.
func (*ClientConfig) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescGZIP(), []int{0}
}

func (x *ClientConfig) GetRoundTripperClient() *anypb.Any {
	if x != nil {
		return x.RoundTripperClient
	}
	return nil
}

func (x *ClientConfig) GetSecurityConfig() *anypb.Any {
	if x != nil {
		return x.SecurityConfig
	}
	return nil
}

func (x *ClientConfig) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

func (x *ClientConfig) GetOutboundTag() string {
	if x != nil {
		return x.OutboundTag
	}
	return ""
}

func (x *ClientConfig) GetServerIdentity() []byte {
	if x != nil {
		return x.ServerIdentity
	}
	return nil
}

type ServerConfig struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	RoundTripperServer *anypb.Any             `protobuf:"bytes,2,opt,name=round_tripper_server,json=roundTripperServer,proto3" json:"round_tripper_server,omitempty"`
	Listen             string                 `protobuf:"bytes,3,opt,name=listen,proto3" json:"listen,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetRoundTripperServer() *anypb.Any {
	if x != nil {
		return x.RoundTripperServer
	}
	return nil
}

func (x *ServerConfig) GetListen() string {
	if x != nil {
		return x.Listen
	}
	return ""
}

var File_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto protoreflect.FileDescriptor

const file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDesc = "" +
	"\n" +
	"]transport/internet/tlsmirror/mirrorenrollment/roundtripperenrollmentconfirmation/config.proto\x12[v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation\x1a common/protoext/extensions.proto\x1a\x19google/protobuf/any.proto\"\xf5\x01\n" +
	"\fClientConfig\x12F\n" +
	"\x14round_tripper_client\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\x12roundTripperClient\x12=\n" +
	"\x0fsecurity_config\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x0esecurityConfig\x12\x12\n" +
	"\x04dest\x18\x03 \x01(\tR\x04dest\x12!\n" +
	"\foutbound_tag\x18\x04 \x01(\tR\voutboundTag\x12'\n" +
	"\x0fserver_identity\x18\x05 \x01(\fR\x0eserverIdentity\"n\n" +
	"\fServerConfig\x12F\n" +
	"\x14round_tripper_server\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x12roundTripperServer\x12\x16\n" +
	"\x06listen\x18\x03 \x01(\tR\x06listenB\xb2\x02\n" +
	"_com.v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmationP\x01Zogithub.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment/roundtripperenrollmentconfirmation\xaa\x02[V2Ray.Core.Transport.Internet.Tlsmirror.MirrorEnrollment.RoundTripperEnrollmentConfirmationb\x06proto3"

var (
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescOnce sync.Once
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescData []byte
)

func file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescGZIP() []byte {
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDesc), len(file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDesc)))
	})
	return file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDescData
}

var file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_goTypes = []any{
	(*ClientConfig)(nil), // 0: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation.ClientConfig
	(*ServerConfig)(nil), // 1: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation.ServerConfig
	(*anypb.Any)(nil),    // 2: google.protobuf.Any
}
var file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation.ClientConfig.round_tripper_client:type_name -> google.protobuf.Any
	2, // 1: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation.ClientConfig.security_config:type_name -> google.protobuf.Any
	2, // 2: v2ray.core.transport.internet.tlsmirror.mirrorenrollment.roundtripperenrollmentconfirmation.ServerConfig.round_tripper_server:type_name -> google.protobuf.Any
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() {
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_init()
}
func file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_init() {
	if File_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDesc), len(file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_msgTypes,
	}.Build()
	File_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto = out.File
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_goTypes = nil
	file_transport_internet_tlsmirror_mirrorenrollment_roundtripperenrollmentconfirmation_config_proto_depIdxs = nil
}
