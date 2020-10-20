package main

import (
	"fmt"
	"os"

	commlog "v2ray.com/core/common/log"
	// _ "v2ray.com/core/infra/conf/command"
	"v2ray.com/core/infra/control/command"
	_ "v2ray.com/core/infra/control/commands"
)

func getCommandName() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return ""
}

func main() {
	// let the v2ctl prints log at stderr
	commlog.RegisterHandler(commlog.NewLogger(commlog.CreateStderrLogWriter()))
	name := getCommandName()
	cmd := command.GetCommand(name)
	if cmd == nil {
		fmt.Fprintln(os.Stderr, "Unknown command:", name)
		fmt.Fprintln(os.Stderr)

		command.PrintUsage()
		os.Exit(-1)
		return
	}
	command.Execute(cmd, os.Args[2:])
}
