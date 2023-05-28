//go:build linux && ((linux && amd64) || (linux && arm64))
// +build linux
// +build linux,amd64 linux,arm64

package all

import (
	_ "github.com/v2fly/v2ray-core/v5/app/tun"
)
