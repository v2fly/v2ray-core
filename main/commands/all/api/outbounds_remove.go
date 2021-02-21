package api

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"github.com/v2fly/v2ray-core/v4/main/commands/helpers"
)

var cmdRemoveOutbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmo [--server=127.0.0.1:8080] <json_file|tag> [json_file] [tag]...",
	Short:       "remove outbounds",
	Long: `
Remove outbounds from V2Ray.

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
	Run: executeRemoveOutbounds,
}

func executeRemoveOutbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	setSharedConfigFlags(cmd)
	c, err := helpers.LoadConfig(cmd.Flag.Args(), apiConfigFormat, apiConfigRecursively)
	if err != nil {
		base.Fatalf("%s", err)
	}
	if len(c.OutboundConfigs) == 0 {
		base.Fatalf("no outbound to remove")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, c := range c.OutboundConfigs {
		fmt.Println("removing:", c.Tag)
		r := &handlerService.RemoveOutboundRequest{
			Tag: c.Tag,
		}
		_, err := client.RemoveOutbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove outbound: %s", err)
		}
	}
}
