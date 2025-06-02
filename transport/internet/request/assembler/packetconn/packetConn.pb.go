package packetconn

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
	state                      protoimpl.MessageState `protogen:"open.v1"`
	UnderlyingTransportSetting *anypb.Any             `protobuf:"bytes,1,opt,name=underlying_transport_setting,json=underlyingTransportSetting,proto3" json:"underlying_transport_setting,omitempty"`
	UnderlyingTransportName    string                 `protobuf:"bytes,2,opt,name=underlying_transport_name,json=underlyingTransportName,proto3" json:"underlying_transport_name,omitempty"`
	MaxWriteDelay              int32                  `protobuf:"varint,3,opt,name=max_write_delay,json=maxWriteDelay,proto3" json:"max_write_delay,omitempty"`
	MaxRequestSize             int32                  `protobuf:"varint,4,opt,name=max_request_size,json=maxRequestSize,proto3" json:"max_request_size,omitempty"`
	PollingIntervalInitial     int32                  `protobuf:"varint,5,opt,name=polling_interval_initial,json=pollingIntervalInitial,proto3" json:"polling_interval_initial,omitempty"`
	unknownFields              protoimpl.UnknownFields
	sizeCache                  protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[0]
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
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP(), []int{0}
}

func (x *ClientConfig) GetUnderlyingTransportSetting() *anypb.Any {
	if x != nil {
		return x.UnderlyingTransportSetting
	}
	return nil
}

func (x *ClientConfig) GetUnderlyingTransportName() string {
	if x != nil {
		return x.UnderlyingTransportName
	}
	return ""
}

func (x *ClientConfig) GetMaxWriteDelay() int32 {
	if x != nil {
		return x.MaxWriteDelay
	}
	return 0
}

func (x *ClientConfig) GetMaxRequestSize() int32 {
	if x != nil {
		return x.MaxRequestSize
	}
	return 0
}

func (x *ClientConfig) GetPollingIntervalInitial() int32 {
	if x != nil {
		return x.PollingIntervalInitial
	}
	return 0
}

type ServerConfig struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	UnderlyingTransportSetting     *anypb.Any             `protobuf:"bytes,1,opt,name=underlying_transport_setting,json=underlyingTransportSetting,proto3" json:"underlying_transport_setting,omitempty"`
	UnderlyingTransportName        string                 `protobuf:"bytes,2,opt,name=underlying_transport_name,json=underlyingTransportName,proto3" json:"underlying_transport_name,omitempty"`
	MaxWriteSize                   int32                  `protobuf:"varint,3,opt,name=max_write_size,json=maxWriteSize,proto3" json:"max_write_size,omitempty"`
	MaxWriteDurationMs             int32                  `protobuf:"varint,4,opt,name=max_write_duration_ms,json=maxWriteDurationMs,proto3" json:"max_write_duration_ms,omitempty"`
	MaxSimultaneousWriteConnection int32                  `protobuf:"varint,5,opt,name=max_simultaneous_write_connection,json=maxSimultaneousWriteConnection,proto3" json:"max_simultaneous_write_connection,omitempty"`
	PacketWritingBuffer            int32                  `protobuf:"varint,6,opt,name=packet_writing_buffer,json=packetWritingBuffer,proto3" json:"packet_writing_buffer,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[1]
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
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetUnderlyingTransportSetting() *anypb.Any {
	if x != nil {
		return x.UnderlyingTransportSetting
	}
	return nil
}

func (x *ServerConfig) GetUnderlyingTransportName() string {
	if x != nil {
		return x.UnderlyingTransportName
	}
	return ""
}

func (x *ServerConfig) GetMaxWriteSize() int32 {
	if x != nil {
		return x.MaxWriteSize
	}
	return 0
}

func (x *ServerConfig) GetMaxWriteDurationMs() int32 {
	if x != nil {
		return x.MaxWriteDurationMs
	}
	return 0
}

func (x *ServerConfig) GetMaxSimultaneousWriteConnection() int32 {
	if x != nil {
		return x.MaxSimultaneousWriteConnection
	}
	return 0
}

func (x *ServerConfig) GetPacketWritingBuffer() int32 {
	if x != nil {
		return x.PacketWritingBuffer
	}
	return 0
}

var File_transport_internet_request_assembler_packetconn_packetConn_proto protoreflect.FileDescriptor

const file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc = "" +
	"\n" +
	"@transport/internet/request/assembler/packetconn/packetConn.proto\x12:v2ray.core.transport.internet.request.assembler.packetconn\x1a common/protoext/extensions.proto\x1a\x19google/protobuf/any.proto\"\xe4\x02\n" +
	"\fClientConfig\x12V\n" +
	"\x1cunderlying_transport_setting\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\x1aunderlyingTransportSetting\x12:\n" +
	"\x19underlying_transport_name\x18\x02 \x01(\tR\x17underlyingTransportName\x12&\n" +
	"\x0fmax_write_delay\x18\x03 \x01(\x05R\rmaxWriteDelay\x12(\n" +
	"\x10max_request_size\x18\x04 \x01(\x05R\x0emaxRequestSize\x128\n" +
	"\x18polling_interval_initial\x18\x05 \x01(\x05R\x16pollingIntervalInitial:4\x82\xb5\x180\n" +
	"\"transport.request.assembler.client\x12\n" +
	"packetconn\"\xb0\x03\n" +
	"\fServerConfig\x12V\n" +
	"\x1cunderlying_transport_setting\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\x1aunderlyingTransportSetting\x12:\n" +
	"\x19underlying_transport_name\x18\x02 \x01(\tR\x17underlyingTransportName\x12$\n" +
	"\x0emax_write_size\x18\x03 \x01(\x05R\fmaxWriteSize\x121\n" +
	"\x15max_write_duration_ms\x18\x04 \x01(\x05R\x12maxWriteDurationMs\x12I\n" +
	"!max_simultaneous_write_connection\x18\x05 \x01(\x05R\x1emaxSimultaneousWriteConnection\x122\n" +
	"\x15packet_writing_buffer\x18\x06 \x01(\x05R\x13packetWritingBuffer:4\x82\xb5\x180\n" +
	"\"transport.request.assembler.server\x12\n" +
	"packetconnB\xcf\x01\n" +
	">com.v2ray.core.transport.internet.request.assembler.packetconnP\x01ZNgithub.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/packetconn\xaa\x02:V2Ray.Core.Transport.Internet.Request.Assembler.Packetconnb\x06proto3"

var (
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescOnce sync.Once
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData []byte
)

func file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP() []byte {
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc), len(file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc)))
	})
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData
}

var file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes = []any{
	(*ClientConfig)(nil), // 0: v2ray.core.transport.internet.request.assembler.packetconn.ClientConfig
	(*ServerConfig)(nil), // 1: v2ray.core.transport.internet.request.assembler.packetconn.ServerConfig
	(*anypb.Any)(nil),    // 2: google.protobuf.Any
}
var file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.request.assembler.packetconn.ClientConfig.underlying_transport_setting:type_name -> google.protobuf.Any
	2, // 1: v2ray.core.transport.internet.request.assembler.packetconn.ServerConfig.underlying_transport_setting:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_internet_request_assembler_packetconn_packetConn_proto_init() }
func file_transport_internet_request_assembler_packetconn_packetConn_proto_init() {
	if File_transport_internet_request_assembler_packetconn_packetConn_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc), len(file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes,
	}.Build()
	File_transport_internet_request_assembler_packetconn_packetConn_proto = out.File
	file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes = nil
	file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs = nil
}
