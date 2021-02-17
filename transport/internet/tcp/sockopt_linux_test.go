// +build linux

package tcp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/testing/servers/tcp"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	. "github.com/v2fly/v2ray-core/v4/transport/internet/tcp"
)

func TestGetOriginalDestination(t *testing.T) {
	tcpServer := tcp.Server{}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	config, err := internet.ToMemoryStreamConfig(nil)
	common.Must(err)
	conn, err := Dial(context.Background(), dest, config)
	common.Must(err)
	defer conn.Close()

	originalDest, err := GetOriginalDestination(conn)
	if !(dest == originalDest || strings.Contains(err.Error(), "failed to call getsockopt")) {
		t.Error("unexpected state")
	}
}
