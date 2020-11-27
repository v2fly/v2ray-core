package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/common/buf"
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

// loadArg loads one arg, maybe an remote url, or local file path
func loadArg(arg string) (out io.Reader, err error) {
	var data []byte
	switch {
	case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
		data, err = fetchHTTPContent(arg)

	case arg == "stdin:":
		data, err = ioutil.ReadAll(os.Stdin)

	default:
		data, err = ioutil.ReadFile(arg)
	}

	if err != nil {
		return
	}
	out = bytes.NewBuffer(data)
	return
}

// fetchHTTPContent dials https for remote content
func fetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, fmt.Errorf("invalid scheme: %s", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial to %s", target)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	content, err := buf.ReadAllToBytes(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response")
	}

	return content, nil
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
