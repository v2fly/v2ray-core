package conf_test

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/runtime/protoiface"
	"v2ray.com/core/app/stats"
	. "v2ray.com/core/infra/conf"
)

func TestChannelConfig(t *testing.T) {
	createParser := func() func(string) (protoiface.MessageV1, error) {
		return func(s string) (protoiface.MessageV1, error) {
			config := new(ChannelConfig)
			if err := json.Unmarshal([]byte(s), config); err != nil {
				return nil, err
			}
			return config.Build()
		}
	}
	runMultiTestCase(t, []TestCase{
		{
			Input:  `{}`,
			Parser: createParser(),
			Output: (*stats.ChannelConfig)(nil),
		},
		{
			Input: `{
				"enabled": true
			}`,
			Parser: createParser(),
			Output: &stats.ChannelConfig{
				SubscriberLimit:  1,
				BufferSize:       16,
				BroadcastTimeout: 100,
			},
		},
		{
			Input: `{
				"enabled": true,
				"subscriberLimit": 0
			}`,
			Parser: createParser(),
			Output: &stats.ChannelConfig{
				SubscriberLimit:  0,
				BufferSize:       16,
				BroadcastTimeout: 100,
			},
		},
		{
			Input: `{
				"subscriberLimit": 0
			}`,
			Parser: createParser(),
			Output: (*stats.ChannelConfig)(nil),
		},
	})
}
