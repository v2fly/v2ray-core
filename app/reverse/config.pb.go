package reverse

import (
	_ "github.com/ghxhy/v2ray-core/v5/common/protoext"
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

type Control_State int32

const (
	Control_ACTIVE Control_State = 0
	Control_DRAIN  Control_State = 1
)

// Enum value maps for Control_State.
var (
	Control_State_name = map[int32]string{
		0: "ACTIVE",
		1: "DRAIN",
	}
	Control_State_value = map[string]int32{
		"ACTIVE": 0,
		"DRAIN":  1,
	}
)

func (x Control_State) Enum() *Control_State {
	p := new(Control_State)
	*p = x
	return p
}

func (x Control_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Control_State) Descriptor() protoreflect.EnumDescriptor {
	return file_app_reverse_config_proto_enumTypes[0].Descriptor()
}

func (Control_State) Type() protoreflect.EnumType {
	return &file_app_reverse_config_proto_enumTypes[0]
}

func (x Control_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Control_State.Descriptor instead.
func (Control_State) EnumDescriptor() ([]byte, []int) {
	return file_app_reverse_config_proto_rawDescGZIP(), []int{0, 0}
}

type Control struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	State         Control_State          `protobuf:"varint,1,opt,name=state,proto3,enum=v2ray.core.app.reverse.Control_State" json:"state,omitempty"`
	Random        []byte                 `protobuf:"bytes,99,opt,name=random,proto3" json:"random,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Control) Reset() {
	*x = Control{}
	mi := &file_app_reverse_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Control) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Control) ProtoMessage() {}

func (x *Control) ProtoReflect() protoreflect.Message {
	mi := &file_app_reverse_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Control.ProtoReflect.Descriptor instead.
func (*Control) Descriptor() ([]byte, []int) {
	return file_app_reverse_config_proto_rawDescGZIP(), []int{0}
}

func (x *Control) GetState() Control_State {
	if x != nil {
		return x.State
	}
	return Control_ACTIVE
}

func (x *Control) GetRandom() []byte {
	if x != nil {
		return x.Random
	}
	return nil
}

type BridgeConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Domain        string                 `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BridgeConfig) Reset() {
	*x = BridgeConfig{}
	mi := &file_app_reverse_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BridgeConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BridgeConfig) ProtoMessage() {}

func (x *BridgeConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_reverse_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BridgeConfig.ProtoReflect.Descriptor instead.
func (*BridgeConfig) Descriptor() ([]byte, []int) {
	return file_app_reverse_config_proto_rawDescGZIP(), []int{1}
}

func (x *BridgeConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *BridgeConfig) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type PortalConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Domain        string                 `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PortalConfig) Reset() {
	*x = PortalConfig{}
	mi := &file_app_reverse_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PortalConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortalConfig) ProtoMessage() {}

func (x *PortalConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_reverse_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortalConfig.ProtoReflect.Descriptor instead.
func (*PortalConfig) Descriptor() ([]byte, []int) {
	return file_app_reverse_config_proto_rawDescGZIP(), []int{2}
}

func (x *PortalConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *PortalConfig) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BridgeConfig  []*BridgeConfig        `protobuf:"bytes,1,rep,name=bridge_config,json=bridgeConfig,proto3" json:"bridge_config,omitempty"`
	PortalConfig  []*PortalConfig        `protobuf:"bytes,2,rep,name=portal_config,json=portalConfig,proto3" json:"portal_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_reverse_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_reverse_config_proto_msgTypes[3]
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
	return file_app_reverse_config_proto_rawDescGZIP(), []int{3}
}

func (x *Config) GetBridgeConfig() []*BridgeConfig {
	if x != nil {
		return x.BridgeConfig
	}
	return nil
}

func (x *Config) GetPortalConfig() []*PortalConfig {
	if x != nil {
		return x.PortalConfig
	}
	return nil
}

var File_app_reverse_config_proto protoreflect.FileDescriptor

var file_app_reverse_config_proto_rawDesc = string([]byte{
	0x0a, 0x18, 0x61, 0x70, 0x70, 0x2f, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x76, 0x32, 0x72, 0x61,
	0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x72, 0x65, 0x76, 0x65, 0x72,
	0x73, 0x65, 0x1a, 0x20, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x65, 0x78, 0x74, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7e, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12,
	0x3b, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25,
	0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e,
	0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x18, 0x63, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x61,
	0x6e, 0x64, 0x6f, 0x6d, 0x22, 0x1e, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0a, 0x0a,
	0x06, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x44, 0x52, 0x41,
	0x49, 0x4e, 0x10, 0x01, 0x22, 0x38, 0x0a, 0x0c, 0x42, 0x72, 0x69, 0x64, 0x67, 0x65, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x74, 0x61, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x38,
	0x0a, 0x0c, 0x50, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x10,
	0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x61, 0x67,
	0x12, 0x16, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0xb6, 0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x49, 0x0a, 0x0d, 0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x76, 0x32, 0x72,
	0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x72, 0x65, 0x76, 0x65,
	0x72, 0x73, 0x65, 0x2e, 0x42, 0x72, 0x69, 0x64, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x0c, 0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x49,
	0x0a, 0x0d, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0x2e, 0x50,
	0x6f, 0x72, 0x74, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0c, 0x70, 0x6f, 0x72,
	0x74, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x3a, 0x16, 0x82, 0xb5, 0x18, 0x12, 0x0a,
	0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x07, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73,
	0x65, 0x42, 0x67, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73,
	0x65, 0x50, 0x01, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x76, 0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65,
	0x2f, 0x76, 0x35, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0xaa,
	0x02, 0x18, 0x56, 0x32, 0x52, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x50, 0x72, 0x6f,
	0x78, 0x79, 0x2e, 0x52, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var (
	file_app_reverse_config_proto_rawDescOnce sync.Once
	file_app_reverse_config_proto_rawDescData []byte
)

func file_app_reverse_config_proto_rawDescGZIP() []byte {
	file_app_reverse_config_proto_rawDescOnce.Do(func() {
		file_app_reverse_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_reverse_config_proto_rawDesc), len(file_app_reverse_config_proto_rawDesc)))
	})
	return file_app_reverse_config_proto_rawDescData
}

