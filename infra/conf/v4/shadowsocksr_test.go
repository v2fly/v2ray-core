package v4_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
	"github.com/v2fly/v2ray-core/v5/proxy/shadowsocksr"
)

func TestShadowsocksRServerConfigParsing(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(v4.ShadowsocksRServerConfig)
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"method": "aes-256-cfb",
				"password": "ssr-password",
				"protocol": "auth_aes128_md5",
				"protocol_param": "64",
				"obfs": "tls1.2_ticket_auth",
				"obfs_param": "cloudflare.com"
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &shadowsocksr.ServerConfig{
				User: &protocol.User{
					Account: serial.ToTypedMessage(&shadowsocksr.Account{
						CipherType:    shadowsocksr.CipherType_AES_256_CFB,
						Password:      "ssr-password",
						Protocol:      "auth_aes128_md5",
						ProtocolParam: "64",
						Obfs:         "tls1.2_ticket_auth",
						ObfsParam:    "cloudflare.com",
					}),
				},
				Network: []net.Network{net.Network_TCP},
			},
		},
		{
			Input: `{
				"method": "chacha20",
				"password": "ssr-password",
				"protocol": "auth_chain_a",
				"obfs": "http_simple"
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &shadowsocksr.ServerConfig{
				User: &protocol.User{
					Account: serial.ToTypedMessage(&shadowsocksr.Account{
						CipherType: shadowsocksr.CipherType_CHACHA20,
						Password:   "ssr-password",
						Protocol:  "auth_chain_a",
						Obfs:      "http_simple",
					}),
				},
				Network: []net.Network{net.Network_TCP},
			},
		},
		{
			Input: `{
				"method": "rc4-md5",
				"password": "ssr-password",
				"protocol": "origin",
				"obfs": "plain"
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &shadowsocksr.ServerConfig{
				User: &protocol.User{
					Account: serial.ToTypedMessage(&shadowsocksr.Account{
						CipherType: shadowsocksr.CipherType_RC4_MD5,
						Password:   "ssr-password",
						Protocol:  "origin",
						Obfs:      "plain",
					}),
				},
				Network: []net.Network{net.Network_TCP},
			},
		},
	})
}
