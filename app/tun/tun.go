//go:build !confonly
// +build !confonly

package tun

import (
	"context"

	"gvisor.dev/gvisor/pkg/tcpip/stack"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/tun/device"
	"github.com/v2fly/v2ray-core/v5/app/tun/device/gvisor"
	"github.com/v2fly/v2ray-core/v5/app/tun/device/udpbridge"
	"github.com/v2fly/v2ray-core/v5/app/tun/tunsorter"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type TUN struct {
	ctx           context.Context
	dispatcher    routing.Dispatcher
	policyManager policy.Manager
	config        *Config

	stack                     *stack.Stack
	device                    device.Device
	preopenedFD               int
	preopenedFDSet            bool
	packetEncodingBypassPorts []net.Port
}

func (t *TUN) Type() interface{} {
	return (*TUN)(nil)
}

func (t *TUN) Start() error {
	deviceOptions := device.Options{
		Name: t.config.Name,
		MTU:  t.config.Mtu,
	}
	if t.preopenedFDSet {
		deviceOptions.PreopenedFD = t.preopenedFD
		deviceOptions.PreopenedFDSet = true
		t.preopenedFD = -1
		t.preopenedFDSet = false
	}

	var tunDevice device.Device
	var err error
	if bridge := t.config.UdpBridge; bridge != nil {
		tunDevice, err = udpbridge.New(deviceOptions, udpbridge.Options{
			ListenAddress: bridge.ListenAddress,
			ListenPort:    bridge.ListenPort,
			PeerAddress:   bridge.PeerAddress,
			PeerPort:      bridge.PeerPort,
			QueueSize:     bridge.QueueSize,
		})
	} else {
		tunDevice, err = gvisor.New(deviceOptions)
	}
	if err != nil {
		return newError("failed to create device").Base(err).AtError()
	}
	t.device = tunDevice

	if t.config.PacketEncoding != packetaddr.PacketAddrType_None {
		writer := device.NewLinkWriterToWriter(tunDevice)
		sorter := tunsorter.NewTunSorter(
			writer,
			t.dispatcher,
			t.config.PacketEncoding,
			t.packetEncodingContext(),
			t.packetEncodingBypassPorts,
		)
		tunDeviceLayered := NewDeviceWithSorter(tunDevice, sorter)
		tunDevice = tunDeviceLayered
	}

	stack, err := t.CreateStack(tunDevice)
	if err != nil {
		if closer, ok := t.device.(interface{ Close() }); ok {
			closer.Close()
		}
		t.device = nil
		return newError("failed to create stack").Base(err).AtError()
	}
	t.stack = stack

	return nil
}

func (t *TUN) packetEncodingContext() context.Context {
	return session.ContextWithInbound(t.ctx, &session.Inbound{Tag: t.config.Tag})
}

func (t *TUN) Close() error {
	if t.stack != nil {
		t.stack.Close()
		t.stack.Wait()
		t.stack = nil
	} else if t.device != nil {
		if closer, ok := t.device.(interface{ Close() }); ok {
			closer.Close()
		}
	}
	t.device = nil
	if t.preopenedFDSet {
		_ = device.ClosePreopenedFD(t.preopenedFD)
		t.preopenedFD = -1
		t.preopenedFDSet = false
	}
	return nil
}

func (t *TUN) Init(ctx context.Context, config *Config, dispatcher routing.Dispatcher, policyManager policy.Manager) error {
	t.ctx = ctx
	t.config = config
	t.dispatcher = dispatcher
	t.policyManager = policyManager
	t.preopenedFD = -1
	t.packetEncodingBypassPorts = make([]net.Port, 0, len(config.PacketEncodingBypassPorts))
	for _, port := range config.PacketEncodingBypassPorts {
		if port == 0 || port > 65535 {
			return newError("invalid packet_encoding_bypass_ports value: ", port).AtError()
		}
		t.packetEncodingBypassPorts = append(t.packetEncodingBypassPorts, net.Port(port))
	}
	if config.PreopenedFd != nil {
		if config.UdpBridge != nil {
			return newError("preopened_fd and udp_bridge cannot be used together").AtError()
		}
		if *config.PreopenedFd < 0 {
			return newError("invalid preopened_fd: ", *config.PreopenedFd).AtError()
		}
		t.preopenedFD = int(*config.PreopenedFd)
		t.preopenedFDSet = true
	}

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
