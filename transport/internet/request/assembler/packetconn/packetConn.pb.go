package packetconn

import (
	_ "github.com/v2fly/v2ray-core/v5/common/protoext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ClientConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UnderlyingTransportSetting *anypb.Any `protobuf:"bytes,1,opt,name=underlying_transport_setting,json=underlyingTransportSetting,proto3" json:"underlying_transport_setting,omitempty"`
	UnderlyingTransportName    string     `protobuf:"bytes,2,opt,name=underlying_transport_name,json=underlyingTransportName,proto3" json:"underlying_transport_name,omitempty"`
	MaxWriteDelay              int32      `protobuf:"varint,3,opt,name=max_write_delay,json=maxWriteDelay,proto3" json:"max_write_delay,omitempty"`
	MaxRequestSize             int32      `protobuf:"varint,4,opt,name=max_request_size,json=maxRequestSize,proto3" json:"max_request_size,omitempty"`
	PollingIntervalInitial     int32      `protobuf:"varint,5,opt,name=polling_interval_initial,json=pollingIntervalInitial,proto3" json:"polling_interval_initial,omitempty"`
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientConfig.ProtoReflect.Descriptor instead.
func (*ClientConfig) Descriptor() ([]byte, []int) {
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP(), []int{0}
}

func (x *ClientConfig) GetUnderlyingTransportSetting() *anypb.Any {
	if x != nil {
		return x.UnderlyingTransportSetting
	}
	return nil
}

func (x *ClientConfig) GetUnderlyingTransportName() string {
	if x != nil {
		return x.UnderlyingTransportName
	}
	return ""
}

func (x *ClientConfig) GetMaxWriteDelay() int32 {
	if x != nil {
		return x.MaxWriteDelay
	}
	return 0
}

func (x *ClientConfig) GetMaxRequestSize() int32 {
	if x != nil {
		return x.MaxRequestSize
	}
	return 0
}

func (x *ClientConfig) GetPollingIntervalInitial() int32 {
	if x != nil {
		return x.PollingIntervalInitial
	}
	return 0
}

type ServerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UnderlyingTransportSetting     *anypb.Any `protobuf:"bytes,1,opt,name=underlying_transport_setting,json=underlyingTransportSetting,proto3" json:"underlying_transport_setting,omitempty"`
	UnderlyingTransportName        string     `protobuf:"bytes,2,opt,name=underlying_transport_name,json=underlyingTransportName,proto3" json:"underlying_transport_name,omitempty"`
	MaxWriteSize                   int32      `protobuf:"varint,3,opt,name=max_write_size,json=maxWriteSize,proto3" json:"max_write_size,omitempty"`
	MaxWriteDurationMs             int32      `protobuf:"varint,4,opt,name=max_write_duration_ms,json=maxWriteDurationMs,proto3" json:"max_write_duration_ms,omitempty"`
	MaxSimultaneousWriteConnection int32      `protobuf:"varint,5,opt,name=max_simultaneous_write_connection,json=maxSimultaneousWriteConnection,proto3" json:"max_simultaneous_write_connection,omitempty"`
	PacketWritingBuffer            int32      `protobuf:"varint,6,opt,name=packet_writing_buffer,json=packetWritingBuffer,proto3" json:"packet_writing_buffer,omitempty"`
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP(), []int{1}
}

func (x *ServerConfig) GetUnderlyingTransportSetting() *anypb.Any {
	if x != nil {
		return x.UnderlyingTransportSetting
	}
	return nil
}

func (x *ServerConfig) GetUnderlyingTransportName() string {
	if x != nil {
		return x.UnderlyingTransportName
	}
	return ""
}

func (x *ServerConfig) GetMaxWriteSize() int32 {
	if x != nil {
		return x.MaxWriteSize
	}
	return 0
}

func (x *ServerConfig) GetMaxWriteDurationMs() int32 {
	if x != nil {
		return x.MaxWriteDurationMs
	}
	return 0
}

func (x *ServerConfig) GetMaxSimultaneousWriteConnection() int32 {
	if x != nil {
		return x.MaxSimultaneousWriteConnection
	}
	return 0
}

func (x *ServerConfig) GetPacketWritingBuffer() int32 {
	if x != nil {
		return x.PacketWritingBuffer
	}
	return 0
}

var File_transport_internet_request_assembler_packetconn_packetConn_proto protoreflect.FileDescriptor

