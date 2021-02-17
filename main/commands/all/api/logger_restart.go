package api

import (
	logService "v2ray.com/core/app/log/command"
	"v2ray.com/core/main/commands/base"
)

var cmdRestartLogger = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api log [--server=127.0.0.1:8080] --restart",
	Short:       "log operations",
	Long: `
Log operations, current supports only '-restart'.

Arguments:

	-restart 
		Restart the logger

	-s, -server 
		The API server address. Default 127.0.0.1:8080

	-t, -timeout
		Timeout seconds to call API. Default 3
`,
	Run: executeRestartLogger,
}

func executeRestartLogger(cmd *base.Command, args []string) {
	var restart bool
	cmd.Flag.BoolVar(&restart, "restart", false, "")
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	if !restart {
		cmd.Usage()
		return
	}

	conn, ctx, close := dialAPIServer()
	defer close()

	client := logService.NewLoggerServiceClient(conn)
	r := &logService.RestartLoggerRequest{}
	_, err := client.RestartLogger(ctx, r)
	if err != nil {
		base.Fatalf("failed to restart logger: %s", err)
	}
}
