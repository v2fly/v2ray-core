package conf_test

import (
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"testing"

	"github.com/v2fly/v2ray-core/v4/app/reverse"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
)

func TestReverseConfig(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(conf.ReverseConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"bridges": [{
					"tag": "test",
					"domain": "test.v2fly.org"
				}]
			}`,
			Parser: loadJSON(creator),
			Output: &reverse.Config{
				BridgeConfig: []*reverse.BridgeConfig{
					{Tag: "test", Domain: "test.v2fly.org"},
				},
			},
		},
		{
			Input: `{
				"portals": [{
					"tag": "test",
					"domain": "test.v2fly.org"
				}]
			}`,
			Parser: loadJSON(creator),
			Output: &reverse.Config{
				PortalConfig: []*reverse.PortalConfig{
					{Tag: "test", Domain: "test.v2fly.org"},
				},
			},
		},
	})
}
