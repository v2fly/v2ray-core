package api

import (
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

// CmdAPI calls an API in an V2Ray process
var CmdAPI = &base.Command{
	UsageLine: "{{.Exec}} api",
	Short:     "call V2Ray API",
	Long: `{{.Exec}} {{.LongName}} provides tools to manipulate V2Ray via its API.
`,
	Commands: []*base.Command{
		cmdLog,
		cmdStats,
		cmdBalancerInfo,
		cmdBalancerOverride,
	},
}
