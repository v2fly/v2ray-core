package command_test

import (
	"context"
	"testing"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/dispatcher"
	"github.com/v2fly/v2ray-core/v4/app/log"
	. "github.com/v2fly/v2ray-core/v4/app/log/command"
	"github.com/v2fly/v2ray-core/v4/app/proxyman"
	_ "github.com/v2fly/v2ray-core/v4/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v4/app/proxyman/outbound"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/serial"
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
