package httprt

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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Http          *HTTPConfig            `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
	AllowHttp     bool                   `protobuf:"varint,2,opt,name=allow_http,json=allowHttp,proto3" json:"allow_http,omitempty"`
	H2PoolSize    int32                  `protobuf:"varint,3,opt,name=h2_pool_size,json=h2PoolSize,proto3" json:"h2_pool_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[0]
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
	return file_transport_internet_request_roundtripper_httprt_config_proto_rawDescGZIP(), []int{0}
}

func (x *ClientConfig) GetHttp() *HTTPConfig {
	if x != nil {
		return x.Http
	}
	return nil
}

func (x *ClientConfig) GetAllowHttp() bool {
	if x != nil {
		return x.AllowHttp
	}
	return false
}

func (x *ClientConfig) GetH2PoolSize() int32 {
	if x != nil {
		return x.H2PoolSize
	}
	return 0
}

type ServerConfig struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	Http                 *HTTPConfig            `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
	NoDecodingSessionTag bool                   `protobuf:"varint,2,opt,name=no_decoding_session_tag,json=noDecodingSessionTag,proto3" json:"no_decoding_session_tag,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[1]
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
	return file_transport_internet_request_roundtripper_httprt_config_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetHttp() *HTTPConfig {
	if x != nil {
		return x.Http
	}
	return nil
}

func (x *ServerConfig) GetNoDecodingSessionTag() bool {
	if x != nil {
		return x.NoDecodingSessionTag
	}
	return false
}

type HTTPConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Path          string                 `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	UrlPrefix     string                 `protobuf:"bytes,2,opt,name=urlPrefix,proto3" json:"urlPrefix,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HTTPConfig) Reset() {
	*x = HTTPConfig{}
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HTTPConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPConfig) ProtoMessage() {}

func (x *HTTPConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPConfig.ProtoReflect.Descriptor instead.
func (*HTTPConfig) Descriptor() ([]byte, []int) {
	return file_transport_internet_request_roundtripper_httprt_config_proto_rawDescGZIP(), []int{2}
}

func (x *HTTPConfig) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *HTTPConfig) GetUrlPrefix() string {
	if x != nil {
		return x.UrlPrefix
	}
	return ""
}

var File_transport_internet_request_roundtripper_httprt_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_roundtripper_httprt_config_proto_rawDesc = "" +
	"\n" +
	";transport/internet/request/roundtripper/httprt/config.proto\x129v2ray.core.transport.internet.request.roundtripper.httprt\x1a common/protoext/extensions.proto\"\xdf\x01\n" +
	"\fClientConfig\x12Y\n" +
	"\x04http\x18\x01 \x01(\v2E.v2ray.core.transport.internet.request.roundtripper.httprt.HTTPConfigR\x04http\x12\x1d\n" +
	"\n" +
	"allow_http\x18\x02 \x01(\bR\tallowHttp\x12 \n" +
	"\fh2_pool_size\x18\x03 \x01(\x05R\n" +
	"h2PoolSize:3\x82\xb5\x18/\n" +
	"%transport.request.roundtripper.client\x12\x06httprt\"\xd5\x01\n" +
	"\fServerConfig\x12Y\n" +
	"\x04http\x18\x01 \x01(\v2E.v2ray.core.transport.internet.request.roundtripper.httprt.HTTPConfigR\x04http\x125\n" +
	"\x17no_decoding_session_tag\x18\x02 \x01(\bR\x14noDecodingSessionTag:3\x82\xb5\x18/\n" +
	"%transport.request.roundtripper.server\x12\x06httprt\">\n" +
	"\n" +
	"HTTPConfig\x12\x12\n" +
	"\x04path\x18\x01 \x01(\tR\x04path\x12\x1c\n" +
	"\turlPrefix\x18\x02 \x01(\tR\turlPrefixB\xcc\x01\n" +
	"=com.v2ray.core.transport.internet.request.roundtripper.httprtP\x01ZMgithub.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/httprt\xaa\x029V2Ray.Core.Transport.Internet.Request.Roundtripper.httprtb\x06proto3"

var (
	file_transport_internet_request_roundtripper_httprt_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_roundtripper_httprt_config_proto_rawDescData []byte
)

func file_transport_internet_request_roundtripper_httprt_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_roundtripper_httprt_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_roundtripper_httprt_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_roundtripper_httprt_config_proto_rawDesc), len(file_transport_internet_request_roundtripper_httprt_config_proto_rawDesc)))
	})
	return file_transport_internet_request_roundtripper_httprt_config_proto_rawDescData
}

var file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_transport_internet_request_roundtripper_httprt_config_proto_goTypes = []any{
	(*ClientConfig)(nil), // 0: v2ray.core.transport.internet.request.roundtripper.httprt.ClientConfig
	(*ServerConfig)(nil), // 1: v2ray.core.transport.internet.request.roundtripper.httprt.ServerConfig
	(*HTTPConfig)(nil),   // 2: v2ray.core.transport.internet.request.roundtripper.httprt.HTTPConfig
}
var file_transport_internet_request_roundtripper_httprt_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.request.roundtripper.httprt.ClientConfig.http:type_name -> v2ray.core.transport.internet.request.roundtripper.httprt.HTTPConfig
	2, // 1: v2ray.core.transport.internet.request.roundtripper.httprt.ServerConfig.http:type_name -> v2ray.core.transport.internet.request.roundtripper.httprt.HTTPConfig
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_internet_request_roundtripper_httprt_config_proto_init() }
func file_transport_internet_request_roundtripper_httprt_config_proto_init() {
	if File_transport_internet_request_roundtripper_httprt_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_roundtripper_httprt_config_proto_rawDesc), len(file_transport_internet_request_roundtripper_httprt_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_roundtripper_httprt_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_roundtripper_httprt_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_roundtripper_httprt_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_roundtripper_httprt_config_proto = out.File
	file_transport_internet_request_roundtripper_httprt_config_proto_goTypes = nil
	file_transport_internet_request_roundtripper_httprt_config_proto_depIdxs = nil
}
