package scenarios

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/tcp"

	_ "github.com/v2fly/v2ray-core/v5/main/distro/all"
)

func TestRRPITOverDTLSUDP(t *testing.T) {
	runRRPITScenario(t, "config/rrpit_client.json", "config/rrpit_server.json", 17894)
}

func TestRRPITOverDTLSUDPMultiChannel(t *testing.T) {
	runRRPITScenario(t, "config/rrpit_multichannel_client.json", "config/rrpit_multichannel_server.json", 17904)
}

func TestRRPITOverDTLSUDPSmallShard(t *testing.T) {
	runRRPITScenario(t, "config/rrpit_smallshard_client.json", "config/rrpit_smallshard_server.json", 18004)
}

func TestRRPITOverDTLSUDPDefaultsAndExplicitAddress(t *testing.T) {
	runRRPITScenario(t, "config/rrpit_defaults_client.json", "config/rrpit_defaults_server.json", 18104)
}

func runRRPITScenario(t *testing.T, clientConfigPath string, serverConfigPath string, socksPort v2net.Port) {
	t.Helper()

	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	coreInst, instMgr := NewInstanceManagerCoreInstance()
	defer coreInst.Close()

	common.Must(instMgr.AddInstance(
		context.TODO(),
		"rrpit_client",
		common.Must2(os.ReadFile(clientConfigPath)).([]byte),
		"jsonv5"))

	common.Must(instMgr.AddInstance(
		context.TODO(),
		"rrpit_server",
		common.Must2(os.ReadFile(serverConfigPath)).([]byte),
		"jsonv5"))

	common.Must(instMgr.StartInstance(context.TODO(), "rrpit_server"))
	common.Must(instMgr.StartInstance(context.TODO(), "rrpit_client"))

	defer func() {
		common.Must(instMgr.StopInstance(context.TODO(), "rrpit_server"))
		common.Must(instMgr.StopInstance(context.TODO(), "rrpit_client"))
		common.Must(instMgr.UntrackInstance(context.TODO(), "rrpit_server"))
		common.Must(instMgr.UntrackInstance(context.TODO(), "rrpit_client"))
		coreInst.Close()
	}()

	if err := testTCPConnViaSocks(socksPort, dest.Port, 4096, 5*time.Second)(); err != nil {
		t.Fatal(err)
	}
}
