package tlstrafficgen

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

type TimeSpec struct {
	state                              protoimpl.MessageState `protogen:"open.v1"`
	BaseNanoseconds                    uint64                 `protobuf:"varint,1,opt,name=base_nanoseconds,json=baseNanoseconds,proto3" json:"base_nanoseconds,omitempty"`
	UniformRandomMultiplierNanoseconds uint64                 `protobuf:"varint,2,opt,name=uniform_random_multiplier_nanoseconds,json=uniformRandomMultiplierNanoseconds,proto3" json:"uniform_random_multiplier_nanoseconds,omitempty"`
	unknownFields                      protoimpl.UnknownFields
	sizeCache                          protoimpl.SizeCache
}

func (x *TimeSpec) Reset() {
	*x = TimeSpec{}
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TimeSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimeSpec) ProtoMessage() {}

func (x *TimeSpec) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimeSpec.ProtoReflect.Descriptor instead.
func (*TimeSpec) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP(), []int{0}
}

func (x *TimeSpec) GetBaseNanoseconds() uint64 {
	if x != nil {
		return x.BaseNanoseconds
	}
	return 0
}

func (x *TimeSpec) GetUniformRandomMultiplierNanoseconds() uint64 {
	if x != nil {
		return x.UniformRandomMultiplierNanoseconds
	}
	return 0
}

type Header struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Values        []string               `protobuf:"bytes,3,rep,name=values,proto3" json:"values,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Header) Reset() {
	*x = Header{}
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP(), []int{1}
}

func (x *Header) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Header) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type TransferCandidate struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Weight        int32                  `protobuf:"varint,1,opt,name=weight,proto3" json:"weight,omitempty"`
	GotoLocation  int64                  `protobuf:"varint,2,opt,name=goto_location,json=gotoLocation,proto3" json:"goto_location,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TransferCandidate) Reset() {
	*x = TransferCandidate{}
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransferCandidate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferCandidate) ProtoMessage() {}

func (x *TransferCandidate) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransferCandidate.ProtoReflect.Descriptor instead.
func (*TransferCandidate) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP(), []int{2}
}

func (x *TransferCandidate) GetWeight() int32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *TransferCandidate) GetGotoLocation() int64 {
	if x != nil {
		return x.GotoLocation
	}
	return 0
}

type Step struct {
	state                        protoimpl.MessageState `protogen:"open.v1"`
	Name                         string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Host                         string                 `protobuf:"bytes,8,opt,name=host,proto3" json:"host,omitempty"`
	Path                         string                 `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Method                       string                 `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	NextStep                     []*TransferCandidate   `protobuf:"bytes,6,rep,name=next_step,json=nextStep,proto3" json:"next_step,omitempty"`
	ConnectionReady              bool                   `protobuf:"varint,7,opt,name=connection_ready,json=connectionReady,proto3" json:"connection_ready,omitempty"`
	Headers                      []*Header              `protobuf:"bytes,9,rep,name=headers,proto3" json:"headers,omitempty"`
	ConnectionRecallExit         bool                   `protobuf:"varint,10,opt,name=connection_recall_exit,json=connectionRecallExit,proto3" json:"connection_recall_exit,omitempty"`
	WaitTime                     *TimeSpec              `protobuf:"bytes,11,opt,name=wait_time,json=waitTime,proto3" json:"wait_time,omitempty"`
	H2DoNotWaitForDownloadFinish bool                   `protobuf:"varint,12,opt,name=h2_do_not_wait_for_download_finish,json=h2DoNotWaitForDownloadFinish,proto3" json:"h2_do_not_wait_for_download_finish,omitempty"`
	unknownFields                protoimpl.UnknownFields
	sizeCache                    protoimpl.SizeCache
}

func (x *Step) Reset() {
	*x = Step{}
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Step) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Step) ProtoMessage() {}

func (x *Step) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Step.ProtoReflect.Descriptor instead.
func (*Step) Descriptor() ([]byte, []int) {
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP(), []int{3}
}

func (x *Step) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Step) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Step) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Step) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Step) GetNextStep() []*TransferCandidate {
	if x != nil {
		return x.NextStep
	}
	return nil
}

func (x *Step) GetConnectionReady() bool {
	if x != nil {
		return x.ConnectionReady
	}
	return false
}

func (x *Step) GetHeaders() []*Header {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Step) GetConnectionRecallExit() bool {
	if x != nil {
		return x.ConnectionRecallExit
	}
	return false
}

func (x *Step) GetWaitTime() *TimeSpec {
	if x != nil {
		return x.WaitTime
	}
	return nil
}

func (x *Step) GetH2DoNotWaitForDownloadFinish() bool {
	if x != nil {
		return x.H2DoNotWaitForDownloadFinish
	}
	return false
}

