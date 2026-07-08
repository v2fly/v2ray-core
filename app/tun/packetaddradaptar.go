package tun

import (
	"sync"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/app/tun/device"
	"github.com/v2fly/v2ray-core/v5/app/tun/tunsorter"
)

func NewDeviceWithSorter(overlay device.Device, sorter *tunsorter.TunSorter) device.Device {
	return &packetAddrDevice{
		Device: overlay,
		sorter: sorter,
	}
}

type packetAddrDevice struct {
	device.Device
	sorter *tunsorter.TunSorter

	dispatcherAccess    sync.RWMutex
	secondaryDispatcher stack.NetworkDispatcher
}

func (p *packetAddrDevice) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	buf := pkt.ToBuffer()
	_, err := p.sorter.OnPacketReceived(buf.Flatten())
	if err != nil {
		p.dispatcherAccess.RLock()
		dispatcher := p.secondaryDispatcher
		p.dispatcherAccess.RUnlock()
		if dispatcher != nil {
			dispatcher.DeliverNetworkPacket(protocol, pkt)
		}
	}
}

func (p *packetAddrDevice) DeliverLinkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	p.dispatcherAccess.RLock()
	dispatcher := p.secondaryDispatcher
	p.dispatcherAccess.RUnlock()
	if dispatcher != nil {
		dispatcher.DeliverLinkPacket(protocol, pkt)
	}
}

func (p *packetAddrDevice) Attach(dispatcher stack.NetworkDispatcher) {
	p.dispatcherAccess.Lock()
	p.secondaryDispatcher = dispatcher
	p.dispatcherAccess.Unlock()
	if dispatcher == nil {
		p.Device.Attach(nil)
		return
	}
	p.Device.Attach(p)
}
