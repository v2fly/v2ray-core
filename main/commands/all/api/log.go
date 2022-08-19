package api

import (
	"io"
	"log"
	"os"

	logService "github.com/v2fly/v2ray-core/v5/app/log/command"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdLog = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} api log [--server=127.0.0.1:8080]",
	Short:       "log operations",
	Long: `
Follow and print logs from v2ray.

> Make sure you have "LoggerService" set in "config.api.services" 
of server config.

> It ignores -timeout flag while following logs

Arguments:

	-restart 
		Restart the logger

	-s, -server <server:port>
		The API server address. Default 127.0.0.1:8080

	-t, -timeout <seconds>
		Timeout seconds to call API. Default 3

Example:

    {{.Exec}} {{.LongName}}
    {{.Exec}} {{.LongName}} --restart
`,
	Run: executeLog,
}

func executeLog(cmd *base.Command, args []string) {
	var restart bool
	cmd.Flag.BoolVar(&restart, "restart", false, "")
	setSharedFlags(cmd)
	cmd.Flag.Parse(args)

	if restart {
		restartLogger()
		return
	}
	followLogger()
}

func restartLogger() {
	conn, ctx, close := dialAPIServer()
	defer close()
	client := logService.NewLoggerServiceClient(conn)
	r := &logService.RestartLoggerRequest{}
	_, err := client.RestartLogger(ctx, r)
	if err != nil {
		base.Fatalf("failed to restart logger: %s", err)
	}
}

func followLogger() {
	conn, ctx, close := dialAPIServerWithoutTimeout()
	defer close()
	client := logService.NewLoggerServiceClient(conn)
	r := &logService.FollowLogRequest{}
	stream, err := client.FollowLog(ctx, r)
	if err != nil {
		base.Fatalf("failed to follow logger: %s", err)
	}
	// work with `v2ray api log | grep expr`
	log.SetOutput(os.Stdout)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			base.Fatalf("failed to fetch log: %s", err)
		}
		log.Println(resp.Message)
	}
}
