package commands

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"v2ray.com/core"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/common/platform"
)

// CmdRun runs V2Ray with config
var CmdRun = &base.Command{
	UsageLine: "{{.Exec}} run [-c config.json] [-confdir dir]",
	Short:     "Run V2Ray with config",
	Long: `
Run V2Ray with config.

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
	CmdRun.Run = executeRun //break init loop
}

var (
	runConfigFiles cmdarg.Arg // "Config file for V2Ray.", the option is customed type, parse in main
	runConfigDir   string
	runFormat      = CmdRun.Flag.String("format", "json", "Format of input file.")

	/* We have to do this here because Golang's Test will also need to parse flag, before
	 * main func in this file is run.
	 */
	_ = func() bool {

		CmdRun.Flag.Var(&runConfigFiles, "config", "Config path for V2Ray.")
		CmdRun.Flag.Var(&runConfigFiles, "c", "Short alias of -config")
		CmdRun.Flag.StringVar(&runConfigDir, "confdir", "", "A dir with multiple json config")

		return true
	}()
)

func executeRun(cmd *base.Command, args []string) {
	printVersion()
	server, err := startV2Ray()
	if err != nil {
		base.Fatalf("Failed to start: %s", err)
	}

	if err := server.Start(); err != nil {
		base.Fatalf("Failed to start: %s", err)
	}
	defer server.Close()

	// Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
	}
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func dirExists(file string) bool {
	if file == "" {
		return false
	}
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

func readConfDir(dirPath string) cmdarg.Arg {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln(err)
	}
	files := make(cmdarg.Arg, 0)
	for _, f := range confs {
		if strings.HasSuffix(f.Name(), ".json") {
			files.Set(path.Join(dirPath, f.Name()))
		}
	}
	return files
}

func getConfigFilePath() cmdarg.Arg {
	if dirExists(runConfigDir) {
		log.Println("Using confdir from arg:", runConfigDir)
		runConfigFiles = append(runConfigFiles, readConfDir(runConfigDir)...)
	} else if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
		log.Println("Using confdir from env:", envConfDir)
		runConfigFiles = append(runConfigFiles, readConfDir(envConfDir)...)
	}
	if len(runConfigFiles) > 0 {
		return runConfigFiles
	}

	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			log.Println("Using default config: ", configFile)
			return cmdarg.Arg{configFile}
		}
	}

	if configFile := platform.GetConfigurationPath(); fileExists(configFile) {
		log.Println("Using config from env: ", configFile)
		return cmdarg.Arg{configFile}
	}

	log.Println("Using config from STDIN")
	return cmdarg.Arg{"stdin:"}
}

func getConfigFormat() string {
	switch strings.ToLower(*runFormat) {
	case "pb", "protobuf":
		return "protobuf"
	default:
		return "json"
	}
}

func startV2Ray() (core.Server, error) {
	configFiles := getConfigFilePath()

	config, err := core.LoadConfig(getConfigFormat(), configFiles[0], configFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", configFiles.String(), "]").Base(err)
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}
