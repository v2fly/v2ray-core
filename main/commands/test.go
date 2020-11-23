package commands

import (
	"fmt"
	"log"

	"v2ray.com/core"
	"v2ray.com/core/commands/base"
)

// CmdTest tests config files
var CmdTest = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} test [-format=json] [-c config.json] [-d dir]",
	Short:       "Test config files",
	Long: `
Test config files, without launching V2Ray server.

Example:

	{{.Exec}} {{.LongName}} -c config.json
	{{.Exec}} {{.LongName}} -d path/to/json_dir

Arguments:

	-c, -config
		Config file for V2Ray. Multiple assign is accepted.

	-d, -confdir
		A dir with multiple json config. Multiple assign is accepted.

	-r
		Load confdir recursively.

	-format
		Format of input files. (default "json")

Use "{{.Exec}} help format-loader" for more information about format.
	`,
	Run: executeTest,
}

func executeTest(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	cmd.Flag.Parse(args)

	extension, err := getLoaderExtension()
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
	if len(configFiles) == 0 {
		if len(configDirs) == 0 {
			cmd.Flag.Usage()
			base.SetExitStatus(1)
			base.Exit()
		}
		base.Fatalf("no config file found with extension: %s", extension)
	}
	printVersion()
	_, err = startV2RayTesting()
	if err != nil {
		base.Fatalf("Test failed: %s", err)
	}
	fmt.Println("Configuration OK.")
}

func startV2RayTesting() (core.Server, error) {
	config, err := core.LoadConfig(*configFormat, configFiles[0], configFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", configFiles.String(), "]").Base(err)
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}
