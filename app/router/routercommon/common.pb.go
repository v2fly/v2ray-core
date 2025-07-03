package routercommon

import (
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

// Type of domain value.
type Domain_Type int32

const (
	// The value is used as is.
	Domain_Plain Domain_Type = 0
	// The value is used as a regular expression.
	Domain_Regex Domain_Type = 1
	// The value is a root domain.
	Domain_RootDomain Domain_Type = 2
	// The value is a domain.
	Domain_Full Domain_Type = 3
)

// Enum value maps for Domain_Type.
var (
	Domain_Type_name = map[int32]string{
		0: "Plain",
		1: "Regex",
		2: "RootDomain",
		3: "Full",
	}
	Domain_Type_value = map[string]int32{
		"Plain":      0,
		"Regex":      1,
		"RootDomain": 2,
		"Full":       3,
	}
)

func (x Domain_Type) Enum() *Domain_Type {
	p := new(Domain_Type)
	*p = x
	return p
}

func (x Domain_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Domain_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_app_router_routercommon_common_proto_enumTypes[0].Descriptor()
}

func (Domain_Type) Type() protoreflect.EnumType {
	return &file_app_router_routercommon_common_proto_enumTypes[0]
}

func (x Domain_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Domain_Type.Descriptor instead.
func (Domain_Type) EnumDescriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{0, 0}
}

// Domain for routing decision.
type Domain struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Domain matching type.
	Type Domain_Type `protobuf:"varint,1,opt,name=type,proto3,enum=v2ray.core.app.router.routercommon.Domain_Type" json:"type,omitempty"`
	// Domain value.
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	// Attributes of this domain. May be used for filtering.
	Attribute     []*Domain_Attribute `protobuf:"bytes,3,rep,name=attribute,proto3" json:"attribute,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Domain) Reset() {
	*x = Domain{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Domain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Domain) ProtoMessage() {}

func (x *Domain) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Domain.ProtoReflect.Descriptor instead.
func (*Domain) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{0}
}

func (x *Domain) GetType() Domain_Type {
	if x != nil {
		return x.Type
	}
	return Domain_Plain
}

func (x *Domain) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Domain) GetAttribute() []*Domain_Attribute {
	if x != nil {
		return x.Attribute
	}
	return nil
}

// IP for routing decision, in CIDR form.
type CIDR struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// IP address, should be either 4 or 16 bytes.
	Ip []byte `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	// Number of leading ones in the network mask.
	Prefix        uint32 `protobuf:"varint,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
	IpAddr        string `protobuf:"bytes,68000,opt,name=ip_addr,json=ipAddr,proto3" json:"ip_addr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CIDR) Reset() {
	*x = CIDR{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CIDR) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CIDR) ProtoMessage() {}

func (x *CIDR) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CIDR.ProtoReflect.Descriptor instead.
func (*CIDR) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{1}
}

func (x *CIDR) GetIp() []byte {
	if x != nil {
		return x.Ip
	}
	return nil
}

func (x *CIDR) GetPrefix() uint32 {
	if x != nil {
		return x.Prefix
	}
	return 0
}

func (x *CIDR) GetIpAddr() string {
	if x != nil {
		return x.IpAddr
	}
	return ""
}

