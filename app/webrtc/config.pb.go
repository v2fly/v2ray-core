package webrtc

import (
	packetaddr "github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
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

type WebRTCTURNServer struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Only UDP TURN URLs are supported for active listeners, for example:
	// turn:turn.example.com:3478?transport=udp
	Url           string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Username      string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Password      string `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WebRTCTURNServer) Reset() {
	*x = WebRTCTURNServer{}
	mi := &file_app_webrtc_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WebRTCTURNServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebRTCTURNServer) ProtoMessage() {}

func (x *WebRTCTURNServer) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebRTCTURNServer.ProtoReflect.Descriptor instead.
func (*WebRTCTURNServer) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{0}
}

func (x *WebRTCTURNServer) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *WebRTCTURNServer) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *WebRTCTURNServer) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

// LocalWebRTCListener is an active WebRTC listener that gathers its own UDP candidates.
type LocalWebRTCListener struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Tag   string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	// The connection will be sent from this outbound tag when supported by the listener runtime.
	ConnectVia         string                    `protobuf:"bytes,2,opt,name=connect_via,json=connectVia,proto3" json:"connect_via,omitempty"`
	PacketEncoding     packetaddr.PacketAddrType `protobuf:"varint,3,opt,name=packet_encoding,json=packetEncoding,proto3,enum=v2ray.core.net.packetaddr.PacketAddrType" json:"packet_encoding,omitempty"`
	StunServers        []string                  `protobuf:"bytes,4,rep,name=stun_servers,json=stunServers,proto3" json:"stun_servers,omitempty"`
	TurnServers        []*WebRTCTURNServer       `protobuf:"bytes,9,rep,name=turn_servers,json=turnServers,proto3" json:"turn_servers,omitempty"`
	UseIpv4            bool                      `protobuf:"varint,5,opt,name=use_ipv4,json=useIpv4,proto3" json:"use_ipv4,omitempty"`
	UseIpv6            bool                      `protobuf:"varint,6,opt,name=use_ipv6,json=useIpv6,proto3" json:"use_ipv6,omitempty"`
	RequestPortBlossom bool                      `protobuf:"varint,7,opt,name=request_port_blossom,json=requestPortBlossom,proto3" json:"request_port_blossom,omitempty"`
	AcceptPortBlossom  bool                      `protobuf:"varint,8,opt,name=accept_port_blossom,json=acceptPortBlossom,proto3" json:"accept_port_blossom,omitempty"`
	// Limits acceptor-side repeated port blossom. 0 uses the default 6 seconds.
	PortBlossomDurationSec                                                   uint32 `protobuf:"varint,10,opt,name=port_blossom_duration_sec,json=portBlossomDurationSec,proto3" json:"port_blossom_duration_sec,omitempty"`
	RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround bool   `protobuf:"varint,11,opt,name=require_remote_finish_candidate_gathering_before_sending_own_candidate_workaround,json=requireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround,proto3" json:"require_remote_finish_candidate_gathering_before_sending_own_candidate_workaround,omitempty"`
	unknownFields                                                            protoimpl.UnknownFields
	sizeCache                                                                protoimpl.SizeCache
}

func (x *LocalWebRTCListener) Reset() {
	*x = LocalWebRTCListener{}
	mi := &file_app_webrtc_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LocalWebRTCListener) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocalWebRTCListener) ProtoMessage() {}

func (x *LocalWebRTCListener) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocalWebRTCListener.ProtoReflect.Descriptor instead.
func (*LocalWebRTCListener) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{1}
}

func (x *LocalWebRTCListener) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *LocalWebRTCListener) GetConnectVia() string {
	if x != nil {
		return x.ConnectVia
	}
	return ""
}

func (x *LocalWebRTCListener) GetPacketEncoding() packetaddr.PacketAddrType {
	if x != nil {
		return x.PacketEncoding
	}
	return packetaddr.PacketAddrType(0)
}

func (x *LocalWebRTCListener) GetStunServers() []string {
	if x != nil {
		return x.StunServers
	}
	return nil
}

func (x *LocalWebRTCListener) GetTurnServers() []*WebRTCTURNServer {
	if x != nil {
		return x.TurnServers
	}
	return nil
}

func (x *LocalWebRTCListener) GetUseIpv4() bool {
	if x != nil {
		return x.UseIpv4
	}
	return false
}

func (x *LocalWebRTCListener) GetUseIpv6() bool {
	if x != nil {
		return x.UseIpv6
	}
	return false
}

func (x *LocalWebRTCListener) GetRequestPortBlossom() bool {
	if x != nil {
		return x.RequestPortBlossom
	}
	return false
}

func (x *LocalWebRTCListener) GetAcceptPortBlossom() bool {
	if x != nil {
		return x.AcceptPortBlossom
	}
	return false
}

func (x *LocalWebRTCListener) GetPortBlossomDurationSec() uint32 {
	if x != nil {
		return x.PortBlossomDurationSec
	}
	return 0
}

func (x *LocalWebRTCListener) GetRequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround() bool {
	if x != nil {
		return x.RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround
	}
	return false
}

// LocalWebRTCSystemListener is a passive WebRTC listener backed by a local packet socket.
type LocalWebRTCSystemListener struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Tag   string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	// If set, then a UDP mux will be created on this port.
	LocalPort uint32 `protobuf:"varint,2,opt,name=local_port,json=localPort,proto3" json:"local_port,omitempty"`
	// The ICE-lite local address.
	Ip                 []byte `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`
	Ipv6               []byte `protobuf:"bytes,4,opt,name=ipv6,proto3" json:"ipv6,omitempty"`
	RequestPortBlossom bool   `protobuf:"varint,7,opt,name=request_port_blossom,json=requestPortBlossom,proto3" json:"request_port_blossom,omitempty"`
	AcceptPortBlossom  bool   `protobuf:"varint,8,opt,name=accept_port_blossom,json=acceptPortBlossom,proto3" json:"accept_port_blossom,omitempty"`
	// Limits acceptor-side repeated port blossom. 0 uses the default 6 seconds.
	PortBlossomDurationSec uint32 `protobuf:"varint,9,opt,name=port_blossom_duration_sec,json=portBlossomDurationSec,proto3" json:"port_blossom_duration_sec,omitempty"`
	IpAddr                 string `protobuf:"bytes,68000,opt,name=ip_addr,json=ipAddr,proto3" json:"ip_addr,omitempty"`
	Ipv6Addr               string `protobuf:"bytes,68001,opt,name=ipv6_addr,json=ipv6Addr,proto3" json:"ipv6_addr,omitempty"`
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *LocalWebRTCSystemListener) Reset() {
	*x = LocalWebRTCSystemListener{}
	mi := &file_app_webrtc_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LocalWebRTCSystemListener) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocalWebRTCSystemListener) ProtoMessage() {}

func (x *LocalWebRTCSystemListener) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocalWebRTCSystemListener.ProtoReflect.Descriptor instead.
func (*LocalWebRTCSystemListener) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{2}
}

func (x *LocalWebRTCSystemListener) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *LocalWebRTCSystemListener) GetLocalPort() uint32 {
	if x != nil {
		return x.LocalPort
	}
	return 0
}

func (x *LocalWebRTCSystemListener) GetIp() []byte {
	if x != nil {
		return x.Ip
	}
	return nil
}

func (x *LocalWebRTCSystemListener) GetIpv6() []byte {
	if x != nil {
		return x.Ipv6
	}
	return nil
}

func (x *LocalWebRTCSystemListener) GetRequestPortBlossom() bool {
	if x != nil {
		return x.RequestPortBlossom
	}
	return false
}

func (x *LocalWebRTCSystemListener) GetAcceptPortBlossom() bool {
	if x != nil {
		return x.AcceptPortBlossom
	}
	return false
}

func (x *LocalWebRTCSystemListener) GetPortBlossomDurationSec() uint32 {
	if x != nil {
		return x.PortBlossomDurationSec
	}
	return 0
}

func (x *LocalWebRTCSystemListener) GetIpAddr() string {
	if x != nil {
		return x.IpAddr
	}
	return ""
}

func (x *LocalWebRTCSystemListener) GetIpv6Addr() string {
	if x != nil {
		return x.Ipv6Addr
	}
	return ""
}

type ClientConfig struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	RoundTripperClient *anypb.Any             `protobuf:"bytes,1,opt,name=round_tripper_client,json=roundTripperClient,proto3" json:"round_tripper_client,omitempty"`
	SecurityConfig     *anypb.Any             `protobuf:"bytes,2,opt,name=security_config,json=securityConfig,proto3" json:"security_config,omitempty"`
	Dest               string                 `protobuf:"bytes,3,opt,name=dest,proto3" json:"dest,omitempty"`
	OutboundTag        string                 `protobuf:"bytes,4,opt,name=outbound_tag,json=outboundTag,proto3" json:"outbound_tag,omitempty"`
	ServerIdentity     []byte                 `protobuf:"bytes,5,opt,name=server_identity,json=serverIdentity,proto3" json:"server_identity,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ClientConfig) Reset() {
	*x = ClientConfig{}
	mi := &file_app_webrtc_config_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientConfig) ProtoMessage() {}

