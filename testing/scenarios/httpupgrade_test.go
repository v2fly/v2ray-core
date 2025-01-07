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

func TestHTTPUpgrade(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	coreInst, InstMgrIfce := NewInstanceManagerCoreInstance()
	defer coreInst.Close()

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_client",
		common.Must2(os.ReadFile("config/httpupgrade_client.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_server",
		common.Must2(os.ReadFile("config/httpupgrade_server.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_server"))
	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_client"))

	defer func() {
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_client"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(17794, dest.Port, 1024, time.Second*2)(); err != nil {
		t.Error(err)
	}
}

func TestHTTPUpgradeWithEarlyData(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	coreInst, InstMgrIfce := NewInstanceManagerCoreInstance()
	defer coreInst.Close()

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_client",
		common.Must2(os.ReadFile("config/httpupgrade_earlydata_client.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_server",
		common.Must2(os.ReadFile("config/httpupgrade_earlydata_server.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_server"))
	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_client"))

	defer func() {
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_client"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(17794, dest.Port, 1024, time.Second*2)(); err != nil {
		t.Error(err)
	}
}

func TestHTTPUpgradeWithShortEarlyData(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	coreInst, InstMgrIfce := NewInstanceManagerCoreInstance()
	defer coreInst.Close()

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_client",
		common.Must2(os.ReadFile("config/httpupgrade_earlydataShortEarlyData_client.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.AddInstance(
		context.TODO(),
		"httpupgrade_server",
		common.Must2(os.ReadFile("config/httpupgrade_earlydataShortEarlyData_server.json")).([]byte),
		"jsonv5"))

	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_server"))
	common.Must(InstMgrIfce.StartInstance(context.TODO(), "httpupgrade_client"))

	defer func() {
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.StopInstance(context.TODO(), "httpupgrade_client"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_server"))
		common.Must(InstMgrIfce.UntrackInstance(context.TODO(), "httpupgrade_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(17794, dest.Port, 1024, time.Second*2)(); err != nil {
		t.Error(err)
	}
}
