package jsonv4

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/main/commands/all/api"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	"github.com/v2fly/v2ray-core/v5/main/commands/helpers"
)

var cmdAddInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api adi [--server=127.0.0.1:8080] [c1.json] [dir1]...",
	Short:       "add inbounds",
	Long: `
Add inbounds to V2Ray.

> Make sure you have "HandlerService" set in "config.api.services" 
of server config.

Arguments:

	-format <format>
		The input format.
		Available values: "auto", "json", "toml", "yaml"
		Default: "auto"

	-r
		Load folders recursively.

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} dir
    {{.Exec}} {{.LongName}} c1.json c2.yaml
`,
	Run: executeAddInbounds,
}

func executeAddInbounds(cmd *base.Command, args []string) {
	api.SetSharedFlags(cmd)
	api.SetSharedConfigFlags(cmd)
	cmd.Flag.Parse(args)
	c, err := helpers.LoadConfig(cmd.Flag.Args(), api.APIConfigFormat, api.APIConfigRecursively)
	if err != nil {
		base.Fatalf("failed to load: %s", err)
	}
	if len(c.InboundConfigs) == 0 {
		base.Fatalf("no valid inbound found")
	}

	conn, ctx, close := api.DialAPIServer()
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
