package conf_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/common/net"
	"github.com/v2fly/v2ray-core/common/protocol"
	"github.com/v2fly/v2ray-core/common/serial"
	. "github.com/v2fly/v2ray-core/infra/conf"
	"github.com/v2fly/v2ray-core/proxy/shadowsocks"
)

func TestShadowsocksServerConfigParsing(t *testing.T) {
	creator := func() Buildable {
		return new(ShadowsocksServerConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"method": "aes-128-cfb",
				"password": "v2ray-password"
			}`,
			Parser: loadJSON(creator),
			Output: &shadowsocks.ServerConfig{
				User: &protocol.User{
					Account: serial.ToTypedMessage(&shadowsocks.Account{
						CipherType: shadowsocks.CipherType_AES_128_CFB,
						Password:   "v2ray-password",
					}),
				},
				Network: []net.Network{net.Network_TCP},
			},
		},
	})
}
