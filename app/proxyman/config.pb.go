package proxyman

import (
	net "github.com/v2fly/v2ray-core/v5/common/net"
	internet "github.com/v2fly/v2ray-core/v5/transport/internet"
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

type KnownProtocols int32

const (
	KnownProtocols_HTTP KnownProtocols = 0
	KnownProtocols_TLS  KnownProtocols = 1
)

// Enum value maps for KnownProtocols.
var (
	KnownProtocols_name = map[int32]string{
		0: "HTTP",
		1: "TLS",
	}
	KnownProtocols_value = map[string]int32{
		"HTTP": 0,
		"TLS":  1,
	}
)

func (x KnownProtocols) Enum() *KnownProtocols {
	p := new(KnownProtocols)
	*p = x
	return p
}

func (x KnownProtocols) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (KnownProtocols) Descriptor() protoreflect.EnumDescriptor {
	return file_app_proxyman_config_proto_enumTypes[0].Descriptor()
}

func (KnownProtocols) Type() protoreflect.EnumType {
	return &file_app_proxyman_config_proto_enumTypes[0]
}

func (x KnownProtocols) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use KnownProtocols.Descriptor instead.
func (KnownProtocols) EnumDescriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{0}
}

type AllocationStrategy_Type int32

const (
	// Always allocate all connection handlers.
	AllocationStrategy_Always AllocationStrategy_Type = 0
	// Randomly allocate specific range of handlers.
	AllocationStrategy_Random AllocationStrategy_Type = 1
	// External. Not supported yet.
	AllocationStrategy_External AllocationStrategy_Type = 2
)

// Enum value maps for AllocationStrategy_Type.
var (
	AllocationStrategy_Type_name = map[int32]string{
		0: "Always",
		1: "Random",
		2: "External",
	}
	AllocationStrategy_Type_value = map[string]int32{
		"Always":   0,
		"Random":   1,
		"External": 2,
	}
)

func (x AllocationStrategy_Type) Enum() *AllocationStrategy_Type {
	p := new(AllocationStrategy_Type)
	*p = x
	return p
}

func (x AllocationStrategy_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AllocationStrategy_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_app_proxyman_config_proto_enumTypes[1].Descriptor()
}

func (AllocationStrategy_Type) Type() protoreflect.EnumType {
	return &file_app_proxyman_config_proto_enumTypes[1]
}

func (x AllocationStrategy_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AllocationStrategy_Type.Descriptor instead.
func (AllocationStrategy_Type) EnumDescriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{1, 0}
}

type SenderConfig_DomainStrategy int32

const (
	SenderConfig_AS_IS   SenderConfig_DomainStrategy = 0
	SenderConfig_USE_IP  SenderConfig_DomainStrategy = 1
	SenderConfig_USE_IP4 SenderConfig_DomainStrategy = 2
	SenderConfig_USE_IP6 SenderConfig_DomainStrategy = 3
)

// Enum value maps for SenderConfig_DomainStrategy.
var (
	SenderConfig_DomainStrategy_name = map[int32]string{
		0: "AS_IS",
		1: "USE_IP",
		2: "USE_IP4",
		3: "USE_IP6",
	}
	SenderConfig_DomainStrategy_value = map[string]int32{
		"AS_IS":   0,
		"USE_IP":  1,
		"USE_IP4": 2,
		"USE_IP6": 3,
	}
)

func (x SenderConfig_DomainStrategy) Enum() *SenderConfig_DomainStrategy {
	p := new(SenderConfig_DomainStrategy)
	*p = x
	return p
}

func (x SenderConfig_DomainStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SenderConfig_DomainStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_app_proxyman_config_proto_enumTypes[2].Descriptor()
}

func (SenderConfig_DomainStrategy) Type() protoreflect.EnumType {
	return &file_app_proxyman_config_proto_enumTypes[2]
}

func (x SenderConfig_DomainStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SenderConfig_DomainStrategy.Descriptor instead.
func (SenderConfig_DomainStrategy) EnumDescriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{6, 0}
}

type InboundConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InboundConfig) Reset() {
	*x = InboundConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InboundConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundConfig) ProtoMessage() {}

