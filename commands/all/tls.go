package all

import (
	"github.com/v2fly/v2ray-core/v4/commands/all/tlscmd"
	"github.com/v2fly/v2ray-core/v4/commands/base"
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
