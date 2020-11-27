package api

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	handlerService "v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/infra/conf/serial"
)

var cmdRemoveInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmi [--server=127.0.0.1:8080] <json_file|tag> [json_file] [tag]...",
	Short:       "Remove inbounds",
	Long: `
Remove inbounds from V2Ray.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json "tag name"
`,
	Run: executeRemoveInbounds,
}

func executeRemoveInbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) == 0 {
		fmt.Println("reading from stdin:")
		unnamedArgs = []string{"stdin:"}
	}

	tags := make([]string, 0)
	for _, arg := range unnamedArgs {
		if r, err := loadArg(arg); err == nil {
			conf, err := serial.DecodeJSONConfig(r)
			if err != nil {
				base.Fatalf("failed to decode %s: %s", arg, err)
			}
			ins := conf.InboundConfigs
			for _, i := range ins {
				tags = append(tags, i.Tag)
			}
		} else {
			// take request as tag
			tags = append(tags, arg)
		}
	}

	if len(tags) == 0 {
		base.Fatalf("no inbound to remove")
	}
	fmt.Println("removing inbounds:", tags)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, tag := range tags {
		fmt.Println("removing:", tag)
		r := &handlerService.RemoveInboundRequest{
			Tag: tag,
		}
		resp, err := client.RemoveInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove inbound: %s", err)
		}
		showResponese(resp)
	}
}
