package api

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"github.com/v2fly/v2ray-core/v4/main/commands/helpers"
)

var cmdAddInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api adi [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "add inbounds",
	Long: `
Add inbounds to V2Ray.

Arguments:

	-format <format>
		Specify the input format.
		Available values: "auto", "json", "toml", "yaml"
		Default: "auto"

	-r
		Load folders recursively.

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json c2.json
`,
	Run: executeAddInbounds,
}

func executeAddInbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	setSharedConfigFlags(cmd)
	cmd.Flag.Parse(args)
	c, err := helpers.LoadConfig(cmd.Flag.Args(), apiConfigFormat, apiConfigRecursively)
	if err != nil {
		base.Fatalf("%s", err)
	}
	if len(c.InboundConfigs) == 0 {
		base.Fatalf("no valid inbound found")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, in := range c.InboundConfigs {
		fmt.Println("adding:", in.Tag)
		i, err := in.Build()
		if err != nil {
			base.Fatalf("failed to build conf: %s", err)
		}
		r := &handlerService.AddInboundRequest{
			Inbound: i,
		}
		_, err = client.AddInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to add inbound: %s", err)
		}
	}
}
