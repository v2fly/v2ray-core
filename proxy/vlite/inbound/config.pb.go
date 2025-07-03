package inbound

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

type UDPProtocolConfig struct {
	state                       protoimpl.MessageState `protogen:"open.v1"`
	Password                    string                 `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	ScramblePacket              bool                   `protobuf:"varint,4,opt,name=scramble_packet,json=scramblePacket,proto3" json:"scramble_packet,omitempty"`
	EnableFec                   bool                   `protobuf:"varint,5,opt,name=enable_fec,json=enableFec,proto3" json:"enable_fec,omitempty"`
	EnableStabilization         bool                   `protobuf:"varint,6,opt,name=enable_stabilization,json=enableStabilization,proto3" json:"enable_stabilization,omitempty"`
	EnableRenegotiation         bool                   `protobuf:"varint,7,opt,name=enable_renegotiation,json=enableRenegotiation,proto3" json:"enable_renegotiation,omitempty"`
	HandshakeMaskingPaddingSize uint32                 `protobuf:"varint,8,opt,name=handshake_masking_padding_size,json=handshakeMaskingPaddingSize,proto3" json:"handshake_masking_padding_size,omitempty"`
	unknownFields               protoimpl.UnknownFields
	sizeCache                   protoimpl.SizeCache
}

func (x *UDPProtocolConfig) Reset() {
	*x = UDPProtocolConfig{}
	mi := &file_proxy_vlite_inbound_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UDPProtocolConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UDPProtocolConfig) ProtoMessage() {}

func (x *UDPProtocolConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proxy_vlite_inbound_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UDPProtocolConfig.ProtoReflect.Descriptor instead.
func (*UDPProtocolConfig) Descriptor() ([]byte, []int) {
	return file_proxy_vlite_inbound_config_proto_rawDescGZIP(), []int{0}
}

func (x *UDPProtocolConfig) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *UDPProtocolConfig) GetScramblePacket() bool {
	if x != nil {
		return x.ScramblePacket
	}
	return false
}

func (x *UDPProtocolConfig) GetEnableFec() bool {
	if x != nil {
		return x.EnableFec
	}
	return false
}

func (x *UDPProtocolConfig) GetEnableStabilization() bool {
	if x != nil {
		return x.EnableStabilization
	}
	return false
}

func (x *UDPProtocolConfig) GetEnableRenegotiation() bool {
	if x != nil {
		return x.EnableRenegotiation
	}
	return false
}

func (x *UDPProtocolConfig) GetHandshakeMaskingPaddingSize() uint32 {
	if x != nil {
		return x.HandshakeMaskingPaddingSize
	}
	return 0
}

var File_proxy_vlite_inbound_config_proto protoreflect.FileDescriptor

const file_proxy_vlite_inbound_config_proto_rawDesc = "" +
	"\n" +
	" proxy/vlite/inbound/config.proto\x12\x1ev2ray.core.proxy.vlite.inbound\x1a common/protoext/extensions.proto\"\xb9\x02\n" +
	"\x11UDPProtocolConfig\x12\x1a\n" +
	"\bpassword\x18\x03 \x01(\tR\bpassword\x12'\n" +
	"\x0fscramble_packet\x18\x04 \x01(\bR\x0escramblePacket\x12\x1d\n" +
	"\n" +
	"enable_fec\x18\x05 \x01(\bR\tenableFec\x121\n" +
	"\x14enable_stabilization\x18\x06 \x01(\bR\x13enableStabilization\x121\n" +
	"\x14enable_renegotiation\x18\a \x01(\bR\x13enableRenegotiation\x12C\n" +
	"\x1ehandshake_masking_padding_size\x18\b \x01(\rR\x1bhandshakeMaskingPaddingSize:\x15\x82\xb5\x18\x11\n" +
	"\ainbound\x12\x06vliteuB{\n" +
	"\"com.v2ray.core.proxy.vlite.inboundP\x01Z2github.com/v2fly/v2ray-core/v5/proxy/vlite/inbound\xaa\x02\x1eV2Ray.Core.Proxy.Vlite.Inboundb\x06proto3"

var (
	file_proxy_vlite_inbound_config_proto_rawDescOnce sync.Once
	file_proxy_vlite_inbound_config_proto_rawDescData []byte
)

func file_proxy_vlite_inbound_config_proto_rawDescGZIP() []byte {
	file_proxy_vlite_inbound_config_proto_rawDescOnce.Do(func() {
		file_proxy_vlite_inbound_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proxy_vlite_inbound_config_proto_rawDesc), len(file_proxy_vlite_inbound_config_proto_rawDesc)))
	})
	return file_proxy_vlite_inbound_config_proto_rawDescData
}

var file_proxy_vlite_inbound_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proxy_vlite_inbound_config_proto_goTypes = []any{
	(*UDPProtocolConfig)(nil), // 0: v2ray.core.proxy.vlite.inbound.UDPProtocolConfig
}
var file_proxy_vlite_inbound_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proxy_vlite_inbound_config_proto_init() }
func file_proxy_vlite_inbound_config_proto_init() {
	if File_proxy_vlite_inbound_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proxy_vlite_inbound_config_proto_rawDesc), len(file_proxy_vlite_inbound_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proxy_vlite_inbound_config_proto_goTypes,
		DependencyIndexes: file_proxy_vlite_inbound_config_proto_depIdxs,
		MessageInfos:      file_proxy_vlite_inbound_config_proto_msgTypes,
	}.Build()
	File_proxy_vlite_inbound_config_proto = out.File
	file_proxy_vlite_inbound_config_proto_goTypes = nil
	file_proxy_vlite_inbound_config_proto_depIdxs = nil
}
