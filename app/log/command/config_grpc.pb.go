package command

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	LoggerService_RestartLogger_FullMethodName = "/v2ray.core.app.log.command.LoggerService/RestartLogger"
	LoggerService_FollowLog_FullMethodName     = "/v2ray.core.app.log.command.LoggerService/FollowLog"
)

// LoggerServiceClient is the client API for LoggerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LoggerServiceClient interface {
	RestartLogger(ctx context.Context, in *RestartLoggerRequest, opts ...grpc.CallOption) (*RestartLoggerResponse, error)
	// Unstable interface
	FollowLog(ctx context.Context, in *FollowLogRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[FollowLogResponse], error)
}

type loggerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLoggerServiceClient(cc grpc.ClientConnInterface) LoggerServiceClient {
	return &loggerServiceClient{cc}
}

func (c *loggerServiceClient) RestartLogger(ctx context.Context, in *RestartLoggerRequest, opts ...grpc.CallOption) (*RestartLoggerResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RestartLoggerResponse)
	err := c.cc.Invoke(ctx, LoggerService_RestartLogger_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *loggerServiceClient) FollowLog(ctx context.Context, in *FollowLogRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[FollowLogResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &LoggerService_ServiceDesc.Streams[0], LoggerService_FollowLog_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[FollowLogRequest, FollowLogResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type LoggerService_FollowLogClient = grpc.ServerStreamingClient[FollowLogResponse]

// LoggerServiceServer is the server API for LoggerService service.
// All implementations must embed UnimplementedLoggerServiceServer
// for forward compatibility.
type LoggerServiceServer interface {
	RestartLogger(context.Context, *RestartLoggerRequest) (*RestartLoggerResponse, error)
	// Unstable interface
	FollowLog(*FollowLogRequest, grpc.ServerStreamingServer[FollowLogResponse]) error
	mustEmbedUnimplementedLoggerServiceServer()
}

// UnimplementedLoggerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLoggerServiceServer struct{}

func (UnimplementedLoggerServiceServer) RestartLogger(context.Context, *RestartLoggerRequest) (*RestartLoggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RestartLogger not implemented")
}
func (UnimplementedLoggerServiceServer) FollowLog(*FollowLogRequest, grpc.ServerStreamingServer[FollowLogResponse]) error {
	return status.Errorf(codes.Unimplemented, "method FollowLog not implemented")
}
func (UnimplementedLoggerServiceServer) mustEmbedUnimplementedLoggerServiceServer() {}
func (UnimplementedLoggerServiceServer) testEmbeddedByValue()                       {}

// UnsafeLoggerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LoggerServiceServer will
// result in compilation errors.
type UnsafeLoggerServiceServer interface {
	mustEmbedUnimplementedLoggerServiceServer()
}

func RegisterLoggerServiceServer(s grpc.ServiceRegistrar, srv LoggerServiceServer) {
	// If the following call pancis, it indicates UnimplementedLoggerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
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
		FullMethod: LoggerService_RestartLogger_FullMethodName,
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
	return srv.(LoggerServiceServer).FollowLog(m, &grpc.GenericServerStream[FollowLogRequest, FollowLogResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type LoggerService_FollowLogServer = grpc.ServerStreamingServer[FollowLogResponse]

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
