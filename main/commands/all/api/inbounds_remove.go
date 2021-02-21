package api

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"github.com/v2fly/v2ray-core/v4/main/commands/helpers"
)

var cmdRemoveInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmi [--server=127.0.0.1:8080] <json_file|tag> [json_file] [tag]...",
	Short:       "remove inbounds",
	Long: `
Remove inbounds from V2Ray.

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

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json "tag name"
`,
	Run: executeRemoveInbounds,
}

func executeRemoveInbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	setSharedConfigFlags(cmd)
	cmd.Flag.Parse(args)
	c, err := helpers.LoadConfig(cmd.Flag.Args(), apiConfigFormat, apiConfigRecursively)
	if err != nil {
		base.Fatalf("%s", err)
	}
	if len(c.InboundConfigs) == 0 {
		base.Fatalf("no inbound to remove")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, c := range c.InboundConfigs {
		fmt.Println("removing:", c.Tag)
		r := &handlerService.RemoveInboundRequest{
			Tag: c.Tag,
		}
		_, err := client.RemoveInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove inbound: %s", err)
		}
	}
}