func (x *ClientConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[3]
	if x != nil {
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
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{3}
}

func (x *ClientConfig) GetRoundTripperClient() *anypb.Any {
	if x != nil {
		return x.RoundTripperClient
	}
	return nil
}

func (x *ClientConfig) GetSecurityConfig() *anypb.Any {
	if x != nil {
		return x.SecurityConfig
	}
	return nil
}

func (x *ClientConfig) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

func (x *ClientConfig) GetOutboundTag() string {
	if x != nil {
		return x.OutboundTag
	}
	return ""
}

func (x *ClientConfig) GetServerIdentity() []byte {
	if x != nil {
		return x.ServerIdentity
	}
	return nil
}

type ServerInverseRoleConfig struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	RoundTripperClient *anypb.Any             `protobuf:"bytes,1,opt,name=round_tripper_client,json=roundTripperClient,proto3" json:"round_tripper_client,omitempty"`
	SecurityConfig     *anypb.Any             `protobuf:"bytes,2,opt,name=security_config,json=securityConfig,proto3" json:"security_config,omitempty"`
	Dest               string                 `protobuf:"bytes,3,opt,name=dest,proto3" json:"dest,omitempty"`
	OutboundTag        string                 `protobuf:"bytes,4,opt,name=outbound_tag,json=outboundTag,proto3" json:"outbound_tag,omitempty"`
	ServerIdentity     []byte                 `protobuf:"bytes,5,opt,name=server_identity,json=serverIdentity,proto3" json:"server_identity,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *ServerInverseRoleConfig) Reset() {
	*x = ServerInverseRoleConfig{}
	mi := &file_app_webrtc_config_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerInverseRoleConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerInverseRoleConfig) ProtoMessage() {}

func (x *ServerInverseRoleConfig) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerInverseRoleConfig.ProtoReflect.Descriptor instead.
func (*ServerInverseRoleConfig) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{4}
}

func (x *ServerInverseRoleConfig) GetRoundTripperClient() *anypb.Any {
	if x != nil {
		return x.RoundTripperClient
	}
	return nil
}

func (x *ServerInverseRoleConfig) GetSecurityConfig() *anypb.Any {
	if x != nil {
		return x.SecurityConfig
	}
	return nil
}

func (x *ServerInverseRoleConfig) GetDest() string {
	if x != nil {
		return x.Dest
	}
	return ""
}

func (x *ServerInverseRoleConfig) GetOutboundTag() string {
	if x != nil {
		return x.OutboundTag
	}
	return ""
}

func (x *ServerInverseRoleConfig) GetServerIdentity() []byte {
	if x != nil {
		return x.ServerIdentity
	}
	return nil
}

type UDPPortForwarderAcceptor struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Port          uint32                 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Ip            []byte                 `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`
	ConnectVia    string                 `protobuf:"bytes,4,opt,name=connect_via,json=connectVia,proto3" json:"connect_via,omitempty"`
	IpAddr        string                 `protobuf:"bytes,68000,opt,name=ip_addr,json=ipAddr,proto3" json:"ip_addr,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UDPPortForwarderAcceptor) Reset() {
	*x = UDPPortForwarderAcceptor{}
	mi := &file_app_webrtc_config_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UDPPortForwarderAcceptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UDPPortForwarderAcceptor) ProtoMessage() {}

