package all

import (
	"v2ray.com/core/commands/all/tlscmd"
	"v2ray.com/core/commands/base"
)

var cmdTLS = &base.Command{
	UsageLine: "{{.Exec}} tls",
	Short:     "TLS tools",
	Long: `{{.Exec}} tls provides tools for TLS.
	`,

	Commands: []*base.Command{
		tlscmd.CmdCert,
		tlscmd.CmdPing,
	},
}
