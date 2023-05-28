package device

import (
	"github.com/v2fly/v2ray-core/v5/common"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type Device interface {
	stack.LinkEndpoint

	common.Closable
}

type Options struct {
	Name string
	MTU  uint32
}

type DeviceConstructor func(Options) (Device, error)