func (x *UDPPortForwarderAcceptor) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UDPPortForwarderAcceptor.ProtoReflect.Descriptor instead.
func (*UDPPortForwarderAcceptor) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{5}
}

func (x *UDPPortForwarderAcceptor) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *UDPPortForwarderAcceptor) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *UDPPortForwarderAcceptor) GetIp() []byte {
	if x != nil {
		return x.Ip
	}
	return nil
}

func (x *UDPPortForwarderAcceptor) GetConnectVia() string {
	if x != nil {
		return x.ConnectVia
	}
	return ""
}

func (x *UDPPortForwarderAcceptor) GetIpAddr() string {
	if x != nil {
		return x.IpAddr
	}
	return ""
}

type UDPPortForwarderRemote struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Tag   string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	// Create a new outbound to accept UDP connection traffic for forwarding.
	AcceptConnectOn string `protobuf:"bytes,4,opt,name=accept_connect_on,json=acceptConnectOn,proto3" json:"accept_connect_on,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *UDPPortForwarderRemote) Reset() {
	*x = UDPPortForwarderRemote{}
	mi := &file_app_webrtc_config_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UDPPortForwarderRemote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UDPPortForwarderRemote) ProtoMessage() {}

func (x *UDPPortForwarderRemote) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UDPPortForwarderRemote.ProtoReflect.Descriptor instead.
func (*UDPPortForwarderRemote) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{6}
}

func (x *UDPPortForwarderRemote) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *UDPPortForwarderRemote) GetAcceptConnectOn() string {
	if x != nil {
		return x.AcceptConnectOn
	}
	return ""
}

type Acceptor struct {
	state                protoimpl.MessageState      `protogen:"open.v1"`
	Tag                  string                      `protobuf:"bytes,3,opt,name=tag,proto3" json:"tag,omitempty"`
	ServerConfig         *anypb.Any                  `protobuf:"bytes,1,opt,name=server_config,json=serverConfig,proto3" json:"server_config,omitempty"`
	PortForwarderAccepts []*UDPPortForwarderAcceptor `protobuf:"bytes,2,rep,name=port_forwarder_accepts,json=portForwarderAccepts,proto3" json:"port_forwarder_accepts,omitempty"`
	AcceptOnTag          string                      `protobuf:"bytes,4,opt,name=accept_on_tag,json=acceptOnTag,proto3" json:"accept_on_tag,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *Acceptor) Reset() {
	*x = Acceptor{}
	mi := &file_app_webrtc_config_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Acceptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Acceptor) ProtoMessage() {}

