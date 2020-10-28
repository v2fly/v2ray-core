package main

//go:generate go run v2ray.com/core/common/errors/errorgen

import (
	"flag"
	"fmt"
	"os"

	"v2ray.com/core/infra/control/command"
	_ "v2ray.com/core/main/distro/all"
)

type null struct{}

func (n *null) Write(p []byte) (int, error) {
	return len(p), nil
}

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

	version := false
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&version, "v", false, "Short alias of -version")
	fs.BoolVar(&version, "version", false, "Show current version of V2Ray.")
	// parse silently, no usage, no error output
	fs.Usage = func() {}
	fs.SetOutput(&null{})

	if err := fs.Parse(os.Args[1:]); err == flag.ErrHelp {
		cmd = command.GetCommand("help")
		command.Execute(cmd, nil)
		os.Exit(0)
	}

	if version {
		cmd = command.GetCommand("version")
		command.Execute(cmd, nil)
		os.Exit(0)
	}

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
