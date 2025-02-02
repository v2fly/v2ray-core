package packetaddr

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

type PacketAddrType int32

const (
	PacketAddrType_None   PacketAddrType = 0
	PacketAddrType_Packet PacketAddrType = 1
)

// Enum value maps for PacketAddrType.
var (
	PacketAddrType_name = map[int32]string{
		0: "None",
		1: "Packet",
	}
	PacketAddrType_value = map[string]int32{
		"None":   0,
		"Packet": 1,
	}
)

func (x PacketAddrType) Enum() *PacketAddrType {
	p := new(PacketAddrType)
	*p = x
	return p
}

func (x PacketAddrType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PacketAddrType) Descriptor() protoreflect.EnumDescriptor {
	return file_common_net_packetaddr_config_proto_enumTypes[0].Descriptor()
}

func (PacketAddrType) Type() protoreflect.EnumType {
	return &file_common_net_packetaddr_config_proto_enumTypes[0]
}

func (x PacketAddrType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PacketAddrType.Descriptor instead.
func (PacketAddrType) EnumDescriptor() ([]byte, []int) {
	return file_common_net_packetaddr_config_proto_rawDescGZIP(), []int{0}
}

var File_common_net_packetaddr_config_proto protoreflect.FileDescriptor

var file_common_net_packetaddr_config_proto_rawDesc = string([]byte{
	0x0a, 0x22, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x61, 0x63,
	0x6b, 0x65, 0x74, 0x61, 0x64, 0x64, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x61, 0x64, 0x64, 0x72, 0x2a,
	0x26, 0x0a, 0x0e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x50,
	0x61, 0x63, 0x6b, 0x65, 0x74, 0x10, 0x01, 0x42, 0x81, 0x01, 0x0a, 0x24, 0x63, 0x6f, 0x6d, 0x2e,
	0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x61, 0x64, 0x64, 0x72,
	0x50, 0x01, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76,
	0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x76, 0x35, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x65, 0x74, 0x61, 0x64, 0x64, 0x72, 0xaa, 0x02, 0x20, 0x56, 0x32, 0x52, 0x61, 0x79,
	0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x4e, 0x65, 0x74,
	0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x61, 0x64, 0x64, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
})

var (
	file_common_net_packetaddr_config_proto_rawDescOnce sync.Once
	file_common_net_packetaddr_config_proto_rawDescData []byte
)

func file_common_net_packetaddr_config_proto_rawDescGZIP() []byte {
	file_common_net_packetaddr_config_proto_rawDescOnce.Do(func() {
		file_common_net_packetaddr_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_net_packetaddr_config_proto_rawDesc), len(file_common_net_packetaddr_config_proto_rawDesc)))
	})
	return file_common_net_packetaddr_config_proto_rawDescData
}

var file_common_net_packetaddr_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_net_packetaddr_config_proto_goTypes = []any{
	(PacketAddrType)(0), // 0: v2ray.core.net.packetaddr.PacketAddrType
}
var file_common_net_packetaddr_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_net_packetaddr_config_proto_init() }
func file_common_net_packetaddr_config_proto_init() {
	if File_common_net_packetaddr_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_net_packetaddr_config_proto_rawDesc), len(file_common_net_packetaddr_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_net_packetaddr_config_proto_goTypes,
		DependencyIndexes: file_common_net_packetaddr_config_proto_depIdxs,
		EnumInfos:         file_common_net_packetaddr_config_proto_enumTypes,
	}.Build()
	File_common_net_packetaddr_config_proto = out.File
	file_common_net_packetaddr_config_proto_goTypes = nil
	file_common_net_packetaddr_config_proto_depIdxs = nil
}
