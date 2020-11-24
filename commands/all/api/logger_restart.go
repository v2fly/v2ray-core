package api

import (
	"context"
	"time"

	"google.golang.org/grpc"
	logService "v2ray.com/core/app/log/command"
	"v2ray.com/core/commands/base"
)

var cmdRestartLogger = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api restartlogger [--server=127.0.0.1:8080]",
	Short:       "Restart the logger",
	Long: `
Restart the logger of V2Ray by calling its API. (timeout 3 seconds)

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	defer conn.Close()

	client := logService.NewLoggerServiceClient(conn)
	r := &logService.RestartLoggerRequest{}
	resp, err := client.RestartLogger(ctx, r)
	if err != nil {
		base.Fatalf("failed to restart logger: %s", err)
	}
	showResponese(responeseToString(resp))
}
