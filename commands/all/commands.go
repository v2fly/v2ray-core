package all

import "v2ray.com/core/commands/base"

// go:generate go run v2ray.com/core/common/errors/errorgen

func init() {
	base.RootCommand.Commands = append(
		base.RootCommand.Commands,
		cmdAPI,
		cmdConvert,
		cmdLove,
		cmdTLS,
		cmdUUID,
		cmdVerify,
		cmdMerge,

		// documents
		docFormat,
		docMerge,
	)
}
