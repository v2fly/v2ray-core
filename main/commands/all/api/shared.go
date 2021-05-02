package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v4"
)

var (
	apiServerAddrPtr string
	apiTimeout       int
	apiJSON          bool
	// ApiConfigFormat is an internal variable
	ApiConfigFormat string
	// ApiConfigRecursively is an internal variable
	ApiConfigRecursively bool
)

// SetSharedFlags is an internal API
func SetSharedFlags(cmd *base.Command) {
	setSharedFlags(cmd)
}

func setSharedFlags(cmd *base.Command) {
	cmd.Flag.StringVar(&apiServerAddrPtr, "s", "127.0.0.1:8080", "")
	cmd.Flag.StringVar(&apiServerAddrPtr, "server", "127.0.0.1:8080", "")
	cmd.Flag.IntVar(&apiTimeout, "t", 3, "")
	cmd.Flag.IntVar(&apiTimeout, "timeout", 3, "")
	cmd.Flag.BoolVar(&apiJSON, "json", false, "")
}

// SetSharedConfigFlags is an internal API
func SetSharedConfigFlags(cmd *base.Command) {
	setSharedConfigFlags(cmd)
}

func setSharedConfigFlags(cmd *base.Command) {
	cmd.Flag.StringVar(&ApiConfigFormat, "format", core.FormatAuto, "")
	cmd.Flag.BoolVar(&ApiConfigRecursively, "r", false, "")
}

// SetSharedFlags is an internal API
func DialAPIServer() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	return dialAPIServer()
}

func dialAPIServer() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	conn = dialAPIServerWithContext(ctx)
	close = func() {
		cancel()
		conn.Close()
	}
	return
}

func dialAPIServerWithoutTimeout() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	ctx = context.Background()
	conn = dialAPIServerWithContext(ctx)
	close = func() {
		conn.Close()
	}
	return
}

func dialAPIServerWithContext(ctx context.Context) (conn *grpc.ClientConn) {
	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
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
