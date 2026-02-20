package wgcommon

import (
	"context"

	"golang.zx2c4.com/wireguard/device"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

func NewWrappedWireguardDevice(ctx context.Context, config *DeviceConfig) (*WrappedWireguardDevice, error) {
	return &WrappedWireguardDevice{
		config: config,
		ctx:    ctx,
	}, nil
}

type WrappedWireguardDevice struct {
	config *DeviceConfig
	ctx    context.Context
	device *device.Device

	tunnel packetswitch.NetworkLayerDevice
	conn   net.PacketConn
}

func (w *WrappedWireguardDevice) Up() error {
	if w.device != nil {
		return w.device.Up()
	}
	return newError("wireguard device do not exist").AtError()
}

// SetTunnel sets the network layer tunnel device for the wrapped WireGuard device.
func (w *WrappedWireguardDevice) SetTunnel(t packetswitch.NetworkLayerDevice) {
	w.tunnel = t
}

// SetConn sets the underlying packet connection used by the wrapped WireGuard device.
func (w *WrappedWireguardDevice) SetConn(c net.PacketConn) {
	w.conn = c
}

func (w *WrappedWireguardDevice) Close() error {
	if w == nil {
		return nil
	}
	// Bring device down if initialized
	if w.device != nil {
		_ = w.device.Down()
		w.device = nil
	}
	// Close tunnel if present
	if w.tunnel != nil {
		_ = w.tunnel.Close()
		w.tunnel = nil
	}
	// Close underlying packet conn if present
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
	return nil
}

func (w *WrappedWireguardDevice) Debug() (string, error) {
	if w.device != nil {
		result, err := w.device.IpcGet()
		if err != nil {
			return "", err
		}
		return filterDebugData(result), nil
	}
	return "", nil
}
