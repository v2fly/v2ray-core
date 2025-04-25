package scenarios

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	xproxy "golang.org/x/net/proxy"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/dispatcher"
	"github.com/ghxhy/v2ray-core/v5/app/log"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	_ "github.com/ghxhy/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/ghxhy/v2ray-core/v5/app/proxyman/outbound"
	"github.com/ghxhy/v2ray-core/v5/app/router"
	"github.com/ghxhy/v2ray-core/v5/common"
	clog "github.com/ghxhy/v2ray-core/v5/common/log"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
	"github.com/ghxhy/v2ray-core/v5/common/uuid"
	"github.com/ghxhy/v2ray-core/v5/proxy/blackhole"
	"github.com/ghxhy/v2ray-core/v5/proxy/dokodemo"
	"github.com/ghxhy/v2ray-core/v5/proxy/freedom"
	v2http "github.com/ghxhy/v2ray-core/v5/proxy/http"
	"github.com/ghxhy/v2ray-core/v5/proxy/socks"
	"github.com/ghxhy/v2ray-core/v5/proxy/vmess"
	"github.com/ghxhy/v2ray-core/v5/proxy/vmess/inbound"
	"github.com/ghxhy/v2ray-core/v5/proxy/vmess/outbound"
	"github.com/ghxhy/v2ray-core/v5/testing/servers/tcp"
	"github.com/ghxhy/v2ray-core/v5/testing/servers/udp"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

func TestPassiveConnection(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
		SendFirst:    []byte("send first"),
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
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
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(serverPort),
	})
	common.Must(err)

	{
		response := make([]byte, 1024)
		nBytes, err := conn.Read(response)
		common.Must(err)
		if string(response[:nBytes]) != "send first" {
			t.Error("unexpected first response: ", string(response[:nBytes]))
		}
	}

	if err := testTCPConn2(conn, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}
}

func TestProxy(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverUserID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: serverUserID.String(),
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

	proxyUserID := protocol.NewID(uuid.New())
	proxyPort := tcp.PickPort()
	proxyConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(proxyPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: proxyUserID.String(),
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
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: serverUserID.String(),
									}),
								},
							},
						},
					},
				}),
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					ProxySettings: &internet.ProxyConfig{
						Tag: "proxy",
					},
				}),
			},
			{
				Tag: "proxy",
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(proxyPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: proxyUserID.String(),
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, proxyConfig, clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}
}

func TestProxyOverKCP(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverUserID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
					StreamSettings: &internet.StreamConfig{
						Protocol: internet.TransportProtocol_MKCP,
					},
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: serverUserID.String(),
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

	proxyUserID := protocol.NewID(uuid.New())
	proxyPort := tcp.PickPort()
	proxyConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(proxyPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&inbound.Config{
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id: proxyUserID.String(),
							}),
						},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					StreamSettings: &internet.StreamConfig{
						Protocol: internet.TransportProtocol_MKCP,
					},
				}),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientConfig := &core.Config{
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
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: serverUserID.String(),
									}),
								},
							},
						},
					},
				}),
				SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{
					ProxySettings: &internet.ProxyConfig{
						Tag: "proxy",
					},
					StreamSettings: &internet.StreamConfig{
						Protocol: internet.TransportProtocol_MKCP,
					},
				}),
			},
			{
				Tag: "proxy",
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(proxyPort),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: proxyUserID.String(),
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, proxyConfig, clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}
}

func TestBlackhole(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	tcpServer2 := tcp.Server{
		MsgProcessor: xor,
	}
	dest2, err := tcpServer2.Start()
	common.Must(err)
	defer tcpServer2.Close()

	serverPort := tcp.PickPort()
	serverPort2 := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
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
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort2),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest2.Address),
					Port:    uint32(dest2.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "direct",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
			{
				Tag:           "blocked",
				ProxySettings: serial.ToTypedMessage(&blackhole.Config{}),
			},
		},
		App: []*anypb.Any{
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						TargetTag: &router.RoutingRule_Tag{
							Tag: "blocked",
						},
						PortRange: net.SinglePortRange(dest2.Port),
					},
				},
			}),
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testTCPConn(serverPort2, 1024, time.Second*5)(); err == nil {
		t.Error("nil error")
	}
}

func TestForward(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&socks.ServerConfig{
					AuthType: socks.AuthType_NO_AUTH,
					Accounts: map[string]string{
						"Test Account": "Test Password",
					},
					Address:    net.NewIPOrDomain(net.LocalHostIP),
					UdpEnabled: false,
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{
					DestinationOverride: &freedom.DestinationOverride{
						Server: &protocol.ServerEndpoint{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(dest.Port),
						},
					},
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		noAuthDialer, err := xproxy.SOCKS5("tcp", net.TCPDestination(net.LocalHostIP, serverPort).NetAddr(), nil, xproxy.Direct)
		common.Must(err)
		conn, err := noAuthDialer.Dial("tcp", "google.com:80")
		common.Must(err)
		defer conn.Close()

		if err := testTCPConn2(conn, 1024, time.Second*5)(); err != nil {
			t.Error(err)
		}
	}
}

func TestUDPConnection(t *testing.T) {
	udpServer := udp.Server{
		MsgProcessor: xor,
	}
	dest, err := udpServer.Start()
	common.Must(err)
	defer udpServer.Close()

	clientPort := udp.PickPort()
	clientConfig := &core.Config{
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
						Network: []net.Network{net.Network_UDP},
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

	servers, err := InitializeServerConfigs(clientConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	if err := testUDPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}

	time.Sleep(20 * time.Second)

	if err := testUDPConn(clientPort, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}
}

func TestDomainSniffing(t *testing.T) {
	sniffingPort := tcp.PickPort()
	httpPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				Tag: "snif",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(sniffingPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
					DomainOverride: []proxyman.KnownProtocols{
						proxyman.KnownProtocols_TLS,
					},
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(net.LocalHostIP),
					Port:    443,
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
			{
				Tag: "http",
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(httpPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag: "redir",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{
					DestinationOverride: &freedom.DestinationOverride{
						Server: &protocol.ServerEndpoint{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(sniffingPort),
						},
					},
				}),
			},
			{
				Tag:           "direct",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
		App: []*anypb.Any{
			serial.ToTypedMessage(&router.Config{
				Rule: []*router.RoutingRule{
					{
						TargetTag: &router.RoutingRule_Tag{
							Tag: "direct",
						},
						InboundTag: []string{"snif"},
					}, {
						TargetTag: &router.RoutingRule_Tag{
							Tag: "redir",
						},
						InboundTag: []string{"http"},
					},
				},
			}),
			serial.ToTypedMessage(&log.Config{
				Error: &log.LogSpecification{Level: clog.Severity_Debug, Type: log.LogType_Console},
			}),
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + httpPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("https://www.github.com/")
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error("unexpected status code: ", resp.StatusCode)
		}
		common.Must(resp.Write(io.Discard))
	}
}

func TestDialV2Ray(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
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

	clientConfig := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
		Inbound: []*core.InboundHandlerConfig{},
		Outbound: []*core.OutboundHandlerConfig{
			{
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
											Type: protocol.SecurityType_AES128_GCM,
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

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	client, err := core.New(clientConfig)
	common.Must(err)

	conn, err := core.Dial(context.Background(), client, dest)
	common.Must(err)
	defer conn.Close()

	if err := testTCPConn2(conn, 1024, time.Second*5)(); err != nil {
		t.Error(err)
	}
}
