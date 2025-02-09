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
	SubscriptionManagerService_ListTrackedSubscription_FullMethodName      = "/v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService/ListTrackedSubscription"
	SubscriptionManagerService_AddTrackedSubscription_FullMethodName       = "/v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService/AddTrackedSubscription"
	SubscriptionManagerService_RemoveTrackedSubscription_FullMethodName    = "/v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService/RemoveTrackedSubscription"
	SubscriptionManagerService_GetTrackedSubscriptionStatus_FullMethodName = "/v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService/GetTrackedSubscriptionStatus"
	SubscriptionManagerService_UpdateTrackedSubscription_FullMethodName    = "/v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService/UpdateTrackedSubscription"
)

// SubscriptionManagerServiceClient is the client API for SubscriptionManagerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SubscriptionManagerServiceClient interface {
	ListTrackedSubscription(ctx context.Context, in *ListTrackedSubscriptionRequest, opts ...grpc.CallOption) (*ListTrackedSubscriptionResponse, error)
	AddTrackedSubscription(ctx context.Context, in *AddTrackedSubscriptionRequest, opts ...grpc.CallOption) (*AddTrackedSubscriptionResponse, error)
	RemoveTrackedSubscription(ctx context.Context, in *RemoveTrackedSubscriptionRequest, opts ...grpc.CallOption) (*RemoveTrackedSubscriptionResponse, error)
	GetTrackedSubscriptionStatus(ctx context.Context, in *GetTrackedSubscriptionStatusRequest, opts ...grpc.CallOption) (*GetTrackedSubscriptionStatusResponse, error)
	UpdateTrackedSubscription(ctx context.Context, in *UpdateTrackedSubscriptionRequest, opts ...grpc.CallOption) (*UpdateTrackedSubscriptionResponse, error)
}

type subscriptionManagerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSubscriptionManagerServiceClient(cc grpc.ClientConnInterface) SubscriptionManagerServiceClient {
	return &subscriptionManagerServiceClient{cc}
}

