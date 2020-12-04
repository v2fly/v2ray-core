package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type serviceHandler func(ctx context.Context, conn *grpc.ClientConn, cmd *base.Command, args []string) string

var (
	apiServerAddrPtr string
	apiTimeout       int
)

func setSharedFlags(cmd *base.Command) {
	cmd.Flag.StringVar(&apiServerAddrPtr, "s", "127.0.0.1:8080", "")
	cmd.Flag.StringVar(&apiServerAddrPtr, "server", "127.0.0.1:8080", "")
	cmd.Flag.IntVar(&apiTimeout, "t", 3, "")
	cmd.Flag.IntVar(&apiTimeout, "timeout", 3, "")
}

func dialAPIServer() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	close = func() {
		cancel()
		conn.Close()
	}
	return
}

func showResponese(m proto.Message) {
	msg := ""
	bs, err := proto.Marshal(m)
	if err != nil {
		msg = err.Error()
	} else {
		msg = string(bs)
		msg = strings.TrimSpace(msg)
	}
	if msg == "" {
		return
	}
	fmt.Println(msg)
}
