package protoext

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
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

type MessageOpt struct {
	state                 protoimpl.MessageState `protogen:"open.v1"`
	Type                  []string               `protobuf:"bytes,1,rep,name=type,proto3" json:"type,omitempty"`
	ShortName             []string               `protobuf:"bytes,2,rep,name=short_name,json=shortName,proto3" json:"short_name,omitempty"`
	TransportOriginalName string                 `protobuf:"bytes,86001,opt,name=transport_original_name,json=transportOriginalName,proto3" json:"transport_original_name,omitempty"`
	// allow_restricted_mode_load allow this config to be loaded in restricted mode
	// this is typically used when a an attacker can control the content
	AllowRestrictedModeLoad bool `protobuf:"varint,86002,opt,name=allow_restricted_mode_load,json=allowRestrictedModeLoad,proto3" json:"allow_restricted_mode_load,omitempty"`
	unknownFields           protoimpl.UnknownFields
	sizeCache               protoimpl.SizeCache
}

func (x *MessageOpt) Reset() {
	*x = MessageOpt{}
	mi := &file_common_protoext_extensions_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageOpt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageOpt) ProtoMessage() {}

func (x *MessageOpt) ProtoReflect() protoreflect.Message {
	mi := &file_common_protoext_extensions_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageOpt.ProtoReflect.Descriptor instead.
func (*MessageOpt) Descriptor() ([]byte, []int) {
	return file_common_protoext_extensions_proto_rawDescGZIP(), []int{0}
}

func (x *MessageOpt) GetType() []string {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *MessageOpt) GetShortName() []string {
	if x != nil {
		return x.ShortName
	}
	return nil
}

func (x *MessageOpt) GetTransportOriginalName() string {
	if x != nil {
		return x.TransportOriginalName
	}
	return ""
}

func (x *MessageOpt) GetAllowRestrictedModeLoad() bool {
	if x != nil {
		return x.AllowRestrictedModeLoad
	}
	return false
}

type FieldOpt struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	AnyWants          []string               `protobuf:"bytes,1,rep,name=any_wants,json=anyWants,proto3" json:"any_wants,omitempty"`
	AllowedValues     []string               `protobuf:"bytes,2,rep,name=allowed_values,json=allowedValues,proto3" json:"allowed_values,omitempty"`
	AllowedValueTypes []string               `protobuf:"bytes,3,rep,name=allowed_value_types,json=allowedValueTypes,proto3" json:"allowed_value_types,omitempty"`
	// convert_time_read_file_into read a file into another field, and clear this field during input parsing
	ConvertTimeReadFileInto string `protobuf:"bytes,4,opt,name=convert_time_read_file_into,json=convertTimeReadFileInto,proto3" json:"convert_time_read_file_into,omitempty"`
	// forbidden marks a boolean to be inaccessible to user
	Forbidden bool `protobuf:"varint,5,opt,name=forbidden,proto3" json:"forbidden,omitempty"`
	// convert_time_resource_loading read a file, and place its resource hash into another field
	ConvertTimeResourceLoading string `protobuf:"bytes,6,opt,name=convert_time_resource_loading,json=convertTimeResourceLoading,proto3" json:"convert_time_resource_loading,omitempty"`
	// convert_time_parse_ip parse a string ip address, and put its binary representation into another field
	ConvertTimeParseIp string `protobuf:"bytes,7,opt,name=convert_time_parse_ip,json=convertTimeParseIp,proto3" json:"convert_time_parse_ip,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *FieldOpt) Reset() {
	*x = FieldOpt{}
	mi := &file_common_protoext_extensions_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FieldOpt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldOpt) ProtoMessage() {}

func (x *FieldOpt) ProtoReflect() protoreflect.Message {
	mi := &file_common_protoext_extensions_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldOpt.ProtoReflect.Descriptor instead.
func (*FieldOpt) Descriptor() ([]byte, []int) {
	return file_common_protoext_extensions_proto_rawDescGZIP(), []int{1}
}

func (x *FieldOpt) GetAnyWants() []string {
	if x != nil {
		return x.AnyWants
	}
	return nil
}

func (x *FieldOpt) GetAllowedValues() []string {
	if x != nil {
		return x.AllowedValues
	}
	return nil
}

func (x *FieldOpt) GetAllowedValueTypes() []string {
	if x != nil {
		return x.AllowedValueTypes
	}
	return nil
}

func (x *FieldOpt) GetConvertTimeReadFileInto() string {
	if x != nil {
		return x.ConvertTimeReadFileInto
	}
	return ""
}

func (x *FieldOpt) GetForbidden() bool {
	if x != nil {
		return x.Forbidden
	}
	return false
}

func (x *FieldOpt) GetConvertTimeResourceLoading() string {
	if x != nil {
		return x.ConvertTimeResourceLoading
	}
	return ""
}

func (x *FieldOpt) GetConvertTimeParseIp() string {
	if x != nil {
		return x.ConvertTimeParseIp
	}
	return ""
}

var file_common_protoext_extensions_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*MessageOpt)(nil),
		Field:         50000,
		Name:          "v2ray.core.common.protoext.message_opt",
		Tag:           "bytes,50000,opt,name=message_opt",
		Filename:      "common/protoext/extensions.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*FieldOpt)(nil),
		Field:         50000,
		Name:          "v2ray.core.common.protoext.field_opt",
		Tag:           "bytes,50000,opt,name=field_opt",
		Filename:      "common/protoext/extensions.proto",
	},
}

// Extension fields to descriptorpb.MessageOptions.
var (
	// optional v2ray.core.common.protoext.MessageOpt message_opt = 50000;
	E_MessageOpt = &file_common_protoext_extensions_proto_extTypes[0]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional v2ray.core.common.protoext.FieldOpt field_opt = 50000;
	E_FieldOpt = &file_common_protoext_extensions_proto_extTypes[1]
)

var File_common_protoext_extensions_proto protoreflect.FileDescriptor

const file_common_protoext_extensions_proto_rawDesc = "" +
	"\n" +
	" common/protoext/extensions.proto\x12\x1av2ray.core.common.protoext\x1a google/protobuf/descriptor.proto\"\xb8\x01\n" +
	"\n" +
	"MessageOpt\x12\x12\n" +
	"\x04type\x18\x01 \x03(\tR\x04type\x12\x1d\n" +
	"\n" +
	"short_name\x18\x02 \x03(\tR\tshortName\x128\n" +
	"\x17transport_original_name\x18\xf1\x9f\x05 \x01(\tR\x15transportOriginalName\x12=\n" +
	"\x1aallow_restricted_mode_load\x18\xf2\x9f\x05 \x01(\bR\x17allowRestrictedModeLoad\"\xd0\x02\n" +
	"\bFieldOpt\x12\x1b\n" +
	"\tany_wants\x18\x01 \x03(\tR\banyWants\x12%\n" +
	"\x0eallowed_values\x18\x02 \x03(\tR\rallowedValues\x12.\n" +
	"\x13allowed_value_types\x18\x03 \x03(\tR\x11allowedValueTypes\x12<\n" +
	"\x1bconvert_time_read_file_into\x18\x04 \x01(\tR\x17convertTimeReadFileInto\x12\x1c\n" +
	"\tforbidden\x18\x05 \x01(\bR\tforbidden\x12A\n" +
	"\x1dconvert_time_resource_loading\x18\x06 \x01(\tR\x1aconvertTimeResourceLoading\x121\n" +
	"\x15convert_time_parse_ip\x18\a \x01(\tR\x12convertTimeParseIp:j\n" +
	"\vmessage_opt\x12\x1f.google.protobuf.MessageOptions\x18І\x03 \x01(\v2&.v2ray.core.common.protoext.MessageOptR\n" +
	"messageOpt:b\n" +
	"\tfield_opt\x12\x1d.google.protobuf.FieldOptions\x18І\x03 \x01(\v2$.v2ray.core.common.protoext.FieldOptR\bfieldOptBo\n" +
	"\x1ecom.v2ray.core.common.protoextP\x01Z.github.com/v2fly/v2ray-core/v5/common/protoext\xaa\x02\x1aV2Ray.Core.Common.ProtoExtb\x06proto3"

var (
	file_common_protoext_extensions_proto_rawDescOnce sync.Once
	file_common_protoext_extensions_proto_rawDescData []byte
)

func file_common_protoext_extensions_proto_rawDescGZIP() []byte {
	file_common_protoext_extensions_proto_rawDescOnce.Do(func() {
		file_common_protoext_extensions_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_protoext_extensions_proto_rawDesc), len(file_common_protoext_extensions_proto_rawDesc)))
	})
	return file_common_protoext_extensions_proto_rawDescData
}

var file_common_protoext_extensions_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_common_protoext_extensions_proto_goTypes = []any{
	(*MessageOpt)(nil),                  // 0: v2ray.core.common.protoext.MessageOpt
	(*FieldOpt)(nil),                    // 1: v2ray.core.common.protoext.FieldOpt
	(*descriptorpb.MessageOptions)(nil), // 2: google.protobuf.MessageOptions
	(*descriptorpb.FieldOptions)(nil),   // 3: google.protobuf.FieldOptions
}
var file_common_protoext_extensions_proto_depIdxs = []int32{
	2, // 0: v2ray.core.common.protoext.message_opt:extendee -> google.protobuf.MessageOptions
	3, // 1: v2ray.core.common.protoext.field_opt:extendee -> google.protobuf.FieldOptions
	0, // 2: v2ray.core.common.protoext.message_opt:type_name -> v2ray.core.common.protoext.MessageOpt
	1, // 3: v2ray.core.common.protoext.field_opt:type_name -> v2ray.core.common.protoext.FieldOpt
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	2, // [2:4] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_protoext_extensions_proto_init() }
func file_common_protoext_extensions_proto_init() {
	if File_common_protoext_extensions_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_protoext_extensions_proto_rawDesc), len(file_common_protoext_extensions_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_common_protoext_extensions_proto_goTypes,
		DependencyIndexes: file_common_protoext_extensions_proto_depIdxs,
		MessageInfos:      file_common_protoext_extensions_proto_msgTypes,
		ExtensionInfos:    file_common_protoext_extensions_proto_extTypes,
	}.Build()
	File_common_protoext_extensions_proto = out.File
	file_common_protoext_extensions_proto_goTypes = nil
	file_common_protoext_extensions_proto_depIdxs = nil
}
