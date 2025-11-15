package engineering

import "github.com/v2fly/v2ray-core/v5/main/commands/base"

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

var cmdEngineering = &base.Command{
	UsageLine: "{{.Exec}} engineering",
	Commands: []*base.Command{
		cmdConvertPb,
		cmdReversePb,
		cmdNonNativeLinkExtract,
		cmdNonNativeLinkExec,
		cmdSubscriptionEntriesExtract,
		cmdEncodeDataURL,
	},
}

func init() {
	base.RegisterCommand(cmdEngineering)
}

func AddCommand(cmd *base.Command) {
	cmdEngineering.Commands = append(cmdEngineering.Commands, cmd)
}
