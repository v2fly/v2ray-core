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
)

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
