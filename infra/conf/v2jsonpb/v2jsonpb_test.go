package v2jsonpb_test

import (
	"strings"
	"testing"

	core "github.com/v2fly/v2ray-core/v5"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/infra/conf/v2jsonpb"
)

func TestLoadFieldNames(t *testing.T) {
	testCases := []struct {
		Input string
		Error string
	}{
		{
			Input: `{"inbound": [{"tag": "in", "receiverSettings": {"type": "v2ray.core.app.proxyman.ReceiverConfig"}}]}`,
		},
		{
			Input: `{"inbound": [{"tag": "in", "receiver_settings": {"type": "v2ray.core.app.proxyman.ReceiverConfig"}}]}`,
		},
		{
			Input: `{"status": {}}`,
			Error: `unknown field "status"`,
		},
		{
			Input: `{"inbound": [{"tag": "in", "status": {}}]}`,
			Error: `unknown field "status"`,
		},
		{
			Input: `{"inbound": [{"tag": "in", "receiverSettings": {"type": "v2ray.core.app.proxyman.ReceiverConfig", "value": {"status": {}}}}]}`,
			Error: `unknown field "status"`,
		},
	}
	for _, testCase := range testCases {
		config, err := core.LoadConfig(v2jsonpb.FormatProtobufV2JSONPB, []byte(testCase.Input))
		if testCase.Error != "" {
			if err == nil || !strings.Contains(err.Error(), testCase.Error) {
				t.Error("unexpected result from ", testCase.Input, ". expected error ", testCase.Error, ", but actually ", err)
			}
			continue
		}
		if err != nil {
			t.Error("failed to load ", testCase.Input, ": ", err)
			continue
		}
		if inbound := config.GetInbound(); len(inbound) != 1 || inbound[0].GetReceiverSettings() == nil {
			t.Error("unexpected config from ", testCase.Input, ": ", config)
		}
	}
}
