package router

import (
	routercommon "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	net "github.com/v2fly/v2ray-core/v5/common/net"
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
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

type DomainStrategy int32

const (
	// Use domain as is.
	DomainStrategy_AsIs DomainStrategy = 0
	// Always resolve IP for domains.
	DomainStrategy_UseIp DomainStrategy = 1
	// Resolve to IP if the domain doesn't match any rules.
	DomainStrategy_IpIfNonMatch DomainStrategy = 2
	// Resolve to IP if any rule requires IP matching.
	DomainStrategy_IpOnDemand DomainStrategy = 3
)

// Enum value maps for DomainStrategy.
var (
	DomainStrategy_name = map[int32]string{
		0: "AsIs",
		1: "UseIp",
		2: "IpIfNonMatch",
		3: "IpOnDemand",
	}
	DomainStrategy_value = map[string]int32{
		"AsIs":         0,
		"UseIp":        1,
		"IpIfNonMatch": 2,
		"IpOnDemand":   3,
	}
)

func (x DomainStrategy) Enum() *DomainStrategy {
	p := new(DomainStrategy)
	*p = x
	return p
}

func (x DomainStrategy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DomainStrategy) Descriptor() protoreflect.EnumDescriptor {
	return file_app_router_config_proto_enumTypes[0].Descriptor()
}

func (DomainStrategy) Type() protoreflect.EnumType {
	return &file_app_router_config_proto_enumTypes[0]
}

func (x DomainStrategy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DomainStrategy.Descriptor instead.
func (DomainStrategy) EnumDescriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{0}
}

type RoutingRule struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to TargetTag:
	//
	//	*RoutingRule_Tag
	//	*RoutingRule_BalancingTag
	TargetTag isRoutingRule_TargetTag `protobuf_oneof:"target_tag"`
	// List of domains for target domain matching.
	Domain []*routercommon.Domain `protobuf:"bytes,2,rep,name=domain,proto3" json:"domain,omitempty"`
	// List of CIDRs for target IP address matching.
	// Deprecated. Use geoip below.
	//
	// Deprecated: Marked as deprecated in app/router/config.proto.
	Cidr []*routercommon.CIDR `protobuf:"bytes,3,rep,name=cidr,proto3" json:"cidr,omitempty"`
	// List of GeoIPs for target IP address matching. If this entry exists, the
	// cidr above will have no effect. GeoIP fields with the same country code are
	// supposed to contain exactly same content. They will be merged during
	// runtime. For customized GeoIPs, please leave country code empty.
	Geoip []*routercommon.GeoIP `protobuf:"bytes,10,rep,name=geoip,proto3" json:"geoip,omitempty"`
	// A range of port [from, to]. If the destination port is in this range, this
	// rule takes effect. Deprecated. Use port_list.
	//
	// Deprecated: Marked as deprecated in app/router/config.proto.
	PortRange *net.PortRange `protobuf:"bytes,4,opt,name=port_range,json=portRange,proto3" json:"port_range,omitempty"`
	// List of ports.
	PortList *net.PortList `protobuf:"bytes,14,opt,name=port_list,json=portList,proto3" json:"port_list,omitempty"`
	// List of networks. Deprecated. Use networks.
	//
	// Deprecated: Marked as deprecated in app/router/config.proto.
	NetworkList *net.NetworkList `protobuf:"bytes,5,opt,name=network_list,json=networkList,proto3" json:"network_list,omitempty"`
	// List of networks for matching.
	Networks []net.Network `protobuf:"varint,13,rep,packed,name=networks,proto3,enum=v2ray.core.common.net.Network" json:"networks,omitempty"`
	// List of CIDRs for source IP address matching.
	//
	// Deprecated: Marked as deprecated in app/router/config.proto.
	SourceCidr []*routercommon.CIDR `protobuf:"bytes,6,rep,name=source_cidr,json=sourceCidr,proto3" json:"source_cidr,omitempty"`
	// List of GeoIPs for source IP address matching. If this entry exists, the
	// source_cidr above will have no effect.
	SourceGeoip []*routercommon.GeoIP `protobuf:"bytes,11,rep,name=source_geoip,json=sourceGeoip,proto3" json:"source_geoip,omitempty"`
	// List of ports for source port matching.
	SourcePortList *net.PortList `protobuf:"bytes,16,opt,name=source_port_list,json=sourcePortList,proto3" json:"source_port_list,omitempty"`
	UserEmail      []string      `protobuf:"bytes,7,rep,name=user_email,json=userEmail,proto3" json:"user_email,omitempty"`
	InboundTag     []string      `protobuf:"bytes,8,rep,name=inbound_tag,json=inboundTag,proto3" json:"inbound_tag,omitempty"`
	Protocol       []string      `protobuf:"bytes,9,rep,name=protocol,proto3" json:"protocol,omitempty"`
	Attributes     string        `protobuf:"bytes,15,opt,name=attributes,proto3" json:"attributes,omitempty"`
	DomainMatcher  string        `protobuf:"bytes,17,opt,name=domain_matcher,json=domainMatcher,proto3" json:"domain_matcher,omitempty"`
	// geo_domain instruct simplified config loader to load geo domain rule and fill in domain field.
	GeoDomain     []*routercommon.GeoSite `protobuf:"bytes,68001,rep,name=geo_domain,json=geoDomain,proto3" json:"geo_domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RoutingRule) Reset() {
	*x = RoutingRule{}
	mi := &file_app_router_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RoutingRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RoutingRule) ProtoMessage() {}

func (x *RoutingRule) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RoutingRule.ProtoReflect.Descriptor instead.
func (*RoutingRule) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{0}
}

func (x *RoutingRule) GetTargetTag() isRoutingRule_TargetTag {
	if x != nil {
		return x.TargetTag
	}
	return nil
}

func (x *RoutingRule) GetTag() string {
	if x != nil {
		if x, ok := x.TargetTag.(*RoutingRule_Tag); ok {
			return x.Tag
		}
	}
	return ""
}

func (x *RoutingRule) GetBalancingTag() string {
	if x != nil {
		if x, ok := x.TargetTag.(*RoutingRule_BalancingTag); ok {
			return x.BalancingTag
		}
	}
	return ""
}

func (x *RoutingRule) GetDomain() []*routercommon.Domain {
	if x != nil {
		return x.Domain
	}
	return nil
}

// Deprecated: Marked as deprecated in app/router/config.proto.
func (x *RoutingRule) GetCidr() []*routercommon.CIDR {
	if x != nil {
		return x.Cidr
	}
	return nil
}

func (x *RoutingRule) GetGeoip() []*routercommon.GeoIP {
	if x != nil {
		return x.Geoip
	}
	return nil
}

// Deprecated: Marked as deprecated in app/router/config.proto.
func (x *RoutingRule) GetPortRange() *net.PortRange {
	if x != nil {
		return x.PortRange
	}
	return nil
}

func (x *RoutingRule) GetPortList() *net.PortList {
	if x != nil {
		return x.PortList
	}
	return nil
}

// Deprecated: Marked as deprecated in app/router/config.proto.
func (x *RoutingRule) GetNetworkList() *net.NetworkList {
	if x != nil {
		return x.NetworkList
	}
	return nil
}

func (x *RoutingRule) GetNetworks() []net.Network {
	if x != nil {
		return x.Networks
	}
	return nil
}

// Deprecated: Marked as deprecated in app/router/config.proto.
func (x *RoutingRule) GetSourceCidr() []*routercommon.CIDR {
	if x != nil {
		return x.SourceCidr
	}
	return nil
}

func (x *RoutingRule) GetSourceGeoip() []*routercommon.GeoIP {
	if x != nil {
		return x.SourceGeoip
	}
	return nil
}

func (x *RoutingRule) GetSourcePortList() *net.PortList {
	if x != nil {
		return x.SourcePortList
	}
	return nil
}

func (x *RoutingRule) GetUserEmail() []string {
	if x != nil {
		return x.UserEmail
	}
	return nil
}

func (x *RoutingRule) GetInboundTag() []string {
	if x != nil {
		return x.InboundTag
	}
	return nil
}

func (x *RoutingRule) GetProtocol() []string {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *RoutingRule) GetAttributes() string {
	if x != nil {
		return x.Attributes
	}
	return ""
}

func (x *RoutingRule) GetDomainMatcher() string {
	if x != nil {
		return x.DomainMatcher
	}
	return ""
}

func (x *RoutingRule) GetGeoDomain() []*routercommon.GeoSite {
	if x != nil {
		return x.GeoDomain
	}
	return nil
}

type isRoutingRule_TargetTag interface {
	isRoutingRule_TargetTag()
}

type RoutingRule_Tag struct {
	// Tag of outbound that this rule is pointing to.
	Tag string `protobuf:"bytes,1,opt,name=tag,proto3,oneof"`
}

type RoutingRule_BalancingTag struct {
	// Tag of routing balancer.
	BalancingTag string `protobuf:"bytes,12,opt,name=balancing_tag,json=balancingTag,proto3,oneof"`
}

func (*RoutingRule_Tag) isRoutingRule_TargetTag() {}

func (*RoutingRule_BalancingTag) isRoutingRule_TargetTag() {}

type BalancingRule struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Tag              string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	OutboundSelector []string               `protobuf:"bytes,2,rep,name=outbound_selector,json=outboundSelector,proto3" json:"outbound_selector,omitempty"`
	Strategy         string                 `protobuf:"bytes,3,opt,name=strategy,proto3" json:"strategy,omitempty"`
	StrategySettings *anypb.Any             `protobuf:"bytes,4,opt,name=strategy_settings,json=strategySettings,proto3" json:"strategy_settings,omitempty"`
	FallbackTag      string                 `protobuf:"bytes,5,opt,name=fallback_tag,json=fallbackTag,proto3" json:"fallback_tag,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *BalancingRule) Reset() {
	*x = BalancingRule{}
	mi := &file_app_router_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BalancingRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BalancingRule) ProtoMessage() {}

