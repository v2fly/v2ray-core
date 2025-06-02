package fakedns

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

type FakeDnsPool struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	IpPool        string                 `protobuf:"bytes,1,opt,name=ip_pool,json=ipPool,proto3" json:"ip_pool,omitempty"` //CIDR of IP pool used as fake DNS IP
	LruSize       int64                  `protobuf:"varint,2,opt,name=lruSize,proto3" json:"lruSize,omitempty"`            //Size of Pool for remembering relationship between domain name and IP address
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FakeDnsPool) Reset() {
	*x = FakeDnsPool{}
	mi := &file_app_dns_fakedns_fakedns_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FakeDnsPool) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeDnsPool) ProtoMessage() {}

func (x *FakeDnsPool) ProtoReflect() protoreflect.Message {
	mi := &file_app_dns_fakedns_fakedns_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeDnsPool.ProtoReflect.Descriptor instead.
func (*FakeDnsPool) Descriptor() ([]byte, []int) {
	return file_app_dns_fakedns_fakedns_proto_rawDescGZIP(), []int{0}
}

func (x *FakeDnsPool) GetIpPool() string {
	if x != nil {
		return x.IpPool
	}
	return ""
}

func (x *FakeDnsPool) GetLruSize() int64 {
	if x != nil {
		return x.LruSize
	}
	return 0
}

type FakeDnsPoolMulti struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Pools         []*FakeDnsPool         `protobuf:"bytes,1,rep,name=pools,proto3" json:"pools,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FakeDnsPoolMulti) Reset() {
	*x = FakeDnsPoolMulti{}
	mi := &file_app_dns_fakedns_fakedns_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FakeDnsPoolMulti) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeDnsPoolMulti) ProtoMessage() {}

func (x *FakeDnsPoolMulti) ProtoReflect() protoreflect.Message {
	mi := &file_app_dns_fakedns_fakedns_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeDnsPoolMulti.ProtoReflect.Descriptor instead.
func (*FakeDnsPoolMulti) Descriptor() ([]byte, []int) {
	return file_app_dns_fakedns_fakedns_proto_rawDescGZIP(), []int{1}
}

func (x *FakeDnsPoolMulti) GetPools() []*FakeDnsPool {
	if x != nil {
		return x.Pools
	}
	return nil
}

var File_app_dns_fakedns_fakedns_proto protoreflect.FileDescriptor

const file_app_dns_fakedns_fakedns_proto_rawDesc = "" +
	"\n" +
	"\x1dapp/dns/fakedns/fakedns.proto\x12\x1av2ray.core.app.dns.fakedns\x1a common/protoext/extensions.proto\"X\n" +
	"\vFakeDnsPool\x12\x17\n" +
	"\aip_pool\x18\x01 \x01(\tR\x06ipPool\x12\x18\n" +
	"\alruSize\x18\x02 \x01(\x03R\alruSize:\x16\x82\xb5\x18\x12\n" +
	"\aservice\x12\afakeDns\"n\n" +
	"\x10FakeDnsPoolMulti\x12=\n" +
	"\x05pools\x18\x01 \x03(\v2'.v2ray.core.app.dns.fakedns.FakeDnsPoolR\x05pools:\x1b\x82\xb5\x18\x17\n" +
	"\aservice\x12\ffakeDnsMultiBo\n" +
	"\x1ecom.v2ray.core.app.dns.fakednsP\x01Z.github.com/v2fly/v2ray-core/v5/app/dns/fakedns\xaa\x02\x1aV2Ray.Core.App.Dns.Fakednsb\x06proto3"

var (
	file_app_dns_fakedns_fakedns_proto_rawDescOnce sync.Once
	file_app_dns_fakedns_fakedns_proto_rawDescData []byte
)

func file_app_dns_fakedns_fakedns_proto_rawDescGZIP() []byte {
	file_app_dns_fakedns_fakedns_proto_rawDescOnce.Do(func() {
		file_app_dns_fakedns_fakedns_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_dns_fakedns_fakedns_proto_rawDesc), len(file_app_dns_fakedns_fakedns_proto_rawDesc)))
	})
	return file_app_dns_fakedns_fakedns_proto_rawDescData
}

var file_app_dns_fakedns_fakedns_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_app_dns_fakedns_fakedns_proto_goTypes = []any{
	(*FakeDnsPool)(nil),      // 0: v2ray.core.app.dns.fakedns.FakeDnsPool
	(*FakeDnsPoolMulti)(nil), // 1: v2ray.core.app.dns.fakedns.FakeDnsPoolMulti
}
var file_app_dns_fakedns_fakedns_proto_depIdxs = []int32{
	0, // 0: v2ray.core.app.dns.fakedns.FakeDnsPoolMulti.pools:type_name -> v2ray.core.app.dns.fakedns.FakeDnsPool
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_app_dns_fakedns_fakedns_proto_init() }
func file_app_dns_fakedns_fakedns_proto_init() {
	if File_app_dns_fakedns_fakedns_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_dns_fakedns_fakedns_proto_rawDesc), len(file_app_dns_fakedns_fakedns_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_dns_fakedns_fakedns_proto_goTypes,
		DependencyIndexes: file_app_dns_fakedns_fakedns_proto_depIdxs,
		MessageInfos:      file_app_dns_fakedns_fakedns_proto_msgTypes,
	}.Build()
	File_app_dns_fakedns_fakedns_proto = out.File
	file_app_dns_fakedns_fakedns_proto_goTypes = nil
	file_app_dns_fakedns_fakedns_proto_depIdxs = nil
}
