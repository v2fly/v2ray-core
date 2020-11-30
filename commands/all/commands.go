package all

import (
	"github.com/v2fly/v2ray-core/v4/commands/all/api"
	"github.com/v2fly/v2ray-core/v4/commands/all/tls"
	"github.com/v2fly/v2ray-core/v4/commands/base"
)

// go:generate go run v2ray.com/core/common/errors/errorgen

func init() {
	base.RootCommand.Commands = append(
		base.RootCommand.Commands,
		api.CmdAPI,
		cmdConvert,
		cmdLove,
		tls.CmdTLS,
		cmdUUID,
		cmdVerify,
		cmdMerge,

		// documents
		docFormat,
		docMerge,
	)
}