func (x *Acceptor) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Acceptor.ProtoReflect.Descriptor instead.
func (*Acceptor) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{7}
}

func (x *Acceptor) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Acceptor) GetServerConfig() *anypb.Any {
	if x != nil {
		return x.ServerConfig
	}
	return nil
}

func (x *Acceptor) GetPortForwarderAccepts() []*UDPPortForwarderAcceptor {
	if x != nil {
		return x.PortForwarderAccepts
	}
	return nil
}

func (x *Acceptor) GetAcceptOnTag() string {
	if x != nil {
		return x.AcceptOnTag
	}
	return ""
}

type Remote struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,3,opt,name=tag,proto3" json:"tag,omitempty"`
	ClientConfig  *anypb.Any             `protobuf:"bytes,1,opt,name=client_config,json=clientConfig,proto3" json:"client_config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Remote) Reset() {
	*x = Remote{}
	mi := &file_app_webrtc_config_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Remote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Remote) ProtoMessage() {}

func (x *Remote) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Remote.ProtoReflect.Descriptor instead.
func (*Remote) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{8}
}

func (x *Remote) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Remote) GetClientConfig() *anypb.Any {
	if x != nil {
		return x.ClientConfig
	}
	return nil
}

type RemoteConnections struct {
	state            protoimpl.MessageState    `protogen:"open.v1"`
	RemoteTag        string                    `protobuf:"bytes,1,opt,name=remote_tag,json=remoteTag,proto3" json:"remote_tag,omitempty"`
	LocalListenerTag string                    `protobuf:"bytes,2,opt,name=local_listener_tag,json=localListenerTag,proto3" json:"local_listener_tag,omitempty"`
	PortForward      []*UDPPortForwarderRemote `protobuf:"bytes,3,rep,name=port_forward,json=portForward,proto3" json:"port_forward,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *RemoteConnections) Reset() {
	*x = RemoteConnections{}
	mi := &file_app_webrtc_config_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoteConnections) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoteConnections) ProtoMessage() {}

