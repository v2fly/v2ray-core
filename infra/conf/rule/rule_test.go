package rule_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/infra/conf/geodata"
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/standard"
	"github.com/v2fly/v2ray-core/v4/infra/conf/rule"
)

func init() {
	const geoipURL = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0o755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
	}
}

func TestToCidrList(t *testing.T) {
	t.Log(os.Getenv("v2ray.location.asset"))

	common.Must(filesystem.CopyFile(platform.GetAssetLocation("geoiptestrouter.dat"), platform.GetAssetLocation("geoip.dat")))

	ips := cfgcommon.StringList([]string{
		"geoip:us",
		"geoip:cn",
		"geoip:!cn",
		"ext:geoiptestrouter.dat:!cn",
		"ext:geoiptestrouter.dat:ca",
		"ext-ip:geoiptestrouter.dat:!cn",
		"ext-ip:geoiptestrouter.dat:!ca",
	})

	cfgctx := cfgcommon.NewConfigureLoadingContext(context.Background())

	if loader, err := geodata.GetGeoDataLoader("standard"); err == nil {
		cfgcommon.SetGeoDataLoader(cfgctx, loader)
	} else {
		t.Fatal(err)
	}

	_, err := rule.ToCidrList(cfgctx, ips)
	if err != nil {
		t.Fatalf("Failed to parse geoip list, got %s", err)
	}
}
