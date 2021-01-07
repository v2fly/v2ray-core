package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	routerService "v2ray.com/core/app/router/command"
	"v2ray.com/core/main/commands/base"
)

var cmdHealthInfo = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api hci [--server=127.0.0.1:8080] [balancerTag]...",
	Short:       "get health check infomation",
	Long: `
Get health check infomation of specified balancers. If no 
balancer tag specified, get infomation of all balancers.

> Make sure you have "RouterService" set in "config.api.services" 
of server config.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3
`,
	Run: executeHealthInfo,
}

func executeHealthInfo(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	r := &routerService.GetHealthInfoRequest{BalancerTags: cmd.Flag.Args()}
	resp, err := client.GetHealthInfo(ctx, r)
	if err != nil {
		base.Fatalf("failed to get health information: %s", err)
	}
	for _, b := range resp.Balancers {
		showBalancerHealth(b)
	}
}

func showBalancerHealth(b *routerService.BalancerHealth) {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("Balancer: %s\n", b.Tag))
	sb.WriteString("  - Selects:\n")
	const format = "    %-4d %-14s %s\n"
	for i, o := range b.Selects {
		sb.WriteString(getHealthLine(format, i+1, o))
	}
	sb.WriteString("  - Outbounds:\n")
	for i, o := range b.Outbounds {
		sb.WriteString(getHealthLine(format, i+1, o))
	}
	os.Stdout.WriteString(sb.String())
}

func getHealthLine(format string, index int, o *routerService.OutboundHealth) string {
	switch {
	case o.RTT < 0:
		return fmt.Sprintf(format, index, "failed", o.Tag)
	case o.RTT == 0:
		return fmt.Sprintf(format, index, "not checked", o.Tag)
	default:
		return fmt.Sprintf(format, index, time.Duration(o.RTT), o.Tag)
	}
}
