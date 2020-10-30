package main

import (
	"flag"
	"fmt"
	"os"

	"v2ray.com/core/commands/base"
	_ "v2ray.com/core/main/distro/all"
)

func main() {
	// TODO: Remove me for v5
	os.Args = getArgsV4Compatible()

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

func getArgsV4Compatible() []string {
	if len(os.Args) == 1 {
		return []string{os.Args[0], "run"}
	}
	if os.Args[1][0] != '-' {
		return os.Args
	}
	version := false
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&version, "version", false, "")
	// parse silently, no usage, no error output
	fs.Usage = func() {}
	fs.SetOutput(&null{})
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println("DEPRECATED: -h, WILL BE REMOVED IN V5.")
		fmt.Println("PLEASE USE: v2ray help")
		fmt.Println()
		return []string{os.Args[0], "help"}
		fmt.Println()
	}
	if version {
		fmt.Println("DEPRECATED: -version, WILL BE REMOVED IN V5.")
		fmt.Println("PLEASE USE: v2ray version")
		fmt.Println()
		return []string{os.Args[0], "version"}
	}
	fmt.Println("COMPATIBLE MODE, DEPRECATED.")
	fmt.Println("PLEASE USE: v2ray run [arguments] INSTEAD.")
	fmt.Println()
	return append([]string{os.Args[0], "run"}, os.Args[1:]...)
}

type null struct{}

func (n *null) Write(p []byte) (int, error) {
	return len(p), nil
}
