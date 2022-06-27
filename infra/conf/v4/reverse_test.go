package v4_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/reverse"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
)

func TestReverseConfig(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(v4.ReverseConfig)
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"bridges": [{
					"tag": "test",
					"domain": "test.v2fly.org"
				}]
			}`,
			Parser: testassist.LoadJSON(creator),
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
			Parser: testassist.LoadJSON(creator),
			Output: &reverse.Config{
				PortalConfig: []*reverse.PortalConfig{
					{Tag: "test", Domain: "test.v2fly.org"},
				},
			},
		},
	})
}
