package roundtripperreverserserver

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

type Config struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	RoundTripperServer *anypb.Any             `protobuf:"bytes,2,opt,name=round_tripper_server,json=roundTripperServer,proto3" json:"round_tripper_server,omitempty"`
	Listen             string                 `protobuf:"bytes,3,opt,name=listen,proto3" json:"listen,omitempty"`
	AccessPassphrase   string                 `protobuf:"bytes,4,opt,name=access_passphrase,json=accessPassphrase,proto3" json:"access_passphrase,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_request_roundtripperreverserserver_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_roundtripperreverserserver_config_proto_msgTypes[0]
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
	return file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetRoundTripperServer() *anypb.Any {
	if x != nil {
		return x.RoundTripperServer
	}
	return nil
}

func (x *Config) GetListen() string {
	if x != nil {
		return x.Listen
	}
	return ""
}

func (x *Config) GetAccessPassphrase() string {
	if x != nil {
		return x.AccessPassphrase
	}
	return ""
}

var File_transport_internet_request_roundtripperreverserserver_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_roundtripperreverserserver_config_proto_rawDesc = "" +
	"\n" +
	"Btransport/internet/request/roundtripperreverserserver/config.proto\x12@v2ray.core.transport.internet.request.roundtripperreverserserver\x1a common/protoext/extensions.proto\x1a\x19google/protobuf/any.proto\"\xc9\x01\n" +
	"\x06Config\x12F\n" +
	"\x14round_tripper_server\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x12roundTripperServer\x12\x16\n" +
	"\x06listen\x18\x03 \x01(\tR\x06listen\x12+\n" +
	"\x11access_passphrase\x18\x04 \x01(\tR\x10accessPassphrase:2\x82\xb5\x18.\n" +
	",transport.request.roundtripperreverserserverB\xe0\x01\n" +
	"Dcom.v2ray.core.transport.internet.request.roundtripperreverserserverP\x01ZTgithub.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripperreverserserver\xaa\x02?V2Ray.Core.Transport.Internet.Request.RoundtripperReverseServerb\x06proto3"

var (
	file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescData []byte
)

func file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_roundtripperreverserserver_config_proto_rawDesc), len(file_transport_internet_request_roundtripperreverserserver_config_proto_rawDesc)))
	})
	return file_transport_internet_request_roundtripperreverserserver_config_proto_rawDescData
}

var file_transport_internet_request_roundtripperreverserserver_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_request_roundtripperreverserserver_config_proto_goTypes = []any{
	(*Config)(nil),    // 0: v2ray.core.transport.internet.request.roundtripperreverserserver.Config
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_transport_internet_request_roundtripperreverserserver_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.transport.internet.request.roundtripperreverserserver.Config.round_tripper_server:type_name -> google.protobuf.Any
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_request_roundtripperreverserserver_config_proto_init() }
func file_transport_internet_request_roundtripperreverserserver_config_proto_init() {
	if File_transport_internet_request_roundtripperreverserserver_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_roundtripperreverserserver_config_proto_rawDesc), len(file_transport_internet_request_roundtripperreverserserver_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_roundtripperreverserserver_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_roundtripperreverserserver_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_roundtripperreverserserver_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_roundtripperreverserserver_config_proto = out.File
	file_transport_internet_request_roundtripperreverserserver_config_proto_goTypes = nil
	file_transport_internet_request_roundtripperreverserserver_config_proto_depIdxs = nil
}
