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

var cmdAddInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api adi [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "Add inbounds",
	Long: `
Add inbounds to V2Ray.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 c1.json c2.json
`,
	Run: executeAddInbounds,
}

func executeAddInbounds(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)
	unnamedArgs := cmd.Flag.Args()
	if len(unnamedArgs) == 0 {
		fmt.Println("reading from stdin:")
		unnamedArgs = []string{"stdin:"}
	}

	ins := make([]conf.InboundDetourConfig, 0)
	for _, arg := range unnamedArgs {
		r, err := loadArg(arg)
		if err != nil {
			base.Fatalf("failed to load %s: %s", arg, err)
		}
		conf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode %s: %s", arg, err)
		}
		ins = append(ins, conf.InboundConfigs...)
	}
	if len(ins) == 0 {
		base.Fatalf("no valid inbound found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	for _, in := range ins {
		fmt.Println("adding:", in.Tag)
		i, err := in.Build()
		if err != nil {
			base.Fatalf("failed to build conf: %s", err)
		}
		r := &handlerService.AddInboundRequest{
			Inbound: i,
		}
		resp, err := client.AddInbound(ctx, r)
		if err != nil {
			base.Fatalf("failed to add inbound: %s", err)
		}
		showResponese(resp)
	}
}
