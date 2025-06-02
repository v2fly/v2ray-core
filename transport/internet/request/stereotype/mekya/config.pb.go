package mekya

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	kcp "github.com/v2fly/v2ray-core/v5/transport/internet/kcp"
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
	state protoimpl.MessageState `protogen:"open.v1"`
	Kcp   *kcp.Config            `protobuf:"bytes,1,opt,name=kcp,proto3" json:"kcp,omitempty"`
	// Client
	MaxWriteDelay          int32 `protobuf:"varint,1003,opt,name=max_write_delay,json=maxWriteDelay,proto3" json:"max_write_delay,omitempty"`
	MaxRequestSize         int32 `protobuf:"varint,1004,opt,name=max_request_size,json=maxRequestSize,proto3" json:"max_request_size,omitempty"`
	PollingIntervalInitial int32 `protobuf:"varint,1005,opt,name=polling_interval_initial,json=pollingIntervalInitial,proto3" json:"polling_interval_initial,omitempty"`
	// Server
	MaxWriteSize                   int32 `protobuf:"varint,2003,opt,name=max_write_size,json=maxWriteSize,proto3" json:"max_write_size,omitempty"`
	MaxWriteDurationMs             int32 `protobuf:"varint,2004,opt,name=max_write_duration_ms,json=maxWriteDurationMs,proto3" json:"max_write_duration_ms,omitempty"`
	MaxSimultaneousWriteConnection int32 `protobuf:"varint,2005,opt,name=max_simultaneous_write_connection,json=maxSimultaneousWriteConnection,proto3" json:"max_simultaneous_write_connection,omitempty"`
	PacketWritingBuffer            int32 `protobuf:"varint,2006,opt,name=packet_writing_buffer,json=packetWritingBuffer,proto3" json:"packet_writing_buffer,omitempty"`
	// Roundtripper
	Url           string `protobuf:"bytes,3001,opt,name=url,proto3" json:"url,omitempty"`
	H2PoolSize    int32  `protobuf:"varint,3003,opt,name=h2_pool_size,json=h2PoolSize,proto3" json:"h2_pool_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_request_stereotype_mekya_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_stereotype_mekya_config_proto_msgTypes[0]
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
	return file_transport_internet_request_stereotype_mekya_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetKcp() *kcp.Config {
	if x != nil {
		return x.Kcp
	}
	return nil
}

func (x *Config) GetMaxWriteDelay() int32 {
	if x != nil {
		return x.MaxWriteDelay
	}
	return 0
}

func (x *Config) GetMaxRequestSize() int32 {
	if x != nil {
		return x.MaxRequestSize
	}
	return 0
}

func (x *Config) GetPollingIntervalInitial() int32 {
	if x != nil {
		return x.PollingIntervalInitial
	}
	return 0
}

func (x *Config) GetMaxWriteSize() int32 {
	if x != nil {
		return x.MaxWriteSize
	}
	return 0
}

func (x *Config) GetMaxWriteDurationMs() int32 {
	if x != nil {
		return x.MaxWriteDurationMs
	}
	return 0
}

func (x *Config) GetMaxSimultaneousWriteConnection() int32 {
	if x != nil {
		return x.MaxSimultaneousWriteConnection
	}
	return 0
}

func (x *Config) GetPacketWritingBuffer() int32 {
	if x != nil {
		return x.PacketWritingBuffer
	}
	return 0
}

func (x *Config) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Config) GetH2PoolSize() int32 {
	if x != nil {
		return x.H2PoolSize
	}
	return 0
}

var File_transport_internet_request_stereotype_mekya_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_stereotype_mekya_config_proto_rawDesc = "" +
	"\n" +
	"8transport/internet/request/stereotype/mekya/config.proto\x126v2ray.core.transport.internet.request.stereotype.mekya\x1a common/protoext/extensions.proto\x1a#transport/internet/kcp/config.proto\"\x82\x04\n" +
	"\x06Config\x12;\n" +
	"\x03kcp\x18\x01 \x01(\v2).v2ray.core.transport.internet.kcp.ConfigR\x03kcp\x12'\n" +
	"\x0fmax_write_delay\x18\xeb\a \x01(\x05R\rmaxWriteDelay\x12)\n" +
	"\x10max_request_size\x18\xec\a \x01(\x05R\x0emaxRequestSize\x129\n" +
	"\x18polling_interval_initial\x18\xed\a \x01(\x05R\x16pollingIntervalInitial\x12%\n" +
	"\x0emax_write_size\x18\xd3\x0f \x01(\x05R\fmaxWriteSize\x122\n" +
	"\x15max_write_duration_ms\x18\xd4\x0f \x01(\x05R\x12maxWriteDurationMs\x12J\n" +
	"!max_simultaneous_write_connection\x18\xd5\x0f \x01(\x05R\x1emaxSimultaneousWriteConnection\x123\n" +
	"\x15packet_writing_buffer\x18\xd6\x0f \x01(\x05R\x13packetWritingBuffer\x12\x11\n" +
	"\x03url\x18\xb9\x17 \x01(\tR\x03url\x12!\n" +
	"\fh2_pool_size\x18\xbb\x17 \x01(\x05R\n" +
	"h2PoolSize:\x1a\x82\xb5\x18\x16\n" +
	"\ttransport\x12\x05mekya\x90\xff)\x01B\xc3\x01\n" +
	":com.v2ray.core.transport.internet.request.stereotype.mekyaP\x01ZJgithub.com/v2fly/v2ray-core/v5/transport/internet/request/stereotype/mekya\xaa\x026V2Ray.Core.Transport.Internet.Request.Stereotype.Mekyab\x06proto3"

var (
	file_transport_internet_request_stereotype_mekya_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_stereotype_mekya_config_proto_rawDescData []byte
)

func file_transport_internet_request_stereotype_mekya_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_stereotype_mekya_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_stereotype_mekya_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_stereotype_mekya_config_proto_rawDesc), len(file_transport_internet_request_stereotype_mekya_config_proto_rawDesc)))
	})
	return file_transport_internet_request_stereotype_mekya_config_proto_rawDescData
}

var file_transport_internet_request_stereotype_mekya_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transport_internet_request_stereotype_mekya_config_proto_goTypes = []any{
	(*Config)(nil),     // 0: v2ray.core.transport.internet.request.stereotype.mekya.Config
	(*kcp.Config)(nil), // 1: v2ray.core.transport.internet.kcp.Config
}
var file_transport_internet_request_stereotype_mekya_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.transport.internet.request.stereotype.mekya.Config.kcp:type_name -> v2ray.core.transport.internet.kcp.Config
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_request_stereotype_mekya_config_proto_init() }
func file_transport_internet_request_stereotype_mekya_config_proto_init() {
	if File_transport_internet_request_stereotype_mekya_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_stereotype_mekya_config_proto_rawDesc), len(file_transport_internet_request_stereotype_mekya_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_stereotype_mekya_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_stereotype_mekya_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_stereotype_mekya_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_stereotype_mekya_config_proto = out.File
	file_transport_internet_request_stereotype_mekya_config_proto_goTypes = nil
	file_transport_internet_request_stereotype_mekya_config_proto_depIdxs = nil
}
