package api

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	statsService "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"github.com/v2fly/v2ray-core/v5/common/units"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdStats = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api stats [--server=127.0.0.1:8080] [pattern]...",
	Short:       "query statistics",
	Long: `
Query statistics from V2Ray.

> Make sure you have "StatsService" set in "config.api.services" 
of server config.

Arguments:

	-regexp
		The patterns are using regexp.

	-reset
		Reset counters to 0 after fetching their values.

	-runtime
		Get runtime statistics.

	-json
		Use json output.

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

	{{.Exec}} {{.LongName}} -runtime
	{{.Exec}} {{.LongName}} node1
	{{.Exec}} {{.LongName}} -json node1 node2
	{{.Exec}} {{.LongName}} -regexp 'node1.+downlink'
`,
	Run: executeStats,
}

func executeStats(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	var (
		runtime bool
		regexp  bool
		reset   bool
	)
	cmd.Flag.BoolVar(&runtime, "runtime", false, "")
	cmd.Flag.BoolVar(&regexp, "regexp", false, "")
	cmd.Flag.BoolVar(&reset, "reset", false, "")
	cmd.Flag.Parse(args)
	unnamed := cmd.Flag.Args()
	if runtime {
		getRuntimeStats(apiJSON)
		return
	}
	getStats(unnamed, regexp, reset, apiJSON)
}

func getRuntimeStats(jsonOutput bool) {
	conn, ctx, close := dialAPIServer()
	defer close()

	client := statsService.NewStatsServiceClient(conn)
	r := &statsService.SysStatsRequest{}
	resp, err := client.GetSysStats(ctx, r)
	if err != nil {
		base.Fatalf("failed to get sys stats: %s", err)
	}
	if jsonOutput {
		showJSONResponse(resp)
		return
	}
	showRuntimeStats(resp)
}

func showRuntimeStats(s *statsService.SysStatsResponse) {
	formats := []string{"%-22s", "%-10s"}
	rows := [][]string{
		{"Up time", (time.Duration(s.Uptime) * time.Second).String()},
		{"Memory obtained", units.ByteSize(s.Sys).String()},
		{"Number of goroutines", fmt.Sprintf("%d", s.NumGoroutine)},
		{"Heap allocated", units.ByteSize(s.Alloc).String()},
		{"Live objects", fmt.Sprintf("%d", s.LiveObjects)},
		{"Heap allocated total", units.ByteSize(s.TotalAlloc).String()},
		{"Heap allocate count", fmt.Sprintf("%d", s.Mallocs)},
		{"Heap free count", fmt.Sprintf("%d", s.Frees)},
		{"Number of GC", fmt.Sprintf("%d", s.NumGC)},
		{"Time of GC pause", (time.Duration(s.PauseTotalNs) * time.Nanosecond).String()},
	}
	sb := new(strings.Builder)
	writeRow(sb, 0, 0,
		[]string{"Item", "Value"},
		formats,
	)
	for i, r := range rows {
		writeRow(sb, 0, i+1, r, formats)
	}
	os.Stdout.WriteString(sb.String())
}

func getStats(patterns []string, regexp, reset, jsonOutput bool) {
	conn, ctx, close := dialAPIServer()
	defer close()

	client := statsService.NewStatsServiceClient(conn)
	r := &statsService.QueryStatsRequest{
		Patterns: patterns,
		Regexp:   regexp,
		Reset_:   reset,
	}
	resp, err := client.QueryStats(ctx, r)
	if err != nil {
		base.Fatalf("failed to query stats: %s", err)
	}
	if jsonOutput {
		showJSONResponse(resp)
		return
	}
	sort.Slice(resp.Stat, func(i, j int) bool {
		return resp.Stat[i].Name < resp.Stat[j].Name
	})
	showStats(resp.Stat)
}

func showStats(stats []*statsService.Stat) {
	if len(stats) == 0 {
		return
	}
	formats := []string{"%-12s", "%s"}
	sum := int64(0)
	sb := new(strings.Builder)
	idx := 0
	writeRow(sb, 0, 0,
		[]string{"Value", "Name"},
		formats,
	)
	for _, stat := range stats {
		// if stat.Value == 0 {
		// 	continue
		// }
		idx++
		sum += stat.Value
		writeRow(
			sb, 0, idx,
			[]string{units.ByteSize(stat.Value).String(), stat.Name},
			formats,
		)
	}
	sb.WriteString(
		fmt.Sprintf("\nTotal: %s\n", units.ByteSize(sum)),
	)
	os.Stdout.WriteString(sb.String())
}
