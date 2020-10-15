// +build !confonly

package admin_test

import (
	"testing"
	"v2ray.com/core/app/admin"
)

func TestDownloadFile(t *testing.T) {
	admin.DownloadFile("http://localhost:20809",
		"https://github.com/v2fly/geoip/raw/release/geoip.dat", "d:/tmp/geoip.dat")

	admin.DownloadFile("http://localhost:20809",
		"https://github.com/v2fly/domain-list-community/raw/release/dlc.dat", "d:/tmp/geosite.dat")


}