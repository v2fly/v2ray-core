package httpupgrade

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
	mi := &file_transport_internet_httpupgrade_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_httpupgrade_config_proto_msgTypes[0]
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
	return file_transport_internet_httpupgrade_config_proto_rawDescGZIP(), []int{0}
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
	state               protoimpl.MessageState `protogen:"open.v1"`
	Path                string                 `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Host                string                 `protobuf:"bytes,2,opt,name=host,proto3" json:"host,omitempty"`
	MaxEarlyData        int32                  `protobuf:"varint,3,opt,name=max_early_data,json=maxEarlyData,proto3" json:"max_early_data,omitempty"`
	EarlyDataHeaderName string                 `protobuf:"bytes,4,opt,name=early_data_header_name,json=earlyDataHeaderName,proto3" json:"early_data_header_name,omitempty"`
	Header              []*Header              `protobuf:"bytes,5,rep,name=header,proto3" json:"header,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_httpupgrade_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_httpupgrade_config_proto_msgTypes[1]
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
	return file_transport_internet_httpupgrade_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Config) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Config) GetMaxEarlyData() int32 {
	if x != nil {
		return x.MaxEarlyData
	}
	return 0
}

func (x *Config) GetEarlyDataHeaderName() string {
	if x != nil {
		return x.EarlyDataHeaderName
	}
	return ""
}

func (x *Config) GetHeader() []*Header {
	if x != nil {
		return x.Header
	}
	return nil
}

var File_transport_internet_httpupgrade_config_proto protoreflect.FileDescriptor

const file_transport_internet_httpupgrade_config_proto_rawDesc = "" +
	"\n" +
	"+transport/internet/httpupgrade/config.proto\x121v2ray.core.transport.internet.request.httpupgrade\x1a common/protoext/extensions.proto\"0\n" +
	"\x06Header\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\"\x80\x02\n" +
	"\x06Config\x12\x12\n" +
	"\x04path\x18\x01 \x01(\tR\x04path\x12\x12\n" +
	"\x04host\x18\x02 \x01(\tR\x04host\x12$\n" +
	"\x0emax_early_data\x18\x03 \x01(\x05R\fmaxEarlyData\x123\n" +
	"\x16early_data_header_name\x18\x04 \x01(\tR\x13earlyDataHeaderName\x12Q\n" +
	"\x06header\x18\x05 \x03(\v29.v2ray.core.transport.internet.request.httpupgrade.HeaderR\x06header: \x82\xb5\x18\x1c\n" +
	"\ttransport\x12\vhttpupgrade\x90\xff)\x01B\x9c\x01\n" +
	"-com.v2ray.core.transport.internet.httpupgradeP\x01Z=github.com/v2fly/v2ray-core/v5/transport/internet/httpupgrade\xaa\x02)V2Ray.Core.Transport.Internet.HttpUpgradeb\x06proto3"

var (
	file_transport_internet_httpupgrade_config_proto_rawDescOnce sync.Once
	file_transport_internet_httpupgrade_config_proto_rawDescData []byte
)

func file_transport_internet_httpupgrade_config_proto_rawDescGZIP() []byte {
	file_transport_internet_httpupgrade_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_httpupgrade_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_httpupgrade_config_proto_rawDesc), len(file_transport_internet_httpupgrade_config_proto_rawDesc)))
	})
	return file_transport_internet_httpupgrade_config_proto_rawDescData
}

var file_transport_internet_httpupgrade_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_httpupgrade_config_proto_goTypes = []any{
	(*Header)(nil), // 0: v2ray.core.transport.internet.request.httpupgrade.Header
	(*Config)(nil), // 1: v2ray.core.transport.internet.request.httpupgrade.Config
}
var file_transport_internet_httpupgrade_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.transport.internet.request.httpupgrade.Config.header:type_name -> v2ray.core.transport.internet.request.httpupgrade.Header
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_httpupgrade_config_proto_init() }
func file_transport_internet_httpupgrade_config_proto_init() {
	if File_transport_internet_httpupgrade_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_httpupgrade_config_proto_rawDesc), len(file_transport_internet_httpupgrade_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_httpupgrade_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_httpupgrade_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_httpupgrade_config_proto_msgTypes,
	}.Build()
	File_transport_internet_httpupgrade_config_proto = out.File
	file_transport_internet_httpupgrade_config_proto_goTypes = nil
	file_transport_internet_httpupgrade_config_proto_depIdxs = nil
}
