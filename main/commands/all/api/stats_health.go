package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	routerService "v2ray.com/core/app/router/command"
	"v2ray.com/core/main/commands/base"
)

var cmdHealthStats = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api health [--server=127.0.0.1:8080]",
	Short:       "Get health statistics of balancers",
	Long: `
Get health statistics of balancers from V2Ray.

> Make sure you have "RouterService" set in "config.api.services" 
of server config.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3

	-b, -balancer
		Tag of the balancer to get statistics for. Get all 
		statistics if not specified.
`,
	Run: executeHealthStats,
}

func executeHealthStats(cmd *base.Command, args []string) {
	var tag string
	setSharedFlags(cmd)
	cmd.Flag.StringVar(&tag, "b", "", "")
	cmd.Flag.StringVar(&tag, "balancer", "", "")
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	r := &routerService.HealthStatsRequest{Tag: tag}
	resp, err := client.GetHealthStats(ctx, r)
	if err != nil {
		base.Fatalf("failed to get sys stats: %s", err)
	}
	for _, s := range resp.Stats {
		showHealthStats(s)
	}
}

func showHealthStats(stat *routerService.HealthStats) {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("Balancer: %s\n", stat.Balancer))
	sb.WriteString("  - Selects:\n")
	for i, o := range stat.Selects {
		sb.WriteString(getHealthStatsLine(i+1, o))
	}
	sb.WriteString("  - Outbounds:\n")
	for i, o := range stat.Outbounds {
		sb.WriteString(getHealthStatsLine(i+1, o))
	}
	os.Stdout.WriteString(sb.String())
}

func getHealthStatsLine(index int, item *routerService.HealthStatItem) string {
	if item.RTT == 0 {
		return fmt.Sprintf("    %-4d %-14s %s\n", index, "failed", item.Outbound)
	}
	return fmt.Sprintf("    %-4d %-14s %s\n", index, time.Duration(item.RTT), item.Outbound)
}
