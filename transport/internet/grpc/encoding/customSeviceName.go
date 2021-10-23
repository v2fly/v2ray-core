//go:build !confonly
// +build !confonly

package encoding

import (
	"context"

	"google.golang.org/grpc"

	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

type ConnHandler interface {
	HandleConn(internet.Connection)
}

func ServerDesc(name string) grpc.ServiceDesc {
	return grpc.ServiceDesc{
		ServiceName: name,
		HandlerType: (*GunServiceServer)(nil),
		Methods:     []grpc.MethodDesc{},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "Tun",
				Handler:       _GunService_Tun_Handler,
				ServerStreams: true,
				ClientStreams: true,
			},
			{
				StreamName:    "TunMulti",
				Handler:       _GunService_TunMulti_Handler,
				ServerStreams: true,
				ClientStreams: true,
			},
			{
				StreamName: "TunRaw",
				Handler: func(srv interface{}, stream grpc.ServerStream) error {
					conn, done := NewRawConn(stream)
					srv.(ConnHandler).HandleConn(conn)
					<-done
					return nil
				},
				ServerStreams: true,
				ClientStreams: true,
			},
		},
		Metadata: "gun.proto",
	}
}

func (c *gunServiceClient) TunCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (GunService_TunClient, error) {
	stream, err := c.cc.NewStream(ctx, &ServerDesc(name).Streams[0], "/"+name+"/Tun", opts...)
	if err != nil {
		return nil, err
	}
	x := &gunServiceTunClient{stream}
	return x, nil
}

func (c *gunServiceClient) TunMultiCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.cc.NewStream(ctx, &ServerDesc(name).Streams[1], "/"+name+"/TunMulti", opts...)
}

func (c *gunServiceClient) TunRawCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.cc.NewStream(ctx, &ServerDesc(name).Streams[2], "/"+name+"/TunRaw", opts...)
}

var _ GunServiceClientX = (*gunServiceClient)(nil)

type GunServiceClientX interface {
	TunCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (GunService_TunClient, error)
	TunMultiCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (grpc.ClientStream, error)
	TunRawCustomName(ctx context.Context, name string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

func RegisterGunServiceServerX(s *grpc.Server, srv GunServiceServer, name string) {
	desc := ServerDesc(name)
	s.RegisterService(&desc, srv)
}
