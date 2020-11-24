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

var cmdAddInbounds = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api adi [--server=127.0.0.1:8080] <c1.json> [c2.json]...",
	Short:       "Add inbounds",
	Long: `
Add inbounds to V2Ray by calling its API. (timeout 3 seconds)

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
		cmd.Usage()
		base.SetExitStatus(1)
		base.Exit()
	}

	ins := make([]conf.InboundDetourConfig, 0)
	for _, arg := range args {
		conf, err := jsonToConfig(arg)
		if err != nil {
			base.Fatalf("failed to read %s: %s", arg, err)
		}
		ins = append(ins, conf.InboundConfigs...)
	}
	if len(ins) == 0 {
		base.Fatalf("no valid inbound found in %s", args)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	resps := make([]string, 0)
	for _, in := range ins {
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
		msg := responeseToString(resp)
		if msg != "" {
			resps = append(resps, msg)
		}
	}
	showResponese(strings.Join(resps, "\n"))
}
