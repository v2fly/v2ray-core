package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

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

func protoToJSONString(m proto.Message, prefix, indent string) (string, error) {
	b := new(strings.Builder)
	e := json.NewEncoder(b)
	e.SetIndent(prefix, indent)
	e.SetEscapeHTML(false)
	err := e.Encode(m)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(b.String()), nil
}

func showJSONResponse(m proto.Message) {
	output, err := protoToJSONString(m, "", "")
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v\n", m)
		base.Fatalf("error encode json: %s", err)
	}
	fmt.Println(output)
}
