package conf_test

import (
	"testing"

	. "github.com/v2fly/v2ray-core/infra/conf"
	"github.com/v2fly/v2ray-core/proxy/http"
)

func TestHttpServerConfig(t *testing.T) {
	creator := func() Buildable {
		return new(HttpServerConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"timeout": 10,
				"accounts": [
					{
						"user": "my-username",
						"pass": "my-password"
					}
				],
				"allowTransparent": true,
				"userLevel": 1
			}`,
			Parser: loadJSON(creator),
			Output: &http.ServerConfig{
				Accounts: map[string]string{
					"my-username": "my-password",
				},
				AllowTransparent: true,
				UserLevel:        1,
				Timeout:          10,
			},
		},
	})
}