func (x *BalancingRule) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BalancingRule.ProtoReflect.Descriptor instead.
func (*BalancingRule) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{1}
}

func (x *BalancingRule) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *BalancingRule) GetOutboundSelector() []string {
	if x != nil {
		return x.OutboundSelector
	}
	return nil
}

func (x *BalancingRule) GetStrategy() string {
	if x != nil {
		return x.Strategy
	}
	return ""
}

func (x *BalancingRule) GetStrategySettings() *anypb.Any {
	if x != nil {
		return x.StrategySettings
	}
	return nil
}

func (x *BalancingRule) GetFallbackTag() string {
	if x != nil {
		return x.FallbackTag
	}
	return ""
}

type StrategyWeight struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Regexp        bool                   `protobuf:"varint,1,opt,name=regexp,proto3" json:"regexp,omitempty"`
	Match         string                 `protobuf:"bytes,2,opt,name=match,proto3" json:"match,omitempty"`
	Value         float32                `protobuf:"fixed32,3,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StrategyWeight) Reset() {
	*x = StrategyWeight{}
	mi := &file_app_router_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StrategyWeight) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StrategyWeight) ProtoMessage() {}

func (x *StrategyWeight) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StrategyWeight.ProtoReflect.Descriptor instead.
func (*StrategyWeight) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{2}
}

func (x *StrategyWeight) GetRegexp() bool {
	if x != nil {
		return x.Regexp
	}
	return false
}

func (x *StrategyWeight) GetMatch() string {
	if x != nil {
		return x.Match
	}
	return ""
}

func (x *StrategyWeight) GetValue() float32 {
	if x != nil {
		return x.Value
	}
	return 0
}

type StrategyRandomConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ObserverTag   string                 `protobuf:"bytes,7,opt,name=observer_tag,json=observerTag,proto3" json:"observer_tag,omitempty"`
	AliveOnly     bool                   `protobuf:"varint,8,opt,name=alive_only,json=aliveOnly,proto3" json:"alive_only,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StrategyRandomConfig) Reset() {
	*x = StrategyRandomConfig{}
	mi := &file_app_router_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StrategyRandomConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StrategyRandomConfig) ProtoMessage() {}