func (x *RemoteConnections) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoteConnections.ProtoReflect.Descriptor instead.
func (*RemoteConnections) Descriptor() ([]byte, []int) {
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{9}
}

func (x *RemoteConnections) GetRemoteTag() string {
	if x != nil {
		return x.RemoteTag
	}
	return ""
}

func (x *RemoteConnections) GetLocalListenerTag() string {
	if x != nil {
		return x.LocalListenerTag
	}
	return ""
}

func (x *RemoteConnections) GetPortForward() []*UDPPortForwarderRemote {
	if x != nil {
		return x.PortForward
	}
	return nil
}

type Config struct {
	state          protoimpl.MessageState       `protogen:"open.v1"`
	Listener       []*LocalWebRTCListener       `protobuf:"bytes,1,rep,name=listener,proto3" json:"listener,omitempty"`
	SystemListener []*LocalWebRTCSystemListener `protobuf:"bytes,2,rep,name=system_listener,json=systemListener,proto3" json:"system_listener,omitempty"`
	Acceptors      []*Acceptor                  `protobuf:"bytes,3,rep,name=acceptors,proto3" json:"acceptors,omitempty"`
	Remotes        []*Remote                    `protobuf:"bytes,4,rep,name=remotes,proto3" json:"remotes,omitempty"`
	Connection     []*RemoteConnections         `protobuf:"bytes,5,rep,name=connection,proto3" json:"connection,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_app_webrtc_config_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_app_webrtc_config_proto_msgTypes[10]
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
	return file_app_webrtc_config_proto_rawDescGZIP(), []int{10}
}

func (x *Config) GetListener() []*LocalWebRTCListener {
	if x != nil {
		return x.Listener
	}
	return nil
}

func (x *Config) GetSystemListener() []*LocalWebRTCSystemListener {
	if x != nil {
		return x.SystemListener
	}
	return nil
}

func (x *Config) GetAcceptors() []*Acceptor {
	if x != nil {
		return x.Acceptors
	}
	return nil
}

func (x *Config) GetRemotes() []*Remote {
	if x != nil {
		return x.Remotes
	}
	return nil
}

func (x *Config) GetConnection() []*RemoteConnections {
	if x != nil {
		return x.Connection
	}
	return nil
}

var File_app_webrtc_config_proto protoreflect.FileDescriptor

const file_app_webrtc_config_proto_rawDesc = "" +
	"\n" +
	"\x17app/webrtc/config.proto\x12\x15v2ray.core.app.webrtc\x1a common/protoext/extensions.proto\x1a\x19google/protobuf/any.proto\x1a\"common/net/packetaddr/config.proto\"\\\n" +
	"\x10WebRTCTURNServer\x12\x10\n" +
	"\x03url\x18\x01 \x01(\tR\x03url\x12\x1a\n" +
	"\busername\x18\x02 \x01(\tR\busername\x12\x1a\n" +
	"\bpassword\x18\x03 \x01(\tR\bpassword\"\x84\x05\n" +
	"\x13LocalWebRTCListener\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x1f\n" +
	"\vconnect_via\x18\x02 \x01(\tR\n" +
	"connectVia\x12R\n" +
	"\x0fpacket_encoding\x18\x03 \x01(\x0e2).v2ray.core.net.packetaddr.PacketAddrTypeR\x0epacketEncoding\x12!\n" +
	"\fstun_servers\x18\x04 \x03(\tR\vstunServers\x12J\n" +
	"\fturn_servers\x18\t \x03(\v2'.v2ray.core.app.webrtc.WebRTCTURNServerR\vturnServers\x12\x19\n" +
	"\buse_ipv4\x18\x05 \x01(\bR\auseIpv4\x12\x19\n" +
	"\buse_ipv6\x18\x06 \x01(\bR\auseIpv6\x120\n" +
	"\x14request_port_blossom\x18\a \x01(\bR\x12requestPortBlossom\x12.\n" +
	"\x13accept_port_blossom\x18\b \x01(\bR\x11acceptPortBlossom\x129\n" +
	"\x19port_blossom_duration_sec\x18\n" +
	" \x01(\rR\x16portBlossomDurationSec\x12\xa3\x01\n" +
	"Qrequire_remote_finish_candidate_gathering_before_sending_own_candidate_workaround\x18\v \x01(\bRHrequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround\"\xdd\x02\n" +
	"\x19LocalWebRTCSystemListener\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x1d\n" +
	"\n" +
	"local_port\x18\x02 \x01(\rR\tlocalPort\x12\x0e\n" +
	"\x02ip\x18\x03 \x01(\fR\x02ip\x12\x12\n" +
	"\x04ipv6\x18\x04 \x01(\fR\x04ipv6\x120\n" +
	"\x14request_port_blossom\x18\a \x01(\bR\x12requestPortBlossom\x12.\n" +
	"\x13accept_port_blossom\x18\b \x01(\bR\x11acceptPortBlossom\x129\n" +
	"\x19port_blossom_duration_sec\x18\t \x01(\rR\x16portBlossomDurationSec\x12#\n" +
	"\aip_addr\x18\xa0\x93\x04 \x01(\tB\b\x82\xb5\x18\x04:\x02ipR\x06ipAddr\x12)\n" +
	"\tipv6_addr\x18\xa1\x93\x04 \x01(\tB\n" +
	"\x82\xb5\x18\x06:\x04ipv6R\bipv6Addr\"\xf5\x01\n" +
	"\fClientConfig\x12F\n" +
	"\x14round_tripper_client\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\x12roundTripperClient\x12=\n" +
	"\x0fsecurity_config\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x0esecurityConfig\x12\x12\n" +
	"\x04dest\x18\x03 \x01(\tR\x04dest\x12!\n" +
	"\foutbound_tag\x18\x04 \x01(\tR\voutboundTag\x12'\n" +
	"\x0fserver_identity\x18\x05 \x01(\fR\x0eserverIdentity\"\x80\x02\n" +
	"\x17ServerInverseRoleConfig\x12F\n" +
	"\x14round_tripper_client\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\x12roundTripperClient\x12=\n" +
	"\x0fsecurity_config\x18\x02 \x01(\v2\x14.google.protobuf.AnyR\x0esecurityConfig\x12\x12\n" +
	"\x04dest\x18\x03 \x01(\tR\x04dest\x12!\n" +
	"\foutbound_tag\x18\x04 \x01(\tR\voutboundTag\x12'\n" +
	"\x0fserver_identity\x18\x05 \x01(\fR\x0eserverIdentity\"\x96\x01\n" +
	"\x18UDPPortForwarderAcceptor\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12\x12\n" +
	"\x04port\x18\x02 \x01(\rR\x04port\x12\x0e\n" +
	"\x02ip\x18\x03 \x01(\fR\x02ip\x12\x1f\n" +
	"\vconnect_via\x18\x04 \x01(\tR\n" +
	"connectVia\x12#\n" +
	"\aip_addr\x18\xa0\x93\x04 \x01(\tB\b\x82\xb5\x18\x04:\x02ipR\x06ipAddr\"V\n" +
	"\x16UDPPortForwarderRemote\x12\x10\n" +
	"\x03tag\x18\x01 \x01(\tR\x03tag\x12*\n" +
	"\x11accept_connect_on\x18\x04 \x01(\tR\x0facceptConnectOn\"\xe2\x01\n" +
	"\bAcceptor\x12\x10\n" +
	"\x03tag\x18\x03 \x01(\tR\x03tag\x129\n" +
	"\rserver_config\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\fserverConfig\x12e\n" +
	"\x16port_forwarder_accepts\x18\x02 \x03(\v2/.v2ray.core.app.webrtc.UDPPortForwarderAcceptorR\x14portForwarderAccepts\x12\"\n" +
	"\raccept_on_tag\x18\x04 \x01(\tR\vacceptOnTag\"U\n" +
	"\x06Remote\x12\x10\n" +
	"\x03tag\x18\x03 \x01(\tR\x03tag\x129\n" +
	"\rclient_config\x18\x01 \x01(\v2\x14.google.protobuf.AnyR\fclientConfig\"\xb2\x01\n" +
	"\x11RemoteConnections\x12\x1d\n" +
	"\n" +
	"remote_tag\x18\x01 \x01(\tR\tremoteTag\x12,\n" +
	"\x12local_listener_tag\x18\x02 \x01(\tR\x10localListenerTag\x12P\n" +
	"\fport_forward\x18\x03 \x03(\v2-.v2ray.core.app.webrtc.UDPPortForwarderRemoteR\vportForward\"\x84\x03\n" +
	"\x06Config\x12F\n" +
	"\blistener\x18\x01 \x03(\v2*.v2ray.core.app.webrtc.LocalWebRTCListenerR\blistener\x12Y\n" +
	"\x0fsystem_listener\x18\x02 \x03(\v20.v2ray.core.app.webrtc.LocalWebRTCSystemListenerR\x0esystemListener\x12=\n" +
	"\tacceptors\x18\x03 \x03(\v2\x1f.v2ray.core.app.webrtc.AcceptorR\tacceptors\x127\n" +
	"\aremotes\x18\x04 \x03(\v2\x1d.v2ray.core.app.webrtc.RemoteR\aremotes\x12H\n" +
	"\n" +
	"connection\x18\x05 \x03(\v2(.v2ray.core.app.webrtc.RemoteConnectionsR\n" +
	"connection:\x15\x82\xb5\x18\x11\n" +
	"\aservice\x12\x06webrtcB`\n" +
	"\x19com.v2ray.core.app.webrtcP\x01Z)github.com/v2fly/v2ray-core/v5/app/webrtc\xaa\x02\x15V2Ray.Core.App.WebRTCb\x06proto3"

var (
	file_app_webrtc_config_proto_rawDescOnce sync.Once
	file_app_webrtc_config_proto_rawDescData []byte
)

func file_app_webrtc_config_proto_rawDescGZIP() []byte {
	file_app_webrtc_config_proto_rawDescOnce.Do(func() {
		file_app_webrtc_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_app_webrtc_config_proto_rawDesc), len(file_app_webrtc_config_proto_rawDesc)))
	})
	return file_app_webrtc_config_proto_rawDescData
}

