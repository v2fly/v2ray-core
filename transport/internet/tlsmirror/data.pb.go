package tlsmirror

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

type EnrollmentConfirmationReq struct {
	state                            protoimpl.MessageState `protogen:"open.v1"`
	ServerIdentifier                 []byte                 `protobuf:"bytes,1,opt,name=server_identifier,json=serverIdentifier,proto3" json:"server_identifier,omitempty"`
	CarrierTlsConnectionClientRandom []byte                 `protobuf:"bytes,2,opt,name=carrier_tls_connection_client_random,json=carrierTlsConnectionClientRandom,proto3" json:"carrier_tls_connection_client_random,omitempty"`
	CarrierTlsConnectionServerRandom []byte                 `protobuf:"bytes,3,opt,name=carrier_tls_connection_server_random,json=carrierTlsConnectionServerRandom,proto3" json:"carrier_tls_connection_server_random,omitempty"`
	IsSelfEnrollment                 bool                   `protobuf:"varint,4,opt,name=is_self_enrollment,json=isSelfEnrollment,proto3" json:"is_self_enrollment,omitempty"`
	unknownFields                    protoimpl.UnknownFields
	sizeCache                        protoimpl.SizeCache
}

func (x *EnrollmentConfirmationReq) Reset() {
	*x = EnrollmentConfirmationReq{}
	mi := &file_transport_internet_tlsmirror_data_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EnrollmentConfirmationReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnrollmentConfirmationReq) ProtoMessage() {}

func (x *EnrollmentConfirmationReq) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_data_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnrollmentConfirmationReq.ProtoReflect.Descriptor instead.
func (*EnrollmentConfirmationReq) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_data_proto_rawDescGZIP(), []int{0}
}

func (x *EnrollmentConfirmationReq) GetServerIdentifier() []byte {
	if x != nil {
		return x.ServerIdentifier
	}
	return nil
}

func (x *EnrollmentConfirmationReq) GetCarrierTlsConnectionClientRandom() []byte {
	if x != nil {
		return x.CarrierTlsConnectionClientRandom
	}
	return nil
}

func (x *EnrollmentConfirmationReq) GetCarrierTlsConnectionServerRandom() []byte {
	if x != nil {
		return x.CarrierTlsConnectionServerRandom
	}
	return nil
}

func (x *EnrollmentConfirmationReq) GetIsSelfEnrollment() bool {
	if x != nil {
		return x.IsSelfEnrollment
	}
	return false
}

type EnrollmentConfirmationResp struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Enrolled      bool                   `protobuf:"varint,1,opt,name=enrolled,proto3" json:"enrolled,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EnrollmentConfirmationResp) Reset() {
	*x = EnrollmentConfirmationResp{}
	mi := &file_transport_internet_tlsmirror_data_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EnrollmentConfirmationResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnrollmentConfirmationResp) ProtoMessage() {}

func (x *EnrollmentConfirmationResp) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_data_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnrollmentConfirmationResp.ProtoReflect.Descriptor instead.
func (*EnrollmentConfirmationResp) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_data_proto_rawDescGZIP(), []int{1}
}

func (x *EnrollmentConfirmationResp) GetEnrolled() bool {
	if x != nil {
		return x.Enrolled
	}
	return false
}

var File_transport_internet_tlsmirror_data_proto protoreflect.FileDescriptor

const file_transport_internet_tlsmirror_data_proto_rawDesc = "" +
	"\n" +
	"'transport/internet/tlsmirror/data.proto\x12'v2ray.core.transport.internet.tlsmirror\"\x96\x02\n" +
	"\x19EnrollmentConfirmationReq\x12+\n" +
	"\x11server_identifier\x18\x01 \x01(\fR\x10serverIdentifier\x12N\n" +
	"$carrier_tls_connection_client_random\x18\x02 \x01(\fR carrierTlsConnectionClientRandom\x12N\n" +
	"$carrier_tls_connection_server_random\x18\x03 \x01(\fR carrierTlsConnectionServerRandom\x12,\n" +
	"\x12is_self_enrollment\x18\x04 \x01(\bR\x10isSelfEnrollment\"8\n" +
	"\x1aEnrollmentConfirmationResp\x12\x1a\n" +
	"\benrolled\x18\x01 \x01(\bR\benrolledB\x96\x01\n" +
	"+com.v2ray.core.transport.internet.tlsmirrorP\x01Z;github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror\xaa\x02'V2Ray.Core.Transport.Internet.Tlsmirrorb\x06proto3"

var (
	file_transport_internet_tlsmirror_data_proto_rawDescOnce sync.Once
	file_transport_internet_tlsmirror_data_proto_rawDescData []byte
)

func file_transport_internet_tlsmirror_data_proto_rawDescGZIP() []byte {
	file_transport_internet_tlsmirror_data_proto_rawDescOnce.Do(func() {
		file_transport_internet_tlsmirror_data_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_data_proto_rawDesc), len(file_transport_internet_tlsmirror_data_proto_rawDesc)))
	})
	return file_transport_internet_tlsmirror_data_proto_rawDescData
}

var file_transport_internet_tlsmirror_data_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_tlsmirror_data_proto_goTypes = []any{
	(*EnrollmentConfirmationReq)(nil),  // 0: v2ray.core.transport.internet.tlsmirror.EnrollmentConfirmationReq
	(*EnrollmentConfirmationResp)(nil), // 1: v2ray.core.transport.internet.tlsmirror.EnrollmentConfirmationResp
}
var file_transport_internet_tlsmirror_data_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transport_internet_tlsmirror_data_proto_init() }
func file_transport_internet_tlsmirror_data_proto_init() {
	if File_transport_internet_tlsmirror_data_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_data_proto_rawDesc), len(file_transport_internet_tlsmirror_data_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tlsmirror_data_proto_goTypes,
		DependencyIndexes: file_transport_internet_tlsmirror_data_proto_depIdxs,
		MessageInfos:      file_transport_internet_tlsmirror_data_proto_msgTypes,
	}.Build()
	File_transport_internet_tlsmirror_data_proto = out.File
	file_transport_internet_tlsmirror_data_proto_goTypes = nil
	file_transport_internet_tlsmirror_data_proto_depIdxs = nil
}