func (x *StrategyRandomConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StrategyRandomConfig.ProtoReflect.Descriptor instead.
func (*StrategyRandomConfig) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{3}
}

func (x *StrategyRandomConfig) GetObserverTag() string {
	if x != nil {
		return x.ObserverTag
	}
	return ""
}

func (x *StrategyRandomConfig) GetAliveOnly() bool {
	if x != nil {
		return x.AliveOnly
	}
	return false
}

type StrategyLeastPingConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ObserverTag   string                 `protobuf:"bytes,7,opt,name=observer_tag,json=observerTag,proto3" json:"observer_tag,omitempty"`
	StickyChoice  bool                   `protobuf:"varint,8,opt,name=sticky_choice,json=stickyChoice,proto3" json:"sticky_choice,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StrategyLeastPingConfig) Reset() {
	*x = StrategyLeastPingConfig{}
	mi := &file_app_router_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StrategyLeastPingConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StrategyLeastPingConfig) ProtoMessage() {}

func (x *StrategyLeastPingConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StrategyLeastPingConfig.ProtoReflect.Descriptor instead.
func (*StrategyLeastPingConfig) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{4}
}

func (x *StrategyLeastPingConfig) GetObserverTag() string {
	if x != nil {
		return x.ObserverTag
	}
	return ""
}

func (x *StrategyLeastPingConfig) GetStickyChoice() bool {
	if x != nil {
		return x.StickyChoice
	}
	return false
}

type StrategyLeastLoadConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// weight settings
	Costs []*StrategyWeight `protobuf:"bytes,2,rep,name=costs,proto3" json:"costs,omitempty"`
	// RTT baselines for selecting, int64 values of time.Duration
	Baselines []int64 `protobuf:"varint,3,rep,packed,name=baselines,proto3" json:"baselines,omitempty"`
	// expected nodes count to select
	Expected int32 `protobuf:"varint,4,opt,name=expected,proto3" json:"expected,omitempty"`
	// max acceptable rtt, filter away high delay nodes. defalut 0
	MaxRTT int64 `protobuf:"varint,5,opt,name=maxRTT,proto3" json:"maxRTT,omitempty"`
	// acceptable failure rate
	Tolerance     float32 `protobuf:"fixed32,6,opt,name=tolerance,proto3" json:"tolerance,omitempty"`
	ObserverTag   string  `protobuf:"bytes,7,opt,name=observer_tag,json=observerTag,proto3" json:"observer_tag,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StrategyLeastLoadConfig) Reset() {
	*x = StrategyLeastLoadConfig{}
	mi := &file_app_router_config_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StrategyLeastLoadConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StrategyLeastLoadConfig) ProtoMessage() {}

func (x *StrategyLeastLoadConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StrategyLeastLoadConfig.ProtoReflect.Descriptor instead.
func (*StrategyLeastLoadConfig) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{5}
}

