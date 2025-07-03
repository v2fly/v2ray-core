package command

import (
	subscription "github.com/v2fly/v2ray-core/v5/app/subscription"
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

type ListTrackedSubscriptionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListTrackedSubscriptionRequest) Reset() {
	*x = ListTrackedSubscriptionRequest{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListTrackedSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTrackedSubscriptionRequest) ProtoMessage() {}

func (x *ListTrackedSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTrackedSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*ListTrackedSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{0}
}

type ListTrackedSubscriptionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Names         []string               `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListTrackedSubscriptionResponse) Reset() {
	*x = ListTrackedSubscriptionResponse{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListTrackedSubscriptionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTrackedSubscriptionResponse) ProtoMessage() {}

func (x *ListTrackedSubscriptionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTrackedSubscriptionResponse.ProtoReflect.Descriptor instead.
func (*ListTrackedSubscriptionResponse) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{1}
}

func (x *ListTrackedSubscriptionResponse) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

type AddTrackedSubscriptionRequest struct {
	state         protoimpl.MessageState     `protogen:"open.v1"`
	Source        *subscription.ImportSource `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddTrackedSubscriptionRequest) Reset() {
	*x = AddTrackedSubscriptionRequest{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddTrackedSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddTrackedSubscriptionRequest) ProtoMessage() {}

func (x *AddTrackedSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddTrackedSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*AddTrackedSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{2}
}

func (x *AddTrackedSubscriptionRequest) GetSource() *subscription.ImportSource {
	if x != nil {
		return x.Source
	}
	return nil
}

type AddTrackedSubscriptionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddTrackedSubscriptionResponse) Reset() {
	*x = AddTrackedSubscriptionResponse{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddTrackedSubscriptionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddTrackedSubscriptionResponse) ProtoMessage() {}

func (x *AddTrackedSubscriptionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddTrackedSubscriptionResponse.ProtoReflect.Descriptor instead.
func (*AddTrackedSubscriptionResponse) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{3}
}

type RemoveTrackedSubscriptionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveTrackedSubscriptionRequest) Reset() {
	*x = RemoveTrackedSubscriptionRequest{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveTrackedSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveTrackedSubscriptionRequest) ProtoMessage() {}

func (x *RemoveTrackedSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveTrackedSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*RemoveTrackedSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{4}
}

func (x *RemoveTrackedSubscriptionRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type RemoveTrackedSubscriptionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveTrackedSubscriptionResponse) Reset() {
	*x = RemoveTrackedSubscriptionResponse{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveTrackedSubscriptionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveTrackedSubscriptionResponse) ProtoMessage() {}

func (x *RemoveTrackedSubscriptionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveTrackedSubscriptionResponse.ProtoReflect.Descriptor instead.
func (*RemoveTrackedSubscriptionResponse) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{5}
}

type UpdateTrackedSubscriptionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateTrackedSubscriptionRequest) Reset() {
	*x = UpdateTrackedSubscriptionRequest{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateTrackedSubscriptionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTrackedSubscriptionRequest) ProtoMessage() {}

func (x *UpdateTrackedSubscriptionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTrackedSubscriptionRequest.ProtoReflect.Descriptor instead.
func (*UpdateTrackedSubscriptionRequest) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateTrackedSubscriptionRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type UpdateTrackedSubscriptionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateTrackedSubscriptionResponse) Reset() {
	*x = UpdateTrackedSubscriptionResponse{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateTrackedSubscriptionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTrackedSubscriptionResponse) ProtoMessage() {}

func (x *UpdateTrackedSubscriptionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTrackedSubscriptionResponse.ProtoReflect.Descriptor instead.
func (*UpdateTrackedSubscriptionResponse) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{7}
}

type GetTrackedSubscriptionStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTrackedSubscriptionStatusRequest) Reset() {
	*x = GetTrackedSubscriptionStatusRequest{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTrackedSubscriptionStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTrackedSubscriptionStatusRequest) ProtoMessage() {}

func (x *GetTrackedSubscriptionStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTrackedSubscriptionStatusRequest.ProtoReflect.Descriptor instead.
func (*GetTrackedSubscriptionStatusRequest) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{8}
}

func (x *GetTrackedSubscriptionStatusRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GetTrackedSubscriptionStatusResponse struct {
	state         protoimpl.MessageState                  `protogen:"open.v1"`
	Status        *subscription.TrackedSubscriptionStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTrackedSubscriptionStatusResponse) Reset() {
	*x = GetTrackedSubscriptionStatusResponse{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTrackedSubscriptionStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTrackedSubscriptionStatusResponse) ProtoMessage() {}

func (x *GetTrackedSubscriptionStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTrackedSubscriptionStatusResponse.ProtoReflect.Descriptor instead.
func (*GetTrackedSubscriptionStatusResponse) Descriptor() ([]byte, []int) {
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{9}
}

func (x *GetTrackedSubscriptionStatusResponse) GetStatus() *subscription.TrackedSubscriptionStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_subscription_subscriptionmanager_command_command_proto_msgTypes[10]
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
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP(), []int{10}
}

var File_app_subscription_subscriptionmanager_command_command_proto protoreflect.FileDescriptor

const file_app_subscription_subscriptionmanager_command_command_proto_rawDesc = "" +
	"\n" +
	":app/subscription/subscriptionmanager/command/command.proto\x127v2ray.core.app.subscription.subscriptionmanager.command\x1a common/protoext/extensions.proto\x1a\x1dapp/subscription/config.proto\x1a'app/subscription/subscription_rpc.proto\" \n" +
	"\x1eListTrackedSubscriptionRequest\"7\n" +
	"\x1fListTrackedSubscriptionResponse\x12\x14\n" +
	"\x05names\x18\x01 \x03(\tR\x05names\"b\n" +
	"\x1dAddTrackedSubscriptionRequest\x12A\n" +
	"\x06source\x18\x01 \x01(\v2).v2ray.core.app.subscription.ImportSourceR\x06source\" \n" +
	"\x1eAddTrackedSubscriptionResponse\"6\n" +
	" RemoveTrackedSubscriptionRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"#\n" +
	"!RemoveTrackedSubscriptionResponse\"6\n" +
	" UpdateTrackedSubscriptionRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"#\n" +
	"!UpdateTrackedSubscriptionResponse\"9\n" +
	"#GetTrackedSubscriptionStatusRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"v\n" +
	"$GetTrackedSubscriptionStatusResponse\x12N\n" +
	"\x06status\x18\x01 \x01(\v26.v2ray.core.app.subscription.TrackedSubscriptionStatusR\x06status\"0\n" +
	"\x06Config:&\x82\xb5\x18\"\n" +
	"\vgrpcservice\x12\x13subscriptionmanager2\xc9\b\n" +
	"\x1aSubscriptionManagerService\x12\xce\x01\n" +
	"\x17ListTrackedSubscription\x12W.v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionRequest\x1aX.v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionResponse\"\x00\x12\xcb\x01\n" +
	"\x16AddTrackedSubscription\x12V.v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionRequest\x1aW.v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionResponse\"\x00\x12\xd4\x01\n" +
	"\x19RemoveTrackedSubscription\x12Y.v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionRequest\x1aZ.v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionResponse\"\x00\x12\xdd\x01\n" +
	"\x1cGetTrackedSubscriptionStatus\x12\\.v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusRequest\x1a].v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusResponse\"\x00\x12\xd4\x01\n" +
	"\x19UpdateTrackedSubscription\x12Y.v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionRequest\x1aZ.v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionResponse\"\x00B\xc2\x01\n" +
	"7com.v2ray.core.subscription.subscriptionmanager.commandP\x01ZKgithub.com/v2fly/v2ray-core/v5/app/subscription/subscriptionmanager/command\xaa\x027V2Ray.Core.App.Subscription.Subscriptionmanager.Commandb\x06proto3"

var (
	file_app_subscription_subscriptionmanager_command_command_proto_rawDescOnce sync.Once
	file_app_subscription_subscriptionmanager_command_command_proto_rawDescData []byte
)

func file_app_subscription_subscriptionmanager_command_command_proto_rawDescGZIP() []byte {
	file_app_subscription_subscriptionmanager_command_command_proto_rawDescOnce.Do(func() {
		file_app_subscription_subscriptionmanager_command_command_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_subscription_subscriptionmanager_command_command_proto_rawDesc), len(file_app_subscription_subscriptionmanager_command_command_proto_rawDesc)))
	})
	return file_app_subscription_subscriptionmanager_command_command_proto_rawDescData
}

var file_app_subscription_subscriptionmanager_command_command_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_app_subscription_subscriptionmanager_command_command_proto_goTypes = []any{
	(*ListTrackedSubscriptionRequest)(nil),       // 0: v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionRequest
	(*ListTrackedSubscriptionResponse)(nil),      // 1: v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionResponse
	(*AddTrackedSubscriptionRequest)(nil),        // 2: v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionRequest
	(*AddTrackedSubscriptionResponse)(nil),       // 3: v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionResponse
	(*RemoveTrackedSubscriptionRequest)(nil),     // 4: v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionRequest
	(*RemoveTrackedSubscriptionResponse)(nil),    // 5: v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionResponse
	(*UpdateTrackedSubscriptionRequest)(nil),     // 6: v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionRequest
	(*UpdateTrackedSubscriptionResponse)(nil),    // 7: v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionResponse
	(*GetTrackedSubscriptionStatusRequest)(nil),  // 8: v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusRequest
	(*GetTrackedSubscriptionStatusResponse)(nil), // 9: v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusResponse
	(*Config)(nil),                                 // 10: v2ray.core.app.subscription.subscriptionmanager.command.Config
	(*subscription.ImportSource)(nil),              // 11: v2ray.core.app.subscription.ImportSource
	(*subscription.TrackedSubscriptionStatus)(nil), // 12: v2ray.core.app.subscription.TrackedSubscriptionStatus
}
var file_app_subscription_subscriptionmanager_command_command_proto_depIdxs = []int32{
	11, // 0: v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionRequest.source:type_name -> v2ray.core.app.subscription.ImportSource
	12, // 1: v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusResponse.status:type_name -> v2ray.core.app.subscription.TrackedSubscriptionStatus
	0,  // 2: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.ListTrackedSubscription:input_type -> v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionRequest
	2,  // 3: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.AddTrackedSubscription:input_type -> v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionRequest
	4,  // 4: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.RemoveTrackedSubscription:input_type -> v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionRequest
	8,  // 5: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.GetTrackedSubscriptionStatus:input_type -> v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusRequest
	6,  // 6: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.UpdateTrackedSubscription:input_type -> v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionRequest
	1,  // 7: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.ListTrackedSubscription:output_type -> v2ray.core.app.subscription.subscriptionmanager.command.ListTrackedSubscriptionResponse
	3,  // 8: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.AddTrackedSubscription:output_type -> v2ray.core.app.subscription.subscriptionmanager.command.AddTrackedSubscriptionResponse
	5,  // 9: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.RemoveTrackedSubscription:output_type -> v2ray.core.app.subscription.subscriptionmanager.command.RemoveTrackedSubscriptionResponse
	9,  // 10: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.GetTrackedSubscriptionStatus:output_type -> v2ray.core.app.subscription.subscriptionmanager.command.GetTrackedSubscriptionStatusResponse
	7,  // 11: v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService.UpdateTrackedSubscription:output_type -> v2ray.core.app.subscription.subscriptionmanager.command.UpdateTrackedSubscriptionResponse
	7,  // [7:12] is the sub-list for method output_type
	2,  // [2:7] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_app_subscription_subscriptionmanager_command_command_proto_init() }
func file_app_subscription_subscriptionmanager_command_command_proto_init() {
	if File_app_subscription_subscriptionmanager_command_command_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_subscription_subscriptionmanager_command_command_proto_rawDesc), len(file_app_subscription_subscriptionmanager_command_command_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_subscription_subscriptionmanager_command_command_proto_goTypes,
		DependencyIndexes: file_app_subscription_subscriptionmanager_command_command_proto_depIdxs,
		MessageInfos:      file_app_subscription_subscriptionmanager_command_command_proto_msgTypes,
	}.Build()
	File_app_subscription_subscriptionmanager_command_command_proto = out.File
	file_app_subscription_subscriptionmanager_command_command_proto_goTypes = nil
	file_app_subscription_subscriptionmanager_command_command_proto_depIdxs = nil
}
