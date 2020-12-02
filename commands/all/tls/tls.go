package tls

import (
	"github.com/v2fly/v2ray-core/v4/commands/base"
)

// CmdTLS holds all tls sub commands
var CmdTLS = &base.Command{
	UsageLine: "{{.Exec}} tls",
	Short:     "TLS tools",
	Long: `{{.Exec}} {{.LongName}} provides tools for TLS.
	`,

	Commands: []*base.Command{
		cmdCert,
		cmdPing,
	},
}