func (c *subscriptionManagerServiceClient) ListTrackedSubscription(ctx context.Context, in *ListTrackedSubscriptionRequest, opts ...grpc.CallOption) (*ListTrackedSubscriptionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListTrackedSubscriptionResponse)
	err := c.cc.Invoke(ctx, SubscriptionManagerService_ListTrackedSubscription_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionManagerServiceClient) AddTrackedSubscription(ctx context.Context, in *AddTrackedSubscriptionRequest, opts ...grpc.CallOption) (*AddTrackedSubscriptionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddTrackedSubscriptionResponse)
	err := c.cc.Invoke(ctx, SubscriptionManagerService_AddTrackedSubscription_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionManagerServiceClient) RemoveTrackedSubscription(ctx context.Context, in *RemoveTrackedSubscriptionRequest, opts ...grpc.CallOption) (*RemoveTrackedSubscriptionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RemoveTrackedSubscriptionResponse)
	err := c.cc.Invoke(ctx, SubscriptionManagerService_RemoveTrackedSubscription_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionManagerServiceClient) GetTrackedSubscriptionStatus(ctx context.Context, in *GetTrackedSubscriptionStatusRequest, opts ...grpc.CallOption) (*GetTrackedSubscriptionStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTrackedSubscriptionStatusResponse)
	err := c.cc.Invoke(ctx, SubscriptionManagerService_GetTrackedSubscriptionStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionManagerServiceClient) UpdateTrackedSubscription(ctx context.Context, in *UpdateTrackedSubscriptionRequest, opts ...grpc.CallOption) (*UpdateTrackedSubscriptionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateTrackedSubscriptionResponse)
	err := c.cc.Invoke(ctx, SubscriptionManagerService_UpdateTrackedSubscription_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SubscriptionManagerServiceServer is the server API for SubscriptionManagerService service.
// All implementations must embed UnimplementedSubscriptionManagerServiceServer
// for forward compatibility.
type SubscriptionManagerServiceServer interface {
	ListTrackedSubscription(context.Context, *ListTrackedSubscriptionRequest) (*ListTrackedSubscriptionResponse, error)
	AddTrackedSubscription(context.Context, *AddTrackedSubscriptionRequest) (*AddTrackedSubscriptionResponse, error)
	RemoveTrackedSubscription(context.Context, *RemoveTrackedSubscriptionRequest) (*RemoveTrackedSubscriptionResponse, error)
	GetTrackedSubscriptionStatus(context.Context, *GetTrackedSubscriptionStatusRequest) (*GetTrackedSubscriptionStatusResponse, error)
	UpdateTrackedSubscription(context.Context, *UpdateTrackedSubscriptionRequest) (*UpdateTrackedSubscriptionResponse, error)
	mustEmbedUnimplementedSubscriptionManagerServiceServer()
}

// UnimplementedSubscriptionManagerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSubscriptionManagerServiceServer struct{}

func (UnimplementedSubscriptionManagerServiceServer) ListTrackedSubscription(context.Context, *ListTrackedSubscriptionRequest) (*ListTrackedSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTrackedSubscription not implemented")
}
func (UnimplementedSubscriptionManagerServiceServer) AddTrackedSubscription(context.Context, *AddTrackedSubscriptionRequest) (*AddTrackedSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddTrackedSubscription not implemented")
}
func (UnimplementedSubscriptionManagerServiceServer) RemoveTrackedSubscription(context.Context, *RemoveTrackedSubscriptionRequest) (*RemoveTrackedSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveTrackedSubscription not implemented")
}
func (UnimplementedSubscriptionManagerServiceServer) GetTrackedSubscriptionStatus(context.Context, *GetTrackedSubscriptionStatusRequest) (*GetTrackedSubscriptionStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTrackedSubscriptionStatus not implemented")
}
func (UnimplementedSubscriptionManagerServiceServer) UpdateTrackedSubscription(context.Context, *UpdateTrackedSubscriptionRequest) (*UpdateTrackedSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTrackedSubscription not implemented")
}
func (UnimplementedSubscriptionManagerServiceServer) mustEmbedUnimplementedSubscriptionManagerServiceServer() {
}
func (UnimplementedSubscriptionManagerServiceServer) testEmbeddedByValue() {}

// UnsafeSubscriptionManagerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SubscriptionManagerServiceServer will
// result in compilation errors.
type UnsafeSubscriptionManagerServiceServer interface {
	mustEmbedUnimplementedSubscriptionManagerServiceServer()
}

func RegisterSubscriptionManagerServiceServer(s grpc.ServiceRegistrar, srv SubscriptionManagerServiceServer) {
	// If the following call pancis, it indicates UnimplementedSubscriptionManagerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SubscriptionManagerService_ServiceDesc, srv)
}

func _SubscriptionManagerService_ListTrackedSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTrackedSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionManagerServiceServer).ListTrackedSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubscriptionManagerService_ListTrackedSubscription_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionManagerServiceServer).ListTrackedSubscription(ctx, req.(*ListTrackedSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionManagerService_AddTrackedSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddTrackedSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionManagerServiceServer).AddTrackedSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubscriptionManagerService_AddTrackedSubscription_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionManagerServiceServer).AddTrackedSubscription(ctx, req.(*AddTrackedSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionManagerService_RemoveTrackedSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveTrackedSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionManagerServiceServer).RemoveTrackedSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubscriptionManagerService_RemoveTrackedSubscription_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionManagerServiceServer).RemoveTrackedSubscription(ctx, req.(*RemoveTrackedSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionManagerService_GetTrackedSubscriptionStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTrackedSubscriptionStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionManagerServiceServer).GetTrackedSubscriptionStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubscriptionManagerService_GetTrackedSubscriptionStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionManagerServiceServer).GetTrackedSubscriptionStatus(ctx, req.(*GetTrackedSubscriptionStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionManagerService_UpdateTrackedSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTrackedSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionManagerServiceServer).UpdateTrackedSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SubscriptionManagerService_UpdateTrackedSubscription_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionManagerServiceServer).UpdateTrackedSubscription(ctx, req.(*UpdateTrackedSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SubscriptionManagerService_ServiceDesc is the grpc.ServiceDesc for SubscriptionManagerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SubscriptionManagerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v2ray.core.app.subscription.subscriptionmanager.command.SubscriptionManagerService",
	HandlerType: (*SubscriptionManagerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListTrackedSubscription",
			Handler:    _SubscriptionManagerService_ListTrackedSubscription_Handler,
		},
		{
			MethodName: "AddTrackedSubscription",
			Handler:    _SubscriptionManagerService_AddTrackedSubscription_Handler,
		},
		{
			MethodName: "RemoveTrackedSubscription",
			Handler:    _SubscriptionManagerService_RemoveTrackedSubscription_Handler,
		},
		{
			MethodName: "GetTrackedSubscriptionStatus",
			Handler:    _SubscriptionManagerService_GetTrackedSubscriptionStatus_Handler,
		},
		{
			MethodName: "UpdateTrackedSubscription",
			Handler:    _SubscriptionManagerService_UpdateTrackedSubscription_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "app/subscription/subscriptionmanager/command/command.proto",
}