func (x *InboundConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundConfig.ProtoReflect.Descriptor instead.
func (*InboundConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{0}
}

type AllocationStrategy struct {
	state protoimpl.MessageState  `protogen:"open.v1"`
	Type  AllocationStrategy_Type `protobuf:"varint,1,opt,name=type,proto3,enum=v2ray.core.app.proxyman.AllocationStrategy_Type" json:"type,omitempty"`
	// Number of handlers (ports) running in parallel.
	// Default value is 3 if unset.
	Concurrency *AllocationStrategy_AllocationStrategyConcurrency `protobuf:"bytes,2,opt,name=concurrency,proto3" json:"concurrency,omitempty"`
	// Number of minutes before a handler is regenerated.
	// Default value is 5 if unset.
	Refresh       *AllocationStrategy_AllocationStrategyRefresh `protobuf:"bytes,3,opt,name=refresh,proto3" json:"refresh,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AllocationStrategy) Reset() {
	*x = AllocationStrategy{}
	mi := &file_app_proxyman_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AllocationStrategy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllocationStrategy) ProtoMessage() {}

func (x *AllocationStrategy) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllocationStrategy.ProtoReflect.Descriptor instead.
func (*AllocationStrategy) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{1}
}

func (x *AllocationStrategy) GetType() AllocationStrategy_Type {
	if x != nil {
		return x.Type
	}
	return AllocationStrategy_Always
}

func (x *AllocationStrategy) GetConcurrency() *AllocationStrategy_AllocationStrategyConcurrency {
	if x != nil {
		return x.Concurrency
	}
	return nil
}

func (x *AllocationStrategy) GetRefresh() *AllocationStrategy_AllocationStrategyRefresh {
	if x != nil {
		return x.Refresh
	}
	return nil
}

type SniffingConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Whether or not to enable content sniffing on an inbound connection.
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	// Override target destination if sniff'ed protocol is in the given list.
	// Supported values are "http", "tls", "fakedns".
	DestinationOverride []string `protobuf:"bytes,2,rep,name=destination_override,json=destinationOverride,proto3" json:"destination_override,omitempty"`
	// Whether should only try to sniff metadata without waiting for client input.
	// Can be used to support SMTP like protocol where server send the first message.
	MetadataOnly  bool `protobuf:"varint,3,opt,name=metadata_only,json=metadataOnly,proto3" json:"metadata_only,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SniffingConfig) Reset() {
	*x = SniffingConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SniffingConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SniffingConfig) ProtoMessage() {}

func (x *SniffingConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SniffingConfig.ProtoReflect.Descriptor instead.
func (*SniffingConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{2}
}

func (x *SniffingConfig) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *SniffingConfig) GetDestinationOverride() []string {
	if x != nil {
		return x.DestinationOverride
	}
	return nil
}

func (x *SniffingConfig) GetMetadataOnly() bool {
	if x != nil {
		return x.MetadataOnly
	}
	return false
}

type ReceiverConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// PortRange specifies the ports which the Receiver should listen on.
	PortRange *net.PortRange `protobuf:"bytes,1,opt,name=port_range,json=portRange,proto3" json:"port_range,omitempty"`
	// Listen specifies the IP address that the Receiver should listen on.
	Listen                     *net.IPOrDomain        `protobuf:"bytes,2,opt,name=listen,proto3" json:"listen,omitempty"`
	AllocationStrategy         *AllocationStrategy    `protobuf:"bytes,3,opt,name=allocation_strategy,json=allocationStrategy,proto3" json:"allocation_strategy,omitempty"`
	StreamSettings             *internet.StreamConfig `protobuf:"bytes,4,opt,name=stream_settings,json=streamSettings,proto3" json:"stream_settings,omitempty"`
	ReceiveOriginalDestination bool                   `protobuf:"varint,5,opt,name=receive_original_destination,json=receiveOriginalDestination,proto3" json:"receive_original_destination,omitempty"`
	// Override domains for the given protocol.
	// Deprecated. Use sniffing_settings.
	//
	// Deprecated: Marked as deprecated in app/proxyman/config.proto.
	DomainOverride   []KnownProtocols `protobuf:"varint,7,rep,packed,name=domain_override,json=domainOverride,proto3,enum=v2ray.core.app.proxyman.KnownProtocols" json:"domain_override,omitempty"`
	SniffingSettings *SniffingConfig  `protobuf:"bytes,8,opt,name=sniffing_settings,json=sniffingSettings,proto3" json:"sniffing_settings,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *ReceiverConfig) Reset() {
	*x = ReceiverConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReceiverConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReceiverConfig) ProtoMessage() {}

func (x *ReceiverConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReceiverConfig.ProtoReflect.Descriptor instead.
func (*ReceiverConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{3}
}

func (x *ReceiverConfig) GetPortRange() *net.PortRange {
	if x != nil {
		return x.PortRange
	}
	return nil
}

func (x *ReceiverConfig) GetListen() *net.IPOrDomain {
	if x != nil {
		return x.Listen
	}
	return nil
}

func (x *ReceiverConfig) GetAllocationStrategy() *AllocationStrategy {
	if x != nil {
		return x.AllocationStrategy
	}
	return nil
}

func (x *ReceiverConfig) GetStreamSettings() *internet.StreamConfig {
	if x != nil {
		return x.StreamSettings
	}
	return nil
}

func (x *ReceiverConfig) GetReceiveOriginalDestination() bool {
	if x != nil {
		return x.ReceiveOriginalDestination
	}
	return false
}

// Deprecated: Marked as deprecated in app/proxyman/config.proto.
func (x *ReceiverConfig) GetDomainOverride() []KnownProtocols {
	if x != nil {
		return x.DomainOverride
	}
	return nil
}

func (x *ReceiverConfig) GetSniffingSettings() *SniffingConfig {
	if x != nil {
		return x.SniffingSettings
	}
	return nil
}

type InboundHandlerConfig struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Tag              string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	ReceiverSettings *anypb.Any             `protobuf:"bytes,2,opt,name=receiver_settings,json=receiverSettings,proto3" json:"receiver_settings,omitempty"`
	ProxySettings    *anypb.Any             `protobuf:"bytes,3,opt,name=proxy_settings,json=proxySettings,proto3" json:"proxy_settings,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *InboundHandlerConfig) Reset() {
	*x = InboundHandlerConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InboundHandlerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InboundHandlerConfig) ProtoMessage() {}

