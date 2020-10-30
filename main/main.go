package main

import (
	"v2ray.com/core/commands/base"
	_ "v2ray.com/core/main/distro/all"
)

func main() {
	base.RootCommand.Long = "A unified platform for anti-censorship."
	base.RootCommand.Commands = append(
		[]*base.Command{
			cmdRun,
			cmdVersion,
		},
		base.RootCommand.Commands...,
	)
	base.Execute()
}
