package command_test

import (
	"context"
	"testing"

	"v2ray.com/core/v4"
	"v2ray.com/core/v4/app/dispatcher"
	"v2ray.com/core/v4/app/log"
	. "v2ray.com/core/v4/app/log/command"
	"v2ray.com/core/v4/app/proxyman"
	_ "v2ray.com/core/v4/app/proxyman/inbound"
	_ "v2ray.com/core/v4/app/proxyman/outbound"
	"v2ray.com/core/v4/common"
	"v2ray.com/core/v4/common/serial"
)

func TestLoggerRestart(t *testing.T) {
	v, err := core.New(&core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	})
	common.Must(err)
	common.Must(v.Start())

	server := &LoggerServer{
		V: v,
	}
	common.Must2(server.RestartLogger(context.Background(), &RestartLoggerRequest{}))
}
