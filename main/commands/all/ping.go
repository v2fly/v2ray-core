package all

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"net/url"

	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/infra/link"
	"v2ray.com/core/infra/vmessping"
	"v2ray.com/core/main/commands/base"
)

var cmdPing = &base.Command{
	UsageLine: "{{.Exec}} ping [argument]",
	Short:     "vmessping, a prober for V2Ray",
	Long: `
A ping prober for links and outbound config.

Arguments:

	-dest
		Test destination url, need 204 for success return. 
		Default "http://www.google.com/gen_204".

	-o
		Timeout seconds for each request. Default 10.

	-i
		Inteval seconds between pings. Default 1.

	-c
		Ping count before stop. Default 9999.

	-n
		Show the node location/outbound ip.

	-v
		Verbose mode (debug log).

	-m
		Use mux outbound.

	-q
		Fast quit on error counts.

> If multiple outbounds found in json file, the first one 
is used.

Examples:

    {{.Exec}} {{.LongName}} "vmess://..."
    {{.Exec}} {{.LongName}} outbound.json
`,
}

func init() {
	cmdPing.Run = executePing
}

var (
	pingVerbose  = cmdPing.Flag.Bool("v", false, "")
	pingShowNode = cmdPing.Flag.Bool("n", false, "")
	pingUsemux   = cmdPing.Flag.Bool("m", false, "")
	pingDesturl  = cmdPing.Flag.String("dest", "http://www.google.com/gen_204", "")
	pingCount    = cmdPing.Flag.Uint("c", 9999, "")
	pingTimeout  = cmdPing.Flag.Uint("o", 10, "")
	pingInteval  = cmdPing.Flag.Uint("i", 1, "")
	pingQuit     = cmdPing.Flag.Uint("q", 0, "")
)

func executePing(cmd *base.Command, args []string) {
	var arg string
	if cmdPing.Flag.NArg() == 0 {
		base.Fatalf("no ping target")
	}
	arg = cmdPing.Flag.Args()[0]

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	var (
		outbound *conf.OutboundDetourConfig
		err      error
	)
	var u *url.URL
	if u, err = url.Parse(arg); err == nil && u.Scheme != "" {
		lk, err := link.Parse(arg)
		if err != nil {
			base.Fatalf("failed to parse link: %s", err)
		}
		fmt.Println("\n" + lk.Detail())
		outbound = lk.ToOutbound()
	} else {
		outbound, err = json2Outbound(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
	}
	if *pingUsemux {
		outbound.MuxSettings = &conf.MuxConfig{}
		outbound.MuxSettings.Enabled = true
		outbound.MuxSettings.Concurrency = 8
	}
	vmessping.PrintVersion()
	ps, err := vmessping.Ping(outbound, *pingCount, *pingDesturl, *pingTimeout, *pingInteval, *pingQuit, osSignals, *pingShowNode, *pingVerbose)
	if err != nil {
		base.Fatalf("ping error: %s", err)
	}
	ps.PrintStats()
	if ps.IsErr() {
		base.Fatalf("")
	}
}

// json2Outbound load json from arg, returns the outbound config
func json2Outbound(arg string) (*conf.OutboundDetourConfig, error) {
	r, err := cmdarg.LoadArg(arg)
	if err != nil {
		return nil, err
	}
	c, err := serial.DecodeJSONConfig(r)
	if err != nil {
		return nil, err
	}
	if c.OutboundConfigs == nil || len(c.OutboundConfigs) == 0 {
		return nil, fmt.Errorf("no valid outbound found in %s", arg)
	}
	out := c.OutboundConfigs[0]
	return &out, nil
}
