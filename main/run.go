package main

import (
	"flag"
	"fmt"
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
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/common/platform"
	"v2ray.com/core/infra/control/command"
)

var (
	configFiles cmdarg.Arg // "Config file for V2Ray.", the option is customed type, parse in main
	configDir   string
	fs          = flag.NewFlagSet("run", flag.ContinueOnError)
	version     = fs.Bool("version", false, "Show current version of V2Ray.")
	test        = fs.Bool("test", false, "Test config file only, without launching V2Ray server.")
	format      = fs.String("format", "json", "Format of input file.")

	/* We have to do this here because Golang's Test will also need to parse flag, before
	 * main func in this file is run.
	 */
	_ = func() bool {

		fs.Var(&configFiles, "config", "Config path for V2Ray.")
		fs.Var(&configFiles, "c", "Short alias of -config")
		fs.StringVar(&configDir, "confdir", "", "A dir with multiple json config")

		return true
	}()
)

type runCommand struct{}

// Name of the command
func (c *runCommand) Name() string {
	return "run"
}

// Description of the command
func (c *runCommand) Description() command.Description {
	return command.Description{
		Short: "run v2ray, the default command",
		Usage: []string{
			"",
			"'run' is the default command, the two command lines below are equivalent:",
			"",
			fmt.Sprintf("  %s %s -c config.json", command.ExecutableName, c.Name()),
			fmt.Sprintf("  %s -c config.json", command.ExecutableName),
			"",
			"More examples:",
			"",
			fmt.Sprintf("  %s -c config.json -c c1.json -c <url>.json -confdir <dir>", command.ExecutableName),
			fmt.Sprintf("  %s -test -c config.json", command.ExecutableName),
			fmt.Sprintf("  %s -version", command.ExecutableName),
			"",
			fmt.Sprintf("For all available commands, run '%s help'", command.ExecutableName),
		},
	}
}

// Execute the command
func (c *runCommand) Execute(args []string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}

	printVersion()

	if *version {
		os.Exit(0)
	}

	server, err := startV2Ray()
	if err != nil {
		fmt.Println(err)
		// Configuration error. Exit with a special value to prevent systemd from restarting.
		os.Exit(23)
	}

	if *test {
		fmt.Println("Configuration OK.")
		os.Exit(0)
	}

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start", err)
		os.Exit(-1)
	}
	defer server.Close()

	// Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
	}
	return nil
}

func init() {
	common.Must(command.RegisterCommand(&runCommand{}))
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

func readConfDir(dirPath string) {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range confs {
		if strings.HasSuffix(f.Name(), ".json") {
			configFiles.Set(path.Join(dirPath, f.Name()))
		}
	}
}

func getConfigFilePath() cmdarg.Arg {
	if dirExists(configDir) {
		log.Println("Using confdir from arg:", configDir)
		readConfDir(configDir)
	} else if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
		log.Println("Using confdir from env:", envConfDir)
		readConfDir(envConfDir)
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

func getConfigFormat() string {
	switch strings.ToLower(*format) {
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

func printVersion() {
	version := core.VersionStatement()
	for _, s := range version {
		fmt.Println(s)
	}
}
