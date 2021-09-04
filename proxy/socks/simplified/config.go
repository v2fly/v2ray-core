package simplified

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/registry"
	"github.com/v2fly/v2ray-core/v4/proxy/socks"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedServer := config.(*ServerConfig)
		fullServer := &socks.ServerConfig{
			AuthType:   socks.AuthType_NO_AUTH,
			Address:    simplifiedServer.Address,
			UdpEnabled: simplifiedServer.UdpEnabled,
		}
		return common.CreateObject(ctx, fullServer)
	}))
	common.Must(registry.RegisterImplementation(new(ServerConfig).ProtoReflect().Descriptor(), nil))

	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedClient := config.(*ClientConfig)
		fullClient := &socks.ClientConfig{
			Server: []*protocol.ServerEndpoint{
				{
					Address: simplifiedClient.Address,
					Port:    simplifiedClient.Port,
				},
			},
		}
		return common.CreateObject(ctx, fullClient)
	}))
	common.Must(registry.RegisterImplementation(new(ClientConfig).ProtoReflect().Descriptor(), nil))
}