func (x *InboundHandlerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InboundHandlerConfig.ProtoReflect.Descriptor instead.
func (*InboundHandlerConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{4}
}

func (x *InboundHandlerConfig) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *InboundHandlerConfig) GetReceiverSettings() *anypb.Any {
	if x != nil {
		return x.ReceiverSettings
	}
	return nil
}

func (x *InboundHandlerConfig) GetProxySettings() *anypb.Any {
	if x != nil {
		return x.ProxySettings
	}
	return nil
}

type OutboundConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OutboundConfig) Reset() {
	*x = OutboundConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OutboundConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutboundConfig) ProtoMessage() {}

func (x *OutboundConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutboundConfig.ProtoReflect.Descriptor instead.
func (*OutboundConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{5}
}

type SenderConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Send traffic through the given IP. Only IP is allowed.
	Via               *net.IPOrDomain             `protobuf:"bytes,1,opt,name=via,proto3" json:"via,omitempty"`
	StreamSettings    *internet.StreamConfig      `protobuf:"bytes,2,opt,name=stream_settings,json=streamSettings,proto3" json:"stream_settings,omitempty"`
	ProxySettings     *internet.ProxyConfig       `protobuf:"bytes,3,opt,name=proxy_settings,json=proxySettings,proto3" json:"proxy_settings,omitempty"`
	MultiplexSettings *MultiplexingConfig         `protobuf:"bytes,4,opt,name=multiplex_settings,json=multiplexSettings,proto3" json:"multiplex_settings,omitempty"`
	DomainStrategy    SenderConfig_DomainStrategy `protobuf:"varint,5,opt,name=domain_strategy,json=domainStrategy,proto3,enum=v2ray.core.app.proxyman.SenderConfig_DomainStrategy" json:"domain_strategy,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *SenderConfig) Reset() {
	*x = SenderConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SenderConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SenderConfig) ProtoMessage() {}

func (x *SenderConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SenderConfig.ProtoReflect.Descriptor instead.
func (*SenderConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{6}
}

func (x *SenderConfig) GetVia() *net.IPOrDomain {
	if x != nil {
		return x.Via
	}
	return nil
}

func (x *SenderConfig) GetStreamSettings() *internet.StreamConfig {
	if x != nil {
		return x.StreamSettings
	}
	return nil
}

func (x *SenderConfig) GetProxySettings() *internet.ProxyConfig {
	if x != nil {
		return x.ProxySettings
	}
	return nil
}

func (x *SenderConfig) GetMultiplexSettings() *MultiplexingConfig {
	if x != nil {
		return x.MultiplexSettings
	}
	return nil
}

func (x *SenderConfig) GetDomainStrategy() SenderConfig_DomainStrategy {
	if x != nil {
		return x.DomainStrategy
	}
	return SenderConfig_AS_IS
}

type MultiplexingConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Whether or not Mux is enabled.
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	// Max number of concurrent connections that one Mux connection can handle.
	Concurrency   uint32 `protobuf:"varint,2,opt,name=concurrency,proto3" json:"concurrency,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MultiplexingConfig) Reset() {
	*x = MultiplexingConfig{}
	mi := &file_app_proxyman_config_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultiplexingConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultiplexingConfig) ProtoMessage() {}

