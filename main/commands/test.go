package commands

import (
	"fmt"
	"log"

	"github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

// CmdTest tests config files
var CmdTest = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} test [-format=json] [-c config.json] [-d dir]",
	Short:       "test config files",
	Long: `
Test config files, without launching V2Ray server.

Arguments:

	-c, -config <file>
		Config file for V2Ray. Multiple assign is accepted.

	-d, -confdir <dir>
		A dir with config files. Multiple assign is accepted.

	-r
		Load confdir recursively.

	-format <format>
		Format of input files. (default "json")

Examples:

	{{.Exec}} {{.LongName}} -c config.json
	{{.Exec}} {{.LongName}} -d path/to/dir

Use "{{.Exec}} help format-loader" for more information about format.
	`,
	Run: executeTest,
}

func executeTest(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	cmd.Flag.Parse(args)

	extension, err := core.GetLoaderExtensions(*configFormat)
	if err != nil {
		base.Fatalf(err.Error())
	}

	if len(configDirs) > 0 {
		dirReader := readConfDir
		if *configDirRecursively {
			dirReader = readConfDirRecursively
		}
		for _, d := range configDirs {
			log.Println("Using confdir from arg:", d)
			configFiles = append(configFiles, dirReader(d, extension)...)
		}
	}
	printVersion()
	_, err = startV2Ray()
	if err != nil {
		base.Fatalf("Test failed: %s", err)
	}
	fmt.Println("Configuration OK.")
}
