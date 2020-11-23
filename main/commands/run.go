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

	"github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/commands/base"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/common/platform"
)

// CmdRun runs V2Ray with config
var CmdRun = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} run [-c config.json] [-confdir dir]",
	Short:       "Run V2Ray with config",
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
	configFiles  cmdarg.Arg // "Config file for V2Ray.", the option is customed type
	configDir    string
	configFormat *string
)

func setConfigFlags(cmd *base.Command) {
	configFormat = cmd.Flag.String("format", "", "")

	cmd.Flag.Var(&configFiles, "config", "")
	cmd.Flag.Var(&configFiles, "c", "")
	cmd.Flag.StringVar(&configDir, "confdir", "", "")
}
func executeRun(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	cmd.Flag.Parse(args)
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
	if dirExists(configDir) {
		log.Println("Using confdir from arg:", configDir)
		configFiles = append(configFiles, readConfDir(configDir)...)
	} else if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
		log.Println("Using confdir from env:", envConfDir)
		configFiles = append(configFiles, readConfDir(envConfDir)...)
	}
	if len(configFiles) > 0 {
		return configFiles
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

func getFormatFromAlias() string {
	switch strings.ToLower(*configFormat) {
	case "pb":
		return "protobuf"
	default:
		return *configFormat
	}
}

func startV2Ray() (core.Server, error) {
	configFiles := getConfigFilePath()

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