func (x *MultiplexingConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultiplexingConfig.ProtoReflect.Descriptor instead.
func (*MultiplexingConfig) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{7}
}

func (x *MultiplexingConfig) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *MultiplexingConfig) GetConcurrency() uint32 {
	if x != nil {
		return x.Concurrency
	}
	return 0
}

type AllocationStrategy_AllocationStrategyConcurrency struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         uint32                 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AllocationStrategy_AllocationStrategyConcurrency) Reset() {
	*x = AllocationStrategy_AllocationStrategyConcurrency{}
	mi := &file_app_proxyman_config_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AllocationStrategy_AllocationStrategyConcurrency) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllocationStrategy_AllocationStrategyConcurrency) ProtoMessage() {}

func (x *AllocationStrategy_AllocationStrategyConcurrency) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllocationStrategy_AllocationStrategyConcurrency.ProtoReflect.Descriptor instead.
func (*AllocationStrategy_AllocationStrategyConcurrency) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{1, 0}
}

func (x *AllocationStrategy_AllocationStrategyConcurrency) GetValue() uint32 {
	if x != nil {
		return x.Value
	}
	return 0
}

type AllocationStrategy_AllocationStrategyRefresh struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         uint32                 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AllocationStrategy_AllocationStrategyRefresh) Reset() {
	*x = AllocationStrategy_AllocationStrategyRefresh{}
	mi := &file_app_proxyman_config_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AllocationStrategy_AllocationStrategyRefresh) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllocationStrategy_AllocationStrategyRefresh) ProtoMessage() {}

