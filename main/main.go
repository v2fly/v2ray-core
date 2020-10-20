package main

//go:generate go run v2ray.com/core/common/errors/errorgen

import (
	"fmt"
	"os"

	"v2ray.com/core/infra/control/command"
	_ "v2ray.com/core/main/distro/all"
)

func getCommandName() string {
	if len(os.Args) > 1 {
		name := os.Args[1]
		if name[0] != '-' {
			return name
		}
	}
	return ""
}

func main() {
	var cmd command.Command
	name := getCommandName()
	if name != "" {
		cmd = command.GetCommand(name)
		if cmd == nil {
			fmt.Fprintln(os.Stderr, "Unknown command:", name)
			fmt.Fprintln(os.Stderr)

			command.PrintUsage()
			os.Exit(-1)
		}
		command.Execute(cmd, os.Args[2:])
	} else {
		// default command
		cmd = command.GetCommand("run")
		command.Execute(cmd, os.Args[1:])
	}
}