var file_app_webrtc_config_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_app_webrtc_config_proto_goTypes = []any{
	(*WebRTCTURNServer)(nil),          // 0: v2ray.core.app.webrtc.WebRTCTURNServer
	(*LocalWebRTCListener)(nil),       // 1: v2ray.core.app.webrtc.LocalWebRTCListener
	(*LocalWebRTCSystemListener)(nil), // 2: v2ray.core.app.webrtc.LocalWebRTCSystemListener
	(*ClientConfig)(nil),              // 3: v2ray.core.app.webrtc.ClientConfig
	(*ServerInverseRoleConfig)(nil),   // 4: v2ray.core.app.webrtc.ServerInverseRoleConfig
	(*UDPPortForwarderAcceptor)(nil),  // 5: v2ray.core.app.webrtc.UDPPortForwarderAcceptor
	(*UDPPortForwarderRemote)(nil),    // 6: v2ray.core.app.webrtc.UDPPortForwarderRemote
	(*Acceptor)(nil),                  // 7: v2ray.core.app.webrtc.Acceptor
	(*Remote)(nil),                    // 8: v2ray.core.app.webrtc.Remote
	(*RemoteConnections)(nil),         // 9: v2ray.core.app.webrtc.RemoteConnections
	(*Config)(nil),                    // 10: v2ray.core.app.webrtc.Config
	(packetaddr.PacketAddrType)(0),    // 11: v2ray.core.net.packetaddr.PacketAddrType
	(*anypb.Any)(nil),                 // 12: google.protobuf.Any
}
var file_app_webrtc_config_proto_depIdxs = []int32{
	11, // 0: v2ray.core.app.webrtc.LocalWebRTCListener.packet_encoding:type_name -> v2ray.core.net.packetaddr.PacketAddrType
	0,  // 1: v2ray.core.app.webrtc.LocalWebRTCListener.turn_servers:type_name -> v2ray.core.app.webrtc.WebRTCTURNServer
	12, // 2: v2ray.core.app.webrtc.ClientConfig.round_tripper_client:type_name -> google.protobuf.Any
	12, // 3: v2ray.core.app.webrtc.ClientConfig.security_config:type_name -> google.protobuf.Any
	12, // 4: v2ray.core.app.webrtc.ServerInverseRoleConfig.round_tripper_client:type_name -> google.protobuf.Any
	12, // 5: v2ray.core.app.webrtc.ServerInverseRoleConfig.security_config:type_name -> google.protobuf.Any
	12, // 6: v2ray.core.app.webrtc.Acceptor.server_config:type_name -> google.protobuf.Any
	5,  // 7: v2ray.core.app.webrtc.Acceptor.port_forwarder_accepts:type_name -> v2ray.core.app.webrtc.UDPPortForwarderAcceptor
	12, // 8: v2ray.core.app.webrtc.Remote.client_config:type_name -> google.protobuf.Any
	6,  // 9: v2ray.core.app.webrtc.RemoteConnections.port_forward:type_name -> v2ray.core.app.webrtc.UDPPortForwarderRemote
	1,  // 10: v2ray.core.app.webrtc.Config.listener:type_name -> v2ray.core.app.webrtc.LocalWebRTCListener
	2,  // 11: v2ray.core.app.webrtc.Config.system_listener:type_name -> v2ray.core.app.webrtc.LocalWebRTCSystemListener
	7,  // 12: v2ray.core.app.webrtc.Config.acceptors:type_name -> v2ray.core.app.webrtc.Acceptor
	8,  // 13: v2ray.core.app.webrtc.Config.remotes:type_name -> v2ray.core.app.webrtc.Remote
	9,  // 14: v2ray.core.app.webrtc.Config.connection:type_name -> v2ray.core.app.webrtc.RemoteConnections
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_app_webrtc_config_proto_init() }
func file_app_webrtc_config_proto_init() {
	if File_app_webrtc_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_app_webrtc_config_proto_rawDesc), len(file_app_webrtc_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_app_webrtc_config_proto_goTypes,
		DependencyIndexes: file_app_webrtc_config_proto_depIdxs,
		MessageInfos:      file_app_webrtc_config_proto_msgTypes,
	}.Build()
	File_app_webrtc_config_proto = out.File
	file_app_webrtc_config_proto_goTypes = nil
	file_app_webrtc_config_proto_depIdxs = nil
}
