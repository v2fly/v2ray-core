package v4_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
	"github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
)

func TestDokodemoConfig(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(v4.DokodemoConfig)
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"address": "8.8.8.8",
				"port": 53,
				"network": "tcp",
				"timeout": 10,
				"followRedirect": true,
				"userLevel": 1
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &dokodemo.Config{
				Address: &net.IPOrDomain{
					Address: &net.IPOrDomain_Ip{
						Ip: []byte{8, 8, 8, 8},
					},
				},
				Port:           53,
				Networks:       []net.Network{net.Network_TCP},
				Timeout:        10,
				FollowRedirect: true,
				UserLevel:      1,
			},
		},
	})
}
