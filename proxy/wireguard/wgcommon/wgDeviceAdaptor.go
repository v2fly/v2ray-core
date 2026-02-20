package wgcommon

import (
	"errors"
	"os"
	"sync"

	"golang.zx2c4.com/wireguard/tun"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

func NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(mtu int, networkLayerSwitch packetswitch.NetworkLayerDevice, batchSize int, inboundChannelSize int) (*NetworkLayerDeviceToWireguardTunDeviceAdaptor, error) {
	if batchSize <= 0 {
		batchSize = 1
	}
	if inboundChannelSize <= 0 {
		inboundChannelSize = 1024
	}
	n := &NetworkLayerDeviceToWireguardTunDeviceAdaptor{
		mtu:                mtu,
		networkLayerSwitch: networkLayerSwitch,
		in:                 make(chan []byte, inboundChannelSize),
		events:             make(chan tun.Event, 4),
		batchSize:          batchSize,
	}
	// Attach writer to the network layer switch so incoming packets are delivered to this adaptor.
	if networkLayerSwitch != nil {
		if err := networkLayerSwitch.OnAttach(&networkLayerWriter{parent: n}); err != nil {
			return nil, err
		}
		// If the underlying device exposes real link events, forward them.
		if src, ok := networkLayerSwitch.(interface{ Events() <-chan tun.Event }); ok {
			go func() {
				for ev := range src.Events() {
					// Acquire lock to synchronize with Close (which closes the events channel).
					n.mu.Lock()
					closed := n.closed
					if !closed {
						// best-effort: do not block if events channel is full
						select {
						case n.events <- ev:
						default:
						}
					}
					n.mu.Unlock()
					if closed {
						return
					}
				}
			}()
		}
	}
	return n, nil
}

// NetworkLayerDeviceToWireguardTunDeviceAdaptor is primarily machine generated.
type NetworkLayerDeviceToWireguardTunDeviceAdaptor struct {
	mtu                int
	networkLayerSwitch packetswitch.NetworkLayerDevice

	mu        sync.RWMutex
	closed    bool
	in        chan []byte
	events    chan tun.Event
	batchSize int
}

// networkLayerWriter adapts packetswitch writes into the adaptor's incoming channel.
type networkLayerWriter struct {
	parent *NetworkLayerDeviceToWireguardTunDeviceAdaptor
}

func (w *networkLayerWriter) Write(packet []byte) (int, error) {
	p := make([]byte, len(packet))
	copy(p, packet)
	w.parent.mu.RLock()
	closed := w.parent.closed
	w.parent.mu.RUnlock()
	if closed {
		return 0, errors.New("device closed")
	}
	select {
	case w.parent.in <- p:
		return len(packet), nil
	default:
		// Channel full, drop packet.
		return 0, errors.New("no buffer space")
	}
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) File() *os.File {
	// No underlying OS file for this adaptor.
	return nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) Read(bufs [][]byte, sizes []int, offset int) (ret int, err error) {
	// Read up to BatchSize packets or until bufs exhausted.
	// NOTE: 'offset' is a byte offset within each buffer (to leave room for transport headers),
	// NOT an index into the bufs slice. WireGuard expects packet data to be written starting at
	// bufs[i][offset:].
	maxCount := n.BatchSize()
	if maxCount <= 0 {
		maxCount = 1
	}
	for i := 0; i < maxCount && i < len(bufs); i++ {
		var b []byte
		var ok bool
		if ret == 0 {
			// first read: block until a packet arrives or channel closes
			b, ok = <-n.in
		} else {
			// subsequent reads: do not block — if no packet available, return what we've got
			select {
			case b, ok = <-n.in:
				// got one
			default:
				return ret, nil
			}
		}
		if !ok {
			// channel closed
			if ret == 0 {
				return 0, os.ErrClosed
			}
			return ret, nil
		}
		to := bufs[i]
		if to == nil {
			// packet consumed but no destination buffer provided, skip copying
			ret++
			continue
		}
		copied := copy(to[offset:], b)
		if sizes != nil && i < len(sizes) {
			sizes[i] = copied
		}
		ret++
	}
	return ret, nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) Write(bufs [][]byte, offset int) (int, error) {
	written := 0
	if n.networkLayerSwitch == nil {
		return 0, errors.New("no network layer writer attached")
	}
	for i := 0; i < len(bufs); i++ {
		b := bufs[i]
		if b == nil || len(b) <= offset {
			continue
		}
		// The offset is a byte offset within each buffer where the actual
		// packet payload starts (after transport headers). Extract only the
		// payload portion. Copy because caller may reuse buffer.
		payload := b[offset:]
		cp := make([]byte, len(payload))
		copy(cp, payload)
		_, err := n.networkLayerSwitch.Write(cp)
		if err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) MTU() (int, error) {
	return n.mtu, nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) Name() (string, error) {
	// No specific name available for this virtual adaptor.
	return "", nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) Events() <-chan tun.Event {
	return n.events
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) Close() error {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return nil
	}
	n.closed = true
	close(n.in)
	// Close events channel to signal no more events.
	close(n.events)
	dev := n.networkLayerSwitch
	n.networkLayerSwitch = nil
	n.mu.Unlock()
	if dev != nil {
		_ = dev.Close()
	}
	return nil
}

func (n *NetworkLayerDeviceToWireguardTunDeviceAdaptor) BatchSize() int {
	return n.batchSize
}
