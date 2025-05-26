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

func TestMekya(t *testing.T) {
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
		"mekya_client",
		common.Must2(os.ReadFile("config/mekya_client.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.AddInstance(
		ctx,
		"mekya_server",
		common.Must2(os.ReadFile("config/mekya_server.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.StartInstance(ctx, "mekya_server"))
	common.Must(InstMgrIfce.StartInstance(ctx, "mekya_client"))

	defer func() {
		common.Must(InstMgrIfce.StopInstance(ctx, "mekya_server"))
		common.Must(InstMgrIfce.StopInstance(ctx, "mekya_client"))
		common.Must(InstMgrIfce.UntrackInstance(ctx, "mekya_server"))
		common.Must(InstMgrIfce.UntrackInstance(ctx, "mekya_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(17774, dest.Port, 1024, time.Second*2)(); err != nil {
		t.Error(err)
	}
}
