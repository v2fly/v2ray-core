package api

import (
	"fmt"
	"os"
	"strings"

	routerService "github.com/ghxhy/v2ray-core/v5/app/router/command"
	"github.com/ghxhy/v2ray-core/v5/main/commands/base"
)

// TODO: support "-json" flag for json output
var cmdBalancerInfo = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api bi [--server=127.0.0.1:8080] [balancer]...",
	Short:       "balancer information",
	Long: `
Get information of specified balancers, including health, strategy 
and selecting. If no balancer tag specified, get information of 
all balancers.

> Make sure you have "RoutingService" set in "config.api.services" 
of server config.

Arguments:

	-json
		Use json output.

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}} --server=127.0.0.1:8080 balancer1 balancer2
`,
	Run: executeBalancerInfo,
}

func executeBalancerInfo(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := routerService.NewRoutingServiceClient(conn)
	r := &routerService.GetBalancerInfoRequest{Tag: cmd.Flag.Arg(0)}
	resp, err := client.GetBalancerInfo(ctx, r)
	if err != nil {
		base.Fatalf("failed to get health information: %s", err)
	}

	if apiJSON {
		showJSONResponse(resp)
		return
	}

	showBalancerInfo(resp.Balancer)
}

func showBalancerInfo(b *routerService.BalancerMsg) {
	const tableIndent = 4
	sb := new(strings.Builder)
	// Override
	if b.Override != nil {
		sb.WriteString("  - Selecting Override:\n")
		for i, s := range []string{b.Override.Target} {
			writeRow(sb, tableIndent, i+1, []string{s}, nil)
		}
	}
	// Selects
	sb.WriteString("  - Selects:\n")

	for i, o := range b.PrincipleTarget.Tag {
		writeRow(sb, tableIndent, i+1, []string{o}, nil)
	}
	os.Stdout.WriteString(sb.String())
}

func getColumnFormats(titles []string) []string {
	w := make([]string, len(titles))
	for i, t := range titles {
		w[i] = fmt.Sprintf("%%-%ds ", len(t))
	}
	return w
}

func writeRow(sb *strings.Builder, indent, index int, values, formats []string) {
	if index == 0 {
		// title line
		sb.WriteString(strings.Repeat(" ", indent+4))
	} else {
		sb.WriteString(fmt.Sprintf("%s%-4d", strings.Repeat(" ", indent), index))
	}
	for i, v := range values {
		format := "%-14s"
		if i < len(formats) {
			format = formats[i]
		}
		sb.WriteString(fmt.Sprintf(format, v))
	}
	sb.WriteByte('\n')
}
