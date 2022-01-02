package commands

import (
	"fmt"

	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

// CmdTest tests config files
var CmdTest = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} test [-format=json] [-c config.json] [-d dir]",
	Short:       "test config files",
	Long: `
Test config files, without launching V2Ray server.

{{.Exec}} will also use the config directory specified by environment 
variable "v2ray.location.confdir". If no config found, it tries 
to load config from one of below:

	1. The default "config.json" in the current directory
	2. The config file from ENV "v2ray.location.config"
	3. The stdin if all failed above

Arguments:

	-c, -config <file>
		Config file for V2Ray. Multiple assign is accepted.

	-d, -confdir <dir>
		A directory with config files. Multiple assign is accepted.

	-r
		Load confdir recursively.

	-format <format>
		Format of config input. (default "auto")

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
	printVersion()
	configFiles = getConfigFilePath()
	_, err := startV2Ray()
	if err != nil {
		base.Fatalf("Test failed: %s", err)
	}
	fmt.Println("Configuration OK.")
}
