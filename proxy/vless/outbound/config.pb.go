// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: proxy/vless/outbound/config.proto

package outbound

import (
	proto "github.com/golang/protobuf/proto"
	protocol "github.com/v2fly/v2ray-core/common/protocol"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vnext []*protocol.ServerEndpoint `protobuf:"bytes,1,rep,name=vnext,proto3" json:"vnext,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proxy_vless_outbound_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vless_outbound_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_proxy_vless_outbound_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetVnext() []*protocol.ServerEndpoint {
	if x != nil {
		return x.Vnext
	}
	return nil
}

var File_proxy_vless_outbound_config_proto protoreflect.FileDescriptor

var file_proxy_vless_outbound_config_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f, 0x76, 0x6c, 0x65, 0x73, 0x73, 0x2f, 0x6f, 0x75,
	0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x1f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x6c, 0x65, 0x73, 0x73, 0x2e, 0x6f, 0x75, 0x74, 0x62,
	0x6f, 0x75, 0x6e, 0x64, 0x1a, 0x21, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x73, 0x70, 0x65,
	0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4a, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x40, 0x0a, 0x05, 0x76, 0x6e, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x2a, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x53, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x05, 0x76, 0x6e,
	0x65, 0x78, 0x74, 0x42, 0x7b, 0x0a, 0x23, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x6c, 0x65, 0x73,
	0x73, 0x2e, 0x6f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x50, 0x01, 0x5a, 0x30, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76,
	0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f,
	0x76, 0x6c, 0x65, 0x73, 0x73, 0x2f, 0x6f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0xaa, 0x02,
	0x1f, 0x56, 0x32, 0x52, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x78,
	0x79, 0x2e, 0x56, 0x6c, 0x65, 0x73, 0x73, 0x2e, 0x4f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proxy_vless_outbound_config_proto_rawDescOnce sync.Once
	file_proxy_vless_outbound_config_proto_rawDescData = file_proxy_vless_outbound_config_proto_rawDesc
)

func file_proxy_vless_outbound_config_proto_rawDescGZIP() []byte {
	file_proxy_vless_outbound_config_proto_rawDescOnce.Do(func() {
		file_proxy_vless_outbound_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_proxy_vless_outbound_config_proto_rawDescData)
	})
	return file_proxy_vless_outbound_config_proto_rawDescData
}

var file_proxy_vless_outbound_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proxy_vless_outbound_config_proto_goTypes = []interface{}{
	(*Config)(nil),                  // 0: v2ray.core.proxy.vless.outbound.Config
	(*protocol.ServerEndpoint)(nil), // 1: v2ray.core.common.protocol.ServerEndpoint
}
var file_proxy_vless_outbound_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.proxy.vless.outbound.Config.vnext:type_name -> v2ray.core.common.protocol.ServerEndpoint
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proxy_vless_outbound_config_proto_init() }
func file_proxy_vless_outbound_config_proto_init() {
	if File_proxy_vless_outbound_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proxy_vless_outbound_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proxy_vless_outbound_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_vless_outbound_config_proto_goTypes,
		DependencyIndexes: file_proxy_vless_outbound_config_proto_depIdxs,
		MessageInfos:      file_proxy_vless_outbound_config_proto_msgTypes,
	}.Build()
	File_proxy_vless_outbound_config_proto = out.File
	file_proxy_vless_outbound_config_proto_rawDesc = nil
	file_proxy_vless_outbound_config_proto_goTypes = nil
	file_proxy_vless_outbound_config_proto_depIdxs = nil
}
