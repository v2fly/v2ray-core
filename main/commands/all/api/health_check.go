package api

import (
	routerService "v2ray.com/core/app/router/command"
	"v2ray.com/core/main/commands/base"
)

var cmdHealthCheck = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api hc [--server=127.0.0.1:8080] [balancerTag]...",
	Short:       "perform health checks",
	Long: `
Perform health checks for specific balancers. if no balancer tag 
specified, check all balancers.

> Make sure you have "RouterService" set in "config.api.services" 
of server config.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3
`,
	Run: executeHealthCheck,
}

func executeHealthCheck(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	r := &routerService.CheckBalancersRequest{BalancerTags: cmd.Flag.Args()}
	_, err := client.CheckBalancers(ctx, r)
	if err != nil {
		base.Fatalf("failed to perform balancer health checks: %s", err)
	}
}
