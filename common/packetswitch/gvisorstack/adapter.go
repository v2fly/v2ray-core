package gvisorstack

import (
	"context"
	"sync"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

func NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(_ context.Context, mtu int, networkLayerSwitch packetswitch.NetworkLayerDevice) *NetworkLayerDeviceToGvisorLinkEndpointAdaptor {
	return &NetworkLayerDeviceToGvisorLinkEndpointAdaptor{
		mtu:                mtu,
		networkLayerSwitch: networkLayerSwitch,
		waitCh:             make(chan struct{}),
	}
}

// NetworkLayerDeviceToGvisorLinkEndpointAdaptor is primarily machine generated.
type NetworkLayerDeviceToGvisorLinkEndpointAdaptor struct {
	mtu                int
	networkLayerSwitch packetswitch.NetworkLayerDevice

	mu         sync.RWMutex
	dispatcher stack.NetworkDispatcher
	attached   bool
	closed     bool
	onClose    func()
	waitCh     chan struct{}
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) MTU() uint32 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return uint32(n.mtu)
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) SetMTU(mtu uint32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.mtu = int(mtu)
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) MaxHeaderLength() uint16 {
	// No additional link-layer header.
	return 0
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) LinkAddress() tcpip.LinkAddress {
	// Not applicable for network-layer device.
	return ""
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) SetLinkAddress(_ tcpip.LinkAddress) {
	// no-op
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) Capabilities() stack.LinkEndpointCapabilities {
	return stack.CapabilityNone
}

// networkLayerWriter adapts packets from NetworkLayerDevice into gVisor Stack.
type networkLayerWriter struct {
	parent *NetworkLayerDeviceToGvisorLinkEndpointAdaptor
}

func (w *networkLayerWriter) Write(packet []byte) (int, error) {
	if len(packet) == 0 {
		return 0, nil
	}

	buf := buffer.MakeWithData(packet)
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buf,
		// Do not call buf.Release here; PacketBuffer.DecRef will release internal buffer.
	})

	// Determine network protocol by IP version.
	ver := packet[0] >> 4
	var proto tcpip.NetworkProtocolNumber
	switch ver {
	case 4:
		proto = ipv4.ProtocolNumber
	case 6:
		proto = ipv6.ProtocolNumber
	default:
		// Unknown network packet, drop.
		pkt.DecRef()
		return 0, nil
	}

	w.parent.mu.RLock()
	d := w.parent.dispatcher
	w.parent.mu.RUnlock()
	if d == nil {
		// No dispatcher attached, drop.
		pkt.DecRef()
		return 0, nil
	}

	// Deliver to network layer. The dispatcher takes ownership of pkt
	// and is responsible for releasing it.
	d.DeliverNetworkPacket(proto, pkt)
	return len(packet), nil
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) Attach(dispatcher stack.NetworkDispatcher) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if dispatcher == nil {
		// Detaching.
		n.dispatcher = nil
		n.attached = false
		return
	}

	n.dispatcher = dispatcher
	writer := &networkLayerWriter{parent: n}
	// Let the network layer device know where to write incoming packets.
	if err := n.networkLayerSwitch.OnAttach(writer); err == nil {
		n.attached = true
	} else {
		// OnAttach failed; keep attached false.
		n.attached = false
	}
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) IsAttached() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.attached
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) Wait() {
	// If closed, return immediately.
	n.mu.RLock()
	closed := n.closed
	ch := n.waitCh
	n.mu.RUnlock()
	if closed {
		return
	}
	// Wait until closed is signaled.
	<-ch
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareNone
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) AddHeader(_ *stack.PacketBuffer) {
	// No link-layer header to add.
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) ParseHeader(_ *stack.PacketBuffer) bool {
	// Nothing to parse; packet is a bare network packet.
	return true
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) Close() {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return
	}
	n.closed = true
	n.attached = false
	n.mu.Unlock()

	// Close underlying network device if any.
	_ = common.Close(n.networkLayerSwitch)

	// Run onClose action if set.
	n.mu.RLock()
	onc := n.onClose
	ch := n.waitCh
	n.mu.RUnlock()
	if onc != nil {
		onc()
	}

	// Signal waiters.
	close(ch)
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) SetOnCloseAction(f func()) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.onClose = f
}

func (n *NetworkLayerDeviceToGvisorLinkEndpointAdaptor) WritePackets(list stack.PacketBufferList) (int, tcpip.Error) {
	// Defensive: if receiver is nil, treat as closed.
	if n == nil {
		return 0, &tcpip.ErrClosedForSend{}
	}
	// Convert each packet to bytes and write to networkLayerSwitch.
	slice := list.AsSlice()
	if len(slice) == 0 {
		return 0, nil
	}

	n.mu.RLock()
	dev := n.networkLayerSwitch
	mtu := n.mtu
	n.mu.RUnlock()
	if dev == nil {
		return 0, &tcpip.ErrClosedForSend{}
	}

	written := 0
	for _, pkt := range slice {
		if pkt == nil {
			continue
		}
		// Get slices and copy into a contiguous buffer.
		slices := pkt.AsSlices()
		total := 0
		for _, s := range slices {
			total += len(s)
		}
		if mtu > 0 && total > mtu {
			return written, &tcpip.ErrMessageTooLong{}
		}
		cp := make([]byte, total)
		off := 0
		for _, s := range slices {
			copy(cp[off:], s)
			off += len(s)
		}
		_, err := dev.Write(cp)
		if err != nil {
			// Map writer error to tcpip error.
			return written, &tcpip.ErrNoBufferSpace{}
		}
		written++
	}

	return written, nil
}
