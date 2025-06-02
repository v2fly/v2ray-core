package testing

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

type TestingMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TestField     string                 `protobuf:"bytes,1,opt,name=test_field,json=testField,proto3" json:"test_field,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TestingMessage) Reset() {
	*x = TestingMessage{}
	mi := &file_common_protoext_testing_test_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TestingMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestingMessage) ProtoMessage() {}

func (x *TestingMessage) ProtoReflect() protoreflect.Message {
	mi := &file_common_protoext_testing_test_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestingMessage.ProtoReflect.Descriptor instead.
func (*TestingMessage) Descriptor() ([]byte, []int) {
	return file_common_protoext_testing_test_proto_rawDescGZIP(), []int{0}
}

func (x *TestingMessage) GetTestField() string {
	if x != nil {
		return x.TestField
	}
	return ""
}

var File_common_protoext_testing_test_proto protoreflect.FileDescriptor

const file_common_protoext_testing_test_proto_rawDesc = "" +
	"\n" +
	"\"common/protoext/testing/test.proto\x12\"v2ray.core.common.protoext.testing\x1a common/protoext/extensions.proto\"U\n" +
	"\x0eTestingMessage\x120\n" +
	"\n" +
	"test_field\x18\x01 \x01(\tB\x11\x82\xb5\x18\r\x12\x04test\x12\x05test2R\ttestField:\x11\x82\xb5\x18\r\n" +
	"\x04demo\n" +
	"\x05demo2B\x84\x01\n" +
	"&com.v2ray.core.common.protoext.testingP\x01Z3github.com/v2fly/v2ray-core/common/protoext/testing\xaa\x02\"V2Ray.Core.Common.ProtoExt.Testingb\x06proto3"

var (
	file_common_protoext_testing_test_proto_rawDescOnce sync.Once
	file_common_protoext_testing_test_proto_rawDescData []byte
)

func file_common_protoext_testing_test_proto_rawDescGZIP() []byte {
	file_common_protoext_testing_test_proto_rawDescOnce.Do(func() {
		file_common_protoext_testing_test_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_common_protoext_testing_test_proto_rawDesc), len(file_common_protoext_testing_test_proto_rawDesc)))
	})
	return file_common_protoext_testing_test_proto_rawDescData
}

var file_common_protoext_testing_test_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_protoext_testing_test_proto_goTypes = []any{
	(*TestingMessage)(nil), // 0: v2ray.core.common.protoext.testing.TestingMessage
}
var file_common_protoext_testing_test_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_protoext_testing_test_proto_init() }
func file_common_protoext_testing_test_proto_init() {
	if File_common_protoext_testing_test_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_common_protoext_testing_test_proto_rawDesc), len(file_common_protoext_testing_test_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_protoext_testing_test_proto_goTypes,
		DependencyIndexes: file_common_protoext_testing_test_proto_depIdxs,
		MessageInfos:      file_common_protoext_testing_test_proto_msgTypes,
	}.Build()
	File_common_protoext_testing_test_proto = out.File
	file_common_protoext_testing_test_proto_goTypes = nil
	file_common_protoext_testing_test_proto_depIdxs = nil
}
