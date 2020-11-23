package commands

import (
	"fmt"
	"log"

	"github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/commands/base"
)

// CmdTest tests config files
var CmdTest = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} test [-format=json] [-c config.json] [-confdir dir]",
	Short:       "Test config files",
	Long: `
Test config files, without launching V2Ray server.

Example:

	{{.Exec}} {{.LongName}} -c config.json

Arguments:

	-c value
		Short alias of -config

	-config value
		Config file for V2Ray. Multiple assign is accepted (only
		json). Latter ones overrides the former ones.

	-confdir string
		A dir with multiple json config

	-format string
		Format of input files. (default "json")
	`,
}

func init() {
	CmdTest.Run = executeTest //break init loop
}

func executeTest(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	cmd.Flag.Parse(args)
	if dirExists(configDir) {
		log.Println("Using confdir from arg:", configDir)
		configFiles = append(configFiles, readConfDir(configDir)...)
	}
	if len(configFiles) == 0 {
		cmd.Flag.Usage()
		base.SetExitStatus(1)
		base.Exit()
	}
	printVersion()
	_, err := startV2RayTesting()
	if err != nil {
		base.Fatalf("Test failed: %s", err)
	}
	fmt.Println("Configuration OK.")
}

func startV2RayTesting() (core.Server, error) {
	config, err := core.LoadConfig(getFormatFromAlias(), configFiles[0], configFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", configFiles.String(), "]").Base(err)
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}
