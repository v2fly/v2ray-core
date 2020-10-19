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

		fmt.Println(command.ExecutableName, "<command>")
		fmt.Println("Available commands:")
		command.PrintUsage()
		os.Exit(-1)
		return
	}
	command.ExecuteCommand(cmd)
}
