package protocol

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
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

type ServerEndpoint struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       *net.IPOrDomain        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	User          []*User                `protobuf:"bytes,3,rep,name=user,proto3" json:"user,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerEndpoint) Reset() {
	*x = ServerEndpoint{}
	mi := &file_common_protocol_server_spec_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerEndpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerEndpoint) ProtoMessage() {}

func (x *ServerEndpoint) ProtoReflect() protoreflect.Message {
	mi := &file_common_protocol_server_spec_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerEndpoint.ProtoReflect.Descriptor instead.
func (*ServerEndpoint) Descriptor() ([]byte, []int) {
	return file_common_protocol_server_spec_proto_rawDescGZIP(), []int{0}
}

func (x *ServerEndpoint) GetAddress() *net.IPOrDomain {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *ServerEndpoint) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ServerEndpoint) GetUser() []*User {
	if x != nil {
		return x.User
	}
	return nil
}

var File_common_protocol_server_spec_proto protoreflect.FileDescriptor

const file_common_protocol_server_spec_proto_rawDesc = "" +
	"\n" +
	"!common/protocol/server_spec.proto\x12\x1av2ray.core.common.protocol\x1a\x18common/net/address.proto\x1a\x1acommon/protocol/user.proto\"\x97\x01\n" +
	"\x0eServerEndpoint\x12;\n" +
	"\aaddress\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\aaddress\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x124\n" +
	"\x04user\x18\x03 \x03(\v2 .v2ray.core.common.protocol.UserR\x04userBo\n" +
	"\x1ecom.v2ray.core.common.protocolP\x01Z.github.com/v2fly/v2ray-core/v5/common/protocol\xaa\x02\x1aV2Ray.Core.Common.Protocolb\x06proto3"

var (
	file_common_protocol_server_spec_proto_rawDescOnce sync.Once
	file_common_protocol_server_spec_proto_rawDescData []byte
)

func file_common_protocol_server_spec_proto_rawDescGZIP() []byte {
	file_common_protocol_server_spec_proto_rawDescOnce.Do(func() {
		file_common_protocol_server_spec_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_protocol_server_spec_proto_rawDesc), len(file_common_protocol_server_spec_proto_rawDesc)))
	})
	return file_common_protocol_server_spec_proto_rawDescData
}

var file_common_protocol_server_spec_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_protocol_server_spec_proto_goTypes = []any{
	(*ServerEndpoint)(nil), // 0: v2ray.core.common.protocol.ServerEndpoint
	(*net.IPOrDomain)(nil), // 1: v2ray.core.common.net.IPOrDomain
	(*User)(nil),           // 2: v2ray.core.common.protocol.User
}
var file_common_protocol_server_spec_proto_depIdxs = []int32{
	1, // 0: v2ray.core.common.protocol.ServerEndpoint.address:type_name -> v2ray.core.common.net.IPOrDomain
	2, // 1: v2ray.core.common.protocol.ServerEndpoint.user:type_name -> v2ray.core.common.protocol.User
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_common_protocol_server_spec_proto_init() }
func file_common_protocol_server_spec_proto_init() {
	if File_common_protocol_server_spec_proto != nil {
		return
	}
	file_common_protocol_user_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_protocol_server_spec_proto_rawDesc), len(file_common_protocol_server_spec_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_protocol_server_spec_proto_goTypes,
		DependencyIndexes: file_common_protocol_server_spec_proto_depIdxs,
		MessageInfos:      file_common_protocol_server_spec_proto_msgTypes,
	}.Build()
	File_common_protocol_server_spec_proto = out.File
	file_common_protocol_server_spec_proto_goTypes = nil
	file_common_protocol_server_spec_proto_depIdxs = nil
}