type GeoIP struct {
	state        protoimpl.MessageState `protogen:"open.v1"`
	CountryCode  string                 `protobuf:"bytes,1,opt,name=country_code,json=countryCode,proto3" json:"country_code,omitempty"`
	Cidr         []*CIDR                `protobuf:"bytes,2,rep,name=cidr,proto3" json:"cidr,omitempty"`
	InverseMatch bool                   `protobuf:"varint,3,opt,name=inverse_match,json=inverseMatch,proto3" json:"inverse_match,omitempty"`
	// resource_hash instruct simplified config converter to load domain from geo file.
	ResourceHash  []byte `protobuf:"bytes,4,opt,name=resource_hash,json=resourceHash,proto3" json:"resource_hash,omitempty"`
	Code          string `protobuf:"bytes,5,opt,name=code,proto3" json:"code,omitempty"`
	FilePath      string `protobuf:"bytes,68000,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoIP) Reset() {
	*x = GeoIP{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoIP) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIP) ProtoMessage() {}

func (x *GeoIP) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoIP.ProtoReflect.Descriptor instead.
func (*GeoIP) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{2}
}

func (x *GeoIP) GetCountryCode() string {
	if x != nil {
		return x.CountryCode
	}
	return ""
}

func (x *GeoIP) GetCidr() []*CIDR {
	if x != nil {
		return x.Cidr
	}
	return nil
}

func (x *GeoIP) GetInverseMatch() bool {
	if x != nil {
		return x.InverseMatch
	}
	return false
}

func (x *GeoIP) GetResourceHash() []byte {
	if x != nil {
		return x.ResourceHash
	}
	return nil
}

func (x *GeoIP) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *GeoIP) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

type GeoIPList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Entry         []*GeoIP               `protobuf:"bytes,1,rep,name=entry,proto3" json:"entry,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoIPList) Reset() {
	*x = GeoIPList{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoIPList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoIPList) ProtoMessage() {}

func (x *GeoIPList) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoIPList.ProtoReflect.Descriptor instead.
func (*GeoIPList) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{3}
}

func (x *GeoIPList) GetEntry() []*GeoIP {
	if x != nil {
		return x.Entry
	}
	return nil
}

type GeoSite struct {
	state       protoimpl.MessageState `protogen:"open.v1"`
	CountryCode string                 `protobuf:"bytes,1,opt,name=country_code,json=countryCode,proto3" json:"country_code,omitempty"`
	Domain      []*Domain              `protobuf:"bytes,2,rep,name=domain,proto3" json:"domain,omitempty"`
	// resource_hash instruct simplified config converter to load domain from geo file.
	ResourceHash  []byte `protobuf:"bytes,3,opt,name=resource_hash,json=resourceHash,proto3" json:"resource_hash,omitempty"`
	Code          string `protobuf:"bytes,4,opt,name=code,proto3" json:"code,omitempty"`
	FilePath      string `protobuf:"bytes,68000,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoSite) Reset() {
	*x = GeoSite{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoSite) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoSite) ProtoMessage() {}

func (x *GeoSite) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoSite.ProtoReflect.Descriptor instead.
func (*GeoSite) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{4}
}

func (x *GeoSite) GetCountryCode() string {
	if x != nil {
		return x.CountryCode
	}
	return ""
}

func (x *GeoSite) GetDomain() []*Domain {
	if x != nil {
		return x.Domain
	}
	return nil
}

func (x *GeoSite) GetResourceHash() []byte {
	if x != nil {
		return x.ResourceHash
	}
	return nil
}

func (x *GeoSite) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *GeoSite) GetFilePath() string {
	if x != nil {
		return x.FilePath
	}
	return ""
}

type GeoSiteList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Entry         []*GeoSite             `protobuf:"bytes,1,rep,name=entry,proto3" json:"entry,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GeoSiteList) Reset() {
	*x = GeoSiteList{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GeoSiteList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoSiteList) ProtoMessage() {}

func (x *GeoSiteList) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoSiteList.ProtoReflect.Descriptor instead.
func (*GeoSiteList) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{5}
}

func (x *GeoSiteList) GetEntry() []*GeoSite {
	if x != nil {
		return x.Entry
	}
	return nil
}