var file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc = []byte{
	0x0a, 0x40, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x65, 0x74, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2f, 0x61, 0x73, 0x73,
	0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x63, 0x6f, 0x6e,
	0x6e, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x3a, 0x76, 0x32, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65,
	0x74, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x6d, 0x62,
	0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x63, 0x6f, 0x6e, 0x6e, 0x1a, 0x20,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x65, 0x78, 0x74, 0x2f,
	0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe4, 0x02, 0x0a, 0x0c,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x56, 0x0a, 0x1c,
	0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x1a, 0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c,
	0x79, 0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x65, 0x74,
	0x74, 0x69, 0x6e, 0x67, 0x12, 0x3a, 0x0a, 0x19, 0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69,
	0x6e, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79,
	0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x26, 0x0a, 0x0f, 0x6d, 0x61, 0x78, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x64, 0x65,
	0x6c, 0x61, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x6d, 0x61, 0x78, 0x57, 0x72,
	0x69, 0x74, 0x65, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x28, 0x0a, 0x10, 0x6d, 0x61, 0x78, 0x5f,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0e, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x69,
	0x7a, 0x65, 0x12, 0x38, 0x0a, 0x18, 0x70, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x5f, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x16, 0x70, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x76, 0x61, 0x6c, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x3a, 0x34, 0x82, 0xb5,
	0x18, 0x30, 0x0a, 0x22, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x2e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x0a, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x63, 0x6f,
	0x6e, 0x6e, 0x22, 0xb0, 0x03, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x56, 0x0a, 0x1c, 0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69, 0x6e,
	0x67, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x73, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52,
	0x1a, 0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x3a, 0x0a, 0x19, 0x75,
	0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17,
	0x75, 0x6e, 0x64, 0x65, 0x72, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x6d, 0x61, 0x78, 0x5f, 0x77,
	0x72, 0x69, 0x74, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0c, 0x6d, 0x61, 0x78, 0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x31, 0x0a,
	0x15, 0x6d, 0x61, 0x78, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x12, 0x6d, 0x61,
	0x78, 0x57, 0x72, 0x69, 0x74, 0x65, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x73,
	0x12, 0x49, 0x0a, 0x21, 0x6d, 0x61, 0x78, 0x5f, 0x73, 0x69, 0x6d, 0x75, 0x6c, 0x74, 0x61, 0x6e,
	0x65, 0x6f, 0x75, 0x73, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x1e, 0x6d, 0x61, 0x78,
	0x53, 0x69, 0x6d, 0x75, 0x6c, 0x74, 0x61, 0x6e, 0x65, 0x6f, 0x75, 0x73, 0x57, 0x72, 0x69, 0x74,
	0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x32, 0x0a, 0x15, 0x70,
	0x61, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x62, 0x75,
	0x66, 0x66, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x13, 0x70, 0x61, 0x63, 0x6b,
	0x65, 0x74, 0x57, 0x72, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x3a,
	0x34, 0x82, 0xb5, 0x18, 0x30, 0x0a, 0x22, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c,
	0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0a, 0x70, 0x61, 0x63, 0x6b, 0x65,
	0x74, 0x63, 0x6f, 0x6e, 0x6e, 0x42, 0xcf, 0x01, 0x0a, 0x3e, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32,
	0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x2e, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x61,
	0x63, 0x6b, 0x65, 0x74, 0x63, 0x6f, 0x6e, 0x6e, 0x50, 0x01, 0x5a, 0x4e, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x32, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x72,
	0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x35, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x2f, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x2f,
	0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x63, 0x6f, 0x6e, 0x6e, 0xaa, 0x02, 0x3a, 0x56, 0x32, 0x52,
	0x61, 0x79, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x2e, 0x50, 0x61, 0x63,
	0x6b, 0x65, 0x74, 0x63, 0x6f, 0x6e, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescOnce sync.Once
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData = file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc
)

func file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescGZIP() []byte {
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescOnce.Do(func() {
		file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData = protoimpl.X.CompressGZIP(file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData)
	})
	return file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDescData
}

var file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes = []interface{}{
	(*ClientConfig)(nil), // 0: v2ray.core.transport.internet.request.assembler.packetconn.ClientConfig
	(*ServerConfig)(nil), // 1: v2ray.core.transport.internet.request.assembler.packetconn.ServerConfig
	(*anypb.Any)(nil),    // 2: google.protobuf.Any
}
var file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs = []int32{
	2, // 0: v2ray.core.transport.internet.request.assembler.packetconn.ClientConfig.underlying_transport_setting:type_name -> google.protobuf.Any
	2, // 1: v2ray.core.transport.internet.request.assembler.packetconn.ServerConfig.underlying_transport_setting:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_transport_internet_request_assembler_packetconn_packetConn_proto_init() }
func file_transport_internet_request_assembler_packetconn_packetConn_proto_init() {
	if File_transport_internet_request_assembler_packetconn_packetConn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes,
		DependencyIndexes: file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs,
		MessageInfos:      file_transport_internet_request_assembler_packetconn_packetConn_proto_msgTypes,
	}.Build()
	File_transport_internet_request_assembler_packetconn_packetConn_proto = out.File
	file_transport_internet_request_assembler_packetconn_packetConn_proto_rawDesc = nil
	file_transport_internet_request_assembler_packetconn_packetConn_proto_goTypes = nil
	file_transport_internet_request_assembler_packetconn_packetConn_proto_depIdxs = nil
}
