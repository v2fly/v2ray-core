package config

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/trojan"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		comprehensiveServer := config.(*ServerConfig)
		fullServer := &trojan.ServerConfig{
			Users: func() (users []*protocol.User) {
				for _, v := range comprehensiveServer.Users {
					account := &trojan.Account{Password: v.Password}
					users = append(users, &protocol.User{
						Level:   v.Level,
						Email:   v.Email,
						Account: serial.ToTypedMessage(account),
					})
				}
				return
			}(),
			PacketEncoding: comprehensiveServer.PacketEncoding,
		}
		return common.CreateObject(ctx, fullServer)
	}))

	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		comprehensiveClient := config.(*ClientConfig)
		fullClient := &trojan.ClientConfig{
			Server: func() (servers []*protocol.ServerEndpoint) {
				for _, v := range comprehensiveClient.Servers {
					servers = append(servers, &protocol.ServerEndpoint{
						Address: v.Address,
						Port:    v.Port,
						User: []*protocol.User{
							{
								Account: serial.ToTypedMessage(&trojan.Account{Password: v.Password}),
							},
						},
					})
				}
				return
			}(),
		}
		return common.CreateObject(ctx, fullClient)
	}))
}
