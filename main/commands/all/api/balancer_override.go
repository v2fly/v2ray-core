package api

import (
	routerService "github.com/v2fly/v2ray-core/v5/app/router/command"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdBalancerOverride = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api bo [--server=127.0.0.1:8080] <-b balancer> outboundTag",
	Short:       "balancer override",
	Long: `
Override a balancer's selection.

> Make sure you have "RoutingService" set in "config.api.services" 
of server config.

Once a balancer's selection is overridden:

- The balancer's selection result will always be outboundTag

Arguments:

	-b, -balancer <tag>
		Tag of the target balancer. Required.

	-r, -remove
		Remove the override

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 -b balancer tag
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 -b balancer -r
`,
	Run: executeBalancerOverride,
}

func executeBalancerOverride(cmd *base.Command, args []string) {
	var (
		balancer string
		remove   bool
	)
	cmd.Flag.StringVar(&balancer, "b", "", "")
	cmd.Flag.StringVar(&balancer, "balancer", "", "")
	cmd.Flag.BoolVar(&remove, "r", false, "")
	cmd.Flag.BoolVar(&remove, "remove", false, "")
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	if balancer == "" {
		base.Fatalf("balancer tag not specified")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	target := ""
	if !remove {
		target = cmd.Flag.Args()[0]
	}
	r := &routerService.OverrideBalancerTargetRequest{
		BalancerTag: balancer,
		Target:      target,
	}

	_, err := client.OverrideBalancerTarget(ctx, r)
	if err != nil {
		base.Fatalf("failed to override balancer: %s", err)
	}
}