func (x *StrategyLeastLoadConfig) GetCosts() []*StrategyWeight {
	if x != nil {
		return x.Costs
	}
	return nil
}

func (x *StrategyLeastLoadConfig) GetBaselines() []int64 {
	if x != nil {
		return x.Baselines
	}
	return nil
}

func (x *StrategyLeastLoadConfig) GetExpected() int32 {
	if x != nil {
		return x.Expected
	}
	return 0
}

func (x *StrategyLeastLoadConfig) GetMaxRTT() int64 {
	if x != nil {
		return x.MaxRTT
	}
	return 0
}

func (x *StrategyLeastLoadConfig) GetTolerance() float32 {
	if x != nil {
		return x.Tolerance
	}
	return 0
}

func (x *StrategyLeastLoadConfig) GetObserverTag() string {
	if x != nil {
		return x.ObserverTag
	}
	return ""
}

type Config struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	DomainStrategy DomainStrategy         `protobuf:"varint,1,opt,name=domain_strategy,json=domainStrategy,proto3,enum=v2ray.core.app.router.DomainStrategy" json:"domain_strategy,omitempty"`
	Rule           []*RoutingRule         `protobuf:"bytes,2,rep,name=rule,proto3" json:"rule,omitempty"`
	BalancingRule  []*BalancingRule       `protobuf:"bytes,3,rep,name=balancing_rule,json=balancingRule,proto3" json:"balancing_rule,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_router_config_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[6]
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
	return file_app_router_config_proto_rawDescGZIP(), []int{6}
}

func (x *Config) GetDomainStrategy() DomainStrategy {
	if x != nil {
		return x.DomainStrategy
	}
	return DomainStrategy_AsIs
}

func (x *Config) GetRule() []*RoutingRule {
	if x != nil {
		return x.Rule
	}
	return nil
}

func (x *Config) GetBalancingRule() []*BalancingRule {
	if x != nil {
		return x.BalancingRule
	}
	return nil
}

type SimplifiedRoutingRule struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to TargetTag:
	//
	//	*SimplifiedRoutingRule_Tag
	//	*SimplifiedRoutingRule_BalancingTag
	TargetTag isSimplifiedRoutingRule_TargetTag `protobuf_oneof:"target_tag"`
	// List of domains for target domain matching.
	Domain []*routercommon.Domain `protobuf:"bytes,2,rep,name=domain,proto3" json:"domain,omitempty"`
	// List of GeoIPs for target IP address matching. If this entry exists, the
	// cidr above will have no effect. GeoIP fields with the same country code are
	// supposed to contain exactly same content. They will be merged during
	// runtime. For customized GeoIPs, please leave country code empty.
	Geoip []*routercommon.GeoIP `protobuf:"bytes,10,rep,name=geoip,proto3" json:"geoip,omitempty"`
	// List of ports.
	PortList string `protobuf:"bytes,14,opt,name=port_list,json=portList,proto3" json:"port_list,omitempty"`
	// List of networks for matching.
	Networks *net.NetworkList `protobuf:"bytes,13,opt,name=networks,proto3" json:"networks,omitempty"`
	// List of GeoIPs for source IP address matching. If this entry exists, the
	// source_cidr above will have no effect.
	SourceGeoip []*routercommon.GeoIP `protobuf:"bytes,11,rep,name=source_geoip,json=sourceGeoip,proto3" json:"source_geoip,omitempty"`
	// List of ports for source port matching.
	SourcePortList string   `protobuf:"bytes,16,opt,name=source_port_list,json=sourcePortList,proto3" json:"source_port_list,omitempty"`
	UserEmail      []string `protobuf:"bytes,7,rep,name=user_email,json=userEmail,proto3" json:"user_email,omitempty"`
	InboundTag     []string `protobuf:"bytes,8,rep,name=inbound_tag,json=inboundTag,proto3" json:"inbound_tag,omitempty"`
	Protocol       []string `protobuf:"bytes,9,rep,name=protocol,proto3" json:"protocol,omitempty"`
	Attributes     string   `protobuf:"bytes,15,opt,name=attributes,proto3" json:"attributes,omitempty"`
	DomainMatcher  string   `protobuf:"bytes,17,opt,name=domain_matcher,json=domainMatcher,proto3" json:"domain_matcher,omitempty"`
	// geo_domain instruct simplified config loader to load geo domain rule and fill in domain field.
	GeoDomain     []*routercommon.GeoSite `protobuf:"bytes,68001,rep,name=geo_domain,json=geoDomain,proto3" json:"geo_domain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SimplifiedRoutingRule) Reset() {
	*x = SimplifiedRoutingRule{}
	mi := &file_app_router_config_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedRoutingRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedRoutingRule) ProtoMessage() {}

func (x *SimplifiedRoutingRule) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimplifiedRoutingRule.ProtoReflect.Descriptor instead.
func (*SimplifiedRoutingRule) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{7}
}

