package jsonv4

import (
	"fmt"

	handlerService "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/main/commands/all/api"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	"github.com/v2fly/v2ray-core/v5/main/commands/helpers"
)

var cmdRemoveInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmi [--server=127.0.0.1:8080] [c1.json] [dir1]...",
	Short:       "remove inbounds",
	Long: `
Remove inbounds from V2Ray.

> Make sure you have "HandlerService" set in "config.api.services" 
of server config.

Arguments:

	-format <format>
		The input format.
		Available values: "auto", "json", "toml", "yaml"
		Default: "auto"

	-r
		Load folders recursively.

	-tags
		The input are tags instead of config files

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} dir
    {{.Exec}} {{.LongName}} c1.json c2.yaml
    {{.Exec}} {{.LongName}} -tags tag1 tag2
`,
	Run: executeRemoveInbounds,
}

func executeRemoveInbounds(cmd *base.Command, args []string) {
	api.SetSharedFlags(cmd)
	api.SetSharedConfigFlags(cmd)
	isTags := cmd.Flag.Bool("tags", false, "")
	cmd.Flag.Parse(args)

	var tags []string
	if *isTags {
		tags = cmd.Flag.Args()
	} else {
		c, err := helpers.LoadConfig(cmd.Flag.Args(), api.APIConfigFormat, api.APIConfigRecursively)
		if err != nil {
			base.Fatalf("failed to load: %s", err)
		}
		tags = make([]string, 0)
		for _, c := range c.InboundConfigs {
			tags = append(tags, c.Tag)
		}
	}
	if len(tags) == 0 {
		base.Fatalf("no inbound to remove")
	}

	conn, ctx, close := api.DialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, tag := range tags {
		fmt.Println("removing:", tag)
		r := &handlerService.RemoveInboundRequest{
			Tag: tag,
		}
		_, err := client.RemoveInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove inbound: %s", err)
		}
	}
}
