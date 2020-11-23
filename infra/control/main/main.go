package main

import (
	_ "v2ray.com/core/commands/all"
	"v2ray.com/core/commands/base"
)

func main() {
	base.RootCommand.Long = "A tool set for V2Ray."
	base.Execute()
}
