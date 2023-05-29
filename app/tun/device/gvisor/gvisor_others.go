//go:build !linux || (linux && !(amd64 || arm64))
// +build !linux linux,!amd64,!arm64

package gvisor

import "github.com/v2fly/v2ray-core/v5/app/tun/device"

func New(options device.Options) (device.Device, error) {
	return nil, newError("not supported").AtError()
}
