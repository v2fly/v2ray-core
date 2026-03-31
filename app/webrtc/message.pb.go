package webrtc

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

type ConnectionRequest struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	ReplyAddressTag []byte                 `protobuf:"bytes,1,opt,name=reply_address_tag,json=replyAddressTag,proto3" json:"reply_address_tag,omitempty"`
	// Supports trickle ICE by allowing incremental candidate updates.
	ConnectionSessionId []byte   `protobuf:"bytes,2,opt,name=connection_session_id,json=connectionSessionId,proto3" json:"connection_session_id,omitempty"`
	SessionDescription  []byte   `protobuf:"bytes,3,opt,name=session_description,json=sessionDescription,proto3" json:"session_description,omitempty"`
	Candidates          [][]byte `protobuf:"bytes,4,rep,name=candidates,proto3" json:"candidates,omitempty"`
	RequestPortBlossom  bool     `protobuf:"varint,5,opt,name=request_port_blossom,json=requestPortBlossom,proto3" json:"request_port_blossom,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *ConnectionRequest) Reset() {
	*x = ConnectionRequest{}
	mi := &file_app_webrtc_message_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionRequest) ProtoMessage() {}

func (x *ConnectionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_message_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectionRequest.ProtoReflect.Descriptor instead.
func (*ConnectionRequest) Descriptor() ([]byte, []int) {
	return file_app_webrtc_message_proto_rawDescGZIP(), []int{0}
}

func (x *ConnectionRequest) GetReplyAddressTag() []byte {
	if x != nil {
		return x.ReplyAddressTag
	}
	return nil
}

func (x *ConnectionRequest) GetConnectionSessionId() []byte {
	if x != nil {
		return x.ConnectionSessionId
	}
	return nil
}

func (x *ConnectionRequest) GetSessionDescription() []byte {
	if x != nil {
		return x.SessionDescription
	}
	return nil
}

func (x *ConnectionRequest) GetCandidates() [][]byte {
	if x != nil {
		return x.Candidates
	}
	return nil
}

func (x *ConnectionRequest) GetRequestPortBlossom() bool {
	if x != nil {
		return x.RequestPortBlossom
	}
	return false
}

type ConnectionResponse struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	ReplyAddressTag    []byte                 `protobuf:"bytes,1,opt,name=reply_address_tag,json=replyAddressTag,proto3" json:"reply_address_tag,omitempty"`
	SessionDescription []byte                 `protobuf:"bytes,3,opt,name=session_description,json=sessionDescription,proto3" json:"session_description,omitempty"`
	Candidates         [][]byte               `protobuf:"bytes,4,rep,name=candidates,proto3" json:"candidates,omitempty"`
	StopPolling        bool                   `protobuf:"varint,5,opt,name=stop_polling,json=stopPolling,proto3" json:"stop_polling,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ConnectionResponse) Reset() {
	*x = ConnectionResponse{}
	mi := &file_app_webrtc_message_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionResponse) ProtoMessage() {}

func (x *ConnectionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_message_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectionResponse.ProtoReflect.Descriptor instead.
func (*ConnectionResponse) Descriptor() ([]byte, []int) {
	return file_app_webrtc_message_proto_rawDescGZIP(), []int{1}
}

func (x *ConnectionResponse) GetReplyAddressTag() []byte {
	if x != nil {
		return x.ReplyAddressTag
	}
	return nil
}

func (x *ConnectionResponse) GetSessionDescription() []byte {
	if x != nil {
		return x.SessionDescription
	}
	return nil
}

func (x *ConnectionResponse) GetCandidates() [][]byte {
	if x != nil {
		return x.Candidates
	}
	return nil
}

func (x *ConnectionResponse) GetStopPolling() bool {
	if x != nil {
		return x.StopPolling
	}
	return false
}

var File_app_webrtc_message_proto protoreflect.FileDescriptor

const file_app_webrtc_message_proto_rawDesc = "" +
	"\n" +
	"\x18app/webrtc/message.proto\x12\x15v2ray.core.app.webrtc\"\xf6\x01\n" +
	"\x11ConnectionRequest\x12*\n" +
	"\x11reply_address_tag\x18\x01 \x01(\fR\x0freplyAddressTag\x122\n" +
	"\x15connection_session_id\x18\x02 \x01(\fR\x13connectionSessionId\x12/\n" +
	"\x13session_description\x18\x03 \x01(\fR\x12sessionDescription\x12\x1e\n" +
	"\n" +
	"candidates\x18\x04 \x03(\fR\n" +
	"candidates\x120\n" +
	"\x14request_port_blossom\x18\x05 \x01(\bR\x12requestPortBlossom\"\xb4\x01\n" +
	"\x12ConnectionResponse\x12*\n" +
	"\x11reply_address_tag\x18\x01 \x01(\fR\x0freplyAddressTag\x12/\n" +
	"\x13session_description\x18\x03 \x01(\fR\x12sessionDescription\x12\x1e\n" +
	"\n" +
	"candidates\x18\x04 \x03(\fR\n" +
	"candidates\x12!\n" +
	"\fstop_polling\x18\x05 \x01(\bR\vstopPollingB`\n" +
	"\x19com.v2ray.core.app.webrtcP\x01Z)github.com/v2fly/v2ray-core/v5/app/webrtc\xaa\x02\x15V2Ray.Core.App.WebRTCb\x06proto3"

var (
	file_app_webrtc_message_proto_rawDescOnce sync.Once
	file_app_webrtc_message_proto_rawDescData []byte
)

func file_app_webrtc_message_proto_rawDescGZIP() []byte {
	file_app_webrtc_message_proto_rawDescOnce.Do(func() {
		file_app_webrtc_message_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_webrtc_message_proto_rawDesc), len(file_app_webrtc_message_proto_rawDesc)))
	})
	return file_app_webrtc_message_proto_rawDescData
}

var file_app_webrtc_message_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_app_webrtc_message_proto_goTypes = []any{
	(*ConnectionRequest)(nil),  // 0: v2ray.core.app.webrtc.ConnectionRequest
	(*ConnectionResponse)(nil), // 1: v2ray.core.app.webrtc.ConnectionResponse
}
var file_app_webrtc_message_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_app_webrtc_message_proto_init() }
func file_app_webrtc_message_proto_init() {
	if File_app_webrtc_message_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_webrtc_message_proto_rawDesc), len(file_app_webrtc_message_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_webrtc_message_proto_goTypes,
		DependencyIndexes: file_app_webrtc_message_proto_depIdxs,
		MessageInfos:      file_app_webrtc_message_proto_msgTypes,
	}.Build()
	File_app_webrtc_message_proto = out.File
	file_app_webrtc_message_proto_goTypes = nil
	file_app_webrtc_message_proto_depIdxs = nil
}
