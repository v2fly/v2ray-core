//go:build !confonly
// +build !confonly

package tun

import (
	"context"

	"gvisor.dev/gvisor/pkg/tcpip/stack"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/tun/device"
	"github.com/v2fly/v2ray-core/v5/app/tun/device/gvisor"
	"github.com/v2fly/v2ray-core/v5/app/tun/tunsorter"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type TUN struct {
	ctx           context.Context
	dispatcher    routing.Dispatcher
	policyManager policy.Manager
	config        *Config

	stack *stack.Stack
}

func (t *TUN) Type() interface{} {
	return (*TUN)(nil)
}

func (t *TUN) Start() error {
	DeviceConstructor := gvisor.New
	tunDevice, err := DeviceConstructor(device.Options{
		Name: t.config.Name,
		MTU:  t.config.Mtu,
	})
	if err != nil {
		return newError("failed to create device").Base(err).AtError()
	}

	if t.config.PacketEncoding != packetaddr.PacketAddrType_None {
		writer := device.NewLinkWriterToWriter(tunDevice)
		sorter := tunsorter.NewTunSorter(writer, t.dispatcher, t.config.PacketEncoding, t.ctx)
		tunDeviceLayered := NewDeviceWithSorter(tunDevice, sorter)
		tunDevice = tunDeviceLayered
	}

	stack, err := t.CreateStack(tunDevice)
	if err != nil {
		return newError("failed to create stack").Base(err).AtError()
	}
	t.stack = stack

	return nil
}

func (t *TUN) Close() error {
	if t.stack != nil {
		t.stack.Close()
		t.stack.Wait()
	}
	return nil
}

func (t *TUN) Init(ctx context.Context, config *Config, dispatcher routing.Dispatcher, policyManager policy.Manager) error {
	t.ctx = ctx
	t.config = config
	t.dispatcher = dispatcher
	t.policyManager = policyManager

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		tun := new(TUN)
		err := core.RequireFeatures(ctx, func(d routing.Dispatcher, p policy.Manager) error {
			return tun.Init(ctx, config.(*Config), d, p)
		})
		return tun, err
	}))
}
