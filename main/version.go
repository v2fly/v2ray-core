package main

import (
	"fmt"

	"v2ray.com/core"
	"v2ray.com/core/commands/base"
)

var cmdVersion = &base.Command{
	UsageLine: "{{.Exec}} version",
	Short:     "Print V2Ray Versions",
	Long: `Version prints the build information for V2Ray executables.
`,
	Run: executeVersion,
}

func executeVersion(cmd *base.Command, args []string) {
	printVersion()
}

func printVersion() {
	version := core.VersionStatement()
	for _, s := range version {
		fmt.Println(s)
	}
}