type Config struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Steps            []*Step                `protobuf:"bytes,1,rep,name=steps,proto3" json:"steps,omitempty"`
	SecuritySettings *anypb.Any             `protobuf:"bytes,2,opt,name=security_settings,json=securitySettings,proto3" json:"security_settings,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes[4]
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
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP(), []int{4}
}

func (x *Config) GetSteps() []*Step {
	if x != nil {
		return x.Steps
	}
	return nil
}

func (x *Config) GetSecuritySettings() *anypb.Any {
	if x != nil {
		return x.SecuritySettings
	}
	return nil
}

var File_transport_internet_tlsmirror_tlstrafficgen_config_proto protoreflect.FileDescriptor

const file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDesc = "" +
	"\n" +
	"7transport/internet/tlsmirror/tlstrafficgen/config.proto\x125v2ray.core.transport.internet.tlsmirror.tlstrafficgen\x1a\x19google/protobuf/any.proto\"\x88\x01\n" +
	"\bTimeSpec\x12)\n" +
	"\x10base_nanoseconds\x18\x01 \x01(\x04R\x0fbaseNanoseconds\x12Q\n" +
	"%uniform_random_multiplier_nanoseconds\x18\x02 \x01(\x04R\"uniformRandomMultiplierNanoseconds\"J\n" +
	"\x06Header\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\x12\x16\n" +
	"\x06values\x18\x03 \x03(\tR\x06values\"P\n" +
	"\x11TransferCandidate\x12\x16\n" +
	"\x06weight\x18\x01 \x01(\x05R\x06weight\x12#\n" +
	"\rgoto_location\x18\x02 \x01(\x03R\fgotoLocation\"\xa3\x04\n" +
	"\x04Step\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x12\n" +
	"\x04host\x18\b \x01(\tR\x04host\x12\x12\n" +
	"\x04path\x18\x02 \x01(\tR\x04path\x12\x16\n" +
	"\x06method\x18\x03 \x01(\tR\x06method\x12e\n" +
	"\tnext_step\x18\x06 \x03(\v2H.v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TransferCandidateR\bnextStep\x12)\n" +
	"\x10connection_ready\x18\a \x01(\bR\x0fconnectionReady\x12W\n" +
	"\aheaders\x18\t \x03(\v2=.v2ray.core.transport.internet.tlsmirror.tlstrafficgen.HeaderR\aheaders\x124\n" +
	"\x16connection_recall_exit\x18\n" +
	" \x01(\bR\x14connectionRecallExit\x12\\\n" +
	"\twait_time\x18\v \x01(\v2?.v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TimeSpecR\bwaitTime\x12H\n" +
	"\"h2_do_not_wait_for_download_finish\x18\f \x01(\bR\x1ch2DoNotWaitForDownloadFinish\"\x9e\x01\n" +
	"\x06Config\x12Q\n" +
	"\x05steps\x18\x01 \x03(\v2;.v2ray.core.transport.internet.tlsmirror.tlstrafficgen.StepR\x05steps\x12A\n" +
	"\x11security_settings\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x10securitySettingsB\xc0\x01\n" +
	"9com.v2ray.core.transport.internet.tlsmirror.tlstrafficgenP\x01ZIgithub.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/tlstrafficgen\xaa\x025V2Ray.Core.Transport.Internet.Tlsmirror.Tlstrafficgenb\x06proto3"

var (
	file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescOnce sync.Once
	file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescData []byte
)

func file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescGZIP() []byte {
	file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescOnce.Do(func() {
		file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDesc), len(file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDesc)))
	})
	return file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDescData
}

var file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_transport_internet_tlsmirror_tlstrafficgen_config_proto_goTypes = []any{
	(*TimeSpec)(nil),          // 0: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TimeSpec
	(*Header)(nil),            // 1: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Header
	(*TransferCandidate)(nil), // 2: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TransferCandidate
	(*Step)(nil),              // 3: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Step
	(*Config)(nil),            // 4: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Config
	(*anypb.Any)(nil),         // 5: google.protobuf.Any
}
var file_transport_internet_tlsmirror_tlstrafficgen_config_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Step.next_step:type_name -> v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TransferCandidate
	1, // 1: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Step.headers:type_name -> v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Header
	0, // 2: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Step.wait_time:type_name -> v2ray.core.transport.internet.tlsmirror.tlstrafficgen.TimeSpec
	3, // 3: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Config.steps:type_name -> v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Step
	5, // 4: v2ray.core.transport.internet.tlsmirror.tlstrafficgen.Config.security_settings:type_name -> google.protobuf.Any
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_transport_internet_tlsmirror_tlstrafficgen_config_proto_init() }
func file_transport_internet_tlsmirror_tlstrafficgen_config_proto_init() {
	if File_transport_internet_tlsmirror_tlstrafficgen_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDesc), len(file_transport_internet_tlsmirror_tlstrafficgen_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_tlsmirror_tlstrafficgen_config_proto_goTypes,
		DependencyIndexes: file_transport_internet_tlsmirror_tlstrafficgen_config_proto_depIdxs,
		MessageInfos:      file_transport_internet_tlsmirror_tlstrafficgen_config_proto_msgTypes,
	}.Build()
	File_transport_internet_tlsmirror_tlstrafficgen_config_proto = out.File
	file_transport_internet_tlsmirror_tlstrafficgen_config_proto_goTypes = nil
	file_transport_internet_tlsmirror_tlstrafficgen_config_proto_depIdxs = nil
}