func (x *SimplifiedRoutingRule) GetTargetTag() isSimplifiedRoutingRule_TargetTag {
	if x != nil {
		return x.TargetTag
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetTag() string {
	if x != nil {
		if x, ok := x.TargetTag.(*SimplifiedRoutingRule_Tag); ok {
			return x.Tag
		}
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetBalancingTag() string {
	if x != nil {
		if x, ok := x.TargetTag.(*SimplifiedRoutingRule_BalancingTag); ok {
			return x.BalancingTag
		}
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetDomain() []*routercommon.Domain {
	if x != nil {
		return x.Domain
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetGeoip() []*routercommon.GeoIP {
	if x != nil {
		return x.Geoip
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetPortList() string {
	if x != nil {
		return x.PortList
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetNetworks() *net.NetworkList {
	if x != nil {
		return x.Networks
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetSourceGeoip() []*routercommon.GeoIP {
	if x != nil {
		return x.SourceGeoip
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetSourcePortList() string {
	if x != nil {
		return x.SourcePortList
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetUserEmail() []string {
	if x != nil {
		return x.UserEmail
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetInboundTag() []string {
	if x != nil {
		return x.InboundTag
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetProtocol() []string {
	if x != nil {
		return x.Protocol
	}
	return nil
}

func (x *SimplifiedRoutingRule) GetAttributes() string {
	if x != nil {
		return x.Attributes
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetDomainMatcher() string {
	if x != nil {
		return x.DomainMatcher
	}
	return ""
}

func (x *SimplifiedRoutingRule) GetGeoDomain() []*routercommon.GeoSite {
	if x != nil {
		return x.GeoDomain
	}
	return nil
}

type isSimplifiedRoutingRule_TargetTag interface {
	isSimplifiedRoutingRule_TargetTag()
}

type SimplifiedRoutingRule_Tag struct {
	// Tag of outbound that this rule is pointing to.
	Tag string `protobuf:"bytes,1,opt,name=tag,proto3,oneof"`
}

type SimplifiedRoutingRule_BalancingTag struct {
	// Tag of routing balancer.
	BalancingTag string `protobuf:"bytes,12,opt,name=balancing_tag,json=balancingTag,proto3,oneof"`
}

func (*SimplifiedRoutingRule_Tag) isSimplifiedRoutingRule_TargetTag() {}

func (*SimplifiedRoutingRule_BalancingTag) isSimplifiedRoutingRule_TargetTag() {}

type SimplifiedConfig struct {
	state          protoimpl.MessageState   `protogen:"open.v1"`
	DomainStrategy DomainStrategy           `protobuf:"varint,1,opt,name=domain_strategy,json=domainStrategy,proto3,enum=v2ray.core.app.router.DomainStrategy" json:"domain_strategy,omitempty"`
	Rule           []*SimplifiedRoutingRule `protobuf:"bytes,2,rep,name=rule,proto3" json:"rule,omitempty"`
	BalancingRule  []*BalancingRule         `protobuf:"bytes,3,rep,name=balancing_rule,json=balancingRule,proto3" json:"balancing_rule,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SimplifiedConfig) Reset() {
	*x = SimplifiedConfig{}
	mi := &file_app_router_config_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimplifiedConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimplifiedConfig) ProtoMessage() {}

func (x *SimplifiedConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_config_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimplifiedConfig.ProtoReflect.Descriptor instead.
func (*SimplifiedConfig) Descriptor() ([]byte, []int) {
	return file_app_router_config_proto_rawDescGZIP(), []int{8}
}

func (x *SimplifiedConfig) GetDomainStrategy() DomainStrategy {
	if x != nil {
		return x.DomainStrategy
	}
	return DomainStrategy_AsIs
}

func (x *SimplifiedConfig) GetRule() []*SimplifiedRoutingRule {
	if x != nil {
		return x.Rule
	}
	return nil
}

func (x *SimplifiedConfig) GetBalancingRule() []*BalancingRule {
	if x != nil {
		return x.BalancingRule
	}
	return nil
}

var File_app_router_config_proto protoreflect.FileDescriptor

const file_app_router_config_proto_rawDesc = "" +
	"\n" +
	"\x17app/router/config.proto\x12\x15v2ray.core.app.router\x1a\x19google/protobuf/any.proto\x1a\x15common/net/port.proto\x1a\x18common/net/network.proto\x1a common/protoext/extensions.proto\x1a$app/router/routercommon/common.proto\"\x80\b\n" +
	"\vRoutingRule\x12\x12\n" +
	"\x03tag\x18\x01 \x01(\tH\x00R\x03tag\x12%\n" +
	"\rbalancing_tag\x18\f \x01(\tH\x00R\fbalancingTag\x12B\n" +
	"\x06domain\x18\x02 \x03(\v2*.v2ray.core.app.router.routercommon.DomainR\x06domain\x12@\n" +
	"\x04cidr\x18\x03 \x03(\v2(.v2ray.core.app.router.routercommon.CIDRB\x02\x18\x01R\x04cidr\x12?\n" +
	"\x05geoip\x18\n" +
	" \x03(\v2).v2ray.core.app.router.routercommon.GeoIPR\x05geoip\x12C\n" +
	"\n" +
	"port_range\x18\x04 \x01(\v2 .v2ray.core.common.net.PortRangeB\x02\x18\x01R\tportRange\x12<\n" +
	"\tport_list\x18\x0e \x01(\v2\x1f.v2ray.core.common.net.PortListR\bportList\x12I\n" +
	"\fnetwork_list\x18\x05 \x01(\v2\".v2ray.core.common.net.NetworkListB\x02\x18\x01R\vnetworkList\x12:\n" +
	"\bnetworks\x18\r \x03(\x0e2\x1e.v2ray.core.common.net.NetworkR\bnetworks\x12M\n" +
	"\vsource_cidr\x18\x06 \x03(\v2(.v2ray.core.app.router.routercommon.CIDRB\x02\x18\x01R\n" +
	"sourceCidr\x12L\n" +
	"\fsource_geoip\x18\v \x03(\v2).v2ray.core.app.router.routercommon.GeoIPR\vsourceGeoip\x12I\n" +
	"\x10source_port_list\x18\x10 \x01(\v2\x1f.v2ray.core.common.net.PortListR\x0esourcePortList\x12\x1d\n" +
	"\n" +
	"user_email\x18\a \x03(\tR\tuserEmail\x12\x1f\n" +
	"\vinbound_tag\x18\b \x03(\tR\n" +
	"inboundTag\x12\x1a\n" +
	"\bprotocol\x18\t \x03(\tR\bprotocol\x12\x1e\n" +
	"\n" +
	"attributes\x18\x0f \x01(\tR\n" +
	"attributes\x12%\n" +
	"\x0edomain_matcher\x18\x11 \x01(\tR\rdomainMatcher\x12L\n" +
	"\n" +
	"geo_domain\x18\xa1\x93\x04 \x03(\v2+.v2ray.core.app.router.routercommon.GeoSiteR\tgeoDomainB\f\n" +
	"\n" +
	"target_tag\"\xd0\x01\n" +
	"\rBalancingRule\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12+\n" +
	"\x11outbound_selector\x18\x02 \x03(\tR\x10outboundSelector\x12\x1a\n" +
	"\bstrategy\x18\x03 \x01(\tR\bstrategy\x12A\n" +
	"\x11strategy_settings\x18\x04 \x01(\v2\x14.google.protobuf.AnyR\x10strategySettings\x12!\n" +
	"\ffallback_tag\x18\x05 \x01(\tR\vfallbackTag\"T\n" +
	"\x0eStrategyWeight\x12\x16\n" +
	"\x06regexp\x18\x01 \x01(\bR\x06regexp\x12\x14\n" +
	"\x05match\x18\x02 \x01(\tR\x05match\x12\x14\n" +
	"\x05value\x18\x03 \x01(\x02R\x05value\"p\n" +
	"\x14StrategyRandomConfig\x12!\n" +
	"\fobserver_tag\x18\a \x01(\tR\vobserverTag\x12\x1d\n" +
	"\n" +
	"alive_only\x18\b \x01(\bR\taliveOnly:\x16\x82\xb5\x18\x12\n" +
	"\bbalancer\x12\x06random\"|\n" +
	"\x17StrategyLeastPingConfig\x12!\n" +
	"\fobserver_tag\x18\a \x01(\tR\vobserverTag\x12#\n" +
	"\rsticky_choice\x18\b \x01(\bR\fstickyChoice:\x19\x82\xb5\x18\x15\n" +
	"\bbalancer\x12\tleastping\"\x84\x02\n" +
	"\x17StrategyLeastLoadConfig\x12;\n" +
	"\x05costs\x18\x02 \x03(\v2%.v2ray.core.app.router.StrategyWeightR\x05costs\x12\x1c\n" +
	"\tbaselines\x18\x03 \x03(\x03R\tbaselines\x12\x1a\n" +
	"\bexpected\x18\x04 \x01(\x05R\bexpected\x12\x16\n" +
	"\x06maxRTT\x18\x05 \x01(\x03R\x06maxRTT\x12\x1c\n" +
	"\ttolerance\x18\x06 \x01(\x02R\ttolerance\x12!\n" +
	"\fobserver_tag\x18\a \x01(\tR\vobserverTag:\x19\x82\xb5\x18\x15\n" +
	"\bbalancer\x12\tleastload\"\xdd\x01\n" +
	"\x06Config\x12N\n" +
	"\x0fdomain_strategy\x18\x01 \x01(\x0e2%.v2ray.core.app.router.DomainStrategyR\x0edomainStrategy\x126\n" +
	"\x04rule\x18\x02 \x03(\v2\".v2ray.core.app.router.RoutingRuleR\x04rule\x12K\n" +
	"\x0ebalancing_rule\x18\x03 \x03(\v2$.v2ray.core.app.router.BalancingRuleR\rbalancingRule\"\xab\x05\n" +
	"\x15SimplifiedRoutingRule\x12\x12\n" +
	"\x03tag\x18\x01 \x01(\tH\x00R\x03tag\x12%\n" +
	"\rbalancing_tag\x18\f \x01(\tH\x00R\fbalancingTag\x12B\n" +
	"\x06domain\x18\x02 \x03(\v2*.v2ray.core.app.router.routercommon.DomainR\x06domain\x12?\n" +
	"\x05geoip\x18\n" +
	" \x03(\v2).v2ray.core.app.router.routercommon.GeoIPR\x05geoip\x12\x1b\n" +
	"\tport_list\x18\x0e \x01(\tR\bportList\x12>\n" +
	"\bnetworks\x18\r \x01(\v2\".v2ray.core.common.net.NetworkListR\bnetworks\x12L\n" +
	"\fsource_geoip\x18\v \x03(\v2).v2ray.core.app.router.routercommon.GeoIPR\vsourceGeoip\x12(\n" +
	"\x10source_port_list\x18\x10 \x01(\tR\x0esourcePortList\x12\x1d\n" +
	"\n" +
	"user_email\x18\a \x03(\tR\tuserEmail\x12\x1f\n" +
	"\vinbound_tag\x18\b \x03(\tR\n" +
	"inboundTag\x12\x1a\n" +
	"\bprotocol\x18\t \x03(\tR\bprotocol\x12\x1e\n" +
	"\n" +
	"attributes\x18\x0f \x01(\tR\n" +
	"attributes\x12%\n" +
	"\x0edomain_matcher\x18\x11 \x01(\tR\rdomainMatcher\x12L\n" +
	"\n" +
	"geo_domain\x18\xa1\x93\x04 \x03(\v2+.v2ray.core.app.router.routercommon.GeoSiteR\tgeoDomainB\f\n" +
	"\n" +
	"target_tag\"\x88\x02\n" +
	"\x10SimplifiedConfig\x12N\n" +
	"\x0fdomain_strategy\x18\x01 \x01(\x0e2%.v2ray.core.app.router.DomainStrategyR\x0edomainStrategy\x12@\n" +
	"\x04rule\x18\x02 \x03(\v2,.v2ray.core.app.router.SimplifiedRoutingRuleR\x04rule\x12K\n" +
	"\x0ebalancing_rule\x18\x03 \x03(\v2$.v2ray.core.app.router.BalancingRuleR\rbalancingRule:\x15\x82\xb5\x18\x11\n" +
	"\aservice\x12\x06router*G\n" +
	"\x0eDomainStrategy\x12\b\n" +
	"\x04AsIs\x10\x00\x12\t\n" +
	"\x05UseIp\x10\x01\x12\x10\n" +
	"\fIpIfNonMatch\x10\x02\x12\x0e\n" +
	"\n" +
	"IpOnDemand\x10\x03B`\n" +
	"\x19com.v2ray.core.app.routerP\x01Z)github.com/v2fly/v2ray-core/v5/app/router\xaa\x02\x15V2Ray.Core.App.Routerb\x06proto3"

var (
	file_app_router_config_proto_rawDescOnce sync.Once
	file_app_router_config_proto_rawDescData []byte
)

func file_app_router_config_proto_rawDescGZIP() []byte {
	file_app_router_config_proto_rawDescOnce.Do(func() {
		file_app_router_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_router_config_proto_rawDesc), len(file_app_router_config_proto_rawDesc)))
	})
	return file_app_router_config_proto_rawDescData
}

var file_app_router_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_app_router_config_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_app_router_config_proto_goTypes = []any{
	(DomainStrategy)(0),             // 0: v2ray.core.app.router.DomainStrategy
	(*RoutingRule)(nil),             // 1: v2ray.core.app.router.RoutingRule
	(*BalancingRule)(nil),           // 2: v2ray.core.app.router.BalancingRule
	(*StrategyWeight)(nil),          // 3: v2ray.core.app.router.StrategyWeight
	(*StrategyRandomConfig)(nil),    // 4: v2ray.core.app.router.StrategyRandomConfig
	(*StrategyLeastPingConfig)(nil), // 5: v2ray.core.app.router.StrategyLeastPingConfig
	(*StrategyLeastLoadConfig)(nil), // 6: v2ray.core.app.router.StrategyLeastLoadConfig
	(*Config)(nil),                  // 7: v2ray.core.app.router.Config
	(*SimplifiedRoutingRule)(nil),   // 8: v2ray.core.app.router.SimplifiedRoutingRule
	(*SimplifiedConfig)(nil),        // 9: v2ray.core.app.router.SimplifiedConfig
	(*routercommon.Domain)(nil),     // 10: v2ray.core.app.router.routercommon.Domain
	(*routercommon.CIDR)(nil),       // 11: v2ray.core.app.router.routercommon.CIDR
	(*routercommon.GeoIP)(nil),      // 12: v2ray.core.app.router.routercommon.GeoIP
	(*net.PortRange)(nil),           // 13: v2ray.core.common.net.PortRange
	(*net.PortList)(nil),            // 14: v2ray.core.common.net.PortList
	(*net.NetworkList)(nil),         // 15: v2ray.core.common.net.NetworkList
	(net.Network)(0),                // 16: v2ray.core.common.net.Network
	(*routercommon.GeoSite)(nil),    // 17: v2ray.core.app.router.routercommon.GeoSite
	(*anypb.Any)(nil),               // 18: google.protobuf.Any
}
var file_app_router_config_proto_depIdxs = []int32{
	10, // 0: v2ray.core.app.router.RoutingRule.domain:type_name -> v2ray.core.app.router.routercommon.Domain
	11, // 1: v2ray.core.app.router.RoutingRule.cidr:type_name -> v2ray.core.app.router.routercommon.CIDR
	12, // 2: v2ray.core.app.router.RoutingRule.geoip:type_name -> v2ray.core.app.router.routercommon.GeoIP
	13, // 3: v2ray.core.app.router.RoutingRule.port_range:type_name -> v2ray.core.common.net.PortRange
	14, // 4: v2ray.core.app.router.RoutingRule.port_list:type_name -> v2ray.core.common.net.PortList
	15, // 5: v2ray.core.app.router.RoutingRule.network_list:type_name -> v2ray.core.common.net.NetworkList
	16, // 6: v2ray.core.app.router.RoutingRule.networks:type_name -> v2ray.core.common.net.Network
	11, // 7: v2ray.core.app.router.RoutingRule.source_cidr:type_name -> v2ray.core.app.router.routercommon.CIDR
	12, // 8: v2ray.core.app.router.RoutingRule.source_geoip:type_name -> v2ray.core.app.router.routercommon.GeoIP
	14, // 9: v2ray.core.app.router.RoutingRule.source_port_list:type_name -> v2ray.core.common.net.PortList
	17, // 10: v2ray.core.app.router.RoutingRule.geo_domain:type_name -> v2ray.core.app.router.routercommon.GeoSite
	18, // 11: v2ray.core.app.router.BalancingRule.strategy_settings:type_name -> google.protobuf.Any
	3,  // 12: v2ray.core.app.router.StrategyLeastLoadConfig.costs:type_name -> v2ray.core.app.router.StrategyWeight
	0,  // 13: v2ray.core.app.router.Config.domain_strategy:type_name -> v2ray.core.app.router.DomainStrategy
	1,  // 14: v2ray.core.app.router.Config.rule:type_name -> v2ray.core.app.router.RoutingRule
	2,  // 15: v2ray.core.app.router.Config.balancing_rule:type_name -> v2ray.core.app.router.BalancingRule
	10, // 16: v2ray.core.app.router.SimplifiedRoutingRule.domain:type_name -> v2ray.core.app.router.routercommon.Domain
	12, // 17: v2ray.core.app.router.SimplifiedRoutingRule.geoip:type_name -> v2ray.core.app.router.routercommon.GeoIP
	15, // 18: v2ray.core.app.router.SimplifiedRoutingRule.networks:type_name -> v2ray.core.common.net.NetworkList
	12, // 19: v2ray.core.app.router.SimplifiedRoutingRule.source_geoip:type_name -> v2ray.core.app.router.routercommon.GeoIP
	17, // 20: v2ray.core.app.router.SimplifiedRoutingRule.geo_domain:type_name -> v2ray.core.app.router.routercommon.GeoSite
	0,  // 21: v2ray.core.app.router.SimplifiedConfig.domain_strategy:type_name -> v2ray.core.app.router.DomainStrategy
	8,  // 22: v2ray.core.app.router.SimplifiedConfig.rule:type_name -> v2ray.core.app.router.SimplifiedRoutingRule
	2,  // 23: v2ray.core.app.router.SimplifiedConfig.balancing_rule:type_name -> v2ray.core.app.router.BalancingRule
	24, // [24:24] is the sub-list for method output_type
	24, // [24:24] is the sub-list for method input_type
	24, // [24:24] is the sub-list for extension type_name
	24, // [24:24] is the sub-list for extension extendee
	0,  // [0:24] is the sub-list for field type_name
}

func init() { file_app_router_config_proto_init() }
func file_app_router_config_proto_init() {
	if File_app_router_config_proto != nil {
		return
	}
	file_app_router_config_proto_msgTypes[0].OneofWrappers = []any{
		(*RoutingRule_Tag)(nil),
		(*RoutingRule_BalancingTag)(nil),
	}
	file_app_router_config_proto_msgTypes[7].OneofWrappers = []any{
		(*SimplifiedRoutingRule_Tag)(nil),
		(*SimplifiedRoutingRule_BalancingTag)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_router_config_proto_rawDesc), len(file_app_router_config_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_router_config_proto_goTypes,
		DependencyIndexes: file_app_router_config_proto_depIdxs,
		EnumInfos:         file_app_router_config_proto_enumTypes,
		MessageInfos:      file_app_router_config_proto_msgTypes,
	}.Build()
	File_app_router_config_proto = out.File
	file_app_router_config_proto_goTypes = nil
	file_app_router_config_proto_depIdxs = nil
}
