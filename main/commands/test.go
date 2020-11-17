package commands

import (
	"fmt"
	"log"

	"v2ray.com/core"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/common/cmdarg"
)

// CmdTest tests config files
var CmdTest = &base.Command{
	UsageLine: "{{.Exec}} test [-format=json] [-c config.json] [-confdir dir]",
	Short:     "Test config files",
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

var (
	testConfigFiles cmdarg.Arg // "Config file for V2Ray.", the option is customed type, parse in main
	testConfigDir   string
	testFormat      = CmdTest.Flag.String("format", "json", "Format of input file.")

	/* We have to do this here because Golang's Test will also need to parse flag, before
	 * main func in this file is run.
	 */
	_ = func() bool {

		CmdTest.Flag.Var(&testConfigFiles, "config", "Config path for V2Ray.")
		CmdTest.Flag.Var(&testConfigFiles, "c", "Short alias of -config")
		CmdTest.Flag.StringVar(&testConfigDir, "confdir", "", "A dir with multiple json config")

		return true
	}()
)

func executeTest(cmd *base.Command, args []string) {
	if dirExists(testConfigDir) {
		log.Println("Using confdir from arg:", runConfigDir)
		testConfigFiles = append(testConfigFiles, readConfDir(testConfigDir)...)
	}
	if len(testConfigFiles) == 0 {
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
	config, err := core.LoadConfig(getConfigFormat(), testConfigFiles[0], testConfigFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", testConfigFiles.String(), "]").Base(err)
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}
