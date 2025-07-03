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

const file_common_net_packetaddr_config_proto_rawDesc = "" +
	"\n" +
	"\"common/net/packetaddr/config.proto\x12\x19v2ray.core.net.packetaddr*&\n" +
	"\x0ePacketAddrType\x12\b\n" +
	"\x04None\x10\x00\x12\n" +
	"\n" +
	"\x06Packet\x10\x01B\x81\x01\n" +
	"$com.v2ray.core.common.net.packetaddrP\x01Z4github.com/v2fly/v2ray-core/v5/common/net/packetaddr\xaa\x02 V2Ray.Core.Common.Net.Packetaddrb\x06proto3"

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
