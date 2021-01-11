package api

import (
	"fmt"
	"os"
	"strings"

	routerService "v2ray.com/core/app/router/command"
	"v2ray.com/core/main/commands/base"
)

var cmdHealthInfo = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api hci [--server=127.0.0.1:8080] [balancerTag]...",
	Short:       "get health information",
	Long: `
Get health check information of specified balancers. If no 
balancer tag specified, get information of all balancers.

> Make sure you have "RoutingService" set in "config.api.services" 
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
		showBalancerInfo(b)
	}
}

func showBalancerInfo(b *routerService.BalancerInfo) {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("Balancer: %s (%s)\n", b.Tag, b.Strategy))
	sb.WriteString("  - Selects:\n")
	writeHealthLine(sb, 0, b.Titles, "Tag")
	for i, o := range b.Selects {
		writeHealthLine(sb, i+1, o.Values, o.Tag)
	}
	scnt := len(b.Selects)
	if len(b.Others) > 0 {
		sb.WriteString("  - Others:\n")
		writeHealthLine(sb, 0, b.Titles, "Tag")
		for i, o := range b.Others {
			writeHealthLine(sb, scnt+i+1, o.Values, o.Tag)
		}
	}
	os.Stdout.WriteString(sb.String())
}

func writeHealthLine(sb *strings.Builder, index int, values []string, tag string) {
	if index == 0 {
		// title line
		sb.WriteString("        ")
	} else {
		sb.WriteString(fmt.Sprintf("    %-4d", index))
	}
	for _, v := range values {
		sb.WriteString(fmt.Sprintf("%-14s", v))
	}
	sb.WriteString(tag)
	sb.WriteByte('\n')
}