func (x *AllocationStrategy_AllocationStrategyRefresh) ProtoReflect() protoreflect.Message {
	mi := &file_app_proxyman_config_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllocationStrategy_AllocationStrategyRefresh.ProtoReflect.Descriptor instead.
func (*AllocationStrategy_AllocationStrategyRefresh) Descriptor() ([]byte, []int) {
	return file_app_proxyman_config_proto_rawDescGZIP(), []int{1, 1}
}

func (x *AllocationStrategy_AllocationStrategyRefresh) GetValue() uint32 {
	if x != nil {
		return x.Value
	}
	return 0
}

var File_app_proxyman_config_proto protoreflect.FileDescriptor

const file_app_proxyman_config_proto_rawDesc = "" +
	"\n" +
	"\x19app/proxyman/config.proto\x12\x17v2ray.core.app.proxyman\x1a\x18common/net/address.proto\x1a\x15common/net/port.proto\x1a\x1ftransport/internet/config.proto\x1a\x19google/protobuf/any.proto\"\x0f\n" +
	"\rInboundConfig\"\xc0\x03\n" +
	"\x12AllocationStrategy\x12D\n" +
	"\x04type\x18\x01 \x01(\x0e20.v2ray.core.app.proxyman.AllocationStrategy.TypeR\x04type\x12k\n" +
	"\vconcurrency\x18\x02 \x01(\v2I.v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyConcurrencyR\vconcurrency\x12_\n" +
	"\arefresh\x18\x03 \x01(\v2E.v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyRefreshR\arefresh\x1a5\n" +
	"\x1dAllocationStrategyConcurrency\x12\x14\n" +
	"\x05value\x18\x01 \x01(\rR\x05value\x1a1\n" +
	"\x19AllocationStrategyRefresh\x12\x14\n" +
	"\x05value\x18\x01 \x01(\rR\x05value\",\n" +
	"\x04Type\x12\n" +
	"\n" +
	"\x06Always\x10\x00\x12\n" +
	"\n" +
	"\x06Random\x10\x01\x12\f\n" +
	"\bExternal\x10\x02\"\x82\x01\n" +
	"\x0eSniffingConfig\x12\x18\n" +
	"\aenabled\x18\x01 \x01(\bR\aenabled\x121\n" +
	"\x14destination_override\x18\x02 \x03(\tR\x13destinationOverride\x12#\n" +
	"\rmetadata_only\x18\x03 \x01(\bR\fmetadataOnly\"\xb4\x04\n" +
	"\x0eReceiverConfig\x12?\n" +
	"\n" +
	"port_range\x18\x01 \x01(\v2 .v2ray.core.common.net.PortRangeR\tportRange\x129\n" +
	"\x06listen\x18\x02 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\x06listen\x12\\\n" +
	"\x13allocation_strategy\x18\x03 \x01(\v2+.v2ray.core.app.proxyman.AllocationStrategyR\x12allocationStrategy\x12T\n" +
	"\x0fstream_settings\x18\x04 \x01(\v2+.v2ray.core.transport.internet.StreamConfigR\x0estreamSettings\x12@\n" +
	"\x1creceive_original_destination\x18\x05 \x01(\bR\x1areceiveOriginalDestination\x12T\n" +
	"\x0fdomain_override\x18\a \x03(\x0e2'.v2ray.core.app.proxyman.KnownProtocolsB\x02\x18\x01R\x0edomainOverride\x12T\n" +
	"\x11sniffing_settings\x18\b \x01(\v2'.v2ray.core.app.proxyman.SniffingConfigR\x10sniffingSettingsJ\x04\b\x06\x10\a\"\xa8\x01\n" +
	"\x14InboundHandlerConfig\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12A\n" +
	"\x11receiver_settings\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x10receiverSettings\x12;\n" +
	"\x0eproxy_settings\x18\x03 \x01(\v2\x14.google.protobuf.AnyR\rproxySettings\"\x10\n" +
	"\x0eOutboundConfig\"\xea\x03\n" +
	"\fSenderConfig\x123\n" +
	"\x03via\x18\x01 \x01(\v2!.v2ray.core.common.net.IPOrDomainR\x03via\x12T\n" +
	"\x0fstream_settings\x18\x02 \x01(\v2+.v2ray.core.transport.internet.StreamConfigR\x0estreamSettings\x12Q\n" +
	"\x0eproxy_settings\x18\x03 \x01(\v2*.v2ray.core.transport.internet.ProxyConfigR\rproxySettings\x12Z\n" +
	"\x12multiplex_settings\x18\x04 \x01(\v2+.v2ray.core.app.proxyman.MultiplexingConfigR\x11multiplexSettings\x12]\n" +
	"\x0fdomain_strategy\x18\x05 \x01(\x0e24.v2ray.core.app.proxyman.SenderConfig.DomainStrategyR\x0edomainStrategy\"A\n" +
	"\x0eDomainStrategy\x12\t\n" +
	"\x05AS_IS\x10\x00\x12\n" +
	"\n" +
	"\x06USE_IP\x10\x01\x12\v\n" +
	"\aUSE_IP4\x10\x02\x12\v\n" +
	"\aUSE_IP6\x10\x03\"P\n" +
	"\x12MultiplexingConfig\x12\x18\n" +
	"\aenabled\x18\x01 \x01(\bR\aenabled\x12 \n" +
	"\vconcurrency\x18\x02 \x01(\rR\vconcurrency*#\n" +
	"\x0eKnownProtocols\x12\b\n" +
	"\x04HTTP\x10\x00\x12\a\n" +
	"\x03TLS\x10\x01Bf\n" +
	"\x1bcom.v2ray.core.app.proxymanP\x01Z+github.com/v2fly/v2ray-core/v5/app/proxyman\xaa\x02\x17V2Ray.Core.App.Proxymanb\x06proto3"

var (
	file_app_proxyman_config_proto_rawDescOnce sync.Once
	file_app_proxyman_config_proto_rawDescData []byte
)

func file_app_proxyman_config_proto_rawDescGZIP() []byte {
	file_app_proxyman_config_proto_rawDescOnce.Do(func() {
		file_app_proxyman_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_proxyman_config_proto_rawDesc), len(file_app_proxyman_config_proto_rawDesc)))
	})
	return file_app_proxyman_config_proto_rawDescData
}

