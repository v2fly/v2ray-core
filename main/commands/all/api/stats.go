package api

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	statsService "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/units"
	"v2ray.com/core/main/commands/base"
)

var cmdStats = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api stats [--server=127.0.0.1:8080] [pattern]...",
	Short:       "query statistics",
	Long: `
Query statistics from V2Ray.

Arguments:

	-regexp
		The patterns are using regexp.

	-reset
		Fetch values then reset statistics counters to 0.

	-runtime
		Get runtime statistics.

	-json
		Use json output.

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
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
		runtime    bool
		regexp     bool
		jsonOutput bool
		reset      bool
	)
	cmd.Flag.BoolVar(&runtime, "runtime", false, "")
	cmd.Flag.BoolVar(&regexp, "regexp", false, "")
	cmd.Flag.BoolVar(&jsonOutput, "json", false, "")
	cmd.Flag.BoolVar(&reset, "reset", false, "")
	cmd.Flag.Parse(args)
	unnamed := cmd.Flag.Args()
	if runtime {
		getRuntimeStats(jsonOutput)
		return
	}
	getStats(unnamed, regexp, reset, jsonOutput)
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
	formts := []string{"%-22s", "%-10s"}
	sb := new(strings.Builder)
	for i, r := range rows {
		writeRow(sb, 0, i+1, r, formts, "")
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
	sum := int64(0)
	sb := new(strings.Builder)
	idx := 0
	for _, stat := range stats {
		// if stat.Value == 0 {
		// 	continue
		// }
		idx++
		sum += stat.Value
		writeRow(
			sb, 0, idx,
			[]string{units.ByteSize(stat.Value).String()},
			[]string{"%-12s"},
			stat.Name,
		)
	}
	sb.WriteString(
		fmt.Sprintf("\nTotal: %s\n", units.ByteSize(sum)),
	)
	os.Stdout.WriteString(sb.String())
}
