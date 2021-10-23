package scenarios

import (
	"context"
	"testing"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/dispatcher"
	"github.com/v2fly/v2ray-core/v4/app/log"
	"github.com/v2fly/v2ray-core/v4/app/policy"
	"github.com/v2fly/v2ray-core/v4/app/proxyman"
	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common"
	clog "github.com/v2fly/v2ray-core/v4/common/log"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol/ntp"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/dokodemo"
	"github.com/v2fly/v2ray-core/v4/proxy/freedom"
	ntpproxy "github.com/v2fly/v2ray-core/v4/proxy/ntp"
	"github.com/v2fly/v2ray-core/v4/testing/servers/udp"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func TestNTP(t *testing.T) {
	serverPort := udp.PickPort()
	serverConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: clog.Severity_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						InboundTag: []string{"ntp_in"},
						TargetTag: &router.RoutingRule_Tag{
							Tag: "ntp_out",
						},
					},
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&policy.Config{}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "ntp_in",
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(net.LocalHostIP),
					Port:     uint32(123),
					Networks: []net.Network{net.Network_UDP},
				}),
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
			{
				Tag:           "ntp_out",
				ProxySettings: serial.ToTypedMessage(&ntpproxy.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	conn, err := internet.DialSystem(context.Background(), net.Destination{
		Network: net.Network_UDP,
		Address: net.LocalHostIP,
		Port:    serverPort,
	}, nil)
	common.Must(err)
	defer conn.Close()

	message, time, err := ntp.Query(conn)
	common.Must(err)
	response := ntp.ParseTime(message, time)
	common.Must(response.Validate())
}
