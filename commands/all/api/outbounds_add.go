package api

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	handlerService "v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/infra/conf"
)

var cmdAddOutbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api ado [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "Add outbounds",
	Long: `
Add outbounds to V2Ray by calling its API. (timeout 3 seconds)

Arguments:

	-server=127.0.0.1:8080 
		The API server address. Default 127.0.0.1:8080

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json c2.json
`,
	Run: executeAddOutbounds,
}

func executeAddOutbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) == 0 {
		cmd.Usage()
		base.SetExitStatus(1)
		base.Exit()
	}

	outs := make([]conf.OutboundDetourConfig, 0)
	for _, arg := range unnamedArgs {
		conf, err := jsonToConfig(arg)
		if err != nil {
			base.Fatalf("failed to read %s: %s", arg, err)
		}
		outs = append(outs, conf.OutboundConfigs...)
	}
	if len(outs) == 0 {
		base.Fatalf("no valid outbound found in %s", args)
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
	for _, out := range outs {
		o, err := out.Build()
		if err != nil {
			base.Fatalf("failed to build conf: %s", err)
		}
		r := &handlerService.AddOutboundRequest{
			Outbound: o,
		}
		resp, err := client.AddOutbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to add outbound: %s", err)
		}
		msg := responeseToString(resp)
		if msg != "" {
			resps = append(resps, msg)
		}
	}
	showResponese(strings.Join(resps, "\n"))
}
