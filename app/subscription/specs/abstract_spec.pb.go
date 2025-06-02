package specs

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type ServerConfiguration struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Protocol          string                 `protobuf:"bytes,1,opt,name=protocol,proto3" json:"protocol,omitempty"`
	ProtocolSettings  *anypb.Any             `protobuf:"bytes,2,opt,name=protocol_settings,json=protocolSettings,proto3" json:"protocol_settings,omitempty"`
	Transport         string                 `protobuf:"bytes,3,opt,name=transport,proto3" json:"transport,omitempty"`
	TransportSettings *anypb.Any             `protobuf:"bytes,4,opt,name=transport_settings,json=transportSettings,proto3" json:"transport_settings,omitempty"`
	Security          string                 `protobuf:"bytes,5,opt,name=security,proto3" json:"security,omitempty"`
	SecuritySettings  *anypb.Any             `protobuf:"bytes,6,opt,name=security_settings,json=securitySettings,proto3" json:"security_settings,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *ServerConfiguration) Reset() {
	*x = ServerConfiguration{}
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfiguration) ProtoMessage() {}

func (x *ServerConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfiguration.ProtoReflect.Descriptor instead.
func (*ServerConfiguration) Descriptor() ([]byte, []int) {
	return file_app_subscription_specs_abstract_spec_proto_rawDescGZIP(), []int{0}
}

func (x *ServerConfiguration) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *ServerConfiguration) GetProtocolSettings() *anypb.Any {
	if x != nil {
		return x.ProtocolSettings
	}
	return nil
}

func (x *ServerConfiguration) GetTransport() string {
	if x != nil {
		return x.Transport
	}
	return ""
}

func (x *ServerConfiguration) GetTransportSettings() *anypb.Any {
	if x != nil {
		return x.TransportSettings
	}
	return nil
}

func (x *ServerConfiguration) GetSecurity() string {
	if x != nil {
		return x.Security
	}
	return ""
}

func (x *ServerConfiguration) GetSecuritySettings() *anypb.Any {
	if x != nil {
		return x.SecuritySettings
	}
	return nil
}

type SubscriptionServerConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Configuration *ServerConfiguration   `protobuf:"bytes,3,opt,name=configuration,proto3" json:"configuration,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubscriptionServerConfig) Reset() {
	*x = SubscriptionServerConfig{}
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscriptionServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionServerConfig) ProtoMessage() {}

func (x *SubscriptionServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionServerConfig.ProtoReflect.Descriptor instead.
func (*SubscriptionServerConfig) Descriptor() ([]byte, []int) {
	return file_app_subscription_specs_abstract_spec_proto_rawDescGZIP(), []int{1}
}

func (x *SubscriptionServerConfig) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SubscriptionServerConfig) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *SubscriptionServerConfig) GetConfiguration() *ServerConfiguration {
	if x != nil {
		return x.Configuration
	}
	return nil
}

type SubscriptionDocument struct {
	state         protoimpl.MessageState      `protogen:"open.v1"`
	Metadata      map[string]string           `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Server        []*SubscriptionServerConfig `protobuf:"bytes,3,rep,name=server,proto3" json:"server,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubscriptionDocument) Reset() {
	*x = SubscriptionDocument{}
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscriptionDocument) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionDocument) ProtoMessage() {}

func (x *SubscriptionDocument) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_specs_abstract_spec_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionDocument.ProtoReflect.Descriptor instead.
func (*SubscriptionDocument) Descriptor() ([]byte, []int) {
	return file_app_subscription_specs_abstract_spec_proto_rawDescGZIP(), []int{2}
}

func (x *SubscriptionDocument) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *SubscriptionDocument) GetServer() []*SubscriptionServerConfig {
	if x != nil {
		return x.Server
	}
	return nil
}

var File_app_subscription_specs_abstract_spec_proto protoreflect.FileDescriptor

