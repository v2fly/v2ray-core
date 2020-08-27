package conf_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/common/net"
	"github.com/v2fly/v2ray-core/common/protocol"
	"github.com/v2fly/v2ray-core/common/serial"
	. "github.com/v2fly/v2ray-core/infra/conf"
	"github.com/v2fly/v2ray-core/proxy/vless"
	"github.com/v2fly/v2ray-core/proxy/vless/inbound"
	"github.com/v2fly/v2ray-core/proxy/vless/outbound"
)

func TestVLessOutbound(t *testing.T) {
	creator := func() Buildable {
		return new(VLessOutboundConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"vnext": [{
					"address": "example.com",
					"port": 443,
					"users": [
						{
							"id": "27848739-7e62-4138-9fd3-098a63964b6b",
							"schedulers": "",
							"encryption": "none",
							"level": 0
						}
					]
				}]
			}`,
			Parser: loadJSON(creator),
			Output: &outbound.Config{
				Receiver: []*protocol.ServerEndpoint{
					{
						Address: &net.IPOrDomain{
							Address: &net.IPOrDomain_Domain{
								Domain: "example.com",
							},
						},
						Port: 443,
						User: []*protocol.User{
							{
								Account: serial.ToTypedMessage(&vless.Account{
									Id:         "27848739-7e62-4138-9fd3-098a63964b6b",
									Schedulers: "",
									Encryption: "none",
								}),
								Level: 0,
							},
						},
					},
				},
			},
		},
	})
}

func TestVLessInbound(t *testing.T) {
	creator := func() Buildable {
		return new(VLessInboundConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"clients": [
					{
						"id": "27848739-7e62-4138-9fd3-098a63964b6b",
						"schedulers": "",
						"level": 0,
						"email": "love@v2fly.org"
					}
				],
				"decryption": "none",
				"fallback": {
					"port": 80
				},
				"fallback_h2": {
					"unix": "@/dev/shm/domain.socket",
					"xver": 2
				}
			}`,
			Parser: loadJSON(creator),
			Output: &inbound.Config{
				User: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&vless.Account{
							Id:         "27848739-7e62-4138-9fd3-098a63964b6b",
							Schedulers: "",
						}),
						Level: 0,
						Email: "love@v2fly.org",
					},
				},
				Decryption: "none",
				Fallback: &inbound.Fallback{
					Addr: &net.IPOrDomain{
						Address: &net.IPOrDomain_Ip{
							Ip: []byte{127, 0, 0, 1},
						},
					},
					Port: 80,
				},
				FallbackH2: &inbound.FallbackH2{
					Addr: &net.IPOrDomain{
						Address: &net.IPOrDomain_Ip{
							Ip: []byte{127, 0, 0, 1},
						},
					},
					Unix: "\x00/dev/shm/domain.socket",
					Xver: 2,
				},
			},
		},
	})
}
