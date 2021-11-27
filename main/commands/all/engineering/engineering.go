package engineering

import "github.com/v2fly/v2ray-core/v4/main/commands/base"

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

var cmdEngineering = &base.Command{
	UsageLine: "{{.Exec}} engineering",
	Commands: []*base.Command{
		cmdConvertPb,
		cmdReversePb,
	},
}

func init() {
	base.RegisterCommand(cmdEngineering)
}