var file_app_proxyman_config_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_app_proxyman_config_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_app_proxyman_config_proto_goTypes = []any{
	(KnownProtocols)(0),                                      // 0: v2ray.core.app.proxyman.KnownProtocols
	(AllocationStrategy_Type)(0),                             // 1: v2ray.core.app.proxyman.AllocationStrategy.Type
	(SenderConfig_DomainStrategy)(0),                         // 2: v2ray.core.app.proxyman.SenderConfig.DomainStrategy
	(*InboundConfig)(nil),                                    // 3: v2ray.core.app.proxyman.InboundConfig
	(*AllocationStrategy)(nil),                               // 4: v2ray.core.app.proxyman.AllocationStrategy
	(*SniffingConfig)(nil),                                   // 5: v2ray.core.app.proxyman.SniffingConfig
	(*ReceiverConfig)(nil),                                   // 6: v2ray.core.app.proxyman.ReceiverConfig
	(*InboundHandlerConfig)(nil),                             // 7: v2ray.core.app.proxyman.InboundHandlerConfig
	(*OutboundConfig)(nil),                                   // 8: v2ray.core.app.proxyman.OutboundConfig
	(*SenderConfig)(nil),                                     // 9: v2ray.core.app.proxyman.SenderConfig
	(*MultiplexingConfig)(nil),                               // 10: v2ray.core.app.proxyman.MultiplexingConfig
	(*AllocationStrategy_AllocationStrategyConcurrency)(nil), // 11: v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyConcurrency
	(*AllocationStrategy_AllocationStrategyRefresh)(nil),     // 12: v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyRefresh
	(*net.PortRange)(nil),                                    // 13: v2ray.core.common.net.PortRange
	(*net.IPOrDomain)(nil),                                   // 14: v2ray.core.common.net.IPOrDomain
	(*internet.StreamConfig)(nil),                            // 15: v2ray.core.transport.internet.StreamConfig
	(*anypb.Any)(nil),                                        // 16: google.protobuf.Any
	(*internet.ProxyConfig)(nil),                             // 17: v2ray.core.transport.internet.ProxyConfig
}
var file_app_proxyman_config_proto_depIdxs = []int32{
	1,  // 0: v2ray.core.app.proxyman.AllocationStrategy.type:type_name -> v2ray.core.app.proxyman.AllocationStrategy.Type
	11, // 1: v2ray.core.app.proxyman.AllocationStrategy.concurrency:type_name -> v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyConcurrency
	12, // 2: v2ray.core.app.proxyman.AllocationStrategy.refresh:type_name -> v2ray.core.app.proxyman.AllocationStrategy.AllocationStrategyRefresh
	13, // 3: v2ray.core.app.proxyman.ReceiverConfig.port_range:type_name -> v2ray.core.common.net.PortRange
	14, // 4: v2ray.core.app.proxyman.ReceiverConfig.listen:type_name -> v2ray.core.common.net.IPOrDomain
	4,  // 5: v2ray.core.app.proxyman.ReceiverConfig.allocation_strategy:type_name -> v2ray.core.app.proxyman.AllocationStrategy
	15, // 6: v2ray.core.app.proxyman.ReceiverConfig.stream_settings:type_name -> v2ray.core.transport.internet.StreamConfig
	0,  // 7: v2ray.core.app.proxyman.ReceiverConfig.domain_override:type_name -> v2ray.core.app.proxyman.KnownProtocols
	5,  // 8: v2ray.core.app.proxyman.ReceiverConfig.sniffing_settings:type_name -> v2ray.core.app.proxyman.SniffingConfig
	16, // 9: v2ray.core.app.proxyman.InboundHandlerConfig.receiver_settings:type_name -> google.protobuf.Any
	16, // 10: v2ray.core.app.proxyman.InboundHandlerConfig.proxy_settings:type_name -> google.protobuf.Any
	14, // 11: v2ray.core.app.proxyman.SenderConfig.via:type_name -> v2ray.core.common.net.IPOrDomain
	15, // 12: v2ray.core.app.proxyman.SenderConfig.stream_settings:type_name -> v2ray.core.transport.internet.StreamConfig
	17, // 13: v2ray.core.app.proxyman.SenderConfig.proxy_settings:type_name -> v2ray.core.transport.internet.ProxyConfig
	10, // 14: v2ray.core.app.proxyman.SenderConfig.multiplex_settings:type_name -> v2ray.core.app.proxyman.MultiplexingConfig
	2,  // 15: v2ray.core.app.proxyman.SenderConfig.domain_strategy:type_name -> v2ray.core.app.proxyman.SenderConfig.DomainStrategy
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_app_proxyman_config_proto_init() }
func file_app_proxyman_config_proto_init() {
	if File_app_proxyman_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_proxyman_config_proto_rawDesc), len(file_app_proxyman_config_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_proxyman_config_proto_goTypes,
		DependencyIndexes: file_app_proxyman_config_proto_depIdxs,
		EnumInfos:         file_app_proxyman_config_proto_enumTypes,
		MessageInfos:      file_app_proxyman_config_proto_msgTypes,
	}.Build()
	File_app_proxyman_config_proto = out.File
	file_app_proxyman_config_proto_goTypes = nil
	file_app_proxyman_config_proto_depIdxs = nil
}
