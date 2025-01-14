package scenarios

import (
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common"
	clog "github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/uuid"
	"github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
	"github.com/v2fly/v2ray-core/v5/proxy/freedom"
	"github.com/v2fly/v2ray-core/v5/proxy/hysteria2"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/inbound"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"github.com/v2fly/v2ray-core/v5/testing/servers/tcp"
	"github.com/v2fly/v2ray-core/v5/testing/servers/udp"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	hyTransport "github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"
	tcpTransport "github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

func TestVMessHysteria2Congestion(t *testing.T) {
	for _, v := range []string{"bbr", "brutal"} {
		testVMessHysteria2(t, v)
	}
}

func testVMessHysteria2(t *testing.T, congestionType string) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := udp.PickPort()
	serverConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),

					StreamSettings: &internet.StreamConfig{
						ProtocolName: "hysteria2",
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									Certificate: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil))},
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								ProtocolName: "hysteria2",
								Settings: serial.ToTypedMessage(&hyTransport.Config{
									Congestion: &hyTransport.Congestion{Type: congestionType, UpMbps: 100, DownMbps: 100},
									Password:   "password",
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id:      userID.String(),
								AlterId: 0,
							}),
						},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					StreamSettings: &internet.StreamConfig{
						ProtocolName: "hysteria2",
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									ServerName:    "www.v2fly.org",
									AllowInsecure: true,
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								ProtocolName: "hysteria2",
								Settings: serial.ToTypedMessage(&hyTransport.Config{
									Congestion: &hyTransport.Congestion{Type: congestionType, UpMbps: 100, DownMbps: 100},
									Password:   "password",
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id:      userID.String(),
										AlterId: 0,
										SecuritySettings: &protocol.SecurityConfig{
											Type: protocol.SecurityType_NONE,
										},
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	if err != nil {
		t.Fatal("Failed to initialize all servers: ", err.Error())
	}
	defer CloseAllServers(servers)

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(testTCPConn(clientPort, 10240*1024, time.Second*40))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteria2Offical(t *testing.T) {
	for _, v := range []bool{true, false} {
		testHysteria2Offical(t, v)
	}
}

func testHysteria2Offical(t *testing.T, isUDP bool) {
	var dest net.Destination
	var err error
	if isUDP {
		udpServer := udp.Server{
			MsgProcessor: xor,
		}
		dest, err = udpServer.Start()
		common.Must(err)
		defer udpServer.Close()
	} else {
		tcpServer := tcp.Server{
			MsgProcessor: xor,
		}
		dest, err = tcpServer.Start()
		common.Must(err)
		defer tcpServer.Close()
	}

	serverPort := udp.PickPort()
	serverConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
					StreamSettings: &internet.StreamConfig{
						ProtocolName: "hysteria2",
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									Certificate: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil))},
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								ProtocolName: "hysteria2",
								Settings: serial.ToTypedMessage(&hyTransport.Config{
									Congestion:      &hyTransport.Congestion{Type: "brutal", UpMbps: 100, DownMbps: 100},
									UseUdpExtension: true,
									Password:        "password",
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&hysteria2.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP, net.Network_UDP},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					StreamSettings: &internet.StreamConfig{
						ProtocolName: "hysteria2",
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									ServerName:    "www.v2fly.org",
									AllowInsecure: true,
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								ProtocolName: "hysteria2",
								Settings: serial.ToTypedMessage(&hyTransport.Config{
									Congestion:      &hyTransport.Congestion{Type: "brutal", UpMbps: 100, DownMbps: 100},
									UseUdpExtension: true,
									Password:        "password",
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&hysteria2.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&hysteria2.Account{}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	if err != nil {
		t.Fatal("Failed to initialize all servers: ", err.Error())
	}
	defer CloseAllServers(servers)

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		if isUDP {
			errg.Go(testUDPConn(clientPort, 1500, time.Second*4))
		} else {
			errg.Go(testTCPConn(clientPort, 10240*1024, time.Second*40))
		}
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteria2OnTCP(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := udp.PickPort()
	serverConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
					StreamSettings: &internet.StreamConfig{
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									Certificate: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil))},
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								Protocol: internet.TransportProtocol_TCP,
								Settings: serial.ToTypedMessage(&tcpTransport.Config{
									HeaderSettings: serial.ToTypedMessage(&http.Config{}),
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&hysteria2.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					StreamSettings: &internet.StreamConfig{
						SecurityType: serial.GetMessageType(&tls.Config{}),
						SecuritySettings: []*anypb.Any{
							serial.ToTypedMessage(
								&tls.Config{
									ServerName:    "www.v2fly.org",
									AllowInsecure: true,
								},
							),
						},
						TransportSettings: []*internet.TransportConfig{
							{
								Protocol: internet.TransportProtocol_TCP,
								Settings: serial.ToTypedMessage(&tcpTransport.Config{
									HeaderSettings: serial.ToTypedMessage(&http.Config{}),
								}),
							},
						},
					},
				}),
				ProxySettings: serial.ToTypedMessage(&hysteria2.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&hysteria2.Account{}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	if err != nil {
		t.Fatal("Failed to initialize all servers: ", err.Error())
	}
	defer CloseAllServers(servers)

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(testTCPConn(clientPort, 10240*1024, time.Second*40))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
