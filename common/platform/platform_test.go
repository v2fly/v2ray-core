package platform_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
)

func init() {
	const geoipURL = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"

	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")

	if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		common.Must(os.MkdirAll(tempPath, 0755))
		geoipBytes, err := common.FetchHTTPContent(geoipURL)
		common.Must(err)
		common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
	}
}

func TestNormalizeEnvName(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "a",
			output: "A",
		},
		{
			input:  "a.a",
			output: "A_A",
		},
		{
			input:  "A.A.B",
			output: "A_A_B",
		},
	}
	for _, test := range cases {
		if v := platform.NormalizeEnvName(test.input); v != test.output {
			t.Error("unexpected output: ", v, " want ", test.output)
		}
	}
}

func TestEnvFlag(t *testing.T) {
	if v := (platform.EnvFlag{
		Name: "xxxxx.y",
	}.GetValueAsInt(10)); v != 10 {
		t.Error("env value: ", v)
	}
}

// TestWrongErrorCheckOnOSStat is a test to detect the misuse of error handling
// in os.Stat, which will lead to failure to find & read geoip & geosite files.
func TestWrongErrorCheckOnOSStat(t *testing.T) {
	theExpectedDir := filepath.Join("usr", "local", "share", "v2ray")
	getAssetLocation := func(file string) string {
		for _, p := range []string{
			filepath.Join(theExpectedDir, file),
		} {
			// errors.Is(fs.ErrNotExist, err) is a mistake supposed Not to
			// be discovered by the Go runtime, which will lead to failure to
			// find & read geoip & geosite files.
			// The correct code is `errors.Is(err, fs.ErrNotExist)`
			if _, err := os.Stat(p); err != nil && errors.Is(fs.ErrNotExist, err) {
				continue
			}
			// asset found
			return p
		}
		return filepath.Join("the", "wrong", "path", "not-exist.txt")
	}

	notExist := getAssetLocation("not-exist.txt")
	if filepath.Dir(notExist) != theExpectedDir {
		t.Error("asset dir:", notExist, "not in", theExpectedDir)
	}
}

func TestGetAssetLocation(t *testing.T) {
	// Test for external geo files
	wd, err := os.Getwd()
	common.Must(err)
	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")
	asset := platform.GetAssetLocation(geoipPath)
	if _, err := os.Stat(asset); err != nil && errors.Is(err, fs.ErrNotExist) {
		t.Error("cannot find external geo file:", asset)
	}

	exec, err := os.Executable()
	common.Must(err)
	loc := platform.GetAssetLocation("t")
	if filepath.Dir(loc) != filepath.Dir(exec) {
		t.Error("asset dir: ", loc, " not in ", exec)
	}

	os.Setenv("v2ray.location.asset", "/v2ray")
	if runtime.GOOS == "windows" {
		if v := platform.GetAssetLocation("t"); v != "\\v2ray\\t" {
			t.Error("asset loc: ", v)
		}
	} else {
		if v := platform.GetAssetLocation("t"); v != "/v2ray/t" {
			t.Error("asset loc: ", v)
		}
	}
}
