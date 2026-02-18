package device

import "gvisor.dev/gvisor/pkg/tcpip/stack"

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type Device interface {
	stack.LinkEndpoint
}

type Options struct {
	Name string
	MTU  uint32
}

type DeviceConstructor func(Options) (Device, error)
