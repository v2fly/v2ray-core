package interconnect

import (
	"context"
	"errors"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

func NewNetworkLayerCable(ctx context.Context) (*NetworkLayerCable, error) {
	return &NetworkLayerCable{
		ctx: ctx,
	}, nil
}

// NetworkLayerCable is primarily Machine Generated
type NetworkLayerCable struct {
	lSideWriter packetswitch.NetworkLayerPacketWriter
	rSideWriter packetswitch.NetworkLayerPacketWriter
	ctx         context.Context
	lock        sync.RWMutex
}

// NetworkLayerCableDevice is Machine Generated
type NetworkLayerCableDevice struct {
	cable  *NetworkLayerCable
	isLeft bool
}

func (c *NetworkLayerCable) GetLSideDevice() *NetworkLayerCableDevice {
	return &NetworkLayerCableDevice{
		cable:  c,
		isLeft: true,
	}
}

func (c *NetworkLayerCable) GetRSideDevice() *NetworkLayerCableDevice {
	return &NetworkLayerCableDevice{
		cable:  c,
		isLeft: false,
	}
}

// OnAttach implements NetworkLayerPacketReader.OnAttach
func (d *NetworkLayerCableDevice) OnAttach(writer packetswitch.NetworkLayerPacketWriter) error {
	if writer == nil {
		return errors.New("nil writer")
	}
	d.cable.lock.Lock()
	defer d.cable.lock.Unlock()
	if d.isLeft {
		if d.cable.lSideWriter != nil {
			return errors.New("left writer already attached")
		}
		d.cable.lSideWriter = writer
	} else {
		if d.cable.rSideWriter != nil {
			return errors.New("right writer already attached")
		}
		d.cable.rSideWriter = writer
	}
	return nil
}

// Write implements NetworkLayerPacketWriter.Write
func (d *NetworkLayerCableDevice) Write(packet []byte) (int, error) {
	d.cable.lock.RLock()
	var peer packetswitch.NetworkLayerPacketWriter
	if d.isLeft {
		peer = d.cable.rSideWriter
	} else {
		peer = d.cable.lSideWriter
	}
	d.cable.lock.RUnlock()
	if peer == nil {
		return 0, errors.New("no peer attached")
	}
	return peer.Write(packet)
}

// Close implements common.Closable.Close
func (d *NetworkLayerCableDevice) Close() error {
	d.cable.lock.Lock()
	defer d.cable.lock.Unlock()
	if d.isLeft {
		d.cable.lSideWriter = nil
	} else {
		d.cable.rSideWriter = nil
	}
	return nil
}