var file_app_reverse_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_app_reverse_config_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_app_reverse_config_proto_goTypes = []any{
	(Control_State)(0),   // 0: v2ray.core.app.reverse.Control.State
	(*Control)(nil),      // 1: v2ray.core.app.reverse.Control
	(*BridgeConfig)(nil), // 2: v2ray.core.app.reverse.BridgeConfig
	(*PortalConfig)(nil), // 3: v2ray.core.app.reverse.PortalConfig
	(*Config)(nil),       // 4: v2ray.core.app.reverse.Config
}
var file_app_reverse_config_proto_depIdxs = []int32{
	0, // 0: v2ray.core.app.reverse.Control.state:type_name -> v2ray.core.app.reverse.Control.State
	2, // 1: v2ray.core.app.reverse.Config.bridge_config:type_name -> v2ray.core.app.reverse.BridgeConfig
	3, // 2: v2ray.core.app.reverse.Config.portal_config:type_name -> v2ray.core.app.reverse.PortalConfig
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_app_reverse_config_proto_init() }
func file_app_reverse_config_proto_init() {
	if File_app_reverse_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_reverse_config_proto_rawDesc), len(file_app_reverse_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_reverse_config_proto_goTypes,
		DependencyIndexes: file_app_reverse_config_proto_depIdxs,
		EnumInfos:         file_app_reverse_config_proto_enumTypes,
		MessageInfos:      file_app_reverse_config_proto_msgTypes,
	}.Build()
	File_app_reverse_config_proto = out.File
	file_app_reverse_config_proto_goTypes = nil
	file_app_reverse_config_proto_depIdxs = nil
}
