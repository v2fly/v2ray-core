package engineering

import "github.com/ghxhy/v2ray-core/v5/main/commands/base"

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

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
