package main

import (
	_ "github.com/v2fly/v2ray-core/v4/commands/all"
	"github.com/v2fly/v2ray-core/v4/commands/base"
)

func main() {
	base.RootCommand.Long = "A tool set for V2Ray."
	base.Execute()
}
