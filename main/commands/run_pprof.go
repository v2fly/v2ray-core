//go:build pprof
// +build pprof

package commands

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

// CmdRun runs V2Ray with config
var CmdRun = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} run [-c config.json] [-d dir] [-pprof port]",
	Short:       "run V2Ray with config",
	Long: `
Run V2Ray with config.

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

	-pprof <port>
		HTTP server port for pprof debugging. (default "8080")

Examples:

	{{.Exec}} {{.LongName}} -c config.json
	{{.Exec}} {{.LongName}} -d path/to/dir

Use "{{.Exec}} help format-loader" for more information about format.
	`,
	Run: executeRun,
}

func executeRun(cmd *base.Command, args []string) {
	setConfigFlags(cmd)
	pprof := cmd.Flag.String("pprof", ":8080", "")
	cmd.Flag.Parse(args)
	printVersion()
	configFiles = getConfigFilePath()
	server, err := startV2Ray()
	if err != nil {
		base.Fatalf("Failed to start: %s", err)
	}

	if addr := *pprof; addr != "" {
		go func() {
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Print(err)
			}
		}()
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
