package geodata_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/geodata"
	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
)

const (
	geoipURL   = "https://raw.githubusercontent.com/v2fly/geoip/release/geoip.dat"
	geositeURL = "https://raw.githubusercontent.com/v2fly/domain-list-community/release/dlc.dat"
)

func init() {
	wd, err := os.Getwd()
	common.Must(err)

	tempPath := filepath.Join(wd, "..", "..", "testing", "temp")
	geoipPath := filepath.Join(tempPath, "geoip.dat")
	geositePath := filepath.Join(tempPath, "geosite.dat")

	os.Setenv("v2ray.location.asset", tempPath)

	common.Must(os.MkdirAll(tempPath, 0755))

	if _, err := os.Stat(platform.GetAssetLocation("geoip.dat")); err != nil && errors.Is(err, fs.ErrNotExist) {
		if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
			geoipBytes, err := common.FetchHTTPContent(geoipURL)
			common.Must(err)
			common.Must(filesystem.WriteFile(geoipPath, geoipBytes))
		}
	}

	if _, err := os.Stat(platform.GetAssetLocation("geosite.dat")); err != nil && errors.Is(err, fs.ErrNotExist) {
		if _, err := os.Stat(geositePath); err != nil && errors.Is(err, fs.ErrNotExist) {
			geositeBytes, err := common.FetchHTTPContent(geositeURL)
			common.Must(err)
			common.Must(filesystem.WriteFile(geositePath, geositeBytes))
		}
	}
}

func TestDecodeGeoIP(t *testing.T) {
	filename := platform.GetAssetLocation("geoip.dat")
	result, err := geodata.Decode(filename, "test")
	if err != nil {
		t.Error(err)
	}

	expected := []byte{10, 4, 84, 69, 83, 84, 18, 8, 10, 4, 127, 0, 0, 0, 16, 8}
	if cmp.Diff(result, expected) != "" {
		t.Errorf("failed to load geoip:test, expected: %v, got: %v", expected, result)
	}
}

func TestDecodeGeoSite(t *testing.T) {
	filename := platform.GetAssetLocation("geosite.dat")
	result, err := geodata.Decode(filename, "test")
	if err != nil {
		t.Error(err)
	}

	expected := []byte{10, 4, 84, 69, 83, 84, 18, 20, 8, 3, 18, 16, 116, 101, 115, 116, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}
	if cmp.Diff(result, expected) != "" {
		t.Errorf("failed to load geosite:test, expected: %v, got: %v", expected, result)
	}
}
