package api

import (
	"context"
	"time"

	"google.golang.org/grpc"
	statsService "v2ray.com/core/app/stats/command"
	"v2ray.com/core/commands/base"
)

var cmdQueryStats = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api statsquery [--server=127.0.0.1:8080]",
	Short:       "Query statistics",
	Long: `
Query statistics from V2Ray by calling its API. (timeout 3 seconds)

Arguments:

	-server=127.0.0.1:8080 
		The API server address. Default 127.0.0.1:8080

	-pattern
		Pattern of the query.

	-reset
		Reset the counter to fetching its value.

Example:

	{{.Exec}} {{.LongName}} --server=127.0.0.1:8080 -pattern "counter_"
`,
	Run: executeQueryStats,
}

func executeQueryStats(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	pattern := cmd.Flag.String("pattern", "", "")
	reset := cmd.Flag.Bool("reset", false, "")
	cmd.Flag.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", *apiServerAddrPtr)
	}
	defer conn.Close()

	client := statsService.NewStatsServiceClient(conn)
	r := &statsService.QueryStatsRequest{
		Pattern: *pattern,
		Reset_:  *reset,
	}
	resp, err := client.QueryStats(ctx, r)
	if err != nil {
		base.Fatalf("failed to query stats: %s", err)
	}
	showResponese(responeseToString(resp))
}
