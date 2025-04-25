package jsonv4

import (
	"fmt"

	handlerService "github.com/ghxhy/v2ray-core/v5/app/proxyman/command"
	"github.com/ghxhy/v2ray-core/v5/main/commands/all/api"
	"github.com/ghxhy/v2ray-core/v5/main/commands/base"
	"github.com/ghxhy/v2ray-core/v5/main/commands/helpers"
)

var cmdRemoveOutbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmo [--server=127.0.0.1:8080] [c1.json] [dir1]...",
	Short:       "remove outbounds",
	Long: `
Remove outbounds from V2Ray.

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
	Run: executeRemoveOutbounds,
}

func executeRemoveOutbounds(cmd *base.Command, args []string) {
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
		for _, c := range c.OutboundConfigs {
			tags = append(tags, c.Tag)
		}
	}
	if len(tags) == 0 {
		base.Fatalf("no outbound to remove")
	}

	conn, ctx, close := api.DialAPIServer()
	defer close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, tag := range tags {
		fmt.Println("removing:", tag)
		r := &handlerService.RemoveOutboundRequest{
			Tag: tag,
		}
		_, err := client.RemoveOutbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove outbound: %s", err)
		}
	}
}
