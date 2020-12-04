package api

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

var cmdAddInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api adi [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "Add inbounds",
	Long: `
Add inbounds to V2Ray.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json c2.json
`,
	Run: executeAddInbounds,
}

func executeAddInbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) == 0 {
		fmt.Println("reading from stdin:")
		unnamedArgs = []string{"stdin:"}
	}

	ins := make([]conf.InboundDetourConfig, 0)
	for _, arg := range unnamedArgs {
		r, err := cmdarg.LoadArg(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
		conf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode %s: %s", arg, err)
		}
		ins = append(ins, conf.InboundConfigs...)
	}
	if len(ins) == 0 {
		base.Fatalf("no valid inbound found")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, in := range ins {
		fmt.Println("adding:", in.Tag)
		i, err := in.Build()
		if err != nil {
			base.Fatalf("failed to build conf: %s", err)
		}
		r := &handlerService.AddInboundRequest{
			Inbound: i,
		}
		resp, err := client.AddInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to add inbound: %s", err)
		}
		showResponese(resp)
	}
}
