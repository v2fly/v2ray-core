package subscription

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

type SubscriptionServer struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	ServerMetadata map[string]string      `protobuf:"bytes,2,rep,name=serverMetadata,proto3" json:"serverMetadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Tag            string                 `protobuf:"bytes,3,opt,name=tag,proto3" json:"tag,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SubscriptionServer) Reset() {
	*x = SubscriptionServer{}
	mi := &file_app_subscription_subscription_rpc_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscriptionServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionServer) ProtoMessage() {}

func (x *SubscriptionServer) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscription_rpc_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionServer.ProtoReflect.Descriptor instead.
func (*SubscriptionServer) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscription_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *SubscriptionServer) GetServerMetadata() map[string]string {
	if x != nil {
		return x.ServerMetadata
	}
	return nil
}

func (x *SubscriptionServer) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

type TrackedSubscriptionStatus struct {
	state            protoimpl.MessageState         `protogen:"open.v1"`
	Servers          map[string]*SubscriptionServer `protobuf:"bytes,1,rep,name=servers,proto3" json:"servers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	DocumentMetadata map[string]string              `protobuf:"bytes,2,rep,name=documentMetadata,proto3" json:"documentMetadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ImportSource     *ImportSource                  `protobuf:"bytes,3,opt,name=importSource,proto3" json:"importSource,omitempty"`
	AddedByApi       bool                           `protobuf:"varint,4,opt,name=added_by_api,json=addedByApi,proto3" json:"added_by_api,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *TrackedSubscriptionStatus) Reset() {
	*x = TrackedSubscriptionStatus{}
	mi := &file_app_subscription_subscription_rpc_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TrackedSubscriptionStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackedSubscriptionStatus) ProtoMessage() {}

func (x *TrackedSubscriptionStatus) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscription_rpc_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackedSubscriptionStatus.ProtoReflect.Descriptor instead.
func (*TrackedSubscriptionStatus) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscription_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *TrackedSubscriptionStatus) GetServers() map[string]*SubscriptionServer {
	if x != nil {
		return x.Servers
	}
	return nil
}

func (x *TrackedSubscriptionStatus) GetDocumentMetadata() map[string]string {
	if x != nil {
		return x.DocumentMetadata
	}
	return nil
}

func (x *TrackedSubscriptionStatus) GetImportSource() *ImportSource {
	if x != nil {
		return x.ImportSource
	}
	return nil
}

func (x *TrackedSubscriptionStatus) GetAddedByApi() bool {
	if x != nil {
		return x.AddedByApi
	}
	return false
}

var File_app_subscription_subscription_rpc_proto protoreflect.FileDescriptor

const file_app_subscription_subscription_rpc_proto_rawDesc = "" +
	"\n" +
	"'app/subscription/subscription_rpc.proto\x12\x1bv2ray.core.app.subscription\x1a\x1dapp/subscription/config.proto\"\xd6\x01\n" +
	"\x12SubscriptionServer\x12k\n" +
	"\x0eserverMetadata\x18\x02 \x03(\v2C.v2ray.core.app.subscription.SubscriptionServer.ServerMetadataEntryR\x0eserverMetadata\x12\x10\n" +
	"\x03tag\x18\x03 \x01(\tR\x03tag\x1aA\n" +
	"\x13ServerMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x97\x04\n" +
	"\x19TrackedSubscriptionStatus\x12]\n" +
	"\aservers\x18\x01 \x03(\v2C.v2ray.core.app.subscription.TrackedSubscriptionStatus.ServersEntryR\aservers\x12x\n" +
	"\x10documentMetadata\x18\x02 \x03(\v2L.v2ray.core.app.subscription.TrackedSubscriptionStatus.DocumentMetadataEntryR\x10documentMetadata\x12M\n" +
	"\fimportSource\x18\x03 \x01(\v2).v2ray.core.app.subscription.ImportSourceR\fimportSource\x12 \n" +
	"\fadded_by_api\x18\x04 \x01(\bR\n" +
	"addedByApi\x1ak\n" +
	"\fServersEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12E\n" +
	"\x05value\x18\x02 \x01(\v2/.v2ray.core.app.subscription.SubscriptionServerR\x05value:\x028\x01\x1aC\n" +
	"\x15DocumentMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01Br\n" +
	"\x1fcom.v2ray.core.app.subscriptionP\x01Z/github.com/v2fly/v2ray-core/v5/app/subscription\xaa\x02\x1bV2Ray.Core.App.Subscriptionb\x06proto3"

var (
	file_app_subscription_subscription_rpc_proto_rawDescOnce sync.Once
	file_app_subscription_subscription_rpc_proto_rawDescData []byte
)

func file_app_subscription_subscription_rpc_proto_rawDescGZIP() []byte {
	file_app_subscription_subscription_rpc_proto_rawDescOnce.Do(func() {
		file_app_subscription_subscription_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_subscription_subscription_rpc_proto_rawDesc), len(file_app_subscription_subscription_rpc_proto_rawDesc)))
	})
	return file_app_subscription_subscription_rpc_proto_rawDescData
}

var file_app_subscription_subscription_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_app_subscription_subscription_rpc_proto_goTypes = []any{
	(*SubscriptionServer)(nil),        // 0: v2ray.core.app.subscription.SubscriptionServer
	(*TrackedSubscriptionStatus)(nil), // 1: v2ray.core.app.subscription.TrackedSubscriptionStatus
	nil,                               // 2: v2ray.core.app.subscription.SubscriptionServer.ServerMetadataEntry
	nil,                               // 3: v2ray.core.app.subscription.TrackedSubscriptionStatus.ServersEntry
	nil,                               // 4: v2ray.core.app.subscription.TrackedSubscriptionStatus.DocumentMetadataEntry
	(*ImportSource)(nil),              // 5: v2ray.core.app.subscription.ImportSource
}
var file_app_subscription_subscription_rpc_proto_depIdxs = []int32{
	2, // 0: v2ray.core.app.subscription.SubscriptionServer.serverMetadata:type_name -> v2ray.core.app.subscription.SubscriptionServer.ServerMetadataEntry
	3, // 1: v2ray.core.app.subscription.TrackedSubscriptionStatus.servers:type_name -> v2ray.core.app.subscription.TrackedSubscriptionStatus.ServersEntry
	4, // 2: v2ray.core.app.subscription.TrackedSubscriptionStatus.documentMetadata:type_name -> v2ray.core.app.subscription.TrackedSubscriptionStatus.DocumentMetadataEntry
	5, // 3: v2ray.core.app.subscription.TrackedSubscriptionStatus.importSource:type_name -> v2ray.core.app.subscription.ImportSource
	0, // 4: v2ray.core.app.subscription.TrackedSubscriptionStatus.ServersEntry.value:type_name -> v2ray.core.app.subscription.SubscriptionServer
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_app_subscription_subscription_rpc_proto_init() }
func file_app_subscription_subscription_rpc_proto_init() {
	if File_app_subscription_subscription_rpc_proto != nil {
		return
	}
	file_app_subscription_config_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_subscription_subscription_rpc_proto_rawDesc), len(file_app_subscription_subscription_rpc_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_subscription_subscription_rpc_proto_goTypes,
		DependencyIndexes: file_app_subscription_subscription_rpc_proto_depIdxs,
		MessageInfos:      file_app_subscription_subscription_rpc_proto_msgTypes,
	}.Build()
	File_app_subscription_subscription_rpc_proto = out.File
	file_app_subscription_subscription_rpc_proto_goTypes = nil
	file_app_subscription_subscription_rpc_proto_depIdxs = nil
}
