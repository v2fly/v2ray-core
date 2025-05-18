package device

import (
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type Device interface {
	stack.LinkEndpoint
	// The Close() method is already included if stack.LinkEndpoint defines it.
	// If you need an explicit Close() method with a specific signature, you can uncomment the following line:
	// Close() error
}

type Options struct {
	Name string
	MTU  uint32
}

type DeviceConstructor func(Options) (Device, error)
