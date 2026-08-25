package wgcommon

import (
	"context"
	"sync"

	"golang.zx2c4.com/wireguard/conn"
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

	tunnel    packetswitch.NetworkLayerDevice
	conn      net.PacketConn
	bind      conn.Bind
	closeOnce sync.Once
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
	w.bind = nil
}

// SetBind sets an explicit WireGuard bind. It is useful for binds that manage
// multiple PacketConn generations over the device lifetime.
func (w *WrappedWireguardDevice) SetBind(bind conn.Bind) {
	w.bind = bind
	w.conn = nil
}

func (w *WrappedWireguardDevice) Close() error {
	if w == nil {
		return nil
	}
	w.closeOnce.Do(func() {
		// Close the device rather than only bringing it down. Down stops peers
		// and the bind, while Close also terminates WireGuard's long-lived worker
		// pools and closes the TUN adaptor.
		if w.device != nil {
			w.device.Close()
		}
		if w.tunnel != nil {
			_ = w.tunnel.Close()
		}
		if w.conn != nil {
			_ = w.conn.Close()
		}
		if w.bind != nil {
			_ = w.bind.Close()
		}
	})
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
