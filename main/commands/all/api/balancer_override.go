package api

import (
	"time"

	routerService "github.com/v2fly/v2ray-core/v4/app/router/command"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

var cmdBalancerOverride = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api bo [--server=127.0.0.1:8080] <-b balancer> selectors...",
	Short:       "balancer override",
	Long: `
Override a balancer's selection in a duration of time.

> Make sure you have "RoutingService" set in "config.api.services" 
of server config.

Once a balancer's selection is overridden:

- The selectors of the balancer won't apply.
- The strategy of the balancer stops selecting qualified nodes 
  according to its settings, doing only the final pick.

Arguments:

	-b, -balancer <tag>
		Tag of the target balancer. Required.

	-v, -validity <duration>
		Time duration of the validity of override, e.g.: 60s, 60m, 
		24h, 1m30s. Default 1h.

	-r, -remove
		Remove the override

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 -b balancer selector1 selector2
    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 -b balancer -r
`,
	Run: executeBalancerOverride,
}

func executeBalancerOverride(cmd *base.Command, args []string) {
	var (
		balancer string
		validity time.Duration
		remove   bool
	)
	cmd.Flag.StringVar(&balancer, "b", "", "")
	cmd.Flag.StringVar(&balancer, "balancer", "", "")
	cmd.Flag.DurationVar(&validity, "v", time.Hour, "")
	cmd.Flag.DurationVar(&validity, "validity", time.Hour, "")
	cmd.Flag.BoolVar(&remove, "r", false, "")
	cmd.Flag.BoolVar(&remove, "remove", false, "")
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	if balancer == "" {
		base.Fatalf("balancer tag not specified")
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	v := int64(0)
	if !remove {
		v = int64(validity)
	}
	client := routerService.NewRoutingServiceClient(conn)
	r := &routerService.OverrideSelectingRequest{
		BalancerTag: balancer,
		Selectors:   cmd.Flag.Args(),
		Validity:    v,
	}
	_, err := client.OverrideSelecting(ctx, r)
	if err != nil {
		base.Fatalf("failed to override balancer: %s", err)
	}
}
