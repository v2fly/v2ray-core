package net

import (
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

// Address of a network host. It may be either an IP address or a domain
// address.
type IPOrDomain struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Address:
	//
	//	*IPOrDomain_Ip
	//	*IPOrDomain_Domain
	Address       isIPOrDomain_Address `protobuf_oneof:"address"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IPOrDomain) Reset() {
	*x = IPOrDomain{}
	mi := &file_common_net_address_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IPOrDomain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPOrDomain) ProtoMessage() {}

func (x *IPOrDomain) ProtoReflect() protoreflect.Message {
	mi := &file_common_net_address_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPOrDomain.ProtoReflect.Descriptor instead.
func (*IPOrDomain) Descriptor() ([]byte, []int) {
	return file_common_net_address_proto_rawDescGZIP(), []int{0}
}

func (x *IPOrDomain) GetAddress() isIPOrDomain_Address {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *IPOrDomain) GetIp() []byte {
	if x != nil {
		if x, ok := x.Address.(*IPOrDomain_Ip); ok {
			return x.Ip
		}
	}
	return nil
}

func (x *IPOrDomain) GetDomain() string {
	if x != nil {
		if x, ok := x.Address.(*IPOrDomain_Domain); ok {
			return x.Domain
		}
	}
	return ""
}

type isIPOrDomain_Address interface {
	isIPOrDomain_Address()
}

type IPOrDomain_Ip struct {
	// IP address. Must by either 4 or 16 bytes.
	Ip []byte `protobuf:"bytes,1,opt,name=ip,proto3,oneof"`
}

type IPOrDomain_Domain struct {
	// Domain address.
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3,oneof"`
}

func (*IPOrDomain_Ip) isIPOrDomain_Address() {}

func (*IPOrDomain_Domain) isIPOrDomain_Address() {}

var File_common_net_address_proto protoreflect.FileDescriptor

var file_common_net_address_proto_rawDesc = string([]byte{
	0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6e, 0x65, 0x74, 0x2f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x76, 0x32, 0x72, 0x61,
	0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x6e, 0x65,
	0x74, 0x22, 0x43, 0x0a, 0x0a, 0x49, 0x50, 0x4f, 0x72, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12,
	0x10, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x02, 0x69,
	0x70, 0x12, 0x18, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x42, 0x09, 0x0a, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x42, 0x60, 0x0a, 0x19, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32,
	0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x6e, 0x65, 0x74, 0x50, 0x01, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x76, 0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f,
	0x72, 0x65, 0x2f, 0x76, 0x35, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6e, 0x65, 0x74,
	0xaa, 0x02, 0x15, 0x56, 0x32, 0x52, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x4e, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_common_net_address_proto_rawDescOnce sync.Once
	file_common_net_address_proto_rawDescData []byte
)

func file_common_net_address_proto_rawDescGZIP() []byte {
	file_common_net_address_proto_rawDescOnce.Do(func() {
		file_common_net_address_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_net_address_proto_rawDesc), len(file_common_net_address_proto_rawDesc)))
	})
	return file_common_net_address_proto_rawDescData
}

var file_common_net_address_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_net_address_proto_goTypes = []any{
	(*IPOrDomain)(nil), // 0: v2ray.core.common.net.IPOrDomain
}
var file_common_net_address_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_net_address_proto_init() }
func file_common_net_address_proto_init() {
	if File_common_net_address_proto != nil {
		return
	}
	file_common_net_address_proto_msgTypes[0].OneofWrappers = []any{
		(*IPOrDomain_Ip)(nil),
		(*IPOrDomain_Domain)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_net_address_proto_rawDesc), len(file_common_net_address_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_net_address_proto_goTypes,
		DependencyIndexes: file_common_net_address_proto_depIdxs,
		MessageInfos:      file_common_net_address_proto_msgTypes,
	}.Build()
	File_common_net_address_proto = out.File
	file_common_net_address_proto_goTypes = nil
	file_common_net_address_proto_depIdxs = nil
}
