package v4_test

import (
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/socketcfg"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	"github.com/v2fly/v2ray-core/v5/transport/internet/headers/noop"
	"github.com/v2fly/v2ray-core/v5/transport/internet/headers/tls"
	"github.com/v2fly/v2ray-core/v5/transport/internet/kcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/quic"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/websocket"
)

func TestSocketConfig(t *testing.T) {
	createParser := func() func(string) (proto.Message, error) {
		return func(s string) (proto.Message, error) {
			config := new(socketcfg.SocketConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"mark": 1,
				"tcpFastOpen": true,
				"tcpFastOpenQueueLength": 1024,
				"mptcp": true
			}`,
			Parser: createParser(),
			Output: &internet.SocketConfig{
				Mark:           1,
				Tfo:            internet.SocketConfig_Enable,
				TfoQueueLength: 1024,
				Mptcp:          internet.MPTCPState_Enable,
			},
		},
	})
}

func TestTransportConfig(t *testing.T) {
	createParser := func() func(string) (proto.Message, error) {
		return func(s string) (proto.Message, error) {
			config := new(v4.TransportConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"tcpSettings": {
					"header": {
						"type": "http",
						"request": {
							"version": "1.1",
							"method": "GET",
							"path": "/b",
							"headers": {
								"a": "b",
								"c": "d"
							}
						},
						"response": {
							"version": "1.0",
							"status": "404",
							"reason": "Not Found"
						}
					}
				},
				"kcpSettings": {
					"mtu": 1200,
					"header": {
						"type": "none"
					}
				},
				"wsSettings": {
					"path": "/t"
				},
				"quicSettings": {
					"key": "abcd",
					"header": {
						"type": "dtls"
					}
				}
			}`,
			Parser: createParser(),
			Output: &transport.Config{
				TransportSettings: []*internet.TransportConfig{
					{
						ProtocolName: "tcp",
						Settings: serial.ToTypedMessage(&tcp.Config{
							HeaderSettings: serial.ToTypedMessage(&http.Config{
								Request: &http.RequestConfig{
									Version: &http.Version{Value: "1.1"},
									Method:  &http.Method{Value: "GET"},
									Uri:     []string{"/b"},
									Header: []*http.Header{
										{Name: "a", Value: []string{"b"}},
										{Name: "c", Value: []string{"d"}},
									},
								},
								Response: &http.ResponseConfig{
									Version: &http.Version{Value: "1.0"},
									Status:  &http.Status{Code: "404", Reason: "Not Found"},
									Header: []*http.Header{
										{
											Name:  "Content-Type",
											Value: []string{"application/octet-stream", "video/mpeg"},
										},
										{
											Name:  "Transfer-Encoding",
											Value: []string{"chunked"},
										},
										{
											Name:  "Connection",
											Value: []string{"keep-alive"},
										},
										{
											Name:  "Pragma",
											Value: []string{"no-cache"},
										},
										{
											Name:  "Cache-Control",
											Value: []string{"private", "no-cache"},
										},
									},
								},
							}),
						}),
					},
					{
						ProtocolName: "mkcp",
						Settings: serial.ToTypedMessage(&kcp.Config{
							Mtu:          &kcp.MTU{Value: 1200},
							HeaderConfig: serial.ToTypedMessage(&noop.Config{}),
						}),
					},
					{
						ProtocolName: "websocket",
						Settings: serial.ToTypedMessage(&websocket.Config{
							Path: "/t",
						}),
					},
					{
						ProtocolName: "quic",
						Settings: serial.ToTypedMessage(&quic.Config{
							Key: "abcd",
							Security: &protocol.SecurityConfig{
								Type: protocol.SecurityType_NONE,
							},
							Header: serial.ToTypedMessage(&tls.PacketConfig{}),
						}),
					},
				},
			},
		},
	})
}