type Domain_Attribute struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Key   string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Types that are valid to be assigned to TypedValue:
	//
	//	*Domain_Attribute_BoolValue
	//	*Domain_Attribute_IntValue
	TypedValue    isDomain_Attribute_TypedValue `protobuf_oneof:"typed_value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Domain_Attribute) Reset() {
	*x = Domain_Attribute{}
	mi := &file_app_router_routercommon_common_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Domain_Attribute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Domain_Attribute) ProtoMessage() {}

func (x *Domain_Attribute) ProtoReflect() protoreflect.Message {
	mi := &file_app_router_routercommon_common_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Domain_Attribute.ProtoReflect.Descriptor instead.
func (*Domain_Attribute) Descriptor() ([]byte, []int) {
	return file_app_router_routercommon_common_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Domain_Attribute) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Domain_Attribute) GetTypedValue() isDomain_Attribute_TypedValue {
	if x != nil {
		return x.TypedValue
	}
	return nil
}

func (x *Domain_Attribute) GetBoolValue() bool {
	if x != nil {
		if x, ok := x.TypedValue.(*Domain_Attribute_BoolValue); ok {
			return x.BoolValue
		}
	}
	return false
}

func (x *Domain_Attribute) GetIntValue() int64 {
	if x != nil {
		if x, ok := x.TypedValue.(*Domain_Attribute_IntValue); ok {
			return x.IntValue
		}
	}
	return 0
}

type isDomain_Attribute_TypedValue interface {
	isDomain_Attribute_TypedValue()
}

type Domain_Attribute_BoolValue struct {
	BoolValue bool `protobuf:"varint,2,opt,name=bool_value,json=boolValue,proto3,oneof"`
}

type Domain_Attribute_IntValue struct {
	IntValue int64 `protobuf:"varint,3,opt,name=int_value,json=intValue,proto3,oneof"`
}

func (*Domain_Attribute_BoolValue) isDomain_Attribute_TypedValue() {}

func (*Domain_Attribute_IntValue) isDomain_Attribute_TypedValue() {}

var File_app_router_routercommon_common_proto protoreflect.FileDescriptor

const file_app_router_routercommon_common_proto_rawDesc = "" +
	"\n" +
	"$app/router/routercommon/common.proto\x12\"v2ray.core.app.router.routercommon\x1a common/protoext/extensions.proto\"\xdd\x02\n" +
	"\x06Domain\x12C\n" +
	"\x04type\x18\x01 \x01(\x0e2/.v2ray.core.app.router.routercommon.Domain.TypeR\x04type\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\x12R\n" +
	"\tattribute\x18\x03 \x03(\v24.v2ray.core.app.router.routercommon.Domain.AttributeR\tattribute\x1al\n" +
	"\tAttribute\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x1f\n" +
	"\n" +
	"bool_value\x18\x02 \x01(\bH\x00R\tboolValue\x12\x1d\n" +
	"\tint_value\x18\x03 \x01(\x03H\x00R\bintValueB\r\n" +
	"\vtyped_value\"6\n" +
	"\x04Type\x12\t\n" +
	"\x05Plain\x10\x00\x12\t\n" +
	"\x05Regex\x10\x01\x12\x0e\n" +
	"\n" +
	"RootDomain\x10\x02\x12\b\n" +
	"\x04Full\x10\x03\"S\n" +
	"\x04CIDR\x12\x0e\n" +
	"\x02ip\x18\x01 \x01(\fR\x02ip\x12\x16\n" +
	"\x06prefix\x18\x02 \x01(\rR\x06prefix\x12#\n" +
	"\aip_addr\x18\xa0\x93\x04 \x01(\tB\b\x82\xb5\x18\x04:\x02ipR\x06ipAddr\"\xfa\x01\n" +
	"\x05GeoIP\x12!\n" +
	"\fcountry_code\x18\x01 \x01(\tR\vcountryCode\x12<\n" +
	"\x04cidr\x18\x02 \x03(\v2(.v2ray.core.app.router.routercommon.CIDRR\x04cidr\x12#\n" +
	"\rinverse_match\x18\x03 \x01(\bR\finverseMatch\x12#\n" +
	"\rresource_hash\x18\x04 \x01(\fR\fresourceHash\x12\x12\n" +
	"\x04code\x18\x05 \x01(\tR\x04code\x122\n" +
	"\tfile_path\x18\xa0\x93\x04 \x01(\tB\x13\x82\xb5\x18\x0f2\rresource_hashR\bfilePath\"L\n" +
	"\tGeoIPList\x12?\n" +
	"\x05entry\x18\x01 \x03(\v2).v2ray.core.app.router.routercommon.GeoIPR\x05entry\"\xdd\x01\n" +
	"\aGeoSite\x12!\n" +
	"\fcountry_code\x18\x01 \x01(\tR\vcountryCode\x12B\n" +
	"\x06domain\x18\x02 \x03(\v2*.v2ray.core.app.router.routercommon.DomainR\x06domain\x12#\n" +
	"\rresource_hash\x18\x03 \x01(\fR\fresourceHash\x12\x12\n" +
	"\x04code\x18\x04 \x01(\tR\x04code\x122\n" +
	"\tfile_path\x18\xa0\x93\x04 \x01(\tB\x13\x82\xb5\x18\x0f2\rresource_hashR\bfilePath\"P\n" +
	"\vGeoSiteList\x12A\n" +
	"\x05entry\x18\x01 \x03(\v2+.v2ray.core.app.router.routercommon.GeoSiteR\x05entryB\x87\x01\n" +
	"&com.v2ray.core.app.router.routercommonP\x01Z6github.com/v2fly/v2ray-core/v5/app/router/routercommon\xaa\x02\"V2Ray.Core.App.Router.Routercommonb\x06proto3"

var (
	file_app_router_routercommon_common_proto_rawDescOnce sync.Once
	file_app_router_routercommon_common_proto_rawDescData []byte
)

func file_app_router_routercommon_common_proto_rawDescGZIP() []byte {
	file_app_router_routercommon_common_proto_rawDescOnce.Do(func() {
		file_app_router_routercommon_common_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_router_routercommon_common_proto_rawDesc), len(file_app_router_routercommon_common_proto_rawDesc)))
	})
	return file_app_router_routercommon_common_proto_rawDescData
}

var file_app_router_routercommon_common_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_app_router_routercommon_common_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_app_router_routercommon_common_proto_goTypes = []any{
	(Domain_Type)(0),         // 0: v2ray.core.app.router.routercommon.Domain.Type
	(*Domain)(nil),           // 1: v2ray.core.app.router.routercommon.Domain
	(*CIDR)(nil),             // 2: v2ray.core.app.router.routercommon.CIDR
	(*GeoIP)(nil),            // 3: v2ray.core.app.router.routercommon.GeoIP
	(*GeoIPList)(nil),        // 4: v2ray.core.app.router.routercommon.GeoIPList
	(*GeoSite)(nil),          // 5: v2ray.core.app.router.routercommon.GeoSite
	(*GeoSiteList)(nil),      // 6: v2ray.core.app.router.routercommon.GeoSiteList
	(*Domain_Attribute)(nil), // 7: v2ray.core.app.router.routercommon.Domain.Attribute
}
var file_app_router_routercommon_common_proto_depIdxs = []int32{
	0, // 0: v2ray.core.app.router.routercommon.Domain.type:type_name -> v2ray.core.app.router.routercommon.Domain.Type
	7, // 1: v2ray.core.app.router.routercommon.Domain.attribute:type_name -> v2ray.core.app.router.routercommon.Domain.Attribute
	2, // 2: v2ray.core.app.router.routercommon.GeoIP.cidr:type_name -> v2ray.core.app.router.routercommon.CIDR
	3, // 3: v2ray.core.app.router.routercommon.GeoIPList.entry:type_name -> v2ray.core.app.router.routercommon.GeoIP
	1, // 4: v2ray.core.app.router.routercommon.GeoSite.domain:type_name -> v2ray.core.app.router.routercommon.Domain
	5, // 5: v2ray.core.app.router.routercommon.GeoSiteList.entry:type_name -> v2ray.core.app.router.routercommon.GeoSite
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_app_router_routercommon_common_proto_init() }
func file_app_router_routercommon_common_proto_init() {
	if File_app_router_routercommon_common_proto != nil {
		return
	}
	file_app_router_routercommon_common_proto_msgTypes[6].OneofWrappers = []any{
		(*Domain_Attribute_BoolValue)(nil),
		(*Domain_Attribute_IntValue)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_router_routercommon_common_proto_rawDesc), len(file_app_router_routercommon_common_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_router_routercommon_common_proto_goTypes,
		DependencyIndexes: file_app_router_routercommon_common_proto_depIdxs,
		EnumInfos:         file_app_router_routercommon_common_proto_enumTypes,
		MessageInfos:      file_app_router_routercommon_common_proto_msgTypes,
	}.Build()
	File_app_router_routercommon_common_proto = out.File
	file_app_router_routercommon_common_proto_goTypes = nil
	file_app_router_routercommon_common_proto_depIdxs = nil
}
