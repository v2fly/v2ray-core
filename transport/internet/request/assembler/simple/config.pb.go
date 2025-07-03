package simple

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

type ClientConfig struct {
	state                    protoimpl.MessageState `protogen:"open.v1"`
	MaxWriteSize             int32                  `protobuf:"varint,1,opt,name=max_write_size,json=maxWriteSize,proto3" json:"max_write_size,omitempty"`
	WaitSubsequentWriteMs    int32                  `protobuf:"varint,2,opt,name=wait_subsequent_write_ms,json=waitSubsequentWriteMs,proto3" json:"wait_subsequent_write_ms,omitempty"`
	InitialPollingIntervalMs int32                  `protobuf:"varint,3,opt,name=initial_polling_interval_ms,json=initialPollingIntervalMs,proto3" json:"initial_polling_interval_ms,omitempty"`
	MaxPollingIntervalMs     int32                  `protobuf:"varint,4,opt,name=max_polling_interval_ms,json=maxPollingIntervalMs,proto3" json:"max_polling_interval_ms,omitempty"`
	MinPollingIntervalMs     int32                  `protobuf:"varint,5,opt,name=min_polling_interval_ms,json=minPollingIntervalMs,proto3" json:"min_polling_interval_ms,omitempty"`
	BackoffFactor            float32                `protobuf:"fixed32,6,opt,name=backoff_factor,json=backoffFactor,proto3" json:"backoff_factor,omitempty"`
	FailedRetryIntervalMs    int32                  `protobuf:"varint,7,opt,name=failed_retry_interval_ms,json=failedRetryIntervalMs,proto3" json:"failed_retry_interval_ms,omitempty"`
	unknownFields            protoimpl.UnknownFields
	sizeCache                protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_transport_internet_request_assembler_simple_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_simple_config_proto_msgTypes[0]
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
	return file_transport_internet_request_assembler_simple_config_proto_rawDescGZIP(), []int{0}
}

func (x *ClientConfig) GetMaxWriteSize() int32 {
	if x != nil {
		return x.MaxWriteSize
	}
	return 0
}

func (x *ClientConfig) GetWaitSubsequentWriteMs() int32 {
	if x != nil {
		return x.WaitSubsequentWriteMs
	}
	return 0
}

func (x *ClientConfig) GetInitialPollingIntervalMs() int32 {
	if x != nil {
		return x.InitialPollingIntervalMs
	}
	return 0
}

func (x *ClientConfig) GetMaxPollingIntervalMs() int32 {
	if x != nil {
		return x.MaxPollingIntervalMs
	}
	return 0
}

func (x *ClientConfig) GetMinPollingIntervalMs() int32 {
	if x != nil {
		return x.MinPollingIntervalMs
	}
	return 0
}

func (x *ClientConfig) GetBackoffFactor() float32 {
	if x != nil {
		return x.BackoffFactor
	}
	return 0
}

func (x *ClientConfig) GetFailedRetryIntervalMs() int32 {
	if x != nil {
		return x.FailedRetryIntervalMs
	}
	return 0
}

type ServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	MaxWriteSize  int32                  `protobuf:"varint,1,opt,name=max_write_size,json=maxWriteSize,proto3" json:"max_write_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_transport_internet_request_assembler_simple_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_simple_config_proto_msgTypes[1]
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
	return file_transport_internet_request_assembler_simple_config_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetMaxWriteSize() int32 {
	if x != nil {
		return x.MaxWriteSize
	}
	return 0
}

var File_transport_internet_request_assembler_simple_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_assembler_simple_config_proto_rawDesc = "" +
	"\n" +
	"8transport/internet/request/assembler/simple/config.proto\x126v2ray.core.transport.internet.request.assembler.simple\x1a common/protoext/extensions.proto\"\xac\x03\n" +
	"\fClientConfig\x12$\n" +
	"\x0emax_write_size\x18\x01 \x01(\x05R\fmaxWriteSize\x127\n" +
	"\x18wait_subsequent_write_ms\x18\x02 \x01(\x05R\x15waitSubsequentWriteMs\x12=\n" +
	"\x1binitial_polling_interval_ms\x18\x03 \x01(\x05R\x18initialPollingIntervalMs\x125\n" +
	"\x17max_polling_interval_ms\x18\x04 \x01(\x05R\x14maxPollingIntervalMs\x125\n" +
	"\x17min_polling_interval_ms\x18\x05 \x01(\x05R\x14minPollingIntervalMs\x12%\n" +
	"\x0ebackoff_factor\x18\x06 \x01(\x02R\rbackoffFactor\x127\n" +
	"\x18failed_retry_interval_ms\x18\a \x01(\x05R\x15failedRetryIntervalMs:0\x82\xb5\x18,\n" +
	"\"transport.request.assembler.client\x12\x06simple\"f\n" +
	"\fServerConfig\x12$\n" +
	"\x0emax_write_size\x18\x01 \x01(\x05R\fmaxWriteSize:0\x82\xb5\x18,\n" +
	"\"transport.request.assembler.server\x12\x06simpleB\xc3\x01\n" +
	":com.v2ray.core.transport.internet.request.assembler.simpleP\x01ZJgithub.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/simple\xaa\x026V2Ray.Core.Transport.Internet.Request.Assembler.Simpleb\x06proto3"

var (
	file_transport_internet_request_assembler_simple_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_assembler_simple_config_proto_rawDescData []byte
)

func file_transport_internet_request_assembler_simple_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_assembler_simple_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_assembler_simple_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembler_simple_config_proto_rawDesc), len(file_transport_internet_request_assembler_simple_config_proto_rawDesc)))
	})
	return file_transport_internet_request_assembler_simple_config_proto_rawDescData
}

var file_transport_internet_request_assembler_simple_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_request_assembler_simple_config_proto_goTypes = []any{
	(*ClientConfig)(nil), // 0: v2ray.core.transport.internet.request.assembler.simple.ClientConfig
	(*ServerConfig)(nil), // 1: v2ray.core.transport.internet.request.assembler.simple.ServerConfig
}
var file_transport_internet_request_assembler_simple_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transport_internet_request_assembler_simple_config_proto_init() }
func file_transport_internet_request_assembler_simple_config_proto_init() {
	if File_transport_internet_request_assembler_simple_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_assembler_simple_config_proto_rawDesc), len(file_transport_internet_request_assembler_simple_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_assembler_simple_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_assembler_simple_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_assembler_simple_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_assembler_simple_config_proto = out.File
	file_transport_internet_request_assembler_simple_config_proto_goTypes = nil
	file_transport_internet_request_assembler_simple_config_proto_depIdxs = nil
}
