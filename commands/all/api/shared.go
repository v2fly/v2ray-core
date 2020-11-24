package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/infra/conf"
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

func jsonToConfig(f string) (*conf.Config, error) {
	c := &conf.Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func responeseToString(m proto.Message) string {
	msg := ""
	bs, err := proto.Marshal(m)
	msg = string(bs)
	if err != nil {
		msg = err.Error()
	}
	msg = strings.Trim(msg, " ")
	return msg
}

func showResponese(r string) {
	r = strings.TrimSpace(r)
	if r == "" {
		return
	}
	fmt.Println(r)
}
