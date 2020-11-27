package api

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	handlerService "v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/conf/serial"
)

var cmdAddOutbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api ado [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "Add outbounds",
	Long: `
Add outbounds to V2Ray.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3

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
		fmt.Println("Reading from STDIN")
		unnamedArgs = []string{"stdin:"}
	}

	outs := make([]conf.OutboundDetourConfig, 0)
	for _, arg := range unnamedArgs {
		r, err := loadArg(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
		conf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode %s: %s", arg, err)
		}
		outs = append(outs, conf.OutboundConfigs...)
	}
	if len(outs) == 0 {
		base.Fatalf("no valid outbound found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, out := range outs {
		fmt.Println("adding:", out.Tag)
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
		showResponese(resp)
	}
}
