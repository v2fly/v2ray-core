package command

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LoggerServiceClient is the client API for LoggerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LoggerServiceClient interface {
	RestartLogger(ctx context.Context, in *RestartLoggerRequest, opts ...grpc.CallOption) (*RestartLoggerResponse, error)
	// Unstable interface
	FollowLog(ctx context.Context, in *FollowLogRequest, opts ...grpc.CallOption) (LoggerService_FollowLogClient, error)
}

type loggerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLoggerServiceClient(cc grpc.ClientConnInterface) LoggerServiceClient {
	return &loggerServiceClient{cc}
}

func (c *loggerServiceClient) RestartLogger(ctx context.Context, in *RestartLoggerRequest, opts ...grpc.CallOption) (*RestartLoggerResponse, error) {
	out := new(RestartLoggerResponse)
	err := c.cc.Invoke(ctx, "/v2ray.core.app.log.command.LoggerService/RestartLogger", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *loggerServiceClient) FollowLog(ctx context.Context, in *FollowLogRequest, opts ...grpc.CallOption) (LoggerService_FollowLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &LoggerService_ServiceDesc.Streams[0], "/v2ray.core.app.log.command.LoggerService/FollowLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &loggerServiceFollowLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type LoggerService_FollowLogClient interface {
	Recv() (*FollowLogResponse, error)
	grpc.ClientStream
}

type loggerServiceFollowLogClient struct {
	grpc.ClientStream
}

func (x *loggerServiceFollowLogClient) Recv() (*FollowLogResponse, error) {
	m := new(FollowLogResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LoggerServiceServer is the server API for LoggerService service.
// All implementations must embed UnimplementedLoggerServiceServer
// for forward compatibility
type LoggerServiceServer interface {
	RestartLogger(context.Context, *RestartLoggerRequest) (*RestartLoggerResponse, error)
	// Unstable interface
	FollowLog(*FollowLogRequest, LoggerService_FollowLogServer) error
	mustEmbedUnimplementedLoggerServiceServer()
}

// UnimplementedLoggerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedLoggerServiceServer struct {
}

func (UnimplementedLoggerServiceServer) RestartLogger(context.Context, *RestartLoggerRequest) (*RestartLoggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RestartLogger not implemented")
}
func (UnimplementedLoggerServiceServer) FollowLog(*FollowLogRequest, LoggerService_FollowLogServer) error {
	return status.Errorf(codes.Unimplemented, "method FollowLog not implemented")
}
func (UnimplementedLoggerServiceServer) mustEmbedUnimplementedLoggerServiceServer() {}

// UnsafeLoggerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LoggerServiceServer will
// result in compilation errors.
type UnsafeLoggerServiceServer interface {
	mustEmbedUnimplementedLoggerServiceServer()
}

func RegisterLoggerServiceServer(s grpc.ServiceRegistrar, srv LoggerServiceServer) {
	s.RegisterService(&LoggerService_ServiceDesc, srv)
}

func _LoggerService_RestartLogger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartLoggerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LoggerServiceServer).RestartLogger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v2ray.core.app.log.command.LoggerService/RestartLogger",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LoggerServiceServer).RestartLogger(ctx, req.(*RestartLoggerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LoggerService_FollowLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(FollowLogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LoggerServiceServer).FollowLog(m, &loggerServiceFollowLogServer{stream})
}

type LoggerService_FollowLogServer interface {
	Send(*FollowLogResponse) error
	grpc.ServerStream
}

type loggerServiceFollowLogServer struct {
	grpc.ServerStream
}

func (x *loggerServiceFollowLogServer) Send(m *FollowLogResponse) error {
	return x.ServerStream.SendMsg(m)
}

// LoggerService_ServiceDesc is the grpc.ServiceDesc for LoggerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LoggerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v2ray.core.app.log.command.LoggerService",
	HandlerType: (*LoggerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RestartLogger",
			Handler:    _LoggerService_RestartLogger_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "FollowLog",
			Handler:       _LoggerService_FollowLog_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app/log/command/config.proto",
}
