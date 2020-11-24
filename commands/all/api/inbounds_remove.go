package api

import (
	"context"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	handlerService "v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/commands/base"
)

var cmdRemoveInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api rmi [--server=127.0.0.1:8080] <json_file|tag> [json_file] [tag]...",
	Short:       "Remove inbounds",
	Long: `
Remove inbounds from V2Ray by calling its API. (timeout 3 seconds)

Arguments:

	-server=127.0.0.1:8080 
		The API server address. Default 127.0.0.1:8080

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
		cmd.Usage()
		base.SetExitStatus(1)
		base.Exit()
	}

	tags := make([]string, 0)
	for _, arg := range unnamedArgs {
		if _, err := os.Stat(arg); err == nil || os.IsExist(err) {
			conf, err := jsonToConfig(arg)
			if err != nil {
				base.Fatalf("failed to read %s: %s", arg, err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", *apiServerAddrPtr)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	resps := make([]string, 0)
	for _, tag := range tags {
		r := &handlerService.RemoveInboundRequest{
			Tag: tag,
		}
		resp, err := client.RemoveInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to remove outbound: %s", err)
		}
		msg := responeseToString(resp)
		if msg != "" {
			resps = append(resps, msg)
		}
	}
	showResponese(strings.Join(resps, "\n"))
}
