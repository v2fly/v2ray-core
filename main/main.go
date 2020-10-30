package main

import (
	"v2ray.com/core/commands/base"
	_ "v2ray.com/core/main/distro/all"
)

func main() {
	base.RootCommand.Long = "A unified platform for anti-censorship."
	base.RegisterCommand(cmdRun)
	base.RegisterCommand(cmdVersion)
	base.SortLessFunc = func(i, j *base.Command) bool {
		left := i.Name()
		right := j.Name()
		if left == "run" {
			return true
		}
		if right == "run" {
			return false
		}
		return left < right
	}
	base.SortCommands()
	base.Execute()
}
