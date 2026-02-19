package wgcommon

import (
	"errors"
	"fmt"
	"strings"

	"golang.zx2c4.com/wireguard/device"
)

func (w *WrappedWireguardDevice) InitDevice() error {
	if w == nil || w.config == nil {
		return errors.New("wireguard: missing config")
	}
	if w.device != nil {
		return errors.New("wireguard: device already initialized")
	}

	// Create a tun device adaptor from the packetswitch network layer device.
	// Use a reasonable default MTU and batch sizes. These can be tuned later.
	tunDev, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(int(w.config.Mtu), w.tunnel, 1, 1024)
	if err != nil {
		return err
	}

	// Create wireguard bind adapter from provided PacketConn
	bind := NewNetPacketConnToWg(w.conn)

	// Create the wireguard device with our logger adapter.
	dev := device.NewDevice(tunDev, bind, NewDeviceLoggerAdapter())
	if dev == nil {
		return errors.New("wireguard: failed to initialize device")
	}
	w.device = dev
	return nil
}

func (w *WrappedWireguardDevice) SetupDeviceWithoutPeers() error {
	if w == nil || w.config == nil {
		return errors.New("wireguard: missing config")
	}
	if w.device == nil {
		return errors.New("wireguard: device not initialized")
	}

	var sb strings.Builder
	if len(w.config.PrivateKey) > 0 {
		sb.WriteString(fmt.Sprintf("private_key=%x\n", w.config.PrivateKey))
	}
	if w.config.ListenPort != 0 {
		sb.WriteString(fmt.Sprintf("listen_port=%d\n", w.config.ListenPort))
	}

	// Terminate operation with a blank line.
	sb.WriteString("\n")

	return w.device.IpcSet(sb.String())
}

func (w *WrappedWireguardDevice) AddOrReplacePeers(peers []*PeerConfig) error {
	if w == nil || w.config == nil {
		return errors.New("wireguard: missing config")
	}
	if w.device == nil {
		return errors.New("wireguard: device not initialized")
	}

	var sb strings.Builder
	// Replace existing peers with the provided list
	sb.WriteString("replace_peers=true\n")

	for _, p := range peers {
		if p == nil || len(p.PublicKey) == 0 {
			// skip empty entries
			continue
		}
		// start peer block
		sb.WriteString(fmt.Sprintf("public_key=%x\n", p.PublicKey))
		if len(p.PresharedKey) > 0 {
			sb.WriteString(fmt.Sprintf("preshared_key=%x\n", p.PresharedKey))
		}
		if p.Endpoint != "" {
			sb.WriteString(fmt.Sprintf("endpoint=%s\n", p.Endpoint))
		}
		if p.PersistentKeepaliveInterval != 0 {
			sb.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", p.PersistentKeepaliveInterval))
		}
		// replace allowed IPs for this peer
		sb.WriteString("replace_allowed_ips=true\n")
		for _, aip := range p.AllowedIps {
			if aip == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("allowed_ip=%s\n", aip))
		}
	}

	// terminate
	sb.WriteString("\n")

	return w.device.IpcSet(sb.String())
}

func (w *WrappedWireguardDevice) RemovePeer(publicKey []byte) error {
	if w == nil {
		return errors.New("wireguard: nil receiver")
	}
	if w.device == nil {
		return errors.New("wireguard: device not initialized")
	}
	if len(publicKey) == 0 {
		return errors.New("wireguard: empty public key")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("public_key=%x\n", publicKey))
	sb.WriteString("remove=true\n")
	sb.WriteString("\n")

	return w.device.IpcSet(sb.String())
}
