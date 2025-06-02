package websocket

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

type Header struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Header) Reset() {
	*x = Header{}
	mi := &file_transport_internet_websocket_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_websocket_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_transport_internet_websocket_config_proto_rawDescGZIP(), []int{0}
}

func (x *Header) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type Config struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// URL path to the WebSocket service. Empty value means root(/).
	Path                 string    `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Header               []*Header `protobuf:"bytes,3,rep,name=header,proto3" json:"header,omitempty"`
	AcceptProxyProtocol  bool      `protobuf:"varint,4,opt,name=accept_proxy_protocol,json=acceptProxyProtocol,proto3" json:"accept_proxy_protocol,omitempty"`
	MaxEarlyData         int32     `protobuf:"varint,5,opt,name=max_early_data,json=maxEarlyData,proto3" json:"max_early_data,omitempty"`
	UseBrowserForwarding bool      `protobuf:"varint,6,opt,name=use_browser_forwarding,json=useBrowserForwarding,proto3" json:"use_browser_forwarding,omitempty"`
	EarlyDataHeaderName  string    `protobuf:"bytes,7,opt,name=early_data_header_name,json=earlyDataHeaderName,proto3" json:"early_data_header_name,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_websocket_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_websocket_config_proto_msgTypes[1]
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
	return file_transport_internet_websocket_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Config) GetHeader() []*Header {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Config) GetAcceptProxyProtocol() bool {
	if x != nil {
		return x.AcceptProxyProtocol
	}
	return false
}

func (x *Config) GetMaxEarlyData() int32 {
	if x != nil {
		return x.MaxEarlyData
	}
	return 0
}

func (x *Config) GetUseBrowserForwarding() bool {
	if x != nil {
		return x.UseBrowserForwarding
	}
	return false
}

func (x *Config) GetEarlyDataHeaderName() string {
	if x != nil {
		return x.EarlyDataHeaderName
	}
	return ""
}

var File_transport_internet_websocket_config_proto protoreflect.FileDescriptor

const file_transport_internet_websocket_config_proto_rawDesc = "" +
	"\n" +
	")transport/internet/websocket/config.proto\x12'v2ray.core.transport.internet.websocket\x1a common/protoext/extensions.proto\"0\n" +
	"\x06Header\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\"\xd6\x02\n" +
	"\x06Config\x12\x12\n" +
	"\x04path\x18\x02 \x01(\tR\x04path\x12G\n" +
	"\x06header\x18\x03 \x03(\v2/.v2ray.core.transport.internet.websocket.HeaderR\x06header\x122\n" +
	"\x15accept_proxy_protocol\x18\x04 \x01(\bR\x13acceptProxyProtocol\x12$\n" +
	"\x0emax_early_data\x18\x05 \x01(\x05R\fmaxEarlyData\x124\n" +
	"\x16use_browser_forwarding\x18\x06 \x01(\bR\x14useBrowserForwarding\x123\n" +
	"\x16early_data_header_name\x18\a \x01(\tR\x13earlyDataHeaderName:$\x82\xb5\x18 \n" +
	"\ttransport\x12\x02ws\x8a\xff)\twebsocket\x90\xff)\x01J\x04\b\x01\x10\x02B\x96\x01\n" +
	"+com.v2ray.core.transport.internet.websocketP\x01Z;github.com/v2fly/v2ray-core/v5/transport/internet/websocket\xaa\x02'V2Ray.Core.Transport.Internet.Websocketb\x06proto3"

var (
	file_transport_internet_websocket_config_proto_rawDescOnce sync.Once
	file_transport_internet_websocket_config_proto_rawDescData []byte
)

func file_transport_internet_websocket_config_proto_rawDescGZIP() []byte {
	file_transport_internet_websocket_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_websocket_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_websocket_config_proto_rawDesc), len(file_transport_internet_websocket_config_proto_rawDesc)))
	})
	return file_transport_internet_websocket_config_proto_rawDescData
}

var file_transport_internet_websocket_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_websocket_config_proto_goTypes = []any{
	(*Header)(nil), // 0: v2ray.core.transport.internet.websocket.Header
	(*Config)(nil), // 1: v2ray.core.transport.internet.websocket.Config
}
var file_transport_internet_websocket_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.transport.internet.websocket.Config.header:type_name -> v2ray.core.transport.internet.websocket.Header
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_websocket_config_proto_init() }
func file_transport_internet_websocket_config_proto_init() {
	if File_transport_internet_websocket_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_websocket_config_proto_rawDesc), len(file_transport_internet_websocket_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_websocket_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_websocket_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_websocket_config_proto_msgTypes,
	}.Build()
	File_transport_internet_websocket_config_proto = out.File
	file_transport_internet_websocket_config_proto_goTypes = nil
	file_transport_internet_websocket_config_proto_depIdxs = nil
}
