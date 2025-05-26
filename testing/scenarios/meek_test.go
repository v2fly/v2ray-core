package scenarios

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/testing/servers/tcp"

	_ "github.com/v2fly/v2ray-core/v5/main/distro/all"
)

func TestMeek(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	coreInst, InstMgrIfce := NewInstanceManagerCoreInstance()
	defer coreInst.Close()

	ctx := context.Background()

	common.Must(InstMgrIfce.AddInstance(
		ctx,
		"meek_client",
		common.Must2(os.ReadFile("config/meek_client.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.AddInstance(
		ctx,
		"meek_server",
		common.Must2(os.ReadFile("config/meek_server.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.StartInstance(ctx, "meek_server"))
	common.Must(InstMgrIfce.StartInstance(ctx, "meek_client"))

	defer func() {
		common.Must(InstMgrIfce.StopInstance(ctx, "meek_server"))
		common.Must(InstMgrIfce.StopInstance(ctx, "meek_client"))
		common.Must(InstMgrIfce.UntrackInstance(ctx, "meek_server"))
		common.Must(InstMgrIfce.UntrackInstance(ctx, "meek_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(17774, dest.Port, 1024, time.Second*2)(); err != nil {
		t.Error(err)
	}
}
