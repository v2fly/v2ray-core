package tls

import (
	"v2ray.com/core/main/commands/base"
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
