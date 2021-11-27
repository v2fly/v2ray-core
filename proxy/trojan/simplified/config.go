package simplified

import (
	"context"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/trojan"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedServer := config.(*ServerConfig)
		fullServer := &trojan.ServerConfig{
			Users: func() (users []*protocol.User) {
				for _, v := range simplifiedServer.Users {
					account := &trojan.Account{Password: v}
					users = append(users, &protocol.User{
						Account: serial.ToTypedMessage(account),
					})
				}
				return
			}(),
		}
		return common.CreateObject(ctx, fullServer)
	}))

	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedClient := config.(*ClientConfig)
		fullClient := &trojan.ClientConfig{
			Server: []*protocol.ServerEndpoint{
				{
					Address: simplifiedClient.Address,
					Port:    simplifiedClient.Port,
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&trojan.Account{Password: simplifiedClient.Password}),
						},
					},
				},
			},
		}
		return common.CreateObject(ctx, fullClient)
	}))
}
