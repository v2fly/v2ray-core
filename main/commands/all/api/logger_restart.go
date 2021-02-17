package api

import (
	logService "github.com/v2fly/v2ray-core/v4/app/log/command"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

var cmdRestartLogger = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api restartlogger [--server=127.0.0.1:8080]",
	Short:       "Restart the logger",
	Long: `
Restart the logger of V2Ray.

Arguments:

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3
`,
	Run: executeRestartLogger,
}

func executeRestartLogger(cmd *base.Command, args []string) {
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	conn, ctx, close := dialAPIServer()
	defer close()

	client := logService.NewLoggerServiceClient(conn)
	r := &logService.RestartLoggerRequest{}
	resp, err := client.RestartLogger(ctx, r)
	if err != nil {
		base.Fatalf("failed to restart logger: %s", err)
	}
	showResponese(resp)
}
