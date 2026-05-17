package gdocsviewer

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

type Config struct {
	state                     protoimpl.MessageState      `protogen:"open.v1"`
	ViewerUrl                 string                      `protobuf:"bytes,1,opt,name=viewer_url,json=viewerUrl,proto3" json:"viewer_url,omitempty"`
	TextUrl                   string                      `protobuf:"bytes,2,opt,name=text_url,json=textUrl,proto3" json:"text_url,omitempty"`
	OriginUrl                 string                      `protobuf:"bytes,3,opt,name=origin_url,json=originUrl,proto3" json:"origin_url,omitempty"`
	ViewerHostHeader          string                      `protobuf:"bytes,4,opt,name=viewer_host_header,json=viewerHostHeader,proto3" json:"viewer_host_header,omitempty"`
	UserAgent                 string                      `protobuf:"bytes,5,opt,name=user_agent,json=userAgent,proto3" json:"user_agent,omitempty"`
	AllowHttp                 bool                        `protobuf:"varint,6,opt,name=allow_http,json=allowHttp,proto3" json:"allow_http,omitempty"`
	H2PoolSize                int32                       `protobuf:"varint,7,opt,name=h2_pool_size,json=h2PoolSize,proto3" json:"h2_pool_size,omitempty"`
	MaxViewerBodyBytes        int32                       `protobuf:"varint,8,opt,name=max_viewer_body_bytes,json=maxViewerBodyBytes,proto3" json:"max_viewer_body_bytes,omitempty"`
	MinRequestIntervalMs      int32                       `protobuf:"varint,9,opt,name=min_request_interval_ms,json=minRequestIntervalMs,proto3" json:"min_request_interval_ms,omitempty"`
	SharedKey                 []byte                      `protobuf:"bytes,10,opt,name=shared_key,json=sharedKey,proto3" json:"shared_key,omitempty"`
	PathPrefix                string                      `protobuf:"bytes,11,opt,name=path_prefix,json=pathPrefix,proto3" json:"path_prefix,omitempty"`
	MaxRequestBytes           int32                       `protobuf:"varint,12,opt,name=max_request_bytes,json=maxRequestBytes,proto3" json:"max_request_bytes,omitempty"`
	MaxResponseBytes          int32                       `protobuf:"varint,13,opt,name=max_response_bytes,json=maxResponseBytes,proto3" json:"max_response_bytes,omitempty"`
	OriginUrlReplacementRules []*OriginUrlReplacementRule `protobuf:"bytes,14,rep,name=origin_url_replacement_rules,json=originUrlReplacementRules,proto3" json:"origin_url_replacement_rules,omitempty"`
	unknownFields             protoimpl.UnknownFields
	sizeCache                 protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes[0]
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
	return file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetViewerUrl() string {
	if x != nil {
		return x.ViewerUrl
	}
	return ""
}

func (x *Config) GetTextUrl() string {
	if x != nil {
		return x.TextUrl
	}
	return ""
}

func (x *Config) GetOriginUrl() string {
	if x != nil {
		return x.OriginUrl
	}
	return ""
}

func (x *Config) GetViewerHostHeader() string {
	if x != nil {
		return x.ViewerHostHeader
	}
	return ""
}

func (x *Config) GetUserAgent() string {
	if x != nil {
		return x.UserAgent
	}
	return ""
}

func (x *Config) GetAllowHttp() bool {
	if x != nil {
		return x.AllowHttp
	}
	return false
}

func (x *Config) GetH2PoolSize() int32 {
	if x != nil {
		return x.H2PoolSize
	}
	return 0
}

func (x *Config) GetMaxViewerBodyBytes() int32 {
	if x != nil {
		return x.MaxViewerBodyBytes
	}
	return 0
}

func (x *Config) GetMinRequestIntervalMs() int32 {
	if x != nil {
		return x.MinRequestIntervalMs
	}
	return 0
}

func (x *Config) GetSharedKey() []byte {
	if x != nil {
		return x.SharedKey
	}
	return nil
}

func (x *Config) GetPathPrefix() string {
	if x != nil {
		return x.PathPrefix
	}
	return ""
}

func (x *Config) GetMaxRequestBytes() int32 {
	if x != nil {
		return x.MaxRequestBytes
	}
	return 0
}

func (x *Config) GetMaxResponseBytes() int32 {
	if x != nil {
		return x.MaxResponseBytes
	}
	return 0
}

func (x *Config) GetOriginUrlReplacementRules() []*OriginUrlReplacementRule {
	if x != nil {
		return x.OriginUrlReplacementRules
	}
	return nil
}

type OriginUrlReplacementRule struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Pattern       string                 `protobuf:"bytes,2,opt,name=pattern,proto3" json:"pattern,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OriginUrlReplacementRule) Reset() {
	*x = OriginUrlReplacementRule{}
	mi := &file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OriginUrlReplacementRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OriginUrlReplacementRule) ProtoMessage() {}

func (x *OriginUrlReplacementRule) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OriginUrlReplacementRule.ProtoReflect.Descriptor instead.
func (*OriginUrlReplacementRule) Descriptor() ([]byte, []int) {
	return file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescGZIP(), []int{1}
}

func (x *OriginUrlReplacementRule) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *OriginUrlReplacementRule) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

var File_transport_internet_request_stereotype_gdocsviewer_config_proto protoreflect.FileDescriptor

const file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDesc = "" +
	"\n" +
	">transport/internet/request/stereotype/gdocsviewer/config.proto\x12<v2ray.core.transport.internet.request.stereotype.gdocsviewer\x1a common/protoext/extensions.proto\"\xaf\x05\n" +
	"\x06Config\x12\x1d\n" +
	"\n" +
	"viewer_url\x18\x01 \x01(\tR\tviewerUrl\x12\x19\n" +
	"\btext_url\x18\x02 \x01(\tR\atextUrl\x12\x1d\n" +
	"\n" +
	"origin_url\x18\x03 \x01(\tR\toriginUrl\x12,\n" +
	"\x12viewer_host_header\x18\x04 \x01(\tR\x10viewerHostHeader\x12\x1d\n" +
	"\n" +
	"user_agent\x18\x05 \x01(\tR\tuserAgent\x12\x1d\n" +
	"\n" +
	"allow_http\x18\x06 \x01(\bR\tallowHttp\x12 \n" +
	"\fh2_pool_size\x18\a \x01(\x05R\n" +
	"h2PoolSize\x121\n" +
	"\x15max_viewer_body_bytes\x18\b \x01(\x05R\x12maxViewerBodyBytes\x125\n" +
	"\x17min_request_interval_ms\x18\t \x01(\x05R\x14minRequestIntervalMs\x12\x1d\n" +
	"\n" +
	"shared_key\x18\n" +
	" \x01(\fR\tsharedKey\x12\x1f\n" +
	"\vpath_prefix\x18\v \x01(\tR\n" +
	"pathPrefix\x12*\n" +
	"\x11max_request_bytes\x18\f \x01(\x05R\x0fmaxRequestBytes\x12,\n" +
	"\x12max_response_bytes\x18\r \x01(\x05R\x10maxResponseBytes\x12\x97\x01\n" +
	"\x1corigin_url_replacement_rules\x18\x0e \x03(\v2V.v2ray.core.transport.internet.request.stereotype.gdocsviewer.OriginUrlReplacementRuleR\x19originUrlReplacementRules: \x82\xb5\x18\x1c\n" +
	"\ttransport\x12\vgdocsviewer\x90\xff)\x01\"H\n" +
	"\x18OriginUrlReplacementRule\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x18\n" +
	"\apattern\x18\x02 \x01(\tR\apatternB\xd5\x01\n" +
	"@com.v2ray.core.transport.internet.request.stereotype.gdocsviewerP\x01ZPgithub.com/v2fly/v2ray-core/v5/transport/internet/request/stereotype/gdocsviewer\xaa\x02<V2Ray.Core.Transport.Internet.Request.Stereotype.Gdocsviewerb\x06proto3"

var (
	file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescOnce sync.Once
	file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescData []byte
)

func file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescGZIP() []byte {
	file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDesc), len(file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDesc)))
	})
	return file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDescData
}

var file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_request_stereotype_gdocsviewer_config_proto_goTypes = []any{
	(*Config)(nil),                   // 0: v2ray.core.transport.internet.request.stereotype.gdocsviewer.Config
	(*OriginUrlReplacementRule)(nil), // 1: v2ray.core.transport.internet.request.stereotype.gdocsviewer.OriginUrlReplacementRule
}
var file_transport_internet_request_stereotype_gdocsviewer_config_proto_depIdxs = []int32{
	1, // 0: v2ray.core.transport.internet.request.stereotype.gdocsviewer.Config.origin_url_replacement_rules:type_name -> v2ray.core.transport.internet.request.stereotype.gdocsviewer.OriginUrlReplacementRule
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transport_internet_request_stereotype_gdocsviewer_config_proto_init() }
func file_transport_internet_request_stereotype_gdocsviewer_config_proto_init() {
	if File_transport_internet_request_stereotype_gdocsviewer_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDesc), len(file_transport_internet_request_stereotype_gdocsviewer_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_stereotype_gdocsviewer_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_stereotype_gdocsviewer_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_stereotype_gdocsviewer_config_proto_msgTypes,
	}.Build()
	File_transport_internet_request_stereotype_gdocsviewer_config_proto = out.File
	file_transport_internet_request_stereotype_gdocsviewer_config_proto_goTypes = nil
	file_transport_internet_request_stereotype_gdocsviewer_config_proto_depIdxs = nil
}
