package command

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

type GetStatsRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Name of the stat counter.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Whether or not to reset the counter to fetching its value.
	Reset_        bool `protobuf:"varint,2,opt,name=reset,proto3" json:"reset,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStatsRequest) Reset() {
	*x = GetStatsRequest{}
	mi := &file_app_stats_command_command_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatsRequest) ProtoMessage() {}

func (x *GetStatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatsRequest.ProtoReflect.Descriptor instead.
func (*GetStatsRequest) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{0}
}

func (x *GetStatsRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetStatsRequest) GetReset_() bool {
	if x != nil {
		return x.Reset_
	}
	return false
}

type Stat struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value         int64                  `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Stat) Reset() {
	*x = Stat{}
	mi := &file_app_stats_command_command_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Stat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Stat) ProtoMessage() {}

func (x *Stat) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Stat.ProtoReflect.Descriptor instead.
func (*Stat) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{1}
}

func (x *Stat) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Stat) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type GetStatsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Stat          *Stat                  `protobuf:"bytes,1,opt,name=stat,proto3" json:"stat,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStatsResponse) Reset() {
	*x = GetStatsResponse{}
	mi := &file_app_stats_command_command_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStatsResponse) ProtoMessage() {}

func (x *GetStatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStatsResponse.ProtoReflect.Descriptor instead.
func (*GetStatsResponse) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{2}
}

func (x *GetStatsResponse) GetStat() *Stat {
	if x != nil {
		return x.Stat
	}
	return nil
}

type QueryStatsRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Deprecated, use Patterns instead
	Pattern       string   `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Reset_        bool     `protobuf:"varint,2,opt,name=reset,proto3" json:"reset,omitempty"`
	Patterns      []string `protobuf:"bytes,3,rep,name=patterns,proto3" json:"patterns,omitempty"`
	Regexp        bool     `protobuf:"varint,4,opt,name=regexp,proto3" json:"regexp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryStatsRequest) Reset() {
	*x = QueryStatsRequest{}
	mi := &file_app_stats_command_command_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryStatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryStatsRequest) ProtoMessage() {}

func (x *QueryStatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryStatsRequest.ProtoReflect.Descriptor instead.
func (*QueryStatsRequest) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{3}
}

func (x *QueryStatsRequest) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *QueryStatsRequest) GetReset_() bool {
	if x != nil {
		return x.Reset_
	}
	return false
}

func (x *QueryStatsRequest) GetPatterns() []string {
	if x != nil {
		return x.Patterns
	}
	return nil
}

func (x *QueryStatsRequest) GetRegexp() bool {
	if x != nil {
		return x.Regexp
	}
	return false
}

type QueryStatsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Stat          []*Stat                `protobuf:"bytes,1,rep,name=stat,proto3" json:"stat,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryStatsResponse) Reset() {
	*x = QueryStatsResponse{}
	mi := &file_app_stats_command_command_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryStatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryStatsResponse) ProtoMessage() {}

func (x *QueryStatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryStatsResponse.ProtoReflect.Descriptor instead.
func (*QueryStatsResponse) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{4}
}

func (x *QueryStatsResponse) GetStat() []*Stat {
	if x != nil {
		return x.Stat
	}
	return nil
}

type SysStatsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SysStatsRequest) Reset() {
	*x = SysStatsRequest{}
	mi := &file_app_stats_command_command_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SysStatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SysStatsRequest) ProtoMessage() {}

func (x *SysStatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SysStatsRequest.ProtoReflect.Descriptor instead.
func (*SysStatsRequest) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{5}
}

type SysStatsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NumGoroutine  uint32                 `protobuf:"varint,1,opt,name=NumGoroutine,proto3" json:"NumGoroutine,omitempty"`
	NumGC         uint32                 `protobuf:"varint,2,opt,name=NumGC,proto3" json:"NumGC,omitempty"`
	Alloc         uint64                 `protobuf:"varint,3,opt,name=Alloc,proto3" json:"Alloc,omitempty"`
	TotalAlloc    uint64                 `protobuf:"varint,4,opt,name=TotalAlloc,proto3" json:"TotalAlloc,omitempty"`
	Sys           uint64                 `protobuf:"varint,5,opt,name=Sys,proto3" json:"Sys,omitempty"`
	Mallocs       uint64                 `protobuf:"varint,6,opt,name=Mallocs,proto3" json:"Mallocs,omitempty"`
	Frees         uint64                 `protobuf:"varint,7,opt,name=Frees,proto3" json:"Frees,omitempty"`
	LiveObjects   uint64                 `protobuf:"varint,8,opt,name=LiveObjects,proto3" json:"LiveObjects,omitempty"`
	PauseTotalNs  uint64                 `protobuf:"varint,9,opt,name=PauseTotalNs,proto3" json:"PauseTotalNs,omitempty"`
	Uptime        uint32                 `protobuf:"varint,10,opt,name=Uptime,proto3" json:"Uptime,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SysStatsResponse) Reset() {
	*x = SysStatsResponse{}
	mi := &file_app_stats_command_command_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SysStatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SysStatsResponse) ProtoMessage() {}

func (x *SysStatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SysStatsResponse.ProtoReflect.Descriptor instead.
func (*SysStatsResponse) Descriptor() ([]byte, []int) {
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{6}
}

func (x *SysStatsResponse) GetNumGoroutine() uint32 {
	if x != nil {
		return x.NumGoroutine
	}
	return 0
}

func (x *SysStatsResponse) GetNumGC() uint32 {
	if x != nil {
		return x.NumGC
	}
	return 0
}

func (x *SysStatsResponse) GetAlloc() uint64 {
	if x != nil {
		return x.Alloc
	}
	return 0
}

func (x *SysStatsResponse) GetTotalAlloc() uint64 {
	if x != nil {
		return x.TotalAlloc
	}
	return 0
}

func (x *SysStatsResponse) GetSys() uint64 {
	if x != nil {
		return x.Sys
	}
	return 0
}

func (x *SysStatsResponse) GetMallocs() uint64 {
	if x != nil {
		return x.Mallocs
	}
	return 0
}

func (x *SysStatsResponse) GetFrees() uint64 {
	if x != nil {
		return x.Frees
	}
	return 0
}

func (x *SysStatsResponse) GetLiveObjects() uint64 {
	if x != nil {
		return x.LiveObjects
	}
	return 0
}

func (x *SysStatsResponse) GetPauseTotalNs() uint64 {
	if x != nil {
		return x.PauseTotalNs
	}
	return 0
}

func (x *SysStatsResponse) GetUptime() uint32 {
	if x != nil {
		return x.Uptime
	}
	return 0
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_stats_command_command_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_stats_command_command_proto_msgTypes[7]
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
	return file_app_stats_command_command_proto_rawDescGZIP(), []int{7}
}

var File_app_stats_command_command_proto protoreflect.FileDescriptor

const file_app_stats_command_command_proto_rawDesc = "" +
	"\n" +
	"\x1fapp/stats/command/command.proto\x12\x1cv2ray.core.app.stats.command\x1a common/protoext/extensions.proto\";\n" +
	"\x0fGetStatsRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x14\n" +
	"\x05reset\x18\x02 \x01(\bR\x05reset\"0\n" +
	"\x04Stat\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x14\n" +
	"\x05value\x18\x02 \x01(\x03R\x05value\"J\n" +
	"\x10GetStatsResponse\x126\n" +
	"\x04stat\x18\x01 \x01(\v2\".v2ray.core.app.stats.command.StatR\x04stat\"w\n" +
	"\x11QueryStatsRequest\x12\x18\n" +
	"\apattern\x18\x01 \x01(\tR\apattern\x12\x14\n" +
	"\x05reset\x18\x02 \x01(\bR\x05reset\x12\x1a\n" +
	"\bpatterns\x18\x03 \x03(\tR\bpatterns\x12\x16\n" +
	"\x06regexp\x18\x04 \x01(\bR\x06regexp\"L\n" +
	"\x12QueryStatsResponse\x126\n" +
	"\x04stat\x18\x01 \x03(\v2\".v2ray.core.app.stats.command.StatR\x04stat\"\x11\n" +
	"\x0fSysStatsRequest\"\xa2\x02\n" +
	"\x10SysStatsResponse\x12\"\n" +
	"\fNumGoroutine\x18\x01 \x01(\rR\fNumGoroutine\x12\x14\n" +
	"\x05NumGC\x18\x02 \x01(\rR\x05NumGC\x12\x14\n" +
	"\x05Alloc\x18\x03 \x01(\x04R\x05Alloc\x12\x1e\n" +
	"\n" +
	"TotalAlloc\x18\x04 \x01(\x04R\n" +
	"TotalAlloc\x12\x10\n" +
	"\x03Sys\x18\x05 \x01(\x04R\x03Sys\x12\x18\n" +
	"\aMallocs\x18\x06 \x01(\x04R\aMallocs\x12\x14\n" +
	"\x05Frees\x18\a \x01(\x04R\x05Frees\x12 \n" +
	"\vLiveObjects\x18\b \x01(\x04R\vLiveObjects\x12\"\n" +
	"\fPauseTotalNs\x18\t \x01(\x04R\fPauseTotalNs\x12\x16\n" +
	"\x06Uptime\x18\n" +
	" \x01(\rR\x06Uptime\"\"\n" +
	"\x06Config:\x18\x82\xb5\x18\x14\n" +
	"\vgrpcservice\x12\x05stats2\xde\x02\n" +
	"\fStatsService\x12k\n" +
	"\bGetStats\x12-.v2ray.core.app.stats.command.GetStatsRequest\x1a..v2ray.core.app.stats.command.GetStatsResponse\"\x00\x12q\n" +
	"\n" +
	"QueryStats\x12/.v2ray.core.app.stats.command.QueryStatsRequest\x1a0.v2ray.core.app.stats.command.QueryStatsResponse\"\x00\x12n\n" +
	"\vGetSysStats\x12-.v2ray.core.app.stats.command.SysStatsRequest\x1a..v2ray.core.app.stats.command.SysStatsResponse\"\x00Bu\n" +
	" com.v2ray.core.app.stats.commandP\x01Z0github.com/v2fly/v2ray-core/v5/app/stats/command\xaa\x02\x1cV2Ray.Core.App.Stats.Commandb\x06proto3"

var (
	file_app_stats_command_command_proto_rawDescOnce sync.Once
	file_app_stats_command_command_proto_rawDescData []byte
)

func file_app_stats_command_command_proto_rawDescGZIP() []byte {
	file_app_stats_command_command_proto_rawDescOnce.Do(func() {
		file_app_stats_command_command_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_stats_command_command_proto_rawDesc), len(file_app_stats_command_command_proto_rawDesc)))
	})
	return file_app_stats_command_command_proto_rawDescData
}

var file_app_stats_command_command_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_app_stats_command_command_proto_goTypes = []any{
	(*GetStatsRequest)(nil),    // 0: v2ray.core.app.stats.command.GetStatsRequest
	(*Stat)(nil),               // 1: v2ray.core.app.stats.command.Stat
	(*GetStatsResponse)(nil),   // 2: v2ray.core.app.stats.command.GetStatsResponse
	(*QueryStatsRequest)(nil),  // 3: v2ray.core.app.stats.command.QueryStatsRequest
	(*QueryStatsResponse)(nil), // 4: v2ray.core.app.stats.command.QueryStatsResponse
	(*SysStatsRequest)(nil),    // 5: v2ray.core.app.stats.command.SysStatsRequest
	(*SysStatsResponse)(nil),   // 6: v2ray.core.app.stats.command.SysStatsResponse
	(*Config)(nil),             // 7: v2ray.core.app.stats.command.Config
}
var file_app_stats_command_command_proto_depIdxs = []int32{
	1, // 0: v2ray.core.app.stats.command.GetStatsResponse.stat:type_name -> v2ray.core.app.stats.command.Stat
	1, // 1: v2ray.core.app.stats.command.QueryStatsResponse.stat:type_name -> v2ray.core.app.stats.command.Stat
	0, // 2: v2ray.core.app.stats.command.StatsService.GetStats:input_type -> v2ray.core.app.stats.command.GetStatsRequest
	3, // 3: v2ray.core.app.stats.command.StatsService.QueryStats:input_type -> v2ray.core.app.stats.command.QueryStatsRequest
	5, // 4: v2ray.core.app.stats.command.StatsService.GetSysStats:input_type -> v2ray.core.app.stats.command.SysStatsRequest
	2, // 5: v2ray.core.app.stats.command.StatsService.GetStats:output_type -> v2ray.core.app.stats.command.GetStatsResponse
	4, // 6: v2ray.core.app.stats.command.StatsService.QueryStats:output_type -> v2ray.core.app.stats.command.QueryStatsResponse
	6, // 7: v2ray.core.app.stats.command.StatsService.GetSysStats:output_type -> v2ray.core.app.stats.command.SysStatsResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_app_stats_command_command_proto_init() }
func file_app_stats_command_command_proto_init() {
	if File_app_stats_command_command_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_stats_command_command_proto_rawDesc), len(file_app_stats_command_command_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_stats_command_command_proto_goTypes,
		DependencyIndexes: file_app_stats_command_command_proto_depIdxs,
		MessageInfos:      file_app_stats_command_command_proto_msgTypes,
	}.Build()
	File_app_stats_command_command_proto = out.File
	file_app_stats_command_command_proto_goTypes = nil
	file_app_stats_command_command_proto_depIdxs = nil
}