const file_app_subscription_specs_abstract_spec_proto_rawDesc = "" +
	"\n" +
	"*app/subscription/specs/abstract_spec.proto\x12!v2ray.core.app.subscription.specs\x1a\x19google/protobuf/any.proto\"\xb6\x02\n" +
	"\x13ServerConfiguration\x12\x1a\n" +
	"\bprotocol\x18\x01 \x01(\tR\bprotocol\x12A\n" +
	"\x11protocol_settings\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x10protocolSettings\x12\x1c\n" +
	"\ttransport\x18\x03 \x01(\tR\ttransport\x12C\n" +
	"\x12transport_settings\x18\x04 \x01(\v2\x14.google.protobuf.AnyR\x11transportSettings\x12\x1a\n" +
	"\bsecurity\x18\x05 \x01(\tR\bsecurity\x12A\n" +
	"\x11security_settings\x18\x06 \x01(\v2\x14.google.protobuf.AnyR\x10securitySettings\"\xac\x02\n" +
	"\x18SubscriptionServerConfig\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12e\n" +
	"\bmetadata\x18\x02 \x03(\v2I.v2ray.core.app.subscription.specs.SubscriptionServerConfig.MetadataEntryR\bmetadata\x12\\\n" +
	"\rconfiguration\x18\x03 \x01(\v26.v2ray.core.app.subscription.specs.ServerConfigurationR\rconfiguration\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x8b\x02\n" +
	"\x14SubscriptionDocument\x12a\n" +
	"\bmetadata\x18\x02 \x03(\v2E.v2ray.core.app.subscription.specs.SubscriptionDocument.MetadataEntryR\bmetadata\x12S\n" +
	"\x06server\x18\x03 \x03(\v2;.v2ray.core.app.subscription.specs.SubscriptionServerConfigR\x06server\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01B\x84\x01\n" +
	"%com.v2ray.core.app.subscription.specsP\x01Z5github.com/v2fly/v2ray-core/v5/app/subscription/specs\xaa\x02!V2Ray.Core.App.Subscription.Specsb\x06proto3"

var (
	file_app_subscription_specs_abstract_spec_proto_rawDescOnce sync.Once
	file_app_subscription_specs_abstract_spec_proto_rawDescData []byte
)

func file_app_subscription_specs_abstract_spec_proto_rawDescGZIP() []byte {
	file_app_subscription_specs_abstract_spec_proto_rawDescOnce.Do(func() {
		file_app_subscription_specs_abstract_spec_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_subscription_specs_abstract_spec_proto_rawDesc), len(file_app_subscription_specs_abstract_spec_proto_rawDesc)))
	})
	return file_app_subscription_specs_abstract_spec_proto_rawDescData
}

var file_app_subscription_specs_abstract_spec_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_app_subscription_specs_abstract_spec_proto_goTypes = []any{
	(*ServerConfiguration)(nil),      // 0: v2ray.core.app.subscription.specs.ServerConfiguration
	(*SubscriptionServerConfig)(nil), // 1: v2ray.core.app.subscription.specs.SubscriptionServerConfig
	(*SubscriptionDocument)(nil),     // 2: v2ray.core.app.subscription.specs.SubscriptionDocument
	nil,                              // 3: v2ray.core.app.subscription.specs.SubscriptionServerConfig.MetadataEntry
	nil,                              // 4: v2ray.core.app.subscription.specs.SubscriptionDocument.MetadataEntry
	(*anypb.Any)(nil),                // 5: google.protobuf.Any
}
var file_app_subscription_specs_abstract_spec_proto_depIdxs = []int32{
	5, // 0: v2ray.core.app.subscription.specs.ServerConfiguration.protocol_settings:type_name -> google.protobuf.Any
	5, // 1: v2ray.core.app.subscription.specs.ServerConfiguration.transport_settings:type_name -> google.protobuf.Any
	5, // 2: v2ray.core.app.subscription.specs.ServerConfiguration.security_settings:type_name -> google.protobuf.Any
	3, // 3: v2ray.core.app.subscription.specs.SubscriptionServerConfig.metadata:type_name -> v2ray.core.app.subscription.specs.SubscriptionServerConfig.MetadataEntry
	0, // 4: v2ray.core.app.subscription.specs.SubscriptionServerConfig.configuration:type_name -> v2ray.core.app.subscription.specs.ServerConfiguration
	4, // 5: v2ray.core.app.subscription.specs.SubscriptionDocument.metadata:type_name -> v2ray.core.app.subscription.specs.SubscriptionDocument.MetadataEntry
	1, // 6: v2ray.core.app.subscription.specs.SubscriptionDocument.server:type_name -> v2ray.core.app.subscription.specs.SubscriptionServerConfig
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_app_subscription_specs_abstract_spec_proto_init() }
func file_app_subscription_specs_abstract_spec_proto_init() {
	if File_app_subscription_specs_abstract_spec_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_subscription_specs_abstract_spec_proto_rawDesc), len(file_app_subscription_specs_abstract_spec_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_subscription_specs_abstract_spec_proto_goTypes,
		DependencyIndexes: file_app_subscription_specs_abstract_spec_proto_depIdxs,
		MessageInfos:      file_app_subscription_specs_abstract_spec_proto_msgTypes,
	}.Build()
	File_app_subscription_specs_abstract_spec_proto = out.File
	file_app_subscription_specs_abstract_spec_proto_goTypes = nil
	file_app_subscription_specs_abstract_spec_proto_depIdxs = nil
}
